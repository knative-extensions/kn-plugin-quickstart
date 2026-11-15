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

func (i Installer) command(args ...string) *exec.Cmd {
	if i.Context != "" {
		args = append([]string{"--context", i.Context}, args...)
	}
	// #nosec G204 -- Fixed executable; arguments are passed directly, never through a shell.
	return exec.Command("kubectl", args...)
}

// Keep the package-level entry points available to callers using the current
// context. Quickstart uses an Installer with an explicit context instead.

// Kourier installs Kourier using the current kubeconfig context.
func Kourier() error { return (Installer{}).Kourier() }

// KourierKind configures Kourier for Kind using the current kubeconfig context.
func KourierKind() error { return (Installer{}).KourierKind() }

// KourierMinikube configures Kourier for Minikube using the current context.
func KourierMinikube() error { return (Installer{}).KourierMinikube() }

// Serving installs Knative Serving using the current kubeconfig context.
func Serving(registries string) error { return (Installer{}).Serving(registries) }

// Eventing installs Knative Eventing using the current kubeconfig context.
func Eventing() error { return (Installer{}).Eventing() }
