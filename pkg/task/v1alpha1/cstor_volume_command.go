/*
Copyright 2019 The OpenEBS Authors

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

// cstorVolumeCommand represents a cstor volume runtask command
//
// NOTE: This is an implementation of CommandRunner
type cstorVolumeCommand struct {
	*RunCommand
}

// instance returns specific cstor volume runtask command implementation based
// on the command's action
func (c *cstorVolumeCommand) instance() (r Runner) {
	switch c.Action {
	case ResizeCommandAction:
		r = &cstorVolumeResize{c}
	default:
		r = &notSupportedActionCommand{c.RunCommand}
	}
	return
}

func (c *cstorVolumeCommand) Run() (r RunCommandResult) {
	return c.instance().Run()
}

// validateOptions checks if the required params are missing
func (c *cstorVolumeCommand) validateOptions() error {
	ip, _ := c.Data["ip"].(string)
	volName, _ := c.Data["volname"].(string)
	capacity, _ := c.Data["capacity"].(string)
	if len(ip) == 0 {
		return errors.Errorf("missing ip address")
	}

	if len(volName) == 0 {
		return errors.Errorf("missing volume name")
	}

	if len(capacity) == 0 {
		return errors.Errorf("missing volume capacity")
	}
	return nil
}

// asCASVolume returns a filled object of CASVolume
func (c *cstorVolumeCommand) asCASVolume() *apis.CASVolume {
	volName, _ := c.Data["volname"].(string)
	capacity, _ := c.Data["capacity"].(string)
	return &apis.CASVolume{
		Spec: apis.CASVolumeSpec{
			Capacity: capacity,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: volName,
		},
	}
}
