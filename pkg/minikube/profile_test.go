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

package minikube

import "testing"

func TestProfileName(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{name: "knative", want: "knative"},
		{name: "custom-name", want: "custom-name"},
		{name: "knative", args: []string{"--profile=custom"}, want: "custom"},
		{name: "knative", args: []string{"--profile", "custom"}, want: "custom"},
		{name: "knative", args: []string{"-p", "custom"}, want: "custom"},
		{name: "knative", args: []string{"--profile=first", "--profile=last"}, want: "last"},
		{name: "knative", args: []string{"--keep-context"}, want: "knative"},
	} {
		if got := profileName(tt.name, tt.args); got != tt.want {
			t.Errorf("profileName(%q, %v) = %q, want %q", tt.name, tt.args, got, tt.want)
		}
	}
}
