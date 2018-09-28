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

import (
	"fmt"

	"github.com/openebs/maya/cmd/cstor-volume-grpc/app/command"
)

// cstorSnapshotCreate represents a cstor snapshot create runtask command
//
// NOTE:
//  This is an implementation of CommandRunner
type cstorSnapshotCreate struct {
	cmd *RunCommand
}

// Run creates cstor snapshot contents
func (c *cstorSnapshotCreate) Run() (r RunCommandResult) {
	// api call to list snapshots and snapshot actions per controller
	ip, _ := c.cmd.Data["ip"].(string)
	volName, _ := c.cmd.Data["volname"].(string)
	snapName, _ := c.cmd.Data["snapname"].(string)
	if len(ip) == 0 {
		return c.cmd.AddError(fmt.Errorf("missing ip address: failed to create cstor snapshot")).Result(nil)
	}

	if len(volName) == 0 {
		return c.cmd.AddError(fmt.Errorf("missing volume name: failed to create cstor snapshot")).Result(nil)
	}

	if len(snapName) == 0 {
		return c.cmd.AddError(fmt.Errorf("missing snapshot name: failed to create cstor snapshot")).Result(nil)
	}

	response, err := command.CreateSnapshot(volName, snapName, ip)
	if err != nil {
		return c.cmd.AddError(err).Result(nil)
	}
	return c.cmd.Result(response)
}
