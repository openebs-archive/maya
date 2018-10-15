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

package citf

import (
	"fmt"

	citfoptions "github.com/openebs/CITF/citf_options"
	"github.com/openebs/CITF/common"
	"github.com/openebs/CITF/config"
	"github.com/openebs/CITF/environments"
	"github.com/openebs/CITF/environments/docker"
	"github.com/openebs/CITF/environments/minikube"
	"github.com/openebs/CITF/utils/k8s"
	"github.com/openebs/CITF/utils/log"
)

var logger log.Logger

// CITF is a struct which will be the driver for all functionalities of this framework
type CITF struct {
	Environment  environments.Environment
	K8S          k8s.K8S
	Docker       docker.Docker
	DebugEnabled bool
	Logger       log.Logger
}

// getEnvironment returns the environment according to the config
func getEnvironment() (environments.Environment, error) {
	switch config.Environment() {
	case common.Minikube:
		return minikube.NewMinikube(), nil
	default:
		return nil, fmt.Errorf("platform: %q is not suppported by CITF", config.Environment())
	}
}

// Reload reloads all the fields of citfInstance according to supplied `citfCreateOptions`
func (citfInstance *CITF) Reload(citfCreateOptions *citfoptions.CreateOptions) error {
	// Here, we don't want to return fatal error since we want to continue
	// executing the function with default configuration even if it fails
	// so we simply log any error and continue
	logger.LogError(config.LoadConf(citfCreateOptions.ConfigPath), "error loading config file")

	if citfCreateOptions.EnvironmentInclude {
		environ, err := getEnvironment()
		if err != nil {
			return err
		}
		citfInstance.Environment = environ
	}

	if citfCreateOptions.K8SInclude {
		k8sInstance, err := k8s.NewK8S()
		if err != nil {
			return err
		}
		citfInstance.K8S = k8sInstance
	}

	if citfCreateOptions.DockerInclude {
		citfInstance.Docker = docker.NewDocker()
	}

	if citfCreateOptions.LoggerInclude {
		citfInstance.Logger = log.Logger{}
	}

	citfInstance.DebugEnabled = config.Debug()
	return nil
}

// NewCITF returns CITF struct filled according to supplied `citfCreateOptions`.
// One need this in order to use any functionality of this framework.
func NewCITF(citfCreateOptions *citfoptions.CreateOptions) (citfInstance CITF, err error) {
	err = citfInstance.Reload(citfCreateOptions)
	return
}
