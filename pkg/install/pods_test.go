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
	"errors"
	"os/exec"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestWaitForPods(t *testing.T) {
	for _, tt := range []struct {
		name   string
		failAt string
		want   []string
	}{
		{name: "creation before readiness", want: []string{"--for=create", "--for=condition=Ready"}},
		{name: "creation failure", failAt: "--for=create", want: []string{"--for=create"}},
		{name: "readiness failure", failAt: "--for=condition=Ready", want: []string{"--for=create", "--for=condition=Ready"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			failure := errors.New("wait failed")
			var conditions []string
			err := waitForPods("test-namespace",
				func(args ...string) *exec.Cmd { return &exec.Cmd{Args: args} },
				func(cmd *exec.Cmd) error {
					var found []string
					for _, arg := range cmd.Args {
						if strings.HasPrefix(arg, "--for=") {
							found = append(found, arg)
						}
					}
					if len(found) != 1 {
						t.Fatalf("want one condition per command, got %v", found)
					}
					if !slices.Contains(cmd.Args, "test-namespace") || !slices.Contains(cmd.Args, "!job-name") {
						t.Errorf("missing namespace or selector: %v", cmd.Args)
					}
					conditions = append(conditions, found[0])
					if found[0] == tt.failAt {
						return failure
					}
					return nil
				},
			)
			if !reflect.DeepEqual(conditions, tt.want) {
				t.Errorf("conditions = %v, want %v", conditions, tt.want)
			}
			if tt.failAt == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else if !errors.Is(err, failure) {
				t.Errorf("error = %v, want original error", err)
			}
		})
	}
}
