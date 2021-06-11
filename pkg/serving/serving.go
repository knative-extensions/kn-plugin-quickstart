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

package serving

import (
	"fmt"
	"os/exec"
)

var servingVersion = "0.23.0"

// InstallKnativeServing installs Serving from Github YAML files
func InstallKnativeServing() {
	fmt.Println("Starting Knative Serving install...")

	crds := exec.Command("kubectl", "apply", "-f", "https://github.com/knative/serving/releases/download/v"+servingVersion+"/serving-crds.yaml")
	err := crds.Run()
	if err != nil {
		fmt.Errorf("%s", err)
	}
	crdWait := exec.Command("kubectl", "wait", "--for=condition=Established", "--all", "crd")
	err = crdWait.Start()
	if err != nil {
		fmt.Errorf("%s", err)
	}
	err = crdWait.Wait()
	if err != nil {
		fmt.Errorf("%s", err)
	}
	fmt.Println("    CRDs installed...")

	core := exec.Command("kubectl", "apply", "-f", "https://github.com/knative/serving/releases/download/v"+servingVersion+"/serving-core.yaml")
	err = core.Run()
	if err != nil {
		fmt.Errorf("%s", err)
	}
	coreWait := exec.Command("kubectl", "wait", "pod", "--timeout=-1s", "--for=condition=Ready", "-l", "'!job-name'", "-n", "knative-serving")
	err = coreWait.Start()
	if err != nil {
		fmt.Errorf("%s", err)
	}
	err = coreWait.Wait()
	if err != nil {
		fmt.Errorf("%s", err)
	}

	fmt.Println("    Core installed...")

	fmt.Println("Finished installing Knative Serving")

}
