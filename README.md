# kn-plugin-quickstart

`kn-plugin-quickstart` is a plugin of the Knative Client, to enable users to quickly set up a local Knative environment from the command line.

## Getting Started

### Installation

You can download the latest binaries from the [Releases](https://github.com/knative-sandbox/kn-plugin-quickstart/releases) page.

There are two ways to run `kn quickstart`:

1. You can run it standalone, just put it on your system path and make sure it is executable.
2. You can install it as a plugin of the `kn` client to run:
    * Follow the [documentation] to install `kn client` if you don't have it
    * Copy the `kn-quickstart` binary to the `~/.config/kn/plugins/` directory and make sure its filename is `kn-quickstart`
    * Run `kn plugin list` to verify that the `kn-quickstart` plugin is installed successfully
    
After the plugin is installed, you can use `kn quickstart` to run its related subcommands.

## Usage

```
Get up and running with a local Knative environment running on KinD.

Usage:
  kn-quickstart [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  kind        Quickstart with Kind
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

## Building from Source

You must [set up your development environment](https://github.com/knative/client/blob/master/docs/DEVELOPMENT.md#prerequisites) before you build `kn-plugin-quickstart`.

Once you've set up your development environment, you can build the plugin by running the following commands:

``` bash
git clone git@github.com:knative-sandbox/kn-plugin-quickstart.git
cd kn-plugin-quickstart
./hack/build.sh
```

