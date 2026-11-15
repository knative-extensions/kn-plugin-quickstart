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

package kind

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReuseSelectsRequestedContext(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake commands use shell scripts")
	}
	for _, tt := range []struct {
		name          string
		failSelection bool
		wantErr       bool
	}{
		{name: "install into requested cluster"},
		{name: "missing context prevents installation", failSelection: true, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			oldName, oldInstall := clusterName, installKnative
			clusterName, installKnative = "demo", true
			t.Cleanup(func() { clusterName, installKnative = oldName, oldInstall })
			dir := t.TempDir()
			calls := filepath.Join(dir, "calls")
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("QUICKSTART_TEST_CALLS", calls)
			fail := ""
			if tt.failSelection {
				fail = "yes"
			}
			t.Setenv("QUICKSTART_TEST_FAIL_SELECTION", fail)
			for name, script := range map[string]string{
				"kind": `#!/bin/sh
case "$*" in
 'version -q') echo v0.31.0;;
 'get clusters -q') echo demo;;
 *) exit 99;;
esac
`,
				"kubectl": `#!/bin/sh
printf '%s\n' "$*" >> "$QUICKSTART_TEST_CALLS"
case "$*" in
 'version --client -o json') echo '{"clientVersion":{"gitVersion":"v1.35.0"}}';;
 *'get namespaces') echo default;;
 'config use-context kind-demo') if [ -n "$QUICKSTART_TEST_FAIL_SELECTION" ]; then exit 1; fi;;
 'get endpointslice '*) echo ready;;
 'apply -f -') while IFS= read -r line; do :; done;;
esac
`,
			} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0755); err != nil {
					t.Fatal(err)
				}
			}
			input, err := os.CreateTemp(dir, "answers")
			if err != nil {
				t.Fatal(err)
			}
			defer input.Close()
			if _, err := input.WriteString("n\n"); err != nil {
				t.Fatal(err)
			}
			if _, err := input.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			oldStdin := os.Stdin
			os.Stdin = input
			t.Cleanup(func() { os.Stdin = oldStdin })
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1.47/info" {
					t.Errorf("unexpected Docker call: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(500)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{}`)
			}))
			defer server.Close()
			t.Setenv("DOCKER_HOST", server.URL)
			t.Setenv("DOCKER_API_VERSION", "1.47")
			t.Setenv("DOCKER_TLS_VERIFY", "")

			err = SetUp("demo", "", true, true, false, "", "", 80)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SetUp() error = %v, wantErr %v", err, tt.wantErr)
			}
			out, err := os.ReadFile(calls)
			if err != nil {
				t.Fatal(err)
			}
			log := string(out)
			namespaceCheck := strings.Index(log, "--context kind-demo get namespaces\n")
			selection := strings.Index(log, "config use-context kind-demo\n")
			apply := strings.Index(log, "apply -f ")
			if namespaceCheck < 0 {
				t.Errorf("namespace check did not target kind-demo: %s", log)
			}
			if selection < namespaceCheck || selection < 0 {
				t.Errorf("requested context was not selected after checking its namespaces: %s", log)
			}
			if tt.failSelection {
				if apply >= 0 {
					t.Errorf("applied manifests after context selection failed: %s", log)
				}
			} else if apply < selection || apply < 0 {
				t.Errorf("manifests were not applied after selecting the requested context: %s", log)
			}
		})
	}
}
