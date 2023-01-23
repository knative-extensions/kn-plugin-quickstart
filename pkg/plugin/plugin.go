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

package plugin

import (
	"os"

	"knative.dev/kn-plugin-quickstart/internal/root"

	knplugin "knative.dev/client-pkg/pkg/kn/plugin"
)

func init() {
	knplugin.InternalPlugins = append(knplugin.InternalPlugins, &plugin{})
}

type plugin struct{}

// Name returns the plugin's name
func (pl *plugin) Name() string {
	return "kn-quickstart"
}

// Execute represents the plugin's entrypoint when called through kn
func (pl *plugin) Execute(args []string) error {
	cmd := root.NewRootCommand()
	oldArgs := os.Args
	defer (func() {
		os.Args = oldArgs
	})()
	os.Args = append([]string{"kn-quickstart"}, args...)
	return cmd.Execute()
}

// Description is displayed in kn's help message
func (pl *plugin) Description() (string, error) {
	return "Get started with Knative", nil
}

// CommandParts defines for plugin is executed from kn
func (pl *plugin) CommandParts() []string {
	return []string{"quickstart"}
}

// Path is empty because its an internal plugins
func (pl *plugin) Path() string {
	return ""
}
