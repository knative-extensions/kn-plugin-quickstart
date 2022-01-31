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

package command

import (
	"fmt"

	"knative.dev/kn-plugin-quickstart/pkg/minikube"

	"github.com/spf13/cobra"
)

// NewMinikubeCommand implements 'kn quickstart minikube' command
func NewMinikubeCommand() *cobra.Command {

	var minikubeCmd = &cobra.Command{
		Use:   "minikube",
		Short: "Quickstart with Minikube",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Running Knative Quickstart using Minikube")
			return minikube.SetUp(name)
		},
	}
	// Set minikubeCmd options
	clusterNameOption(minikubeCmd, "knative")
	return minikubeCmd
}
