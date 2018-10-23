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
	"fmt"
	"os"

	"github.com/openebs/CITF/common"
)

// runPostStartCommandsForMinikube runs the commands required when run minikube as --vm-driver=none
// Assumption: Environment variables `USER` and `HOME` is well defined.
func (minikube Minikube) runPostStartCommandsForMinikubeNoneDriver() {
	userName := os.Getenv("USER")
	homeDir := os.Getenv("HOME")
	commands := []string{
		"mv /root/.kube " + homeDir + "/.kube",
		"chown -R " + userName + " " + homeDir + "/.kube",
		"chgrp -R " + userName + " " + homeDir + "/.kube",
		"mv /root/.minikube " + homeDir + "/.minikube",
		"chown -R " + userName + " " + homeDir + "/.minikube",
		"chgrp -R " + userName + " " + homeDir + "/.minikube",
	}

	for _, command := range commands {
		fmt.Printf("Running %q\n", command)
		output, err := execCommand(command)
		logger.PrintErrorf(err, "running %q failed", command)
		logger.PrintNonErrorf(err, "run %q successfully. Output: %s", command, output)
	}
}

// StartMinikube method starts minikube with `--vm-driver=none` option.
func (minikube Minikube) StartMinikube() error {
	err := runCommand(common.Minikube + " start --vm-driver=none")
	// We can also use following:
	// "minikube start --vm-driver=none --feature-gates=MountPropagation=true --cpus=1 --memory=1024 --v=3 --alsologtostderr"
	if err != nil {
		return fmt.Errorf("error occurred while starting minikube. Error: %+v", err)
	}

	envChangeMinikubeNoneUser := os.Getenv("CHANGE_MINIKUBE_NONE_USER")
	logger.PrintfDebugMessage("Environ CHANGE_MINIKUBE_NONE_USER = %q", envChangeMinikubeNoneUser)

	if envChangeMinikubeNoneUser == "true" {
		// Below commands shall automatically run in this case.
		logger.PrintlnDebugMessage("Returning from setup.")
		return nil
	}

	minikube.waitForDotKubeDirToBeCreated()

	minikube.waitForDotMinikubeDirToBeCreated()

	minikube.runPostStartCommandsForMinikubeNoneDriver()

	return nil
}

// Setup checks if a teardown is required before minikube start
// if so it does that and then start the minikube.
// It does nothing when minikube is already running.
// it prints status too.
func (minikube Minikube) Setup() error {
	minikubeStatus, err := minikube.Status()

	logger.PrintfDebugMessageIfError(err, "error occurred while checking minikube status")
	logger.PrintfDebugMessageIfNotError(err, common.Minikube+" status: %q", minikubeStatus)

	teardownRequired := false
	startRequired := false

	// I won't use common.Minikube here because I am not really using name here,
	// this is just another string which appears in the output of minikube status command
	status, ok := minikubeStatus["minikube"]
	if !ok {
		fmt.Println("\"minikube\" not present in status. May be minikube is not accessible. Aborting...")
		os.Exit(1)
	}
	if status == "" { // This means cluster itself is not there
		fmt.Println("cluster is not up. will start the machine")
		startRequired = true // So, Start the minikube
	} else if status == "Stopped" { // Cluster is there but it is stopped
		fmt.Println("minikube cluster is present but not \"Running\", so will tearing down the machine then start again.")
		teardownRequired = true // We need to teardown it first
		startRequired = true    // Then also we need to start the machine
	} else if status != "Running" { // If cluster is there and machine is not in "Stopped" or "Running" state
		// Then there is a problem
		fmt.Printf("minikube is in unknown state. State: %q. Aborting...", status)
		os.Exit(1)
	} else { // Else minikube is Running so we need not do anything.
		fmt.Println("minikube is already Running.")
	}

	// If we figured out that a teardown is needed then do so
	if teardownRequired {
		err = minikube.Teardown()
		if err != nil {
			return fmt.Errorf("error occurred while deleting existing machine. Error: %+v", err)
		}
		fmt.Println("minikube deleted.")
	}

	// If we figured out that a start is needed then do so
	if startRequired {
		minikube.StartMinikube()
	}
	return nil
}
