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
	"strconv"
	"strings"
	"syscall"
	"time"

	"knative.dev/kn-plugin-quickstart/pkg/install"
)

type MinikubeConfig struct {
	clusterName            string
	kubernetesVersion      string
	clusterVersionOverride bool
	minikubeVersion        float64
	cpus                   string
	memory                 string
	installKnative         bool
	exposeIp               string
}

var config = MinikubeConfig{
	clusterName:            "minikube-knative",
	kubernetesVersion:      "1.30.5",
	clusterVersionOverride: false,
	minikubeVersion:        1.35,
	cpus:                   "4",
	memory:                 "8g",
	installKnative:         true,
	exposeIp:               "127.0.0.1",
}

var domain = config.exposeIp + ".nip.io"
var defaultRegistryPort = "5000"

// SetUp creates a local Minikube cluster and installs all the relevant Knative components
func SetUp(
	name, kVersion string,
	installServing, installEventing bool,
	registryPort string,
	extraMountHostPath string,
	extraMountContainerPath string,
) error {
	start := time.Now()

	// if neither --install-serving nor --install-eventing is set, assume both
	if !installServing && !installEventing {
		installServing = true
		installEventing = true
	}

	// kubectl is required, fail if not found
	if _, err := exec.LookPath("kubectl"); err != nil {
		fmt.Println("ERROR: kubectl is required for quickstart")
		fmt.Println("Download from https://kubectl.docs.kubernetes.io/installation/kubectl/")
		os.Exit(1)
	}

	config.clusterName = name
	if kVersion != "" {
		config.kubernetesVersion = kVersion
		config.clusterVersionOverride = true
	}

	if err := createMinikubeCluster(extraMountHostPath, extraMountContainerPath); err != nil {
		return fmt.Errorf("❌ creating cluster: %w", err)
	}

	fmt.Println()

	if err := enableRegistryAddon(registryPort); err != nil {
		return fmt.Errorf("❌ enabling registry addon: %w", err)
	}

	// Configure Minikube registry helper for image pushing
	if err := setupMinikubeRegistryHelper(); err != nil {
		return fmt.Errorf("❌ setting up registry helper: %w", err)
	}
	fmt.Println("✅ Minikube is now configured with registry support!")

	// Pause to let the user review before continuing.
	time.Sleep(2 * time.Second)
	fmt.Println("\n⏎ Press the Enter key to continue")
	fmt.Scanln()

	if config.installKnative {
		if installServing {
			if err := install.Serving(""); err != nil {
				return fmt.Errorf("❌ install serving: %w", err)
			}
			if err := kourierMinikube(); err != nil {
				return fmt.Errorf("❌ configure kourier: %w", err)
			}
		}
		if installEventing {
			if err := install.Eventing(); err != nil {
				return fmt.Errorf("❌ install eventing: %w", err)
			}
		}
	}

	if err := listAllPods(); err != nil {
		fmt.Printf("Warning: could not list pods: %v\n", err)
	}

	// Start Minikube Tunnel to expose LoadBalancer
	if err := startMinikubeTunnel(); err != nil {
		return fmt.Errorf("❌ failed to start Minikube tunnel: %w", err)
	}

	setProfileCmd := exec.Command("minikube", "profile", config.clusterName)
	if err := setProfileCmd.Run(); err != nil {
		return fmt.Errorf("❌ setting minikube profile: %w", err)
	}

	finish := time.Since(start).Round(time.Second)
	fmt.Printf("🎯 Minikube default profile set to %s\n", config.clusterName)
	fmt.Printf("🚀 Knative install took: %s\n", finish)
	fmt.Println("🎉 Now have some fun with Serverless and Event Driven Apps!")

	return nil
}

func kourierMinikube() error {
	fmt.Println("🕸  Configuring Kourier for Minikube...")
	if err := install.RetryingApply("https://github.com/knative/net-kourier/releases/latest/download/kourier.yaml"); err != nil {
		return fmt.Errorf("❌ default domain: %w", err)
	}

	fmt.Println("    Waiting for Kourier pods to be ready...")
	if err := install.WaitForPodsReady("knative-serving"); err != nil {
		return fmt.Errorf("❌ core: %w", err)
	}

	if err := patchKourierLoadBalancer(); err != nil {
		return fmt.Errorf("❌ failed to patch Kourier with LoadBalancer: %w", err)
	}
	fmt.Println("    Kourier installation and patching complete!")

	if err := install.WaitForPodsReady("knative-serving"); err != nil {
		return fmt.Errorf("❌ core: %w", err)
	}

	if err := patchKnativeDomain(); err != nil {
		return fmt.Errorf("❌ patching knative domain: %w", err)
	}

	if err := install.WaitForPodsReady("knative-serving"); err != nil {
		return fmt.Errorf("❌ core: %w", err)
	}

	if err := configureKnativeIngress(); err != nil {
		return fmt.Errorf("❌ Failed to configure Knative ingress: %w", err)
	}

	if err := install.WaitForPodsReady("knative-serving"); err != nil {
		return fmt.Errorf("❌ core: %w", err)
	}

	fmt.Println("    Domain DNS set up...")
	fmt.Println("    Finished configuring Kourier")
	return nil
}

