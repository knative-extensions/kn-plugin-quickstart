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


# Documentation about how the functions here are hooked into the release process
# and how to use it can be found at https://github.com/knative/hack

source $(dirname "$0")/global_vars.sh
source $(dirname "$0")/../vendor/knative.dev/hack/release.sh

function build_release() {
  export GO111MODULE=on
  export CGO_ENABLED=0
  echo "🚧 🐧 Building for Linux (amd64)"
  GOOS=linux GOARCH=amd64 go build -mod=vendor -o ./${PLUGIN}-linux-amd64 ./cmd/...
  echo "🚧 💪 Building for Linux (arm64)"
  GOOS=linux GOARCH=arm64 go build -mod=vendor -o ./${PLUGIN}-linux-arm64 ./cmd/...
  echo "🚧 🍏 Building for macOS"
  GOOS=darwin GOARCH=amd64 go build -mod=vendor -o ./${PLUGIN}-darwin-amd64 ./cmd/...
  echo "🚧 🎠 Building for Windows"
  GOOS=windows GOARCH=amd64 go build -mod=vendor -o ./${PLUGIN}-windows-amd64.exe ./cmd/...
  echo "🚧 Z  Building for Linux(s390x)"
  GOOS=linux GOARCH=s390x go build -mod=vendor -o ./${PLUGIN}-linux-s390x ./cmd/...
  echo "🚧 P  Building for Linux (ppc64le)"
  GOOS=linux GOARCH=ppc64le go build -mod=vendor -o ./${PLUGIN}-linux-ppc64le ./cmd/...
  ARTIFACTS_TO_PUBLISH="${PLUGIN}-darwin-amd64 ${PLUGIN}-linux-amd64 ${PLUGIN}-linux-arm64 ${PLUGIN}-windows-amd64.exe ${PLUGIN}-linux-s390x ${PLUGIN}-linux-ppc64le"
  sha256sum ${ARTIFACTS_TO_PUBLISH} > checksums.txt
  ARTIFACTS_TO_PUBLISH="${ARTIFACTS_TO_PUBLISH} checksums.txt"
  echo "🧮     Checksum:"
  cat checksums.txt
}

main "$@"
