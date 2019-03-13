/*
Copyright 2018 The OpenEBS Authors

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

package task

import (
	"errors"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/template"
	api_core_v1 "k8s.io/api/core/v1"
)

// TaskExec will consist of exec configuration
type TaskExec struct {
	// Command - command that we want to execute remotely in given container
	Command string `json:"command"`
	// ContainerName - in which container we want to exec
	ContainerName string `json:"containerName"`
	// Stdin - Stdin option of pod exec it can be true or false
	Stdin bool `json:"stdin"`
	// Stdout - Stdout option of pod exec it can be true or false
	Stdout bool `json:"stdout"`
	// Stderr - Stderr option of pod exec it can be true or false
	Stderr bool `json:"stderr"`
	// TTY - TTY option of pod exec it can be true or false
	// TTY bool `json:"tty"`
}

// asTaskExec runs go template against the yaml document & converts it to a TaskPatch type
func asTaskExec(context, yml string, values map[string]interface{}) (exec TaskExec, err error) {
	b, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		return
	}

	// unmarshall into TaskExec
	err = yaml.Unmarshal(b, &exec)
	return
}

type taskExecExecutor struct {
	exec TaskExec
}

// isValidExecTask checks container name and command is present or not
func isValidExecTask(exec TaskExec) error {
	if exec.Command == "" {
		return errors.New("command not present")
	}
	if exec.ContainerName == "" {
		return errors.New("container name not present")
	}
	return nil
}

func newTaskExecExecutor(exec TaskExec) (*taskExecExecutor, error) {
	if err := isValidExecTask(exec); err != nil {
		return nil, errors.New("Failed to create exec executor invalid exec config : " + err.Error())
	}

	return &taskExecExecutor{
		exec: exec,
	}, nil
}

// toPodExecOptions puts PodExecOptions from runtask and
// returns PodExecOptions object pointer
func (e *taskExecExecutor) toPodExecOptions() *api_core_v1.PodExecOptions {
	return &api_core_v1.PodExecOptions{
		Command:   strings.Fields(e.exec.Command),
		Container: e.exec.ContainerName,
		Stdin:     e.exec.Stdin,
		Stdout:    e.exec.Stdout,
		Stderr:    e.exec.Stdout,
		TTY:       false,
	}
}
