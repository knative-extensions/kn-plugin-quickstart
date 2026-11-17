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
	"testing"
	"time"
)

func TestRetryCommand(t *testing.T) {
	for _, tt := range []struct {
		name      string
		failures  int
		wantCalls int
		wantErr   bool
	}{
		{name: "immediate success", wantCalls: 1},
		{name: "transient failure", failures: 1, wantCalls: 2},
		{name: "attempts exhausted", failures: 3, wantCalls: 3, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			failure := errors.New("apply failed")
			calls, sleeps := 0, 0
			seen := map[*exec.Cmd]bool{}
			err := retryCommand(
				func() *exec.Cmd { return &exec.Cmd{} },
				func(cmd *exec.Cmd) error {
					if seen[cmd] {
						t.Error("reused a command")
					}
					seen[cmd] = true
					calls++
					if calls <= tt.failures {
						return failure
					}
					return nil
				},
				func(delay time.Duration) {
					sleeps++
					if delay != 10*time.Second {
						t.Errorf("delay = %s, want 10s", delay)
					}
				},
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, failure) {
				t.Errorf("error = %v, want original error", err)
			}
			if calls != tt.wantCalls {
				t.Errorf("calls = %d, want %d", calls, tt.wantCalls)
			}
			if sleeps != tt.wantCalls-1 {
				t.Errorf("sleeps = %d, want %d", sleeps, tt.wantCalls-1)
			}
		})
	}
}
