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
	"reflect"
	"testing"
)

func TestInstallerCommand(t *testing.T) {
	for _, context := range []string{"", "knative", "kind-custom", "custom-profile"} {
		t.Run("context="+context, func(t *testing.T) {
			installer := Installer{Context: context}
			want := []string{"kubectl"}
			if context != "" {
				want = append(want, "--context", context)
			}
			want = append(want, "get", "pods")
			cmd := installer.command("get", "pods")
			if !reflect.DeepEqual(cmd.Args, want) {
				t.Errorf("args = %v, want %v", cmd.Args, want)
			}
			// Each command must carry its own context, rather than relying on a prior
			// `kubectl config use-context` call that would override --keep-context.
			again := installer.command("get", "pods")
			if !reflect.DeepEqual(again.Args, want) {
				t.Errorf("second command args = %v, want %v", again.Args, want)
			}
		})
	}
}
