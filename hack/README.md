## Hacky scripts

This directory contains all the build and CI scripts that are used in building
the plugin. They are centered around
[knative/hack](https://github.com/knative/hack) which is included in the
`vendor` directory by using a placeholder go dependency in `tools.go` (and in
`go.mod`).

The only exception is the main `build.sh` which is an adapted clone of the
Knative client
[build.sh](https://github.com/knative/client/blob/master/hack/build.sh).

The following scripts are provided:

- `global_vars.sh` is the central place that you need to adapt for your specific
  plugin. It also should be the _only_ place that need to be adapted.
- `build.sh` to be used for the regular development workflow. Use
  `hack/build.sh --help` for an overview of the available options.

- `release.sh` is used by the Knative CI to create a release during an
  auto-release step which is triggered when you create a release branch or push
  a new commit to an existing release branch. For more details about the release
  process see ...

- `update`
