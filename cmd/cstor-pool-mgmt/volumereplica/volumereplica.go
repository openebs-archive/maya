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
	"os"
	"strings"
	"time"

	"github.com/openebs/maya/pkg/alertlog"

	"encoding/json"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/debug"
	"github.com/openebs/maya/pkg/hash"
	"github.com/openebs/maya/pkg/util"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/klog"
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
	IsIOReceiverCreated       int    `json:"isIOReceiverCreated"`
	RunningIONum              int    `json:"runningIONum"`
	CheckpointedIONum         int    `json:"checkpointedIONum"`
	DegradedCheckpointedIONum int    `json:"degradedCheckpointedIONum"`
	CheckpointedTime          int    `json:"checkpointedTime"`
	RebuildBytes              int    `json:"rebuildBytes"`
	RebuildCnt                int    `json:"rebuildCnt"`
	RebuildDoneCnt            int    `json:"rebuildDoneCnt"`
	RebuildFailedCnt          int    `json:"rebuildFailedCnt"`
	Quorum                    int    `json:"quorum"`
}

// RunnerVar the runner variable for executing binaries.
var RunnerVar util.Runner

// PoolNameFromCVR gets the name of cstorpool from cstorvolumereplica label
// if not found then gets cstorpoolinstance name from the OPENEBS_IO_POOL_NAME
// env
func PoolNameFromCVR(cvr *apis.CStorVolumeReplica) string {
	poolname := cvr.Labels[CStorPoolUIDKey]
	if strings.TrimSpace(poolname) == "" {
		poolname = os.Getenv(string("OPENEBS_IO_POOL_NAME"))
		if strings.TrimSpace(poolname) == "" {
			return ""
		}
	}
	return PoolPrefix + poolname
}

// PoolNameFromBackup gets the name of cstorpool from cstorvolumereplica label
// if not found then gets cstorpoolinstance name from the OPENEBS_IO_POOL_NAME
// env
func PoolNameFromBackup(bkp *apis.CStorBackup) string {
	poolname := bkp.Labels[CStorPoolUIDKey]
	if strings.TrimSpace(poolname) == "" {
		poolname = os.Getenv(string("OPENEBS_IO_POOL_NAME"))
		if strings.TrimSpace(poolname) == "" {
			return ""
		}
	}
	return PoolPrefix + poolname

}

// PoolNameFromRestore gets the name of cstorPool from cstorvolumereplica label
// if not found then gets cstorPoolInstance name from the OPENEBS_IO_POOL_NAME
// env
func PoolNameFromRestore(rst *apis.CStorRestore) string {
	poolname := rst.Labels[CStorPoolUIDKey]
	if strings.TrimSpace(poolname) == "" {
		poolname = os.Getenv(string("OPENEBS_IO_POOL_NAME"))
		if strings.TrimSpace(poolname) == "" {
			return ""
		}
	}
	return PoolPrefix + poolname
}

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
	if len(cVR.Labels["cstorpool.openebs.io/uid"]) == 0 &&
		len(cVR.Labels["cstorpoolinstance.openebs.io/uid"]) == 0 {
		err = fmt.Errorf("Pool cannot be empty")
		return err
	}
	if len(cVR.Labels["cstorpool.openebs.io/uid"]) != 0 &&
		len(cVR.Labels["cstorpoolinstance.openebs.io/uid"]) != 0 {
		err = fmt.Errorf("Both pool types related labels are available")
		return err
	}
	return nil
}

