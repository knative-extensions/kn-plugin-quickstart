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
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// kubectlMinVersion is the minimum kubectl version we recommend. We gate on it
// because waitForPodsReady relies on `kubectl wait --for=create` with a label
// selector, which only works correctly on kubectl >= 1.33 (see
// kubernetes/kubernetes#128662).
//
// kubectlMinVersion must match DefaultKubernetesMinVersion in
// https://github.com/knative/pkg/blob/main/version/version.go (without the
// leading "v"), matching the kind and minikube cluster versions. The
// verify-min-k8s-version CI check enforces this. We only compare the major and
// minor components at runtime; the patch component is kept so the literal
// matches upstream's format for that check.
var kubectlMinVersion = "1.34.0"

// kubectlGitVersion matches the "vMAJOR.MINOR" prefix of kubectl's reported
// git version, e.g. "v1.34.0" or "v1.33.2-eks-1234".
var kubectlGitVersion = regexp.MustCompile(`^v?(\d+)\.(\d+)`)

// kubectlMinMajor / kubectlMinMinor are parsed from kubectlMinVersion.
var kubectlMinMajor, kubectlMinMinor = mustParseMinVersion(kubectlMinVersion)

// mustParseMinVersion parses the leading MAJOR.MINOR out of kubectlMinVersion.
// It panics on a malformed literal, which can only happen if kubectlMinVersion
// is edited to an invalid value (a programming/CI error, caught by tests).
func mustParseMinVersion(v string) (int, int) {
	m := kubectlGitVersion.FindStringSubmatch(v)
	if m == nil {
		panic(fmt.Sprintf("kn-plugin-quickstart: malformed kubectlMinVersion %q", v))
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	return major, minor
}

// Component versions are generated at buildtime via the hack/build.sh script
var ServingVersion string
var KourierVersion string
var EventingVersion string

// Kourier installs Kourier networking layer from Github YAML files
func Kourier() error {
	fmt.Println("🕸️ Installing Kourier networking layer v" + KourierVersion + " ...")

	if err := retryingApply("https://github.com/knative-sandbox/net-kourier/releases/download/knative-v" + KourierVersion + "/kourier.yaml"); err != nil {
		return fmt.Errorf("wait: %w", err)
	}
	if err := waitForPodsReady("kourier-system"); err != nil {
		return fmt.Errorf("kourier: %w", err)
	}
	if err := waitForPodsReady("knative-serving"); err != nil {
		return fmt.Errorf("serving: %w", err)
	}
	fmt.Println("    Kourier installed...")

	ingress := exec.Command("kubectl", "patch", "configmap/config-network", "--namespace", "knative-serving", "--type", "merge", "--patch", "{\"data\":{\"ingress.class\":\"kourier.ingress.networking.knative.dev\"}}")
	if err := runCommand(ingress); err != nil {
		return fmt.Errorf("ingress error: %w", err)
	}
	fmt.Println("    Ingress patched...")

	fmt.Println("    Finished installing Kourier Networking layer")

	return nil
}

// KourierKind runs the kind-specific setup for Kourier
func KourierKind() error {
	fmt.Println("🕸️ Configuring Kourier for Kind...")

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

	domainDns := exec.Command("kubectl", "patch", "configmap", "-n", "knative-serving", "config-domain", "-p", "{\"data\": {\"127.0.0.1.sslip.io\": \"\"}}")
	if err := runCommand(domainDns); err != nil {
		return fmt.Errorf("domain dns: %w", err)
	}
	fmt.Println("    Domain DNS set up...")
	fmt.Println("    Finished configuring Kourier")

	return nil
}

// KourierMinikube runs the minikube-specific setup for Kourier
func KourierMinikube() error {
	fmt.Println("🕸️ Configuring Kourier for Minikube...")

	if err := retryingApply("https://github.com/knative/serving/releases/download/knative-v" + ServingVersion + "/serving-default-domain.yaml"); err != nil {
		return fmt.Errorf("default domain: %w", err)
	}
	if err := waitForPodsReady("knative-serving"); err != nil {
		return fmt.Errorf("core: %w", err)
	}

	fmt.Println("    Domain DNS set up...")

	fmt.Println("    Finished configuring Kourier")
	return nil
}

// Serving installs Knative Serving from Github YAML files
func Serving(registries string) error {
	fmt.Println("🍿 Installing Knative Serving v" + ServingVersion + " ...")
	baseURL := "https://github.com/knative/serving/releases/download/knative-v" + ServingVersion

	if err := retryingApply(baseURL + "/serving-crds.yaml"); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	if err := waitForCRDsEstablished(); err != nil {
		return fmt.Errorf("crds: %w", err)
	}
	fmt.Println("    CRDs installed...")

	if err := retryingApply(baseURL + "/serving-core.yaml"); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	if err := waitForPodsReady("knative-serving"); err != nil {
		return fmt.Errorf("core: %w", err)
	}

	fmt.Println("    Core installed...")

	// Wait for webhook to be ready before attempting to patch configmaps
	if err := waitForWebhookReady(); err != nil {
		return fmt.Errorf("webhook: %w", err)
	}

	if registries != "" {
		configPatch := fmt.Sprintf(`{"data":{"registries-skipping-tag-resolving":"%s"}}`, registries)
		ignoreRegistry := exec.Command("kubectl", "patch", "configmap", "-n", "knative-serving", "config-deployment", "-p", configPatch)
		if err := runCommand(ignoreRegistry); err != nil {
			return fmt.Errorf("tag resolving configuration: %w", err)
		}
		fmt.Println("    Enabled local registry deployment...")
	}

	fmt.Println("    Finished installing Knative Serving")

	return nil
}

// Eventing installs Knative Eventing from Github YAML files
func Eventing() error {
	fmt.Println("🔥 Installing Knative Eventing v" + EventingVersion + " ... ")
	baseURL := "https://github.com/knative/eventing/releases/download/knative-v" + EventingVersion

	if err := retryingApply(baseURL + "/eventing-crds.yaml"); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	if err := waitForCRDsEstablished(); err != nil {
		return fmt.Errorf("crds: %w", err)
	}
	fmt.Println("    CRDs installed...")

	if err := retryingApply(baseURL + "/eventing-core.yaml"); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	if err := waitForPodsReady("knative-eventing"); err != nil {
		return fmt.Errorf("core: %w", err)
	}
	fmt.Println("    Core installed...")

	if err := retryingApply(baseURL + "/in-memory-channel.yaml"); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	if err := waitForPodsReady("knative-eventing"); err != nil {
		return fmt.Errorf("channel: %w", err)
	}
	fmt.Println("    In-memory channel installed...")

	if err := retryingApply(baseURL + "/mt-channel-broker.yaml"); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	if err := waitForPodsReady("knative-eventing"); err != nil {
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
	fmt.Println("    Finished installing Knative Eventing")

	return nil
}

func runCommand(c *exec.Cmd) error {
	if out, err := c.CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return err
	}
	return nil
}

// retryingApply retries a kubectl apply call with the given path 3 times, sleeping
// for 10s between each try.
func retryingApply(path string) error {
	cmd := exec.Command("kubectl", "apply", "-f", path)
	var err error
	for i := 0; i < 3; i++ {
		err = runCommand(cmd)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Second)
	}
	return err
}