func createMinikubeCluster(
	extraMountHostPath string,
	extraMountContainerPath string,
) error {
	if err := checkMinikubeVersion(); err != nil {
		return fmt.Errorf("❌ minikube version: %w", err)
	}
	if err := checkForExistingCluster(extraMountHostPath, extraMountContainerPath); err != nil {
		return fmt.Errorf("❌ existing cluster: %w", err)
	}
	return nil
}

func checkMinikubeVersion() error {
	versionCheck := exec.Command("minikube", "version", "--short")
	out, err := versionCheck.CombinedOutput()
	if err != nil {
		return fmt.Errorf("❌ minikube version: %w", err)
	}
	fmt.Printf("Minikube version is: %s\n", string(out))

	userMinikubeVersion, err := parseMinikubeVersion(string(out))
	if err != nil {
		return fmt.Errorf("❌ parsing minikube version: %w", err)
	}
	if userMinikubeVersion < config.minikubeVersion {
		var resp string
		fmt.Printf("WARNING: We recommend at least Minikube v%.2f, while you are using v%.2f\n", config.minikubeVersion, userMinikubeVersion)
		fmt.Println("You can download a newer version from https://github.com/kubernetes/minikube/releases/")
		fmt.Print("Continue anyway? (not recommended) [y/N]: ")
		fmt.Scanf("%s", &resp)
		if strings.ToLower(resp) != "y" {
			fmt.Println("Installation stopped. Please upgrade minikube and run again")
			os.Exit(0)
		}
	}
	return nil
}

func checkForExistingCluster(
	extraMountHostPath string,
	extraMountContainerPath string,
) error {
	getClusters := exec.Command("minikube", "profile", "list")
	out, err := getClusters.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(out), "MK_USAGE_NO_PROFILE") {
			return fmt.Errorf("❌ check cluster: %w", err)
		}
	}
	r := regexp.MustCompile(config.clusterName)
	matches := r.Match(out)
	if matches {
		var resp string
		fmt.Printf("⚠️ Knative Cluster %s already installed.\nDelete and recreate [y/N]: ", config.clusterName)
		fmt.Scanf("%s", &resp)
		if strings.ToLower(resp) != "y" {
			fmt.Println("Installation skipped")
			checkKnativeNamespace := exec.Command("kubectl", "get", "namespaces")
			output, err := checkKnativeNamespace.CombinedOutput()
			namespaces := string(output)
			if err != nil {
				fmt.Println(string(output))
				return fmt.Errorf("❌ check existing cluster: %w", err)
			}
			if strings.Contains(namespaces, "knative") {
				fmt.Print("⚠️ Knative installation already exists.\nDelete and recreate the cluster [y/N]: ")
				fmt.Scanf("%s", &resp)
				if strings.ToLower(resp) != "y" {
					fmt.Println("Skipping installation")
					config.installKnative = false
					return nil
				} else {
					return recreateCluster(extraMountHostPath, extraMountContainerPath)
				}
			}
			return nil
		}
		return recreateCluster(extraMountHostPath, extraMountContainerPath)
	}
	return createNewCluster(extraMountHostPath, extraMountContainerPath)
}

func createNewCluster(
	extraMountHostPath string,
	extraMountContainerPath string,
) error {
	fmt.Println("☸ Creating Minikube cluster...")

	if !config.clusterVersionOverride {
		if kVersion, ok := getMinikubeConfig("kubernetes-version"); ok {
			config.kubernetesVersion = kVersion
		}
	}

	// Get user configs for memory/cpus if they exist.
	if cpus, ok := getMinikubeConfig("cpus"); ok {
		config.cpus = cpus
	}
	if memory, ok := getMinikubeConfig("memory"); ok {
		config.memory = memory
	}

	fmt.Println("ℹ️  Using the Docker minikube driver")
	fmt.Println("If you wish to use a different driver, please configure minikube using")
	fmt.Println("    minikube config set driver <your-driver>")
	fmt.Println()

	createClusterCmd := exec.Command(
		"minikube", "start",
		"--kubernetes-version", config.kubernetesVersion,
		"--driver", "docker", // using the Docker driver as default
		"--apiserver-ips", config.exposeIp,
		"--cpus", config.cpus,
		"--memory", config.memory,
		"--profile", config.clusterName,
		"--wait", "all",
		"--insecure-registry", "10.0.0.0/24",
		"--cache-images=false",
		"--extra-config=apiserver.service-node-port-range=1-65535",
	)

	// If extra mount paths are specified, pass them to minikube.
	if extraMountHostPath != "" && extraMountContainerPath != "" {
		createClusterCmd.Args = append(
			createClusterCmd.Args,
			"--mount",
			"--mount-string="+extraMountHostPath+":"+extraMountContainerPath,
		)
	}

	fmt.Println("Creating Minikube cluster with the following options:")
	fmt.Printf("  Default Profile: %s\n", config.clusterName)
	fmt.Printf("  Kubernetes version: %s\n", config.kubernetesVersion)
	fmt.Printf("  Driver: Docker\n")
	fmt.Printf("  CPUs: %s\n", config.cpus)
	fmt.Printf("  Memory: %s\n", config.memory)
	fmt.Printf("  Knative domain: %s\n", domain)
	fmt.Println()

	if err := runCommandWithOutput(createClusterCmd); err != nil {
		return fmt.Errorf("❌ minikube create: %w", err)
	}

	return nil
}

