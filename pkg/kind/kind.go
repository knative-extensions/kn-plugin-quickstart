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
	"regexp"
	"time"

	"knative.dev/kn-plugin-quickstart/pkg/install"
)

var kubernetesVersion = "v1.21.1@sha256:fae9a58f17f18f06aeac9772ca8b5ac680ebbed985e266f711d936e91d113bad"
var clusterName = "knative-quickstart"
var kindVersion = "v0.11"

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

	if err := checkKindVersion(); err != nil {
		return fmt.Errorf("kind version: %w", err)
	}
	if err := checkForExistingCluster(); err != nil {
		return fmt.Errorf("existing cluster: %w", err)
	}

	return nil
}

// checkKindVersion validates that the user has the correct version of Kind installed.
// If not, it prompts the user to download a newer version before continuing.
func checkKindVersion() error {

	versionCheck := exec.Command("kind", "version")
	out, err := versionCheck.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kind version: %w", err)
	}
	fmt.Printf("Kind version is: %s\n", string(out))

	r := regexp.MustCompile(kindVersion)
	matches := r.Match(out)
	if !matches {
		var resp string
		fmt.Printf("WARNING: Please make sure you are using Kind version %s.x", kindVersion)
		fmt.Println("Download from https://github.com/kubernetes-sigs/kind/releases")
		fmt.Print("Do you want to continue at your own risk [Y/n]: ")
		fmt.Scanf("%s", &resp)
		if resp == "n" || resp == "N" {
			fmt.Println("Installation stopped. Please upgrade kind and run again")
			os.Exit(0)
		}
	}

	return nil
}

// checkForExistingCluster checks if the user already has a Kind cluster. If so, it provides
// the option of deleting the existing cluster and recreating it. If not, it proceeds to
// creating a new cluster
func checkForExistingCluster() error {

	getClusters := exec.Command("kind", "get", "clusters", "-q")
	out, err := getClusters.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		fmt.Errorf("check cluster: %w", err)
	}
	// TODO Add tests for regex
	r := regexp.MustCompile(`(?m)^knative-quickstart\n`)
	matches := r.Match(out)
	if matches {
		var resp string
		fmt.Print("Knative Cluster kind-" + clusterName + " already installed.\nDelete and recreate [y/N]: ")
		fmt.Scanf("%s", &resp)
		if resp == "y" || resp == "Y" {
			fmt.Println("deleting cluster...")
			deleteCluster := exec.Command("kind", "delete", "cluster", "--name", clusterName)
			if err := deleteCluster.Run(); err != nil {
				return fmt.Errorf("delete cluster: %w", err)
			}
			if err := createNewCluster(); err != nil {
				return fmt.Errorf("new cluster: %w", err)
			}
		} else {
			fmt.Println("Cluster create skipped")
			return nil
		}
	} else {
		if err := createNewCluster(); err != nil {
			return fmt.Errorf("new cluster: %w", err)
		}
	}

	return nil
}

// createNewCluster creates a new Kind cluster
func createNewCluster() error {

	fmt.Println("Creating Kind cluster...")
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
	if err := runCommand(createCluster); err != nil {
		return fmt.Errorf("kind create: %w", err)
	}

	if err := kindConfig.Close(); err != nil {
		return fmt.Errorf("kind create: %w", err)
	}

	// sleep for 10s to allow initial cluster creation, then wait until all pods in kube-system namespace are ready
	fmt.Println("    Waiting on cluster to be ready...")
	time.Sleep(10 * time.Second)
	clusterWait := exec.Command("kubectl", "wait", "pod", "--timeout=-1s", "--for=condition=Ready", "-l", "!job-name", "-n", "kube-system")
	if err := runCommand(clusterWait); err != nil {
		return fmt.Errorf("kind ready: %w", err)
	}

	fmt.Println("Cluster created")
	return nil
}

func runCommand(c *exec.Cmd) error {
	if out, err := c.CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return err
	}
	return nil
}
