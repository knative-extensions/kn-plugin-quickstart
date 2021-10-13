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
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"knative.dev/kn-plugin-quickstart/pkg/install"
)

var kubernetesVersion = "v1.21.1@sha256:fae9a58f17f18f06aeac9772ca8b5ac680ebbed985e266f711d936e91d113bad"
var clusterName = "knative"
var kindVersion = 0.11

// SetUp creates a local Kind cluster and installs all the relevant Knative components
func SetUp() error {
	start := time.Now()
	if err := createKindCluster(); err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	if err := install.Serving(); err != nil {
		return fmt.Errorf("install serving: %w", err)
	}
	if err := install.Kourier(); err != nil {
		return fmt.Errorf("install kourier: %w", err)
	}
	if err := install.KourierKind(); err != nil {
		return fmt.Errorf("configure kourier: %w", err)
	}
	if err := install.Eventing(); err != nil {
		return fmt.Errorf("install eventing: %w", err)
	}
	finish := time.Since(start).Round(time.Second)
	fmt.Printf("🚀 Knative install took: %s \n", finish)
	fmt.Println("🎉 Now have some fun with Serverless and Event Driven Apps!")
	return nil
}

func createKindCluster() error {

	fmt.Println("✅ Checking dependencies...")
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

	versionCheck := exec.Command("kind", "version", "-q")
	out, err := versionCheck.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kind version: %w", err)
	}
	fmt.Printf("    Kind version is: %s\n", string(out))

	userKindVersion, err := parseKindVersion(string(out))
	if err != nil {
		fmt.Errorf("parsing kind version: %w", err)
	}
	if userKindVersion < kindVersion {
		var resp string
		fmt.Printf("WARNING: Please make sure you are using Kind version %.2f or later", kindVersion)
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
	if err != nil {
		return fmt.Errorf("check cluster: %w", err)
	}
	// TODO Add tests for regex
	r := regexp.MustCompile(`(?m)^knative\n`)
	matches := r.Match(out)
	if matches {
		var resp string
		fmt.Print("Knative Cluster kind-" + clusterName + " already installed.\nDelete and recreate [y/N]: ")
		fmt.Scanf("%s", &resp)
		if resp == "y" || resp == "Y" {
			fmt.Println("\n    Deleting cluster...")
			deleteCluster := exec.Command("kind", "delete", "cluster", "--name", clusterName)
			if err := deleteCluster.Run(); err != nil {
				return fmt.Errorf("delete cluster: %w", err)
			}
			if err := createNewCluster(); err != nil {
				return fmt.Errorf("new cluster: %w", err)
			}
		} else {
			fmt.Println("\n    Installation skipped")
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

	fmt.Println("☸ Creating Kind cluster...")
	config := fmt.Sprintf(`
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: %s
nodes:
- role: control-plane
  image: kindest/node:%s
  extraPortMappings:
  - containerPort: 31080
    listenAddress: 127.0.0.1
    hostPort: 80`, clusterName, kubernetesVersion)

	createCluster := exec.Command("kind", "create", "cluster", "--wait=120s", "--config=-")
	createCluster.Stdin = strings.NewReader(config)
	if err := runCommand(createCluster); err != nil {
		return fmt.Errorf("kind create: %w", err)
	}

	fmt.Println("    Cluster ready")
	return nil
}

func runCommand(c *exec.Cmd) error {
	if out, err := c.CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return err
	}
	return nil
}

func parseKindVersion(v string) (float64, error) {
	strippedVersion := strings.TrimLeft(strings.TrimRight(v, "\n"), "v")
	dotVersion := strings.Split(strippedVersion, ".")
	floatVersion, err := strconv.ParseFloat(dotVersion[0]+"."+dotVersion[1], 64)
	if err != nil {
		return 0, err
	}
	return floatVersion, nil
}