func recreateCluster(
	extraMountHostPath string,
	extraMountContainerPath string,
) error {
	fmt.Println("🗑️  deleting cluster...")
	deleteCluster := exec.Command("minikube", "delete", "--profile", config.clusterName)
	if err := deleteCluster.Run(); err != nil {
		return fmt.Errorf("❌ delete cluster: %w", err)
	}
	if err := createNewCluster(extraMountHostPath, extraMountContainerPath); err != nil {
		return fmt.Errorf("❌ new cluster: %w", err)
	}
	return nil
}

func parseMinikubeVersion(v string) (float64, error) {
	strippedVersion := strings.TrimLeft(strings.TrimRight(v, "\n"), "v")
	dotVersion := strings.Split(strippedVersion, ".")
	floatVersion, err := strconv.ParseFloat(dotVersion[0]+"."+dotVersion[1], 64)
	if err != nil {
		return 0, err
	}
	return floatVersion, nil
}

func getMinikubeConfig(k string) (string, bool) {
	getConfig := exec.Command("minikube", "config", "get", k)
	v, err := getConfig.Output()
	if err == nil {
		return strings.TrimRight(string(v), "\n"), true
	}
	return "", false
}

func runCommandWithOutput(c *exec.Cmd) error {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("❌ piping output: %w", err)
	}
	fmt.Println()
	return nil
}

func setupMinikubeRegistryHelper() error {
	fmt.Println("🔄 Configuring Minikube registry helper...")

	// Set Minikube's internal registry as the Docker environment.
	cmd := exec.Command("minikube", "-p", config.clusterName, "docker-env")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("❌ retrieving minikube docker-env: %w", err)
	}

	// Apply the environment variables.
	envVars := strings.Split(string(output), "\n")
	for _, line := range envVars {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(parts[1], `"`)
			os.Setenv(key, value)
		}
	}

	return nil
}

func enableRegistryAddon(registryPort string) error {
	fmt.Printf("🔌 Enabling registry addon with port %s...\n", registryPort)

	cmd := exec.Command("minikube", "addons", "enable", "registry", "--profile", config.clusterName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ failed to enable registry addon: %v", err)
	}

	if registryPort != defaultRegistryPort {
		patchPayload := fmt.Sprintf(
			`[{"op": "replace", "path": "/spec/template/spec/containers/0/ports/0/hostPort", "value": %s}]`,
			registryPort,
		)
		fmt.Printf("🔧 Patching registry-proxy daemonset to host port %s...\n", registryPort)
		patchCmd := exec.Command("kubectl", "-n", "kube-system", "patch", "daemonset", "registry-proxy", "--type=json", "-p", patchPayload)
		patchCmd.Stdout = os.Stdout
		patchCmd.Stderr = os.Stderr
		if err := patchCmd.Run(); err != nil {
			return fmt.Errorf("❌ failed to patch registry-proxy daemonset: %v", err)
		}
	}

	return nil
}

func retryCommand(cmd *exec.Cmd, maxRetries int, delay time.Duration) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		output, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		fmt.Printf("❌ Retry %d/%d failed: %s\n", i+1, maxRetries, strings.TrimSpace(string(output)))
		time.Sleep(delay)
	}
	return err
}

