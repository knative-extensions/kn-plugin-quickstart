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

package kind

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"knative.dev/kn-plugin-quickstart/pkg/util"

	"knative.dev/kn-plugin-quickstart/pkg/install"
)

var kubernetesVersion = "v1.21.1@sha256:fae9a58f17f18f06aeac9772ca8b5ac680ebbed985e266f711d936e91d113bad"
var clusterName = "knative"

// SetUp creates a local Kind cluster and installs all the relevant Knative components
func SetUp() error {
	if err := createKindCluster(); err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	if err := install.Serving(); err != nil {
		return fmt.Errorf("install serving: %w", err)
	}
	if err := install.Kourier(); err != nil {
		return fmt.Errorf("install kourier: %w", err)
	}
	if err := install.Eventing(); err != nil {
		return fmt.Errorf("install eventing: %w", err)
	}
	return nil
}

func createKindCluster() error {
	fmt.Println("Creating Kind cluster...")

	// Get kind config file
	kindConfig, err := ioutil.TempFile(os.TempDir(), "kind-config-*.yaml")
	if err != nil {
		return fmt.Errorf("kind create: %w", err)
	}

	defer os.Remove(kindConfig.Name())

	configRaw := "kind: Cluster\n" +
		"apiVersion: kind.x-k8s.io/v1alpha4\n" +
		"name: " + clusterName + "\n" +
		"nodes:\n" +
		"- role: control-plane\n" +
		"  image: kindest/node:" + kubernetesVersion + "\n" +
		"  extraPortMappings:\n" +
		"  - containerPort: 31080\n" +
		"    listenAddress: 127.0.0.1\n" +
		"    hostPort: 80"
	config := []byte(configRaw)

	if _, err := kindConfig.Write(config); err != nil {
		return fmt.Errorf("kind create: %w", err)
	}

	configFile := kindConfig.Name()
	createCluster := exec.Command("kind", "create", "cluster", "--config", configFile)
	if err := util.RunCommand(createCluster); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := kindConfig.Close(); err != nil {
		return fmt.Errorf("kind create: %w", err)
	}

	fmt.Println("    Waiting on cluster to be ready...")
	time.Sleep(10 * time.Second)

	clusterWait := exec.Command("kubectl", "wait", "pod", "--timeout=-1s", "--for=condition=Ready", "-l", "!job-name", "-n", "kube-system")
	if err := util.RunCommand(clusterWait); err != nil {
		return fmt.Errorf("%w", err)
	}

	fmt.Println("Cluster created")
	return nil
}