// CreateVolumeReplica creates cStor replica(zfs volumes).
func CreateVolumeReplica(cStorVolumeReplica *apis.CStorVolumeReplica, fullVolName string, quorum bool) error {
	var cmd []string
	isClone := cStorVolumeReplica.Labels[string(apis.CloneEnableKEY)] == "true"
	snapName := ""

	if debug.EI.IsZFSCreateErrorInjected() {
		return errors.New("ZFS create error via injection")
	}

	if isClone {
		srcVolume := cStorVolumeReplica.Annotations[string(apis.SourceVolumeKey)]
		snapName = cStorVolumeReplica.Annotations[string(apis.SnapshotNameKey)]
		// Get the dataset name from volume name
		dataset := strings.Split(fullVolName, "/")[0]
		klog.Infof("Creating clone volume: %s from snapshot %s", fullVolName, srcVolume+"@"+snapName)
		// zfs snapshots are named as dataset/volname@snapname
		cmd = buildVolumeCloneCommand(cStorVolumeReplica, dataset+"/"+srcVolume+"@"+snapName, fullVolName)
	} else {
		// Parse capacity unit on CVR to support backward compatibility
		volCapacity := parseCapacityUnit(cStorVolumeReplica.Spec.Capacity)
		cStorVolumeReplica.Spec.Capacity = volCapacity
		cmd = buildVolumeCreateCommand(cStorVolumeReplica, fullVolName, quorum)
	}

	stdoutStderr, err := RunnerVar.RunCombinedOutput(VolumeReplicaOperator, cmd...)
	if err != nil {
		if isClone {
			klog.Errorf("Unable to create clone volume: %s for snapshot %s. error : %v", fullVolName, snapName, string(stdoutStderr))
			alertlog.Logger.Errorw("",
				"eventcode", "cstor.volume.replica.clone.create.failure",
				"msg", "Failed to create CStor volume replica clone",
				"rname", fullVolName,
			)
		} else {
			klog.Errorf("Unable to create volume %s. error : %v", fullVolName, string(stdoutStderr))
			alertlog.Logger.Errorw("",
				"eventcode", "cstor.volume.replica.create.failure",
				"msg", "Failed to create CStor volume replica",
				"rname", fullVolName,
			)
		}

		return errors.Wrapf(err, "failed to create volume: %s", fullVolName)
	}

	if isClone {
		alertlog.Logger.Infow("",
			"eventcode", "cstor.volume.replica.clone.create.success",
			"msg", "Successfully created CStor volume replica clone",
			"rname", fullVolName,
		)
	} else {
		alertlog.Logger.Infow("",
			"eventcode", "cstor.volume.replica.create.success",
			"msg", "Successfully created CStor volume replica",
			"rname", fullVolName,
		)
	}

	return nil
}

// builldVolumeCreateCommand returns volume create command along with attributes as a string array
func buildVolumeCreateCommand(cStorVolumeReplica *apis.CStorVolumeReplica, fullVolName string, quorum bool) []string {
	var createVolCmd []string

	openebsVolname := "io.openebs:volname=" + cStorVolumeReplica.ObjectMeta.Name
	openebsTargetIP := "io.openebs:targetip=" + cStorVolumeReplica.Spec.TargetIP
	// ZvolWorkers represents number of threads that executes client IOs
	openebsZvolWorkers := "io.openebs:zvol_workers=" + cStorVolumeReplica.Spec.ZvolWorkers
	// ReplicaId represents unique identification number for volume
	openebsZvolID := "io.openebs:zvol_replica_id=" + fmt.Sprintf("%v", cStorVolumeReplica.Spec.ReplicaID)

	quorumValue := "quorum=on"
	if !quorum {
		quorumValue = "quorum=off"
	}

	// set volume property
	createVolCmd = append(createVolCmd, CreateCmd,
		"-b",
		"4K",
		"-s",
		"-o", "compression=on",
		"-o", quorumValue,
		"-o", openebsZvolID,
		"-o", openebsVolname,
	)

	if len(cStorVolumeReplica.Spec.ZvolWorkers) != 0 {
		createVolCmd = append(createVolCmd,
			"-o", openebsZvolWorkers,
		)
	}

	if cStorVolumeReplica.Annotations["isRestoreVol"] != "true" {
		createVolCmd = append(createVolCmd,
			"-o", openebsTargetIP,
		)
	}

	// append volume size and volume name
	return append(createVolCmd, "-V", cStorVolumeReplica.Spec.Capacity, fullVolName)
}

