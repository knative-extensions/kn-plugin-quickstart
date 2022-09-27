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

function build_flags() {
  local base="${1}"
  local now="$(date -u '+%Y-%m-%d %H:%M:%S')"
  local rev="$(git rev-parse --short HEAD)"
  local version="${TAG:-}"
  # Use vYYYYMMDD-local-<hash> for the version string, if not passed.
  if [[ -z "${version}" ]]; then
    # Get the commit, excluding any tags but keeping the "dirty" flag
    local commit="$(git describe --always --dirty --match '^$')"
    [[ -n "${commit}" ]] || (echo "error getting the current commit" && exit 1)
    version="v$(date +%Y%m%d)-local-${commit}"
  fi
  # Knative component versions
  local branch="`git branch --show-current | cut -d '-' -s -f2`"
  local serving="`git ls-remote --tags --ref https://github.com/knative/serving.git | grep -F "${branch}" | cut -d '-' -f2 | cut -d 'v' -f2 | sort -Vr | head -n 1`"
  local kourier="`git ls-remote --tags --ref https://github.com/knative-sandbox/net-kourier.git | grep -F "${branch}" | cut -d '-' -f2 | cut -d 'v' -f2 | sort -Vr | head -n 1`"
  local eventing="`git ls-remote --tags --ref https://github.com/knative/eventing.git | grep -F "${branch}" | cut -d '-' -f2 | cut -d 'v' -f2 | sort -Vr | head -n 1`"


  echo "-X '${VERSION_PACKAGE}.BuildDate=${now}' -X ${VERSION_PACKAGE}.Version=${version} -X ${VERSION_PACKAGE}.GitRevision=${rev} -X ${COMPONENT_PACKAGE}.ServingVersion=${serving} -X ${COMPONENT_PACKAGE}.KourierVersion=${kourier} -X ${COMPONENT_PACKAGE}.EventingVersion=${eventing}"
}
