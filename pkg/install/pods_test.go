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
	"slices"
	"testing"
)

func TestWaitForPods(t *testing.T) {
	for _, tt := range []struct {
		name   string
		failAt string
		want   []string
	}{
		{name: "creation before readiness", want: []string{"create", "condition=Ready"}},
		{name: "creation failure", failAt: "create", want: []string{"create"}},
		{name: "readiness failure", failAt: "condition=Ready", want: []string{"create", "condition=Ready"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			failure := errors.New("wait failed")
			var conditions []string
			err := waitForPods(func(condition string) error {
				conditions = append(conditions, condition)
				if condition == tt.failAt {
					return failure
				}
				return nil
			})
			if !slices.Equal(conditions, tt.want) {
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
