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
	VolumeReplicaOperator = "zfs"
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

// CreateVolume creates cStor replica(zfs volumes).
func CreateVolume(cStorVolumeReplica *apis.CStorVolumeReplica, fullVolName string) error {
	createVolAttr := createVolumeBuilder(cStorVolumeReplica, fullVolName)
	stdoutStderr, err := RunnerVar.RunCombinedOutput(VolumeReplicaOperator, createVolAttr...)
	if err != nil {
		glog.Errorf("Unable to create volume: %v", string(stdoutStderr))
		return err
	}
	return nil
}

// createVolumeBuilder builds volume creations command to run.
func createVolumeBuilder(cStorVolumeReplica *apis.CStorVolumeReplica, fullVolName string) []string {
	var createVolAttr []string

	openebsVolname := "io.openebs:volname=" + cStorVolumeReplica.ObjectMeta.Name

	openebsTargetIP := "io.openebs:targetip=" + cStorVolumeReplica.Spec.TargetIP

	createVolAttr = append(createVolAttr, "create",
		"-b", "4K", "-s", "-o", "compression=on",
		"-V", cStorVolumeReplica.Spec.Capacity, fullVolName,
		"-o", openebsTargetIP, "-o", openebsVolname)

	return createVolAttr
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
	deleteVolStr := []string{"destroy", fullVolName}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(VolumeReplicaOperator, deleteVolStr...)
	if err != nil {
		glog.Errorf("Unable to delete volume : %v", string(stdoutStderr))
		return err
	}
	return nil
}
