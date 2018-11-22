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

	"encoding/json"
	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
)

const (
	// VolumeReplicaOperator is the name of the tool that makes volume-related operations.
	VolumeReplicaOperator = "zfs"
	// BinaryCapacityUnitSuffix is the suffix for binary capacity unit.
	BinaryCapacityUnitSuffix = "i"
	// CreateCmd is the create command for zfs volume.
	CreateCmd = "create"
	// CloneCmd is the zfs volume clone command.
	CloneCmd = "clone"
	// StatsCmd is the zfs volume stats command.
	StatsCmd = "stats"
	// ZfsStatusDegraded is the degraded state of zfs volume.
	ZfsStatusDegraded = "Degraded"
	// ZfsStatusOffline is the offline state of zfs volume.
	ZfsStatusOffline = "Offline"
	// ZfsStatusHealthy is the healthy state of zfs volume.
	ZfsStatusHealthy = "Healthy"
	// ZpoolStatusRebuilding is the rebuilding state of zfs volume.
	ZfsStatusRebuilding = "Rebuilding"
)
const (
	// CStorPoolUIDKey is the key for csp object uid which is present in cvr labels.
	CStorPoolUIDKey = "cstorpool.openebs.io/uid"
	// PvNameKey is the key for pv object uid which is present in cvr labels.
	PvNameKey = "cstorvolume.openebs.io/name"
	// PoolPrefix is the prefix of zpool name.
	PoolPrefix = "cstor-"
)

// CvrStats struct is zfs volume status output JSON contract.
type CvrStats struct {
	// Stats is an array which holds zfs volume related stats
	Stats []Stats `json:"stats"`
}

// Stats contain the zfs volume related stats.
type Stats struct {
	// Name of the zfs volume.
	Name string `json:"name"`
	// Status of the zfs volume.
	Status string `json:"status"`
	// RebuildStatus of the zfs volume.
	RebuildStatus             string `json:"rebuildStatus"`
	IsIOAckSenderCreated      int    `json:"isIOAckSenderCreated"`
	isIOReceiverCreated       int    `json:"isIOReceiverCreated"`
	RunningIONum              int    `json:"runningIONum"`
	CheckpointedIONum         int    `json:"checkpointedIONum"`
	DegradedCheckpointedIONum int    `json:"degradedCheckpointedIONum"`
	CheckpointedTime          int    `json:"checkpointedTime"`
	RebuildBytes              int    `json:"rebuildBytes"`
	RebuildCnt                int    `json:"rebuildCnt"`
	RebuildDoneCnt            int    `json:"rebuildDoneCnt"`
	RebuildFailedCnt          int    `json:"rebuildFailedCnt"`
}

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
	isClone := cStorVolumeReplica.Labels[string(apis.CloneEnableKEY)] == "true"
	snapName := ""
	if isClone {
		srcVolume := cStorVolumeReplica.Annotations[string(apis.SourceVolumeKey)]
		snapName = cStorVolumeReplica.Annotations[string(apis.SnapshotNameKey)]
		// Get the dataset name from volume name
		dataset := strings.Split(fullVolName, "/")[0]
		glog.Infof("Creating clone volume: %s from snapshot %s", fullVolName, srcVolume+"@"+snapName)
		// zfs snapshots are named as dataset/volname@snapname
		cmd = builldVolumeCloneCommand(cStorVolumeReplica, dataset+"/"+srcVolume+"@"+snapName, fullVolName)
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
	deleteVolStr := []string{"destroy", "-R", fullVolName}
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

// Status function gives the status of cvr which extracted and mapped to a set of cvr statuses
// after getting the zfs volume status
func Status(volumeName string) (string, error) {
	statusPoolStr := []string{StatsCmd, volumeName}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(VolumeReplicaOperator, statusPoolStr...)
	if err != nil {
		glog.Errorf("Unable to get volume stats: %v", string(stdoutStderr))
		return "", fmt.Errorf("Unable to get volume stats: %s", err.Error())
	}
	volumeStats := &CvrStats{}
	err = json.Unmarshal(stdoutStderr, volumeStats)
	if err != nil {
		return "", fmt.Errorf("Unable to unmarshal volume stats:%s", err)
	}
	volumeStatus := volumeStats.Stats[0].Status
	if strings.TrimSpace(volumeStatus) == "" {
		glog.Warning("Empty status of volume on volume stats")
	}
	cvrStatus := ZfsToCvrStatusMapper(volumeStatus)
	return cvrStatus, nil
}

// GetVolumeName finds the zctual zfs volume name for the given cvr.
func GetVolumeName(cVR *apis.CStorVolumeReplica) (string, error) {
	var volumeName string
	// Get the corresponding CSP UID for this CVR
	if cVR.Labels == nil {
		return "", fmt.Errorf("no labels found on cvr object")
	}
	cspUID := cVR.Labels[CStorPoolUIDKey]
	if strings.TrimSpace(cspUID) == "" {
		return "", fmt.Errorf("csp uid not found on cvr label")
	}
	pvName := cVR.Labels[PvNameKey]
	if strings.TrimSpace(pvName) == "" {
		return "", fmt.Errorf("pv name not found on cvr label")
	}
	volumeName = PoolPrefix + cspUID + "/" + pvName
	return volumeName, nil
}

// ZfsToCvrStatusMapper maps zfs status to defined cvr status.
func ZfsToCvrStatusMapper(zfsstatus string) string {
	if zfsstatus == ZfsStatusHealthy {
		return string(apis.CVRStatusOnline)
	}
	if zfsstatus == ZfsStatusOffline {
		return string(apis.CVRStatusOffline)
	}
	if zfsstatus == ZfsStatusDegraded {
		return string(apis.CVRStatusDegraded)
	}
	if zfsstatus == ZfsStatusRebuilding {
		return string(apis.CVRStatusRebuilding)
	}
	return string(apis.CVRStatusError)
}
