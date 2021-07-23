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
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

var servingVersion = "0.24.0"
var kourierVersion = "0.24.0"
var eventingVersion = "0.24.1"

// Kourier installs Kourier networking layer from Github YAML files
func Kourier() error {

	fmt.Println("Starting Kourier Networking layer " + kourierVersion + " install...")

	kourier := exec.Command("kubectl", "apply", "-f", "https://github.com/knative-sandbox/net-kourier/releases/download/v"+kourierVersion+"/kourier.yaml")
	// retries installing kourier if it fails, see discussion in:
	// https://github.com/knative-sandbox/kn-plugin-quickstart/pull/58
	for i := 0; i <= 3; i++ {
		if err := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
			return runCommand(kourier) == nil, nil
		}); err != nil {
			if i >= 3 {
				return fmt.Errorf("wait: %w", err)
			}
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}

	kourierWait := exec.Command("kubectl", "wait", "pod", "--timeout=-1s", "--for=condition=Ready", "-l", "!job-name", "-n", "kourier-system")
	if err := runCommand(kourierWait); err != nil {
		return fmt.Errorf("kourier: %w", err)
	}
	servingWait := exec.Command("kubectl", "wait", "pod", "--timeout=-1s", "--for=condition=Ready", "-l", "!job-name", "-n", "knative-serving")
	if err := runCommand(servingWait); err != nil {
		return fmt.Errorf("serving: %w", err)
	}
	fmt.Println("    Kourier installed...")

	ingress := exec.Command("kubectl", "patch", "configmap/config-network", "--namespace", "knative-serving", "--type", "merge", "--patch", "{\"data\":{\"ingress.class\":\"kourier.ingress.networking.knative.dev\"}}")
	if err := runCommand(ingress); err != nil {
		return fmt.Errorf("ingress error: %w", err)
	}
	fmt.Println("    Ingress patched...")

	config := `apiVersion: v1
kind: Service
metadata:
  name: kourier-ingress
  namespace: kourier-system
  labels:
    networking.knative.dev/ingress-provider: kourier
spec:
  type: NodePort
  selector:
    app: 3scale-kourier-gateway
  ports:
    - name: http2
      nodePort: 31080
      port: 80
      targetPort: 8080`

	kourierIngress := exec.Command("kubectl", "apply", "-f", "-")
	kourierIngress.Stdin = strings.NewReader(config)
	if err := runCommand(kourierIngress); err != nil {
		return fmt.Errorf("kourier service: %w", err)
	}

	fmt.Println("    Kourier service installed...")

	domainDns := exec.Command("kubectl", "patch", "configmap", "-n", "knative-serving", "config-domain", "-p", "{\"data\": {\"127.0.0.1.nip.io\": \"\"}}")
	if err := domainDns.Run(); err != nil {
		return fmt.Errorf("domain dns: %w", err)
	}
	fmt.Println("    Domain DNS set up...")

	fmt.Println("Finished installing Networking layer")

	return nil
}

// Serving installs Knative Serving from Github YAML files
func Serving() error {
	fmt.Println("Starting Knative Serving " + servingVersion + " install...")

	crds := exec.Command("kubectl", "apply", "-f", "https://github.com/knative/serving/releases/download/v"+servingVersion+"/serving-crds.yaml")
	if err := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		return runCommand(crds) == nil, nil
	}); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	crdWait := exec.Command("kubectl", "wait", "--for=condition=Established", "--all", "crd")
	if err := runCommand(crdWait); err != nil {
		return fmt.Errorf("crds: %w", err)
	}
	fmt.Println("    CRDs installed...")

	core := exec.Command("kubectl", "apply", "-f", "https://github.com/knative/serving/releases/download/v"+servingVersion+"/serving-core.yaml")
	if err := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		return runCommand(core) == nil, nil
	}); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	coreWait := exec.Command("kubectl", "wait", "pod", "--timeout=-1s", "--for=condition=Ready", "-l", "!job-name", "-n", "knative-serving")
	if err := runCommand(coreWait); err != nil {
		return fmt.Errorf("core: %w", err)
	}

	fmt.Println("    Core installed...")

	fmt.Println("Finished installing Knative Serving")

	return nil
}

// Eventing installs Knative Eventing from Github YAML files
func Eventing() error {
	fmt.Println("Starting Knative Eventing " + eventingVersion + " install...")

	crds := exec.Command("kubectl", "apply", "-f", "https://github.com/knative/eventing/releases/download/v"+eventingVersion+"/eventing-crds.yaml")
	if err := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		return runCommand(crds) == nil, nil
	}); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	crdWait := exec.Command("kubectl", "wait", "--for=condition=Established", "--all", "crd")
	if err := runCommand(crdWait); err != nil {
		return fmt.Errorf("crds: %w", err)
	}
	fmt.Println("    CRDs installed...")

	core := exec.Command("kubectl", "apply", "-f", "https://github.com/knative/eventing/releases/download/v"+eventingVersion+"/eventing-core.yaml")
	if err := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		return runCommand(core) == nil, nil
	}); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	coreWait := exec.Command("kubectl", "wait", "pod", "--timeout=-1s", "--for=condition=Ready", "-l", "!job-name", "-n", "knative-eventing")
	if err := runCommand(coreWait); err != nil {
		return fmt.Errorf("core: %w", err)
	}
	fmt.Println("    Core installed...")

	channel := exec.Command("kubectl", "apply", "-f", "https://github.com/knative/eventing/releases/download/v"+eventingVersion+"/in-memory-channel.yaml")
	if err := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		return runCommand(channel) == nil, nil
	}); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	channelWait := exec.Command("kubectl", "wait", "pod", "--timeout=-1s", "--for=condition=Ready", "-l", "!job-name", "-n", "knative-eventing")
	if err := runCommand(channelWait); err != nil {
		return fmt.Errorf("channel: %w", err)
	}
	fmt.Println("    In-memory channel installed...")

	broker := exec.Command("kubectl", "apply", "-f", "https://github.com/knative/eventing/releases/download/v"+eventingVersion+"/mt-channel-broker.yaml")
	if err := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		return runCommand(broker) == nil, nil
	}); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	brokerWait := exec.Command("kubectl", "wait", "pod", "--timeout=-1s", "--for=condition=Ready", "-l", "!job-name", "-n", "knative-eventing")
	if err := runCommand(brokerWait); err != nil {
		return fmt.Errorf("broker: %w", err)
	}
	fmt.Println("    Mt-channel broker installed...")

	config := `apiVersion: eventing.knative.dev/v1
kind: broker
metadata:
 name: example-broker
 namespace: default`

	exampleBroker := exec.Command("kubectl", "apply", "-f", "-")
	exampleBroker.Stdin = strings.NewReader(config)
	if err := runCommand(exampleBroker); err != nil {
		return fmt.Errorf("example broker: %w", err)
	}

	fmt.Println("    Example broker installed...")

	return nil
}

func runCommand(c *exec.Cmd) error {
	if out, err := c.CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return err
	}
	return nil
}
