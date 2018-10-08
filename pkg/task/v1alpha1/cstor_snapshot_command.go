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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// cstorSnapshotCommand represents a cstor snapshot runtask command
//
// NOTE:
//  This is an implementation of CommandRunner
type cstorSnapshotCommand struct {
	*RunCommand
}

// instance returns specific cstor snapshot runtask command implementation based
// on the command's action
func (c *cstorSnapshotCommand) instance() (r Runner) {
	switch c.Action {
	case CreateCommandAction:
		r = &cstorSnapshotCreate{c}
	case DeleteCommandAction:
		r = &cstorSnapshotDelete{c}
	default:
		r = &notSupportedActionCommand{c.RunCommand}
	}
	return
}

// Run executes various cstor volume related operations
func (c *cstorSnapshotCommand) Run() (r RunCommandResult) {
	return c.instance().Run()
}

// validateOptions checks if the required params are missing
func (c *cstorSnapshotCommand) validateOptions() error {
	ip, _ := c.Data["ip"].(string)
	volName, _ := c.Data["volname"].(string)
	snapName, _ := c.Data["snapname"].(string)
	if len(ip) == 0 {
		return errors.Errorf("missing ip address")
	}

	if len(volName) == 0 {
		return errors.Errorf("missing volume name")
	}

	if len(snapName) == 0 {
		return errors.Errorf("missing snapshot name")
	}
	return nil
}

// casSnapshot returns a filled object of CASSnapshot
func (c *cstorSnapshotCommand) casSnapshot() *apis.CASSnapshot {
	volName, _ := c.Data["volname"].(string)
	snapName, _ := c.Data["snapname"].(string)
	return &apis.CASSnapshot{
		Spec: apis.SnapshotSpec{
			VolumeName: volName,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: snapName,
		},
	}
}
