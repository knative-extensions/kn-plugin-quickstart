# Copyright 2020 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Global variables that you should adapt according to the plugin that
# you are creating

# Name of the plugin
PLUGIN="kn-quickstart"

# Directories containing go code which needs to be formatted
SOURCE_DIRS="cmd pkg internal"

# Directory which should be compiled
MAIN_SOURCE_DIR="cmd"

# Package which holds the version variables
VERSION_PACKAGE="knative.dev/kn-plugin-quickstart/internal/command"
