# kn-plugin-quickstart

`kn-plugin-quickstart` is a plugin of the Knative Client, to enable users to quickly set up a local Knative environment from the command line.

## Description

tbd

## Build and Install

You must [set up your development environment](https://github.com/knative/client/blob/master/docs/DEVELOPMENT.md#prerequisites) before you build `kn-plugin-quickstart`.

### Building

Once you've set up your development environment, you can build the plugin by running the following commands:

``` bash
git clone git@github.com:knative-sandbox/kn-plugin-quickstart.git
cd kn-plugin-quickstart
./hack/build.sh
```

You'll get an executable plugin binary named `kn-plugin-quickstart` in your current directory. To use as a stand-alone binary, check the available commands by running `./kn-quickstart -h`.

### Installing

If you'd like to use the plugin with the `kn` CLI, install the plugin by copying the executable file under the `kn` plugins directory by running the following:

``` bash
mkdir -p ~/.config/kn/plugins
cp kn-quickstart ~/.config/kn/plugins
```

Check if the plugin is loaded by running `kn -h` (`quickstart` should appear in the list of plugin commands in the output).

To run the plugin, use `kn quickstart`, for example:

`kn quickstart -h`


## Examples

### Quickstart with KinD

Set up a local Knative cluster using [KinD](https://kind.sigs.k8s.io/):

``` bash
kn quickstart kind
```
