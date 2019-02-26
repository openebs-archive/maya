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
	"github.com/golang/glog"
	cstor "github.com/openebs/maya/pkg/volume/cstor/v1alpha1"
	"github.com/pkg/errors"
)

// cstorVolumeResize represents a cstor volume resize runtask command
type cstorVolumeResize struct {
	*cstorVolumeCommand
}

// Run creates cstor volume resize contents
func (c *cstorVolumeResize) Run() (r RunCommandResult) {
	glog.Infof("After cas template execution: %v", c)
	err := c.validateOptions()
	if err != nil {
		return c.AddError(errors.Errorf("failed to resize the cstor volume: %s", err)).Result(nil)
	}
	ip, _ := c.Data["ip"].(string)
	volOps := cstor.Cstor()
	volOps.IP = ip
	volOps.Vol = c.casVolumeResize()

	response, err := volOps.Resize()
	if err != nil {
		return c.AddError(err).Result(nil)
	}
	return c.Result(response)
}
