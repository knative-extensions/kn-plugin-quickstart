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

import (
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
		cluster       string
		args          []string
		profile       string
		failSelection bool
	}{
		{name: "default profile", cluster: "knative", profile: "knative"},
		{name: "named profile", cluster: "demo", profile: "demo"},
		{name: "passthrough profile override", cluster: "unused", args: []string{"--profile=demo"}, profile: "demo"},
		{name: "missing context prevents installation", cluster: "demo", profile: "demo", failSelection: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			oldName, oldInstall, oldArgs := clusterName, installKnative, customMinikubeArgs
			installKnative, customMinikubeArgs = true, nil
			t.Cleanup(func() { clusterName, installKnative, customMinikubeArgs = oldName, oldInstall, oldArgs })
			dir := t.TempDir()
			calls := filepath.Join(dir, "calls")
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("QUICKSTART_TEST_CALLS", calls)
			t.Setenv("QUICKSTART_TEST_PROFILE", tt.profile)
			fail := ""
			if tt.failSelection {
				fail = "yes"
			}
			t.Setenv("QUICKSTART_TEST_FAIL_SELECTION", fail)
			for name, script := range map[string]string{
				"minikube": `#!/bin/sh
case "$*" in
 'version --short') echo v1.38.0;;
 'profile list') echo "$QUICKSTART_TEST_PROFILE";;
 *) exit 99;;
esac
`,
				"kubectl": `#!/bin/sh
printf '%s\n' "$*" >> "$QUICKSTART_TEST_CALLS"
case "$*" in
 'version --client -o json') echo '{"clientVersion":{"gitVersion":"v1.35.0"}}';;
 *'get namespaces') echo default;;
 "config use-context $QUICKSTART_TEST_PROFILE") if [ -n "$QUICKSTART_TEST_FAIL_SELECTION" ]; then exit 1; fi;;
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
			if _, err := input.WriteString("n\n\n"); err != nil {
				t.Fatal(err)
			}
			if _, err := input.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			oldStdin := os.Stdin
			os.Stdin = input
			t.Cleanup(func() { os.Stdin = oldStdin })

			err = SetUp(tt.cluster, "", true, true, tt.args)
			if (err != nil) != tt.failSelection {
				t.Fatalf("SetUp() error = %v, wantErr %v", err, tt.failSelection)
			}
			out, err := os.ReadFile(calls)
			if err != nil {
				t.Fatal(err)
			}
			log := string(out)
			namespaceCheck := strings.Index(log, "--context "+tt.profile+" get namespaces\n")
			selection := strings.Index(log, "config use-context "+tt.profile+"\n")
			apply := strings.Index(log, "apply -f ")
			if namespaceCheck < 0 {
				t.Errorf("namespace check did not target %s: %s", tt.profile, log)
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
