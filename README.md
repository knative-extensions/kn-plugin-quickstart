# kn-plugin-quickstart

**[This component is BETA](https://github.com/knative/community/tree/main/mechanics/MATURITY-LEVELS.md)**

`kn-plugin-quickstart` is a plugin of the Knative Client, to enable users to quickly set up a local Knative environment from the command line.

## Getting Started

Note: In order to use the `quickstart` plugin, you must install the [Kubernetes CLI `kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl) and either [`kind`](https://kind.sigs.k8s.io/docs/user/quick-start) or [`minikube`](https://minikube.sigs.k8s.io/docs/start/).

### Installation

You can download the latest binaries from the [Releases](https://github.com/knative-sandbox/kn-plugin-quickstart/releases) page.

There are two ways to run `kn quickstart`:

1. You can run it standalone, just put it on your system path and make sure it is executable.
2. You can install it as a plugin of the `kn` client to run:
    * Follow the [documentation](https://github.com/knative/client/blob/main/docs/README.md#installing-kn) to install `kn client` if you don't have it
    * Copy the `kn-quickstart` binary to a directory on your `PATH` (for example, `/usr/local/bin`) and make sure its filename is `kn-quickstart`
    * Run `kn plugin list` to verify that the `kn-quickstart` plugin is installed successfully

After the plugin is installed, you can use `kn quickstart` to run its related subcommands.

## Usage

```
Get up and running with a local Knative environment

Usage:
  kn-quickstart [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  kind        Quickstart with Kind
  minikube    Quickstart with Minikube
  version     Prints the plugin version

Flags:
  -h, --help   help for kn-quickstart

Use "kn-quickstart [command] --help" for more information about a command.
```

### Quickstart with KinD

Set up a local Knative cluster using [KinD](https://kind.sigs.k8s.io/):

``` bash
kn quickstart kind
```

### Quickstart with Minikube

Set up a local Knative cluster using [Minikube](https://minikube.sigs.k8s.io/):

```bash
kn quickstart minikube
```

Note: for Windows/Mac users, after the above command completes, you will need to run the following in a separate terminal window:

``` bash
minikube tunnel --profile minikube-knative
```

## Building from Source

You must [set up your development environment](https://github.com/knative/client/blob/master/docs/DEVELOPMENT.md#prerequisites) before you build `kn-plugin-quickstart`.

Once you've set up your development environment, you can build the plugin by running the following commands:

``` bash
git clone git@github.com:knative-sandbox/kn-plugin-quickstart.git
cd kn-plugin-quickstart
./hack/build.sh
```

## Using the Nightlies

You can grab the latest nightly binary executable for:

- [macOS](https://storage.googleapis.com/knative-nightly/kn-plugin-quickstart/latest/kn-quickstart-darwin-amd64)
- [Linux](https://storage.googleapis.com/knative-nightly/kn-plugin-quickstart/latest/kn-quickstart-linux-amd64)
- [Windows](https://storage.googleapis.com/knative-nightly/kn-plugin-quickstart/latest/kn-quickstart-windows-amd64.exe)

Add the binary to the system PATH and ensure that it is executable.

