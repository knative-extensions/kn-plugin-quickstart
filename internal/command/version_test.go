// Copyright © 2020 The Knative Authors
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
	"bytes"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

var versionOutputTemplate = `Version:      %s
Build Date:   %s
Git Revision: %s
`

const (
	fakeVersion     = "fake-version"
	fakeBuildDate   = "fake-build-date"
	fakeGitRevision = "fake-git-revision"
)

func TestVersionSetup(t *testing.T) {
	versionCmd := NewVersionCommand()
	assert.Equal(t, versionCmd.Use, "version")
	assert.Equal(t, versionCmd.Short, "Prints the plugin version")
	assert.Assert(t, versionCmd.RunE != nil)
}

func TestVersionOutput(t *testing.T) {
	Version = fakeVersion
	BuildDate = fakeBuildDate
	GitRevision = fakeGitRevision
	expectedOutput := fmt.Sprintf(versionOutputTemplate, fakeVersion, fakeBuildDate, fakeGitRevision)

	out, err := runVersionCmd()
	assert.NilError(t, err)
	assert.Equal(t, out, expectedOutput)
}

func runVersionCmd() (string, error) {
	versionCmd := NewVersionCommand()

	output := new(bytes.Buffer)
	versionCmd.SetOut(output)
	err := versionCmd.Execute()
	return output.String(), err
}
