#!/usr/bin/env bash

# Copyright 2026 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Verifies that the hardcoded Kubernetes minimum version in pkg/kind/kind.go
# and pkg/minikube/minikube.go matches DefaultKubernetesMinVersion in the
# version of knative.dev/pkg pinned in our go.sum. We keep a copy locally
# rather than importing the package to avoid pulling in client-go and its
# transitive dependency tree.
#
# If --write is passed, the script rewrites the local files in place when
# they drift from upstream (used by the auto-bump workflow). Default is
# read-only verification.

set -euo pipefail

WRITE=0
if [[ "${1:-}" == "--write" ]]; then
  WRITE=1
fi

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

pkg_version="$(go list -mod=mod -m -f '{{.Version}}' knative.dev/pkg)"
# Pseudo-versions look like v0.0.0-20260507212125-df317a52d112; the trailing
# field is the commit sha. Real tags (e.g. v0.34.0) work as refs directly.
if [[ "$pkg_version" == v0.0.0-*-* ]]; then
  ref="${pkg_version##*-}"
else
  ref="$pkg_version"
fi

upstream_url="https://raw.githubusercontent.com/knative/pkg/${ref}/version/version.go"
upstream_version="$(curl -fsSL "$upstream_url" \
  | sed -nE 's/.*DefaultKubernetesMinVersion[[:space:]]*=[[:space:]]*"v([0-9.]+)".*/\1/p' \
  | head -n1)"

if [[ -z "$upstream_version" ]]; then
  echo "ERROR: could not parse DefaultKubernetesMinVersion from $upstream_url" >&2
  exit 1
fi

extract() {
  sed -nE "$2" "$REPO_ROOT/$1" | head -n1
}

kind_version="$(extract pkg/kind/kind.go \
  's/.*kubernetesVersion[[:space:]]*=[[:space:]]*"kindest\/node:v([0-9.]+)".*/\1/p')"
minikube_version="$(extract pkg/minikube/minikube.go \
  's/.*kubernetesVersion[[:space:]]*=[[:space:]]*"([0-9.]+)".*/\1/p')"

fail=0
for entry in "kind:$kind_version:pkg/kind/kind.go" "minikube:$minikube_version:pkg/minikube/minikube.go"; do
  name="${entry%%:*}"
  rest="${entry#*:}"
  local_version="${rest%%:*}"
  file="${rest#*:}"
  if [[ -z "$local_version" ]]; then
    echo "ERROR: could not parse kubernetesVersion from $file" >&2
    fail=1
  elif [[ "$local_version" != "$upstream_version" ]]; then
    if [[ $WRITE -eq 1 ]]; then
      sed -i.bak -E "s/(kubernetesVersion[[:space:]]*=[[:space:]]*\"(kindest\/node:v)?)${local_version}/\1${upstream_version}/" "$REPO_ROOT/$file"
      rm -f "$REPO_ROOT/$file.bak"
      echo "UPDATED: $name kubernetesVersion in $file: $local_version -> $upstream_version"
    else
      echo "ERROR: $name kubernetesVersion in $file is $local_version, upstream is $upstream_version (knative/pkg@$ref)" >&2
      fail=1
    fi
  else
    echo "OK: $name kubernetesVersion ($local_version) matches upstream (knative/pkg@$ref)"
  fi
done

if [[ $fail -ne 0 ]]; then
  echo
  echo "Run './hack/verify-min-k8s-version.sh --write' to update locally, or wait for the auto-bump workflow." >&2
  exit 1
fi