// waitForCRDsEstablished waits for all CRDs to be established.
func waitForCRDsEstablished() error {
	return runCommand(exec.Command("kubectl", "wait", "--for=condition=Established", "--all", "crd"))
}

// CheckKubectlVersion validates that the user has a recent enough version of
// kubectl installed. If not, it warns the user and prompts them to continue,
// mirroring the behavior of the kind and minikube version checks.
func CheckKubectlVersion() error {
	versionCheck := exec.Command("kubectl", "version", "--client", "-o", "json")
	out, err := versionCheck.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get kubectl version: %w", err)
	}

	major, minor, err := parseKubectlVersion(string(out))
	if err != nil {
		return fmt.Errorf("unable to parse kubectl version: %w", err)
	}
	fmt.Printf("    kubectl version is: v%d.%d\n", major, minor)

	if major < kubectlMinMajor || (major == kubectlMinMajor && minor < kubectlMinMinor) {
		var resp string
		fmt.Printf("WARNING: We recommend at least kubectl v%d.%d, while you are using v%d.%d\n", kubectlMinMajor, kubectlMinMinor, major, minor)
		fmt.Println("You can download a newer version from https://kubernetes.io/docs/tasks/tools/install-kubectl")
		fmt.Print("Continue anyway? (not recommended) [y/N]: ")
		fmt.Scanf("%s", &resp)
		if resp != "y" && resp != "Y" {
			fmt.Println("Installation stopped. Please upgrade kubectl and run again")
			os.Exit(0)
		}
	}

	return nil
}

