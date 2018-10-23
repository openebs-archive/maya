/*
Copyright 2018 The OpenEBS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package docker

import (
	"os"
	"strings"

	"github.com/openebs/CITF/common"
	"github.com/openebs/CITF/utils/log"
	sysutil "github.com/openebs/CITF/utils/system"
)

var logger log.Logger

var (
	useSudo     = true // Default value to use sudo
	execCommand = sysutil.ExecCommandWithSudo
	runCommand  = sysutil.RunCommandWithSudo
)

func init() {
	// `sudo` use detection
	useSudoEnv := strings.ToLower(strings.TrimSpace(os.Getenv("USE_SUDO")))
	if useSudoEnv == "true" { // If it is mentioned in the environment variable to use sudo
		useSudo = true // use sudo then
	} else if useSudoEnv == "false" { // Else if it is mentioned in the environment variable not to use sudo
		useSudo = false // do not use sudo
	} // Else use default value mentioned above

	if !useSudo {
		execCommand = sysutil.ExecCommand
		runCommand = sysutil.RunCommand
	}
}

// Docker is a struct which will be the driver for all the methods related to docker
// Docker implements github.com/openebs/CITF/Environment interface
type Docker struct{}

// NewDocker returns Docker struct
func NewDocker() Docker {
	return Docker{}
}

// Name returns the name of the environment, In this case common.Docker
func (docker Docker) Name() string {
	return common.Docker
}
