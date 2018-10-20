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

package minikube

import (
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
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
	// Check if `minikube` is present
	minikubePath, err := sysutil.BinPathFromPathEnv(common.Minikube)
	if minikubePath == "" {
		logger.LogFatalf(err, "%q not found in current directory or in directories represented by PATH environment variable", common.Minikube)
	}
	glog.Infof("%q found on path: %q", common.Minikube, minikubePath)

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

// Minikube is a struct which will be the driver for all the methods related to minikube
// Minikube implements github.com/openebs/CITF/Environment interface
type Minikube struct {
	// Timeout is the timeout that will be used throughout the minikube package
	// for timeout in any operation if requires.
	Timeout time.Duration

	// WaitTimeUnit is the time duration, which will be used throughout package
	// if it needs to wait for some sub-task. (It is small timeout)
	WaitTimeUnit time.Duration
}

// NewMinikube returns a Minikube struct
func NewMinikube() Minikube {
	return Minikube{
		Timeout:      time.Minute,
		WaitTimeUnit: time.Second,
	}
}

// Name returns the name of the environment, In this case common.Minikube
func (minikube Minikube) Name() string {
	return common.Minikube
}