func patchKnativeDomain() error {
	fmt.Printf("🩹 Patching Knative config-domain\n")
	domainPatch := fmt.Sprintf(`{"data": {"%s": ""}}`, domain)

	cmd := exec.Command("kubectl", "get", "pods", "-n", "knative-serving",
		"-l", "app=webhook",
		"-o", "jsonpath={.items[0].status.phase}")

	if err := retryCommand(cmd, 10, 5*time.Second); err != nil {
		return fmt.Errorf("❌ webhook pod not ready after several retries: %w", err)
	}

	fmt.Println("    Webhook pod is running, proceeding with patch...")

	patchCmd := exec.Command("kubectl", "patch", "configmap", "config-domain", "-n", "knative-serving",
		"--type=merge", "-p", domainPatch)

	if err := retryCommand(patchCmd, 3, 10*time.Second); err != nil {
		return fmt.Errorf("❌ patching config-domain failed after retries: %w", err)
	}

	fmt.Printf("    Configured Knative domain to use %s\n", domain)
	return nil
}

func listAllPods() error {
	cmd := exec.Command("minikube", "--profile", config.clusterName, "kubectl", "--", "get", "pods", "-A")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("❌ error executing 'minikube kubectl -- get pods -A': %w\nOutput: %s", err, output)
	}

	fmt.Printf("📦 All pods in %s:\n", config.clusterName)
	outputLines := strings.Split(string(output), "\n")
	for _, line := range outputLines {
		fmt.Printf("    %s\n", line)
	}
	return nil
}

func patchKourierLoadBalancer() error {
	fmt.Println("⚖️ Ensure Kourier is using LoadBalancer...")

	getTypeCmd := exec.Command("kubectl", "get", "svc", "kourier", "-n", "kourier-system",
		"-o", "jsonpath={.spec.type}")

	serviceTypeBytes, err := getTypeCmd.Output()
	if err != nil {
		return fmt.Errorf("❌ failed to get Kourier service type: %w", err)
	}

	serviceType := strings.TrimSpace(string(serviceTypeBytes))

	if serviceType == "LoadBalancer" {
		fmt.Println("    Kourier is already set to LoadBalancer. No patching needed.")
		return nil
	}

	fmt.Println("    Patching Kourier service to LoadBalancer...")
	patchCmd := exec.Command("kubectl", "patch", "svc", "kourier", "-n", "kourier-system",
		"--type=merge", "-p", `{"spec": {"type": "LoadBalancer"}}`)

	if err := retryCommand(patchCmd, 3, 10*time.Second); err != nil {
		return fmt.Errorf("❌ failed to patch Kourier service after retries: %w", err)
	}

	fmt.Println("    Kourier service is using LoadBalancer.")
	return nil
}

func startMinikubeTunnel() error {
	fmt.Println("🚇 Setup Minikube tunnel in the background to expose LoadBalancer services...")

	// Find existing Minikube tunnel process
	checkCmd := exec.Command("pgrep", "-f", "minikube tunnel")
	_, err := checkCmd.Output()
	if err == nil {
		fmt.Println("    🛑 Existing Minikube tunnel found, stopping it...")
		killCmd := exec.Command("pkill", "-f", "minikube tunnel")
		if err := killCmd.Run(); err != nil {
			return fmt.Errorf("❌ Failed to stop existing Minikube tunnel: %w", err)
		}
		time.Sleep(2 * time.Second)
	}

	fmt.Println("    🟢 Starting Minikube tunnel in the background...")

	// Use `nohup` to ensure it runs persistently, with `sudo` for permissions
	tunnelCmd := exec.Command("nohup", "sudo", "minikube", "tunnel", "-p", config.clusterName)
	tunnelCmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} // Detach from parent
	tunnelCmd.Stdout = nil
	tunnelCmd.Stderr = nil

	if err := tunnelCmd.Start(); err != nil {
		return fmt.Errorf("❌ Failed to start Minikube tunnel: %w", err)
	}

	fmt.Println("    ✅ Minikube tunnel is now running in the background.")
	return nil
}

func configureKnativeIngress() error {
	fmt.Println("🔄 Configuring Knative to use Kourier as the only ingress class...")

	fmt.Println("    🚨 Removing existing config-network ConfigMap to enforce Kourier...")

	deleteCmd := exec.Command("kubectl", "delete", "cm", "config-network", "-n", "knative-serving")
	if err := retryCommand(deleteCmd, 3, 10*time.Second); err != nil {
		fmt.Printf("⚠️ Warning: Could not delete config-network (may not exist yet): %v\n", err)
	}

	fmt.Println("    🗑️  ConfigMap deleted. Now recreating with Kourier as the only ingress class...")
	createCmd := exec.Command("kubectl", "create", "configmap", "config-network", "-n", "knative-serving",
		"--from-literal=ingress-class=kourier.ingress.networking.knative.dev")
	if err := retryCommand(createCmd, 3, 10*time.Second); err != nil {
		return fmt.Errorf("❌ Failed to create config-network ConfigMap after retries: %w", err)
	}

	fmt.Println("    ✅ Knative is now using Kourier as the only ingress class.")

	return nil
}
