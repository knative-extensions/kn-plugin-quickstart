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

# ==========================================
# Unit and Build tests

# We can't use MD checks now as they will propagate into
# the plugins' vendor/ dir
# (the filter in markdown_build_tests() in test-infra/scripts/presumit-tests.sh is
# not strong enough
export DISABLE_MD_LINTING=1
export DISABLE_MD_LINK_CHECK=1

export PRESUBMIT_TEST_FAIL_FAST=1
export GO111MODULE=on
source $(dirname "$0")/../vendor/knative.dev/hack/presubmit-tests.sh

# Run cross platform build to ensure kn compiles for Linux, macOS and Windows
function post_build_tests() {
  local failed=0
  header "Ensuring cross platform build"
  ./hack/build.sh -x || failed=1
  if (( failed )); then
    results_banner "Cross platform build" ${failed}
    exit ${failed}
  fi
}

# Run the unit tests with an additional flag '-mod=vendor' to avoid
# downloading the deps in unit tests CI job
function unit_tests() {
  report_go_test -race ./... || failed=1
}

# We use the default build and integration test runners.
main "$@"
