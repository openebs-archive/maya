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

package v1alpha1

// cstorSnapshotCommand represents a cstor snapshot runtask command
//
// NOTE:
//  This is an implementation of CommandRunner
type cstorSnapshotCommand struct {
	cmd *RunCommand
}

// instance returns specific cstor snapshot runtask command implementation based
// on the command's action
func (c *cstorSnapshotCommand) instance() (r CommandRunner) {
	switch c.cmd.Action {
	case CreateCommandAction:
		r = &cstorSnapshotCreate{cmd: c.cmd}
	default:
		r = &notSupportedActionCommand{cmd: c.cmd}
	}
	return
}

// Run executes various cstor volume related operations
func (c *cstorSnapshotCommand) Run() (r RunCommandResult) {
	return c.instance().Run()
}
