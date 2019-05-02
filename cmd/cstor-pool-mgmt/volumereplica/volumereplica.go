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
	"time"

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
	// BackupCmd is the zfs send command
	BackupCmd = "send"
	// RestoreCmd is the zfs volume send command.
	RestoreCmd = "recv"
	// StatsCmd is the zfs volume stats command.
	StatsCmd = "stats"
	// ZfsStatusDegraded is the degraded state of zfs volume.
	ZfsStatusDegraded = "Degraded"
	// ZfsStatusOffline is the offline state of zfs volume.
	ZfsStatusOffline = "Offline"
	// ZfsStatusHealthy is the healthy state of zfs volume.
	ZfsStatusHealthy = "Healthy"
	// ZfsStatusRebuilding is the rebuilding state of zfs volume.
	ZfsStatusRebuilding = "Rebuilding"
	// MaxBackupRetryCount is a max number of retry should be performed during backup transfer
	MaxBackupRetryCount = 10
	// BackupRetryDelay is time(in seconds) to wait before the next attempt for backup transfer
	BackupRetryDelay = 5
	// MaxRestoreRetryCount is a max number of retry should be performed during restore transfer
	MaxRestoreRetryCount = 10
	// RestoreRetryDelay is time(in seconds) to wait before the next attempt for restore transfer
	RestoreRetryDelay = 5
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
func CreateVolumeReplica(cStorVolumeReplica *apis.CStorVolumeReplica, fullVolName string, quorum bool) error {
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
		cmd = builldVolumeCreateCommand(cStorVolumeReplica, fullVolName, quorum)
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
func builldVolumeCreateCommand(cStorVolumeReplica *apis.CStorVolumeReplica, fullVolName string, quorum bool) []string {
	var createVolCmd []string

	openebsVolname := "io.openebs:volname=" + cStorVolumeReplica.ObjectMeta.Name
	openebsTargetIP := "io.openebs:targetip=" + cStorVolumeReplica.Spec.TargetIP
	// ZvolWorkers represents number of threads that executes client IOs
	openebsZvolWorkers := "io.openebs:zvol_workers=" + cStorVolumeReplica.Spec.ZvolWorkers

	quorumValue := "quorum=on"
	if !quorum {
		quorumValue = "quorum=off"
	}

	// set volume property
	createVolCmd = append(createVolCmd, CreateCmd,
		"-b", "4K", "-s", "-o", "compression=on", "-o", quorumValue, "-o", openebsVolname)

	if len(cStorVolumeReplica.Spec.ZvolWorkers) != 0 {
		createVolCmd = append(createVolCmd, "-o", openebsZvolWorkers)
	}

	if cStorVolumeReplica.Annotations["isRestoreVol"] != "true" {
		createVolCmd = append(createVolCmd, "-o", openebsTargetIP)
	}

	// append volume size and volume name
	return append(createVolCmd, "-V", cStorVolumeReplica.Spec.Capacity, fullVolName)
}

// builldVolumeCloneCommand returns volume clone command along with attributes as a string array
func builldVolumeCloneCommand(cStorVolumeReplica *apis.CStorVolumeReplica, snapName, fullVolName string) []string {
	var cloneVolCmd []string

	openebsVolname := "io.openebs:volname=" + cStorVolumeReplica.ObjectMeta.Name
	openebsTargetIP := "io.openebs:targetip=" + cStorVolumeReplica.Spec.TargetIP
	// ZvolWorkers represents number of threads that executes client IOs
	openebsZvolWorkers := "io.openebs:zvol_workers=" + cStorVolumeReplica.Spec.ZvolWorkers

	if len(cStorVolumeReplica.Spec.ZvolWorkers) != 0 {
		cloneVolCmd = append(cloneVolCmd, CloneCmd,
			"-o", "compression=on", "-o", openebsTargetIP, "-o", "quorum=on",
			"-o", openebsZvolWorkers, "-o", openebsVolname, snapName, fullVolName)
	} else {
		cloneVolCmd = append(cloneVolCmd, CloneCmd,
			"-o", "compression=on", "-o", openebsTargetIP, "-o", "quorum=on",
			"-o", openebsVolname, snapName, fullVolName)
	}
	return cloneVolCmd
}

// CreateVolumeBackup sends cStor snapshots to remote location specified by backupcstor.
func CreateVolumeBackup(bkp *apis.BackupCStor) error {
	var cmd []string
	var retryCount int
	var err error

	// Parse capacity unit on CVR to support backward compatibility
	cmd = builldVolumeBackupCommand(bkp.ObjectMeta.Labels["cstorpool.openebs.io/uid"], bkp.Spec.VolumeName, bkp.Spec.PrevSnapName, bkp.Spec.SnapName, bkp.Spec.BackupDest)

	glog.Infof("Backup Command for volume: %v created, Cmd: %v\n", bkp.Spec.VolumeName, cmd)

	for retryCount < MaxBackupRetryCount {
		stdoutStderr, err := RunnerVar.RunCombinedOutput("/usr/local/bin/execute.sh", cmd...)
		if err != nil {
			glog.Errorf("Unable to start backup %s. error : %v retry:%v :%s", bkp.Spec.VolumeName, string(stdoutStderr), retryCount, err.Error())
			retryCount++
			time.Sleep(BackupRetryDelay * time.Second)
			continue
		}
		break
	}
	return err
}

// builldVolumeBackupCommand returns volume create command along with attributes as a string array
func builldVolumeBackupCommand(poolName, fullVolName, oldSnapName, newSnapName, backupDest string) []string {
	var startBackupCmd []string

	bkpAddr := strings.Split(backupDest, ":")
	if oldSnapName == "" {
		startBackupCmd = append(startBackupCmd, VolumeReplicaOperator, BackupCmd, "cstor-"+poolName+"/"+fullVolName+"@"+newSnapName, "| nc -w 3 "+bkpAddr[0]+" "+bkpAddr[1])
	} else {
		startBackupCmd = append(startBackupCmd, VolumeReplicaOperator, BackupCmd,
			"-i", "cstor-"+poolName+"/"+fullVolName+"@"+oldSnapName, "cstor-"+poolName+"/"+fullVolName+"@"+newSnapName, "| nc -w 3 "+bkpAddr[0]+" "+bkpAddr[1])
	}
	return startBackupCmd
}

// CreateVolumeRestore receive cStor snapshots from remote location(zfs volumes).
func CreateVolumeRestore(rst *apis.CStorRestore) error {
	var cmd []string
	var retryCount int
	var err error

	cmd = builldVolumeRestoreCommand(rst.ObjectMeta.Labels["cstorpool.openebs.io/uid"], rst.Spec.VolumeName, rst.Spec.RestoreSrc)

	glog.Infof("Restore Command for volume: %v created, Cmd: %v\n", rst.Spec.VolumeName, cmd)

	for retryCount < MaxRestoreRetryCount {
		stdoutStderr, err := RunnerVar.RunCombinedOutput("/usr/local/bin/execute.sh", cmd...)
		if err != nil {
			glog.Errorf("Unable to start restore %s. error : %v.. trying again", rst.Spec.VolumeName, string(stdoutStderr))
			time.Sleep(RestoreRetryDelay * time.Second)
			retryCount++
			continue
		}
		break
	}
	return err
}

// builldVolumeRestoreCommand returns restore command along with attributes as a string array
func builldVolumeRestoreCommand(poolName, fullVolName, restoreSrc string) []string {
	var restorecmd []string

	restorAddr := strings.Split(restoreSrc, ":")
	restorecmd = append(restorecmd, "nc -w 3 "+restorAddr[0]+" "+restorAddr[1]+" | ", VolumeReplicaOperator, RestoreCmd,
		" -F cstor-"+poolName+"/"+fullVolName)

	return restorecmd
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

// Capacity finds the capacity of the volume.
// The ouptut of command executed is as follows:
/*
root@cstor-sparse-pool-6dft-5b5c78ccc7-dls8s:/# zfs get used,logicalused cstor-d82bd105-f3a8-11e8-87fd-42010a800087/pvc-1b2a7d4b-f3a9-11e8-87fd-42010a800087
NAME                                                                                 PROPERTY     VALUE  SOURCE
cstor-d82bd105-f3a8-11e8-87fd-42010a800087/pvc-1b2a7d4b-f3a9-11e8-87fd-42010a800087  used         6K     -
cstor-d82bd105-f3a8-11e8-87fd-42010a800087/pvc-1b2a7d4b-f3a9-11e8-87fd-42010a800087  logicalused  6K     -
*/
func Capacity(volName string) (*apis.CStorVolumeCapacityAttr, error) {
	capacityVolStr := []string{"get", "used,logicalused", volName}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(VolumeReplicaOperator, capacityVolStr...)
	if err != nil {
		glog.Errorf("Unable to get volume capacity: %v", string(stdoutStderr))
		return nil, err
	}
	poolCapacity := capacityOutputParser(string(stdoutStderr))
	if strings.TrimSpace(poolCapacity.TotalAllocated) == "" || strings.TrimSpace(poolCapacity.Used) == "" {
		return nil, fmt.Errorf("unable to get volume capacity from capacity parser")
	}
	return poolCapacity, nil
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
	var volumeStatus string
	if volumeStats != nil && len(volumeStats.Stats) != 0 {
		volumeStatus = volumeStats.Stats[0].Status
	}
	if strings.TrimSpace(volumeStatus) == "" {
		glog.Warningf("Empty volume status for volume stats: '%+v'", volumeStats)
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

/*
root@cstor-sparse-pool-6dft-5b5c78ccc7-dls8s:/# zfs get used,logicalused cstor-d82bd105-f3a8-11e8-87fd-42010a800087/pvc-1b2a7d4b-f3a9-11e8-87fd-42010a800087
NAME                                                                                 PROPERTY     VALUE  SOURCE
cstor-d82bd105-f3a8-11e8-87fd-42010a800087/pvc-1b2a7d4b-f3a9-11e8-87fd-42010a800087  used         6K     -
cstor-d82bd105-f3a8-11e8-87fd-42010a800087/pvc-1b2a7d4b-f3a9-11e8-87fd-42010a800087  logicalused  6K     -
*/
// capacityOutputParser parse output of `zfs get` command to extract the capacity of the pool.
// ToDo: Need to find some better way e.g contract for zfs command outputs.
func capacityOutputParser(output string) *apis.CStorVolumeCapacityAttr {
	var outputStr []string
	// Initialize capacity object.
	// 'TotalAllocated' value(on cvr) is filled from the value of 'used' property in 'zfs get' output.
	// 'Used' value(on cvr) is filled from the value of 'logicalused' property in 'zfs get' output.
	capacity := &apis.CStorVolumeCapacityAttr{
		"",
		"",
	}
	if strings.TrimSpace(string(output)) != "" {
		outputStr = strings.Split(string(output), "\n")
		if !(len(outputStr) < 3) {
			poolCapacityArrAlloc := strings.Fields(outputStr[1])
			poolCapacityArrUsed := strings.Fields(outputStr[2])
			// If the array 'poolCapacityArrAlloc' and 'poolCapacityArrUsed' is having elements greater than
			// or less than 4 it might give wrong values and throw out of bound exception.
			if len(poolCapacityArrAlloc) == 4 && len(poolCapacityArrUsed) == 4 {
				capacity.TotalAllocated = strings.TrimSpace(poolCapacityArrAlloc[2])
				capacity.Used = strings.TrimSpace(poolCapacityArrUsed[2])
			}
		}
	}
	return capacity
}
