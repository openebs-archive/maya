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

package volumereplica

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
)

// VolumeReplicaOperator is the name of the tool that makes
// volume-related operations.
const (
	VolumeReplicaOperator    = "zfs"
	BinaryCapacityUnitSuffix = "i"
	VolumeTypeClone          = "clone"
	CreateCmd                = "create"
	CloneCmd                 = "clone"
)

// RunnerVar the runner variable for executing binaries.
var RunnerVar util.Runner

// CheckValidVolumeReplica checks for validity of cStor replica resource.
func CheckValidVolumeReplica(cVR *apis.CStorVolumeReplica) error {
	var err error
	if len(cVR.Labels["cstorvolume.openebs.io/name"]) == 0 {
		err = fmt.Errorf("Volume Name/UID cannot be empty")
		return err
	}
	if len(cVR.Spec.TargetIP) == 0 {
		err = fmt.Errorf("TargetIP cannot be empty")
		return err
	}
	if len(cVR.Spec.Capacity) == 0 {
		err = fmt.Errorf("Capacity cannot be empty")
		return err
	}
	if len(cVR.Labels["cstorpool.openebs.io/uid"]) == 0 {
		err = fmt.Errorf("Pool cannot be empty")
		return err
	}
	return nil
}

// CreateVolumeReplica creates cStor replica(zfs volumes).
func CreateVolumeReplica(cStorVolumeReplica *apis.CStorVolumeReplica, fullVolName string) error {
	cmd := []string{}
	isClone := false
	snapName := ""
	if len(cStorVolumeReplica.Spec.Type) == 0 || cStorVolumeReplica.Spec.Type == VolumeTypeClone {
		isClone = true
		snapName = cStorVolumeReplica.Spec.SnapName
		glog.Infof("Creating clone volume: %s of snapshot %s", string(fullVolName), string(snapName))
		cmd = builldVolumeCloneCommand(cStorVolumeReplica, snapName, fullVolName)
	} else {
		// Parse capacity unit on CVR to support backward compatibility
		volCapacity := parseCapacityUnit(cStorVolumeReplica.Spec.Capacity)
		cStorVolumeReplica.Spec.Capacity = volCapacity
		cmd = builldVolumeCreateCommand(cStorVolumeReplica, fullVolName)
	}

	stdoutStderr, err := RunnerVar.RunCombinedOutput(VolumeReplicaOperator, cmd...)
	if err != nil {
		if isClone {
			glog.Errorf("Unable to create clone volume: %s for snapshot %s. error : %v", fullVolName, snapName, string(stdoutStderr))
		} else {
			glog.Errorf("Unable to create volume %s. error : %v", fullVolName, string(stdoutStderr))
		}

		return err
	}
	return nil
}

// builldVolumeCreateCommand returns volume create command along with attributes as a string array
func builldVolumeCreateCommand(cStorVolumeReplica *apis.CStorVolumeReplica, fullVolName string) []string {
	var createVolCmd []string

	openebsVolname := "io.openebs:volname=" + cStorVolumeReplica.ObjectMeta.Name

	openebsTargetIP := "io.openebs:targetip=" + cStorVolumeReplica.Spec.TargetIP

	createVolCmd = append(createVolCmd, CreateCmd,
		"-b", "4K", "-s", "-o", "compression=on",
		"-o", openebsTargetIP, "-o", openebsVolname,
		"-V", cStorVolumeReplica.Spec.Capacity, fullVolName)

	return createVolCmd
}

// builldVolumeCloneCommand returns volume clone command along with attributes as a string array
func builldVolumeCloneCommand(cStorVolumeReplica *apis.CStorVolumeReplica, snapName, fullVolName string) []string {
	var cloneVolCmd []string

	openebsVolname := "io.openebs:volname=" + cStorVolumeReplica.ObjectMeta.Name

	openebsTargetIP := "io.openebs:targetip=" + cStorVolumeReplica.Spec.TargetIP

	cloneVolCmd = append(cloneVolCmd, CloneCmd,
		"-o", "compression=on", "-o", openebsTargetIP,
		"-o", openebsVolname, snapName, fullVolName)

	return cloneVolCmd
}

// GetVolumes returns the slice of volumes.
func GetVolumes() ([]string, error) {
	volStrCmd := []string{"get", "-Hp", "name", "-o", "name"}
	volnameByte, err := RunnerVar.RunStdoutPipe(VolumeReplicaOperator, volStrCmd...)
	if err != nil || string(volnameByte) == "" {
		glog.Errorf("Unable to get volumes:%v", string(volnameByte))
		return []string{}, err
	}
	noisyVolname := string(volnameByte)
	sepNoisyVolName := strings.Split(noisyVolname, "\n")
	var volNames []string
	for _, volName := range sepNoisyVolName {
		volName = strings.TrimSpace(volName)
		volNames = append(volNames, volName)
	}
	return volNames, nil
}

// DeleteVolume deletes the specified volume.
func DeleteVolume(fullVolName string) error {
	deleteVolStr := []string{"destroy", "-r", fullVolName}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(VolumeReplicaOperator, deleteVolStr...)
	if err != nil {
		// If volume is missing then do not return error
		if strings.Contains(err.Error(), "dataset does not exist") {
			glog.Infof("Assuming volume deletion successful for error: %v", string(stdoutStderr))
			return nil
		}
		glog.Errorf("Unable to delete volume : %v", string(stdoutStderr))
		return err
	}
	return nil
}

// parseCapacityUnit add support for backward compatibility with respect to capacity units
func parseCapacityUnit(capacity string) string {
	// ToDo Use parsing factor for Ki->K,Gi->G, etc conversion
	if strings.HasSuffix(capacity, BinaryCapacityUnitSuffix) {
		newCapacity := strings.TrimSuffix(capacity, BinaryCapacityUnitSuffix)
		return newCapacity
	}
	return capacity
}
