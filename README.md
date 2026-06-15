# kn-plugin-quickstart

`kn-plugin-quickstart` is a plugin of the Knative Client, to enable users to quickly set up a local Knative environment from the command line.

## Getting Started

On Mac first run these commands:

```sh
brew install kubectl
brew install minikube 
brew install knative/client/kn
```

You will also need bash >=4. On mac, install the latest bash like so:

```sh
brew install go
brew install bash
# this assumes you use zsh as your default shell. replace with the shell of your choice!
echo 'export PATH="$(brew --prefix)/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
```

## Building from Source

You must build from source to get some custom PolyAPI kn plugin quickstart goodness.

``` bash
gh repo clone polyapi/kn-plugin-quickstart
cd kn-plugin-quickstart
./hack/build.sh
chmod +x ./kn-quickstart
# replace /usr/local/bin with your preferred bin, anywhere on your path should work, this is mac default
sudo mv ./kn-quickstart /usr/local/bin/
```

> **Note:** You can also build the plugin with pinned versions.
``` bash
SERVING_VERSION=1.19.6 \
EVENTING_VERSION=1.19.3 \
KOURIER_VERSION=1.19.5 \
hack/build.sh --fast
```

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

### Quickstart with Minikube

> `minikube` IS RECOMMENDED. `kind` is for advanced users.

Set up a local Knative cluster using [Minikube](https://minikube.sigs.k8s.io/):

```bash
kn quickstart minikube

# OR with extra mount
kn-quickstart minikube --extraMountHostPath /home/myname/foo --extraMountContainerPath /foo
```

Note: for Windows/Mac users, after the above command completes, you will need to run the following in a separate terminal window:

``` bash
minikube tunnel --profile minikube-knative
```

### Quickstart with KinD

> `minikube` IS RECOMMENDED. `kind` is for advanced users.

Set up a local Knative cluster using [KinD](https://kind.sigs.k8s.io/):

``` bash
kn quickstart kind
```
Kind can be configured with a [local container image registry](https://kind.sigs.k8s.io/docs/user/local-registry/) by passing the `--registry` flag:

```bash
kn quickstart kind --registry
```

Note: we automatically configure tag resolution for the local registry when this flag is passed

Kind can also be configured with an [extra mount](https://kind.sigs.k8s.io/docs/user/configuration#extra-mounts) so your containers can access files on your local machine.

```bash
kn quickstart kind --extraMountHostPath /home/myname/foo --extraMountContainerPath /foo
```