// builldVolumeCloneCommand returns volume clone command along with attributes as a string array
func buildVolumeCloneCommand(cStorVolumeReplica *apis.CStorVolumeReplica, snapName, fullVolName string) []string {
	var cloneVolCmd []string

	openebsVolname := "io.openebs:volname=" + cStorVolumeReplica.ObjectMeta.Name
	openebsTargetIP := "io.openebs:targetip=" + cStorVolumeReplica.Spec.TargetIP
	// ZvolWorkers represents number of threads that executes client IOs
	openebsZvolWorkers := "io.openebs:zvol_workers=" + cStorVolumeReplica.Spec.ZvolWorkers
	// ReplicaId represents unique identification number for volume
	openebsZvolID := "io.openebs:zvol_replica_id=" + fmt.Sprintf("%v", cStorVolumeReplica.Spec.ReplicaID)

	if len(cStorVolumeReplica.Spec.ZvolWorkers) != 0 {
		cloneVolCmd = append(cloneVolCmd, CloneCmd,
			"-o", "compression=on",
			"-o", openebsTargetIP,
			"-o", "quorum=on",
			"-o", openebsZvolWorkers,
			"-o", openebsVolname,
			"-o", openebsZvolID,
			snapName, fullVolName)
	} else {
		cloneVolCmd = append(cloneVolCmd, CloneCmd,
			"-o", "compression=on",
			"-o", openebsTargetIP,
			"-o", "quorum=on",
			"-o", openebsZvolID,
			"-o", openebsVolname,
			snapName, fullVolName)
	}
	return cloneVolCmd
}

// CreateVolumeBackup sends cStor snapshots to remote location specified by cstorbackup.
func CreateVolumeBackup(bkp *apis.CStorBackup) error {
	var cmd []string
	var retryCount int
	var err error
	var stdoutStderr []byte

	// Parse capacity unit on CVR to support backward compatibility
	cmd = buildVolumeBackupCommand(PoolNameFromBackup(bkp), bkp.Spec.VolumeName, bkp.Spec.PrevSnapName, bkp.Spec.SnapName, bkp.Spec.BackupDest)

	klog.Infof("Backup Command for volume: %v created, Cmd: %v\n", bkp.Spec.VolumeName, cmd)

	for retryCount < MaxBackupRetryCount {
		stdoutStderr, err = RunnerVar.RunCombinedOutput("/usr/local/bin/execute.sh", cmd...)
		if err != nil {
			klog.Errorf("Unable to start backup %s. error : %v retry:%v :%s", bkp.Spec.VolumeName, string(stdoutStderr), retryCount, err.Error())
			retryCount++
			time.Sleep(BackupRetryDelay * time.Second)
			continue
		}
		break
	}
	if err != nil {
		alertlog.Logger.Errorw("",
			"eventcode", "cstor.volume.backup.create.failure",
			"msg", "Failed to create backup CStor volume",
			"rname", bkp.Spec.VolumeName,
		)
	} else {
		alertlog.Logger.Infow("",
			"eventcode", "cstor.volume.backup.create.success",
			"msg", "Successfully created backup CStor volume",
			"rname", bkp.Spec.VolumeName,
		)
	}
	return err
}

// builldVolumeBackupCommand returns volume create command along with attributes as a string array
func buildVolumeBackupCommand(poolName, fullVolName, oldSnapName, newSnapName, backupDest string) []string {
	var startBackupCmd []string

	bkpAddr := strings.Split(backupDest, ":")
	if oldSnapName == "" {
		startBackupCmd = append(startBackupCmd, VolumeReplicaOperator, BackupCmd, poolName+"/"+fullVolName+"@"+newSnapName, "| nc -w 3 "+bkpAddr[0]+" "+bkpAddr[1])
	} else {
		startBackupCmd = append(startBackupCmd, VolumeReplicaOperator, BackupCmd,
			"-i", poolName+"/"+fullVolName+"@"+oldSnapName, poolName+"/"+fullVolName+"@"+newSnapName, "| nc -w 3 "+bkpAddr[0]+" "+bkpAddr[1])
	}
	return startBackupCmd
}

