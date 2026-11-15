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
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWaitForPodsReady(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake kubectl uses a shell script")
	}
	create := "wait pod --timeout=10m --for=create -l !job-name -n knative-serving"
	ready := "wait pod --timeout=10m --for=condition=Ready -l !job-name -n knative-serving"
	for _, tt := range []struct {
		name          string
		failCondition string
		wantCalls     []string
		wantErr       bool
	}{
		{name: "creation precedes readiness", wantCalls: []string{create, ready}},
		{name: "creation failure stops readiness wait", failCondition: "--for=create", wantCalls: []string{create}, wantErr: true},
		{name: "readiness failure is returned", failCondition: "--for=condition=Ready", wantCalls: []string{create, ready}, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			calls := filepath.Join(dir, "calls")
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("QUICKSTART_TEST_CALLS", calls)
			t.Setenv("QUICKSTART_TEST_FAIL_CONDITION", tt.failCondition)
			script := `#!/bin/sh
printf '%s\n' "$*" >> "$QUICKSTART_TEST_CALLS"
conditions=0
for arg in "$@"; do
 case "$arg" in --for=*) conditions=$((conditions + 1));; esac
 if [ "$arg" = "$QUICKSTART_TEST_FAIL_CONDITION" ]; then exit 1; fi
done
# kubectl only honors the last --for value; reject this misuse explicitly.
[ "$conditions" -eq 1 ]
`
			if err := os.WriteFile(filepath.Join(dir, "kubectl"), []byte(script), 0755); err != nil {
				t.Fatal(err)
			}
			err := waitForPodsReady("knative-serving")
			if (err != nil) != tt.wantErr {
				t.Errorf("waitForPodsReady() error = %v, wantErr %v", err, tt.wantErr)
			}
			out, err := os.ReadFile(calls)
			if err != nil {
				t.Fatal(err)
			}
			want := strings.Join(tt.wantCalls, "\n") + "\n"
			if string(out) != want {
				t.Errorf("kubectl calls = %q, want %q", out, want)
			}
		})
	}
}
