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
	"strconv"
	"strings"
	"testing"
)

func TestRetryingApply(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake kubectl uses a shell script")
	}
	for _, tt := range []struct {
		name      string
		failures  string
		wantCalls int
		wantErr   bool
	}{
		{name: "first attempt succeeds", failures: "0", wantCalls: 1},
		{name: "transient failure", failures: "1", wantCalls: 2},
		{name: "attempts exhausted", failures: "3", wantCalls: 3, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			calls := filepath.Join(dir, "calls")
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("QUICKSTART_TEST_CALLS", calls)
			t.Setenv("QUICKSTART_TEST_FAILURES", tt.failures)
			script := `#!/bin/sh
count=0
if [ -f "$QUICKSTART_TEST_CALLS" ]; then
 read -r count < "$QUICKSTART_TEST_CALLS"
fi
count=$((count + 1))
printf '%s\n' "$count" > "$QUICKSTART_TEST_CALLS"
if [ "$*" != 'apply -f test.yaml' ]; then exit 99; fi
if [ "$count" -le "$QUICKSTART_TEST_FAILURES" ]; then exit 1; fi
`
			if err := os.WriteFile(filepath.Join(dir, "kubectl"), []byte(script), 0755); err != nil {
				t.Fatal(err)
			}
			err := retryingApply("test.yaml")
			if (err != nil) != tt.wantErr {
				t.Fatalf("retryingApply() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != "exit status 1" {
				t.Fatalf("got %v, want the final kubectl error", err)
			}
			out, readErr := os.ReadFile(calls)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if got := strings.TrimSpace(string(out)); got != strconv.Itoa(tt.wantCalls) {
				t.Errorf("kubectl ran %s times, want %d", got, tt.wantCalls)
			}
		})
	}
}