// CreateVolumeRestore receive cStor snapshots from remote location(zfs volumes).
func CreateVolumeRestore(rst *apis.CStorRestore) error {
	var cmd []string
	var retryCount int
	var err error
	var stdoutStderr []byte

	cmd = buildVolumeRestoreCommand(PoolNameFromRestore(rst), rst.Spec.VolumeName, rst.Spec.RestoreSrc)

	klog.Infof("Restore Command for volume: %v created, Cmd: %v\n", rst.Spec.VolumeName, cmd)

	for retryCount < MaxRestoreRetryCount {
		stdoutStderr, err = RunnerVar.RunCombinedOutput("/usr/local/bin/execute.sh", cmd...)
		if err != nil {
			klog.Errorf("Unable to start restore %s. error : %v.. trying again", rst.Spec.VolumeName, string(stdoutStderr))
			time.Sleep(RestoreRetryDelay * time.Second)
			retryCount++
			continue
		}
		break
	}
	if err != nil {
		alertlog.Logger.Errorw("",
			"eventcode", "cstor.volume.restore.failure",
			"msg", "Failed to restore CStor volume",
			"rname", rst.Spec.VolumeName,
		)
	} else {
		alertlog.Logger.Infow("",
			"eventcode", "cstor.volume.restore.success",
			"msg", "Successfully restored CStor volume",
			"rname", rst.Spec.VolumeName,
		)
	}
	return err
}

// builldVolumeRestoreCommand returns restore command along with attributes as a string array
func buildVolumeRestoreCommand(poolName, fullVolName, restoreSrc string) []string {
	var restorecmd []string

	restorAddr := strings.Split(restoreSrc, ":")
	restorecmd = append(restorecmd, "nc -w 3 "+restorAddr[0]+" "+restorAddr[1]+" | ", VolumeReplicaOperator, RestoreCmd,
		" -F "+poolName+"/"+fullVolName)

	return restorecmd
}

// GetVolumes returns the slice of volumes.
func GetVolumes() ([]string, error) {
	if debug.EI.IsZFSGetErrorInjected() {
		return []string{}, errors.New("ZFS get error via injection")
	}

	volStrCmd := []string{"get", "-Hp", "name", "-o", "name"}
	volnameByte, err := RunnerVar.RunStdoutPipe(VolumeReplicaOperator, volStrCmd...)
	if err != nil || string(volnameByte) == "" {
		klog.Errorf("Unable to get volumes:%v", string(volnameByte))
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
	if debug.EI.IsZFSDeleteErrorInjected() {
		return errors.New("ZFS delete error via injection")
	}

	deleteVolStr := []string{"destroy", "-R", fullVolName}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(VolumeReplicaOperator, deleteVolStr...)
	if err != nil {
		// If volume is missing then do not return error
		if strings.Contains(string(stdoutStderr), "dataset does not exist") {
			klog.Infof("Assuming volume deletion successful for error: %v", string(stdoutStderr))
			return nil
		}
		klog.Errorf("Unable to delete volume : %v", string(stdoutStderr))
		alertlog.Logger.Errorw("",
			"eventcode", "cstor.volume.delete.failure",
			"msg", "Failed to delete CStor volume",
			"rname", fullVolName,
		)
		return errors.Wrapf(err, "failed to delete volume.. %s", string(stdoutStderr))
	}
	alertlog.Logger.Infow("",
		"eventcode", "cstor.volume.delete.success",
		"msg", "Successfully deleted CStor volume",
		"rname", fullVolName,
	)
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
		klog.Errorf("Unable to get volume capacity: %v", string(stdoutStderr))
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
		klog.Errorf("Unable to get volume stats: %v", string(stdoutStderr))
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
		klog.Warningf("Empty volume status for volume stats: '%+v'", volumeStats)
	}
	cvrStatus := ZfsToCvrStatusMapper(volumeStatus, volumeStats.Stats[0].Quorum)
	return cvrStatus, nil
}