// parseKubectlVersion extracts the client major and minor version from the
// JSON output of `kubectl version --client -o json`. It reads the gitVersion
// field (e.g. "v1.34.0") rather than the major/minor fields, which some
// distributions emit with non-numeric suffixes (e.g. minor "33+").
func parseKubectlVersion(jsonOut string) (int, int, error) {
	// Pull gitVersion out of the JSON without a full struct decode so we stay
	// resilient to extra fields; fall back to matching any vX.Y in the blob.
	gitVersion := ""
	if idx := strings.Index(jsonOut, `"gitVersion"`); idx >= 0 {
		rest := jsonOut[idx:]
		if start := strings.Index(rest, `:`); start >= 0 {
			rest = rest[start+1:]
			if open := strings.Index(rest, `"`); open >= 0 {
				rest = rest[open+1:]
				if close := strings.Index(rest, `"`); close >= 0 {
					gitVersion = rest[:close]
				}
			}
		}
	}

	m := kubectlGitVersion.FindStringSubmatch(gitVersion)
	if m == nil {
		return 0, 0, fmt.Errorf("could not find a version in kubectl output: %q", gitVersion)
	}
	major, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, 0, err
	}
	minor, err := strconv.Atoi(m[2])
	if err != nil {
		return 0, 0, err
	}
	return major, minor, nil
}

// waitForPodsReady waits for all pods in the given namespace to be ready.
//
// We pass both --for=create and --for=condition=Ready because kubectl wait
// exits immediately with "no matching resources found" when no pods match the
// selector yet (e.g. the deployment controller hasn't created them). --for=create
// is always evaluated first, so this waits for the pods to appear and then for
// them to become Ready. This requires kubectl >= 1.33 for label-selector waits,
// which is within the versions this plugin already supports.
func waitForPodsReady(ns string) error {
	return runCommand(exec.Command("kubectl", "wait", "pod", "--timeout=10m", "--for=create", "--for=condition=Ready", "-l", "!job-name", "-n", ns))
}

// waitForWebhookReady waits for the Knative Serving webhook to be ready.
func waitForWebhookReady() error {
	fmt.Println("    Waiting for webhook to be ready...")

	// Retry for up to 2 minutes (12 attempts with 10s intervals)
	for range 12 {
		// Check if the webhook service has ready endpointslices
		// NOTE: We use 'kubectl get' instead of 'kubectl wait' because kubectl wait's JSONPath
		// doesn't support checking if ANY endpoint is ready (wildcard [*] fails with multiple endpoints)
		checkEndpointSlices := exec.Command("kubectl", "get", "endpointslice",
			"-l", "kubernetes.io/service-name=webhook",
			"-n", "knative-serving",
			"-o", "jsonpath={.items[*].endpoints[?(@.conditions.ready==true)]}")

		output, err := checkEndpointSlices.CombinedOutput()
		if err == nil && strings.TrimSpace(string(output)) != "" {
			fmt.Println("    Webhook is ready...")
			return nil
		}

		fmt.Println("    Webhook not ready yet, waiting...")
		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("timeout waiting for webhook to be ready")
}
