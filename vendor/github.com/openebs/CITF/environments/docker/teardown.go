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
	"fmt"
	"strings"
)

// Teardown stops all the docker containers present on the machine
func (docker Docker) Teardown() error {
	// CAUTION: This function call stops all docker containers
	containersStr, err := execCommand("docker ps -q")
	if err != nil {
		return fmt.Errorf("error while getting container id. Error: %+v", err)
	}
	if containersStr != "" {
		containers := strings.Fields(containersStr)
		for _, container := range containers {
			err = runCommand("docker stop -f " + container)
			logger.LogErrorf(err, "error occurred while stopping docker container: %s", container)
			logger.PrintNonErrorf(err, "Stopped container: %s", container)
		}
	}
	return nil
}
