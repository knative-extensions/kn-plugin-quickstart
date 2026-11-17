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

import (
	"slices"
	"testing"
)

func TestInstallerKubectl(t *testing.T) {
	for _, context := range []string{"", "kind-custom"} {
		t.Run("context="+context, func(t *testing.T) {
			installer := Installer{Context: context}
			want := []string{"kubectl"}
			if context != "" {
				want = append(want, "--context", context)
			}
			want = append(want, "get", "pods")
			cmd := installer.kubectl("get", "pods")
			if !slices.Equal(cmd.Args, want) {
				t.Errorf("args = %v, want %v", cmd.Args, want)
			}
		})
	}
}
