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
	"os/exec"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/cstorops/pool"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// VolumeReplicaOperator is the name of the tool that makes
// volume-related operations.
const (
	VolumeReplicaOperator = "zfs"
)

// CheckValidVolumeReplica checks for validity of cStor replica resource.
func CheckValidVolumeReplica(cStorVolumeReplicaUpdated *apis.CStorVolumeReplica) error {
	if cStorVolumeReplicaUpdated.Spec.VolName == "" {
		return fmt.Errorf("Volume name cannot be empty")
	}
	if cStorVolumeReplicaUpdated.Spec.Capacity == "" {
		return fmt.Errorf("Capacity cannot be empty")
	}
	return nil
}

// CreateVolume creates cStor replica(zfs volumes).
func CreateVolume(cStorVolumeReplicaUpdated *apis.CStorVolumeReplica, fullvolname string) error {

	volCmd := createVolumeBuilder(cStorVolumeReplicaUpdated, fullvolname)
	stdoutStderr, err := volCmd.CombinedOutput()
	if err != nil {
		glog.Errorf("stdoutStderr: %v-%v", string(stdoutStderr), err)
		return err
	}
	glog.Infof("Volume creation successful : %v", fullvolname)
	return nil

}

// createVolumeBuilder builds volume creations command to run.
func createVolumeBuilder(cStorVolumeReplicaUpdated *apis.CStorVolumeReplica, fullvolname string) *exec.Cmd {
	var createVolAttr []string
	createVolAttr = append(createVolAttr, "create", "-s",
		"-V", cStorVolumeReplicaUpdated.Spec.Capacity, fullvolname)
	volCmd := exec.Command(VolumeReplicaOperator, createVolAttr...)
	glog.V(4).Infof("volCmd : ", volCmd)
	return volCmd
}

// GetVolumes returns the slice of volumes
func GetVolumes() []string {
	poolname, err := pool.GetPoolName()
	volStrCmd := VolumeReplicaOperator + " get volsize | grep " + poolname
	volcmd := exec.Command("bash", "-c", volStrCmd)
	stderr, err := volcmd.CombinedOutput()
	if err != nil {
		fmt.Errorf("Unable to get vol info :%v ", err)
	}

	noisyVolname := string(stderr)

	Volnames := strings.Split(noisyVolname, "\n")
	var Volnam []string
	var output []string
	for _, volname := range Volnames {
		Volnam = strings.Split(volname, "volsize")
		vol := strings.TrimSpace(Volnam[0])
		output = append(output, vol)
	}
	return output
}

// DeleteVolume deletes the specified volume
func DeleteVolume(fullVolName string) error {
	deleteVolStr := VolumeReplicaOperator + " destroy -f " + fullVolName
	deleteVolCmd := exec.Command("bash", "-c", deleteVolStr)
	_, err := deleteVolCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Unable to delete volume :%v ", err)
	}
	return nil
}
