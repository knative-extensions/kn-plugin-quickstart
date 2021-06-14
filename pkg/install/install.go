// Copyright © 2021 The Knative Authors
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
	"fmt"
	"os/exec"
)

var servingVersion = "0.23.0"

// Serving installs Knative Serving from Github YAML files
func Serving() error {
	fmt.Println("Starting Knative Serving install...")

	crds := exec.Command("kubectl", "apply", "-f", "https://github.com/knative/serving/releases/download/v"+servingVersion+"/serving-crds.yaml")
	if err := crds.Run(); err != nil {
		return fmt.Errorf("apply: %w", err)
	}

	crdWait := exec.Command("kubectl", "wait", "--for=condition=Established", "--all", "crd")
	if err := crdWait.Run(); err != nil {
		return fmt.Errorf("wait: %w", err)
	}
	fmt.Println("    CRDs installed...")

	core := exec.Command("kubectl", "apply", "-f", "https://github.com/knative/serving/releases/download/v"+servingVersion+"/serving-core.yaml")
	if err := core.Run(); err != nil {
		return fmt.Errorf("core apply: %w", err)
	}

	coreWait := exec.Command("kubectl", "wait", "pod", "--timeout=-1s", "--for=condition=Ready", "-l", "!job-name", "-n", "knative-serving")
	if err := coreWait.Run(); err != nil {
		return fmt.Errorf("core wait: %w", err)
	}

	fmt.Println("    Core installed...")

	fmt.Println("Finished installing Knative Serving")

	return nil
}

// Eventing installs Knative Eventing from Github YAML files
// TODO
func Eventing() {
	fmt.Println("TODO: Installing Knative Eventing...")
}