// GetVolumeName finds the zctual zfs volume name for the given cvr.
func GetVolumeName(cVR *apis.CStorVolumeReplica) (string, error) {
	var volumeName string
	// Get the corresponding CSP UID for this CVR
	if cVR.Labels == nil {
		return "", fmt.Errorf("no labels found on cvr object")
	}
	poolname := PoolNameFromCVR(cVR)
	pvName := cVR.Labels[PvNameKey]
	if strings.TrimSpace(pvName) == "" {
		return "", fmt.Errorf("pv name not found on cvr label")
	}
	volumeName = poolname + "/" + pvName
	return volumeName, nil
}

// ZfsToCvrStatusMapper maps zfs status to defined cvr status.
func ZfsToCvrStatusMapper(zfsstatus string, quorum int) string {
	switch zfsstatus {
	case ZfsStatusHealthy:
		return string(apis.CVRStatusOnline)
	case ZfsStatusOffline:
		return string(apis.CVRStatusOffline)
	case ZfsStatusDegraded:
		if quorum == 1 {
			return string(apis.CVRStatusDegraded)
		}
		return string(apis.CVRStatusNewReplicaDegraded)
	case ZfsStatusRebuilding:
		if quorum == 1 {
			return string(apis.CVRStatusRebuilding)
		}
		return string(apis.CVRStatusReconstructingNewReplica)
	default:
		return string(apis.CVRStatusError)
	}
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
		TotalAllocated: "",
		Used:           "",
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

// GenerateReplicaID generate new replicaID for given CVR
func GenerateReplicaID(cvr *apis.CStorVolumeReplica) error {
	if len(cvr.Spec.ReplicaID) != 0 {
		return errors.Errorf("ReplicaID for cvr(%s) is already generated", cvr.Name)
	}

	csum, err := hash.Hash(cvr.UID)
	if err != nil {
		return err
	}
	cvr.Spec.ReplicaID = strings.ToUpper(csum)
	return nil
}

// GetReplicaIDFromZFS returns replicaID for provided volume name by executing
// ZFS commands
func GetReplicaIDFromZFS(volumeName string) (string, error) {
	ret, err := zfs.NewVolumeGetProperty().
		WithScriptedMode(true).
		WithParsableMode(true).
		WithField("value").
		WithProperty("io.openebs:zvol_replica_id").
		WithDataset(volumeName).
		Execute()
	if err != nil {
		return "", errors.Errorf("Failed to get replicaID %s", err)
	}
	return strings.Split(string(ret), "\n")[0], nil
}

// SetReplicaID set replicaID to volume
func SetReplicaID(cvr *apis.CStorVolumeReplica) error {
	var err error

	vol, err := GetVolumeName(cvr)
	if err != nil {
		return err
	}

	sid, err := GetReplicaIDFromZFS(vol)
	if err != nil {
		return err
	}

	if len(sid) == 0 {
		lr, err := zfs.NewVolumeSetProperty().
			WithProperty("io.openebs:zvol_replica_id", cvr.Spec.ReplicaID).
			WithDataset(vol).
			Execute()
		if err != nil {
			return errors.Errorf("Failed to set replicaID %s %s", err, string(lr))
		}
	} else if cvr.Spec.ReplicaID != sid {
		return errors.Errorf("ReplicaID mismatch.. actual(%s) on-disk(%s)", cvr.Spec.ReplicaID, sid)
	}

	return nil
}

// GetAndUpdateReplicaID update replicaID for CVR and set it to volume
func GetAndUpdateReplicaID(cvr *apis.CStorVolumeReplica) error {
	if len(cvr.Spec.ReplicaID) == 0 {
		if err := GenerateReplicaID(cvr); err != nil {
			return errors.Errorf("CVR(%s) replicaID generation error %s",
				cvr.Name, err)
		}
	}

	if err := SetReplicaID(cvr); err != nil {
		return errors.Errorf("Failed to set ReplicaID for CVR(%s).. %s", cvr.Name, err)
	}
	return nil
}
