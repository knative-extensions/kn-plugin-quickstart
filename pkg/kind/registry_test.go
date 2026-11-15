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
	"sync"
	"testing"
)

func TestRegistryChangesFollowClusterDecision(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake commands use shell scripts")
	}
	for _, tt := range []struct {
		name           string
		existing       bool
		answers        string
		registry       bool
		wantRegistries int
	}{
		{name: "decline recreation", existing: true, answers: "n\nn\n", registry: true},
		{name: "new cluster", registry: true, wantRegistries: 1},
		{name: "confirm recreation", existing: true, answers: "y\n", registry: true, wantRegistries: 1},
		{name: "confirm second recreation prompt", existing: true, answers: "n\ny\n", registry: true, wantRegistries: 1},
		{name: "recreate without registry leaves shared registry alone", existing: true, answers: "y\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			oldName, oldInstall := clusterName, installKnative
			clusterName, installKnative = "knative", true
			t.Cleanup(func() { clusterName, installKnative = oldName, oldInstall })
			dir := t.TempDir()
			commands := filepath.Join(dir, "commands")
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("QUICKSTART_TEST_COMMANDS", commands)
			existing := ""
			if tt.existing {
				existing = "knative"
			}
			t.Setenv("QUICKSTART_TEST_CLUSTER", existing)
			scripts := map[string]string{
				"kind": `#!/bin/sh
printf 'kind %s\n' "$*" >> "$QUICKSTART_TEST_COMMANDS"
case "$*" in
 'version -q') echo v0.31.0;;
 'get clusters -q') if [ -n "$QUICKSTART_TEST_CLUSTER" ]; then echo "$QUICKSTART_TEST_CLUSTER"; fi;;
 'get nodes --name knative') echo knative-control-plane;;
 'create cluster '*) while IFS= read -r line; do :; done;;
 'delete cluster --name knative') :;;
 *) exit 99;;
esac
`,
				"kubectl": `#!/bin/sh
printf 'kubectl %s\n' "$*" >> "$QUICKSTART_TEST_COMMANDS"
if [ "$1" = '--context' ]; then shift 2; fi
case "$*" in
 'version --client -o json') echo '{"clientVersion":{"gitVersion":"v1.35.0"}}';;
 'get namespaces') echo knative-serving;;
 'apply -f -') while IFS= read -r line; do :; done;;
 *) exit 99;;
esac
`,
			}
			for name, script := range scripts {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0755); err != nil {
					t.Fatal(err)
				}
			}
			input, err := os.CreateTemp(dir, "answers")
			if err != nil {
				t.Fatal(err)
			}
			defer input.Close()
			if _, err := input.WriteString(tt.answers); err != nil {
				t.Fatal(err)
			}
			if _, err := input.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			oldStdin := os.Stdin
			os.Stdin = input
			t.Cleanup(func() { os.Stdin = oldStdin })

			var mu sync.Mutex
			counts := map[string]int{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				path := strings.TrimPrefix(r.URL.Path, "/v1.47")
				key := r.Method + " " + path
				mu.Lock()
				counts[key]++
				mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				switch key {
				case "GET /info":
					fmt.Fprint(w, `{}`)
				case "POST /images/create":
					fmt.Fprintln(w, `{"status":"Already exists","id":"registry:2"}`)
				case "DELETE /containers/kind-registry", "POST /containers/registry-id/start", "POST /exec/exec-id/start", "POST /networks/kind/connect":
					w.WriteHeader(http.StatusNoContent)
				case "POST /containers/create":
					// No destructive registry operation should happen before cluster creation.
					out, err := os.ReadFile(commands)
					if err != nil || !strings.Contains(string(out), "kind create cluster ") {
						t.Error("registry was replaced before creating the cluster")
					}
					w.WriteHeader(http.StatusCreated)
					fmt.Fprint(w, `{"Id":"registry-id"}`)
				case "POST /containers/knative-control-plane/exec":
					w.WriteHeader(http.StatusCreated)
					fmt.Fprint(w, `{"Id":"exec-id"}`)
				default:
					t.Errorf("unexpected Docker call: %s", key)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			defer server.Close()
			t.Setenv("DOCKER_HOST", server.URL)
			t.Setenv("DOCKER_API_VERSION", "1.47")
			t.Setenv("DOCKER_TLS_VERIFY", "")

			if err := createKindCluster(tt.registry, "", "", 80); err != nil {
				t.Fatal(err)
			}
			mu.Lock()
			defer mu.Unlock()
			for _, call := range []string{"POST /images/create", "DELETE /containers/kind-registry", "POST /containers/create", "POST /containers/registry-id/start", "POST /networks/kind/connect"} {
				if got := counts[call]; got != tt.wantRegistries {
					t.Errorf("%s called %d times, want %d", call, got, tt.wantRegistries)
				}
			}
		})
	}
}
