#!/usr/bin/env bash

# Copyright 2021 The Knative Authors
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

# Documentation about this script and how to use it can be found
# at https://github.com/knative/hack

PLUGIN="kn-quickstart"
VERSION_PACKAGE="knative.dev/kn-plugin-quickstart/internal/command"
COMPONENT_PACKAGE="knative.dev/kn-plugin-quickstart/pkg/install"

source $(dirname $0)/../vendor/knative.dev/hack/release.sh

function build_release() {
  source $(dirname $0)/build-flags.sh
  local ld_flags="$(build_flags $(dirname $0)/..)"
  local version="${TAG}"
  # Use vYYYYMMDD-<hash>-local for the version string, if not passed.
  [[ -z "${version}" ]] && version="v${BUILD_TAG}-local"
  echo "Building version: $version"
  echo "Using flags: $ld_flags"

  export GO111MODULE=on
  export CGO_ENABLED=0
  echo "🚧 🐧 Building for Linux (amd64)"
  GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./${PLUGIN}-linux-amd64 ./cmd/...
  echo "🚧 💪 Building for Linux (arm64)"
  GOOS=linux GOARCH=arm64 go build -mod=vendor -ldflags "${ld_flags}" -o ./${PLUGIN}-linux-arm64 ./cmd/...
  echo "🚧 🍏 Building for macOS"
  GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./${PLUGIN}-darwin-amd64 ./cmd/...
  echo "🚧 🍎 Building for macOS (arm64)"
  GOOS=darwin GOARCH=arm64 go build -mod=vendor -ldflags "${ld_flags}" -o ./${PLUGIN}-darwin-arm64 ./cmd/...
  echo "🚧 🎠 Building for Windows"
  GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./${PLUGIN}-windows-amd64.exe ./cmd/...
  echo "🚧 Z  Building for Linux(s390x)"
  GOOS=linux GOARCH=s390x go build -mod=vendor -ldflags "${ld_flags}" -o ./${PLUGIN}-linux-s390x ./cmd/...
  echo "🚧 P  Building for Linux (ppc64le)"
  GOOS=linux GOARCH=ppc64le go build -mod=vendor -ldflags "${ld_flags}" -o ./${PLUGIN}-linux-ppc64le ./cmd/...
  ARTIFACTS_TO_PUBLISH="${PLUGIN}-darwin-amd64 ${PLUGIN}-darwin-arm64 ${PLUGIN}-linux-amd64 ${PLUGIN}-linux-arm64 ${PLUGIN}-windows-amd64.exe ${PLUGIN}-linux-s390x ${PLUGIN}-linux-ppc64le"
  sha256sum ${ARTIFACTS_TO_PUBLISH} > checksums.txt
  ARTIFACTS_TO_PUBLISH="${ARTIFACTS_TO_PUBLISH} checksums.txt"
  echo "🧮     Checksum:"
  cat checksums.txt
}

main $@
