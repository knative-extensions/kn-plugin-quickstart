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

package minikube

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"knative.dev/kn-plugin-quickstart/pkg/install"
)

var clusterName = "minikube-knative"
var minikubeVersion = "v1.23"

// SetUp creates a local Minikube cluster and installs all the relevant Knative components
func SetUp() error {
	if err := createMinikubeCluster(); err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	if err := install.Serving(); err != nil {
		return fmt.Errorf("install serving: %w", err)
	}
	if err := install.Kourier(); err != nil {
		return fmt.Errorf("install kourier: %w", err)
	}
	if err := install.KourierMinikube(); err != nil {
		return fmt.Errorf("configure kourier: %w", err)
	}
	if err := install.Eventing(); err != nil {
		return fmt.Errorf("install eventing: %w", err)
	}
	return nil
}

func createMinikubeCluster() error {
	if err := checkMinikubeVersion(); err != nil {
		return fmt.Errorf("minikube version: %w", err)
	}
	if err := checkForExistingCluster(); err != nil {
		return fmt.Errorf("existing cluster: %w", err)
	}
	return nil
}

// checkMinikubeVersion validates that the user has the correct version of Minikube installed.
// If not, it prompts the user to download a newer version before continuing.
func checkMinikubeVersion() error {
	versionCheck := exec.Command("minikube", "version", "--short")
	out, err := versionCheck.CombinedOutput()
	if err != nil {
		return fmt.Errorf("minikube version: %w", err)
	}
	fmt.Printf("Minikube version is: %s\n", string(out))

	r := regexp.MustCompile(minikubeVersion)
	matches := r.Match(out)
	if !matches {
		var resp string
		fmt.Printf("WARNING: Please make sure you are using Minikube version %s.x\n", minikubeVersion)
		fmt.Println("Download from https://github.com/kubernetes/minikube/releases/")
		fmt.Print("Do you want to continue at your own risk [Y/n]: ")
		fmt.Scanf("%s", &resp)
		if resp == "n" || resp == "N" {
			fmt.Println("Installation stopped. Please upgrade minikube and run again")
			os.Exit(0)
		}
	}

	return nil
}

// checkForExistingCluster checks if the user already has a Minikube cluster. If so, it provides
// the option of deleting the existing cluster and recreating it. If not, it proceeds to
// creating a new cluster
func checkForExistingCluster() error {

	getClusters := exec.Command("minikube", "profile", "list")
	out, err := getClusters.CombinedOutput()
	if err != nil {
		// there are no existing minikube profiles, the listing profiles command will error
		// if there were no profiles, we simply want to create a new one and not stop the install
		// so if the error contains a "no profile found" string, we ignore it and continue onwards
		if !strings.Contains(string(out), "No minikube profile was found") {
			return fmt.Errorf("check cluster: %w", err)
		}
	}
	// TODO Add tests for regex
	r := regexp.MustCompile(clusterName)
	matches := r.Match(out)
	if matches {
		var resp string
		fmt.Print("Knative Cluster " + clusterName + " already installed.\nDelete and recreate [y/N]: ")
		fmt.Scanf("%s", &resp)
		if resp == "y" || resp == "Y" {
			fmt.Println("deleting cluster...")
			deleteCluster := exec.Command("minikube", "delete", "--profile", clusterName)
			if err := deleteCluster.Run(); err != nil {
				return fmt.Errorf("delete cluster: %w", err)
			}
			if err := createNewCluster(); err != nil {
				return fmt.Errorf("new cluster: %w", err)
			}
		} else {
			fmt.Println("Installation skipped")
			return nil
		}
	} else {
		if err := createNewCluster(); err != nil {
			return fmt.Errorf("new cluster: %w", err)
		}
	}

	return nil
}

// createNewCluster creates a new Minikube cluster
func createNewCluster() error {

	fmt.Println("Creating Minikube cluster...")

	createCluster := exec.Command("minikube", "start", "--profile", clusterName, "--wait", "all")
	if err := runCommand(createCluster); err != nil {
		return fmt.Errorf("minikube create: %w", err)
	}

	// sleep for 10s to allow initial cluster creation, then wait until all pods in kube-system namespace are ready
	fmt.Println("    Waiting on cluster to be ready...")
	clusterWait := exec.Command("kubectl", "wait", "pod", "--timeout=-1s", "--for=condition=Ready", "-l", "!job-name", "-n", "kube-system")
	if err := runCommand(clusterWait); err != nil {
		return fmt.Errorf("minikube ready: %w", err)
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
