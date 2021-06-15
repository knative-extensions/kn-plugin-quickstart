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

package kind

import (
	"fmt"

	"knative.dev/kn-plugin-quickstart/pkg/install"
)

// SetUp creates a local Kind cluster and installs all the relevant Knative components
func SetUp() error {
	if err := createKindCluster(); err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	if err := install.Serving(); err != nil {
		return fmt.Errorf("apply: %w", err)
	}
	if err := install.Eventing(); err != nil {
		return fmt.Errorf("install eventing: %w", err)
	}
	return nil
}

func createKindCluster() error {
	fmt.Println("TODO: Creating Kind cluster...")
	return nil
}
