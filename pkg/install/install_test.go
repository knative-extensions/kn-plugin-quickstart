// Copyright © 2026 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package install

import "testing"

func TestParseKubectlVersion(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantMajor int
		wantMinor int
		wantErr   bool
	}{{
		name:      "standard release",
		json:      `{"clientVersion":{"major":"1","minor":"34","gitVersion":"v1.34.0"}}`,
		wantMajor: 1,
		wantMinor: 34,
	}, {
		name:      "patch version",
		json:      `{"clientVersion":{"major":"1","minor":"33","gitVersion":"v1.33.2"}}`,
		wantMajor: 1,
		wantMinor: 33,
	}, {
		name: "vendor suffix in gitVersion",
		// EKS and other distributions append a suffix; we must read the
		// numeric prefix, not choke on it.
		json:      `{"clientVersion":{"major":"1","minor":"33+","gitVersion":"v1.33.2-eks-1234567"}}`,
		wantMajor: 1,
		wantMinor: 33,
	}, {
		name:    "no version present",
		json:    `{"clientVersion":{"buildDate":"2026-01-01"}}`,
		wantErr: true,
	}, {
		name:    "garbage",
		json:    `not json at all`,
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, err := parseKubectlVersion(tt.json)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got major=%d minor=%d", major, minor)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if major != tt.wantMajor || minor != tt.wantMinor {
				t.Errorf("got v%d.%d, want v%d.%d", major, minor, tt.wantMajor, tt.wantMinor)
			}
		})
	}
}
