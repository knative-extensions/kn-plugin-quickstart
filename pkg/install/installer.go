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

import "os/exec"

// Installer targets every installation command at Context without changing the
// kubeconfig's current context. Its zero value uses kubectl's current context.
type Installer struct {
	Context string
}

func (i Installer) kubectl(args ...string) *exec.Cmd {
	if i.Context != "" {
		args = append([]string{"--context", i.Context}, args...)
	}
	// #nosec G204 -- Fixed executable; arguments are passed directly, never through a shell.
	return exec.Command("kubectl", args...)
}
