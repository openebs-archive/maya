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

package pool

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	zpool "github.com/openebs/maya/pkg/apis/openebs.io/zpool/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/pkg/errors"
)

var (
	poolTypeCommand  = map[string]string{"mirrored": "mirror", "raidz": "raidz", "raidz2": "raidz2"}
	defaultGroupSize = map[string]int{"striped": 1, "mirrored": 2, "raidz": 3, "raidz2": 6}
)

// PoolOperator is the name of the tool that makes pool-related operations.
const (
	StatusNoPoolsAvailable = "no pools available"
	ZpoolStatusDegraded    = "DEGRADED"
	ZpoolStatusFaulted     = "FAULTED"
	ZpoolStatusOffline     = "OFFLINE"
	ZpoolStatusOnline      = "ONLINE"
	ZpoolStatusRemoved     = "REMOVED"
	ZpoolStatusUnavail     = "UNAVAIL"
)

//PoolAddEventHandled is a flag representing if the pool has been initially imported or created
var PoolAddEventHandled = false

// PoolNamePrefix is a typed string to store pool name prefix
type PoolNamePrefix string

// ImportedCStorPools is a map of imported cstor pools API config identified via their UID
var ImportedCStorPools map[string]*apis.CStorPool

// CStorZpools is a map of imported cstor pools config at backend identified via their UID
var CStorZpools map[string]zpool.Topology

// PoolPrefix is prefix for pool name
const (
	PoolPrefix PoolNamePrefix = "cstor-"
)

// RunnerVar the runner variable for executing binaries.
var RunnerVar util.Runner

// ImportPool imports cStor pool if already present.
func ImportPool(cStorPool *apis.CStorPool, cachefileFlag bool) error {
	importAttr := importPoolBuilder(cStorPool, cachefileFlag)
	stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, importAttr...)
	if err != nil {
		glog.Errorf("Unable to import pool: %v, %v", err.Error(), string(stdoutStderr))
		return err
	}
	glog.Info("Importing Pool Successful")
	return nil
}

// importPoolBuilder is to build pool import command.
func importPoolBuilder(cStorPool *apis.CStorPool, cachefileFlag bool) []string {
	// populate pool import attributes.
	var importAttr []string
	importAttr = append(importAttr, "import")
	if cStorPool.Spec.PoolSpec.CacheFile != "" && cachefileFlag {
		importAttr = append(importAttr, "-c", cStorPool.Spec.PoolSpec.CacheFile,
			"-o", cStorPool.Spec.PoolSpec.CacheFile)
	}
	importAttr = append(importAttr, string(PoolPrefix)+string(cStorPool.ObjectMeta.UID))
	return importAttr
}

// CreatePool creates a new cStor pool.
func CreatePool(cStorPool *apis.CStorPool, blockDeviceList []string) error {
	createAttr := createPoolBuilder(cStorPool, blockDeviceList)
	glog.V(4).Info("createAttr : ", createAttr)

	stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, createAttr...)
	if err != nil {
		glog.Errorf("Unable to create pool: %v", string(stdoutStderr))
		return err
	}
	return nil
}

// createPoolBuilder is to build create pool command.
func createPoolBuilder(cStorPool *apis.CStorPool, blockDeviceList []string) []string {
	// populate pool creation attributes.
	var createAttr []string
	// When block devices of other file formats, say ext4, are used to create cstorpool,
	// it errors out with normal zpool create. To avoid that, we go for forceful create.
	createAttr = append(createAttr, "create", "-f")
	if cStorPool.Spec.PoolSpec.CacheFile != "" {
		cachefile := "cachefile=" + cStorPool.Spec.PoolSpec.CacheFile
		createAttr = append(createAttr, "-o", cachefile)
	}

	openebsPoolname := "io.openebs:poolname=" + cStorPool.Name
	createAttr = append(createAttr, "-O", openebsPoolname)

	poolNameUID := string(PoolPrefix) + string(cStorPool.ObjectMeta.UID)
	createAttr = append(createAttr, poolNameUID)
	poolType := cStorPool.Spec.PoolSpec.PoolType
	if poolType == "striped" {
		createAttr = append(createAttr, blockDeviceList...)
		return createAttr
	}
	// To generate pool of the following types:
	// mirrored (grouped by multiples of 2): mirror blockdevice1 blockdevice2 mirror blockdevice3 blockdevice4
	// raidz (grouped by multiples of 3): raidz blockdevice1 blockdevice2 blockdevice3 raidz blockdevice 4 blockdevice5 blockdevice6
	// raidz2 (grouped by multiples of 6): raidz2 blockdevice1 blockdevice2 blockdevice3 blockdevice4 blockdevice5 blockdevice6
	for i, bd := range blockDeviceList {
		if i%defaultGroupSize[poolType] == 0 {
			createAttr = append(createAttr, poolTypeCommand[poolType])
		}
		createAttr = append(createAttr, bd)
	}

	return createAttr
}

// ValidatePool checks for validity of CStorPool resource.
func ValidatePool(cStorPool *apis.CStorPool, devID []string) error {
	poolUID := cStorPool.ObjectMeta.UID
	if len(poolUID) == 0 {
		return fmt.Errorf("Poolname/UID cannot be empty")
	}
	diskCount := len(devID)
	poolType := cStorPool.Spec.PoolSpec.PoolType
	if diskCount < defaultGroupSize[poolType] {
		return errors.Errorf(
			"csp validation failed: expected {%d} blockdevices got {%d}, for pool type {%s}",
			defaultGroupSize[poolType],
			diskCount,
			poolType,
		)
	}
	if diskCount%defaultGroupSize[poolType] != 0 {
		return errors.Errorf(
			"csp validation failed: expected multiples of {%d} blockdevices required got {%d}, for pool type {%s}",
			defaultGroupSize[poolType],
			diskCount,
			poolType,
		)
	}
	return nil
}

// GetPoolName return the pool already created.
func GetPoolName() ([]string, error) {
	GetPoolStr := []string{"get", "-Hp", "name", "-o", "name"}
	poolNameByte, err := RunnerVar.RunStdoutPipe(zpool.PoolOperator, GetPoolStr...)
	if err != nil || len(string(poolNameByte)) == 0 {
		return []string{}, err
	}
	noisyPoolName := string(poolNameByte)
	sepNoisyPoolName := strings.Split(noisyPoolName, "\n")
	var poolNames []string
	for _, poolName := range sepNoisyPoolName {
		poolName = strings.TrimSpace(poolName)
		poolNames = append(poolNames, poolName)
	}
	return poolNames, nil
}

// DeletePool destroys the pool created.
func DeletePool(poolName string) error {
	deletePoolStr := []string{"destroy", poolName}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, deletePoolStr...)
	if err != nil {
		glog.Errorf("Unable to delete pool: %v", string(stdoutStderr))
		return err
	}
	return nil
}

// Capacity finds the capacity of the pool.
// The ouptut of command executed is as follows:
/*
root@cstor-sparse-pool-o8bw-6869f69cc8-jhs6c:/# zpool get size,free,allocated cstor-2ebe403a-f2e2-11e8-87fd-42010a800087
NAME                                        PROPERTY   VALUE  SOURCE
cstor-2ebe403a-f2e2-11e8-87fd-42010a800087  size       9.94G  -
cstor-2ebe403a-f2e2-11e8-87fd-42010a800087  free       9.94G  -
cstor-2ebe403a-f2e2-11e8-87fd-42010a800087  allocated  202K   -
*/
func Capacity(poolName string) (*apis.CStorPoolCapacityAttr, error) {
	capacityPoolStr := []string{"get", "size,free,allocated", poolName}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, capacityPoolStr...)
	if err != nil {
		glog.Errorf("Unable to get pool capacity: %v", string(stdoutStderr))
		return nil, err
	}
	poolCapacity := capacityOutputParser(string(stdoutStderr))
	if strings.TrimSpace(poolCapacity.Used) == "" || strings.TrimSpace(poolCapacity.Free) == "" {
		return nil, fmt.Errorf("Unable to get pool capacity from capacity parser")
	}
	return poolCapacity, nil
}

// PoolStatus finds the status of the pool.
// The ouptut of command(`zpool status <pool-name>`) executed is as follows:

/*
		  pool: cstor-530c9c4f-e0df-11e8-94a8-42010a80013b
	 state: ONLINE
	  scan: none requested
	config:

		NAME                                        STATE     READ WRITE CKSUM
		cstor-530c9c4f-e0df-11e8-94a8-42010a80013b  ONLINE       0     0     0
		  scsi-0Google_PersistentDisk_ashu-disk2    ONLINE       0     0     0

	errors: No known data errors
*/
// The output is then parsed by poolStatusOutputParser function to get the status of the pool
func Status(poolName string) (string, error) {
	var poolStatus string
	statusPoolStr := []string{"status", poolName}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, statusPoolStr...)
	if err != nil {
		glog.Errorf("Unable to get pool status: %v", string(stdoutStderr))
		return "", err
	}
	poolStatus = poolStatusOutputParser(string(stdoutStderr))
	if poolStatus == ZpoolStatusDegraded {
		return string(apis.CStorPoolStatusDegraded), nil
	} else if poolStatus == ZpoolStatusFaulted {
		return string(apis.CStorPoolStatusOffline), nil
	} else if poolStatus == ZpoolStatusOffline {
		return string(apis.CStorPoolStatusOffline), nil
	} else if poolStatus == ZpoolStatusOnline {
		return string(apis.CStorPoolStatusOnline), nil
	} else if poolStatus == ZpoolStatusRemoved {
		return string(apis.CStorPoolStatusDegraded), nil
	} else if poolStatus == ZpoolStatusUnavail {
		return string(apis.CStorPoolStatusOffline), nil
	} else {
		return string(apis.CStorPoolStatusError), nil
	}
}

// poolStatusOutputParser parse output of `zpool status` command to extract the status of the pool.
// ToDo: Need to find some better way e.g contract for zpool command outputs.
func poolStatusOutputParser(output string) string {
	var outputStr []string
	var poolStatus string
	if strings.TrimSpace(string(output)) != "" {
		outputStr = strings.Split(string(output), "\n")
		if !(len(outputStr) < 2) {
			poolStatusArr := strings.Split(outputStr[1], ":")
			if !(len(outputStr) < 2) {
				poolStatus = strings.TrimSpace(poolStatusArr[1])
			}
		}
	}
	return poolStatus
}

// capacityOutputParser parse output of `zpool get` command to extract the capacity of the pool.
// ToDo: Need to find some better way e.g contract for zpool command outputs.
func capacityOutputParser(output string) *apis.CStorPoolCapacityAttr {
	var outputStr []string
	// Initialize capacity object.
	capacity := &apis.CStorPoolCapacityAttr{
		"",
		"",
		"",
	}
	if strings.TrimSpace(string(output)) != "" {
		outputStr = strings.Split(string(output), "\n")
		if !(len(outputStr) < 4) {
			poolCapacityArrTotal := strings.Fields(outputStr[1])
			poolCapacityArrFree := strings.Fields(outputStr[2])
			poolCapacityArrAlloc := strings.Fields(outputStr[3])
			if !(len(poolCapacityArrTotal) < 4 || len(poolCapacityArrFree) < 4) || len(poolCapacityArrAlloc) < 4 {
				capacity.Total = strings.TrimSpace(poolCapacityArrTotal[2])
				capacity.Free = strings.TrimSpace(poolCapacityArrFree[2])
				capacity.Used = strings.TrimSpace(poolCapacityArrAlloc[2])
			}
		}
	}
	return capacity
}

// SetCachefile is to set the cachefile for pool.
func SetCachefile(cStorPool *apis.CStorPool) error {
	poolNameUID := string(PoolPrefix) + string(cStorPool.ObjectMeta.UID)
	setCachefileStr := []string{"set", "cachefile=" + cStorPool.Spec.PoolSpec.CacheFile,
		poolNameUID}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, setCachefileStr...)
	if err != nil {
		glog.Errorf("Unable to set cachefile: %v", string(stdoutStderr))
		return err
	}
	return nil
}

// CheckForZreplInitial is blocking call for checking status of zrepl in cstor-pool container.
func CheckForZreplInitial(ZreplRetryInterval time.Duration) {
	for {
		_, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, "status")
		if err != nil {
			time.Sleep(ZreplRetryInterval)
			glog.Errorf("zpool status returned error in zrepl startup : %v", err)
			glog.Infof("Waiting for zpool replication container to start...")
			continue
		}
		break
	}
}

// CheckForZreplContinuous is continuous health checker for status of zrepl in cstor-pool container.
func CheckForZreplContinuous(ZreplRetryInterval time.Duration) {
	for {
		out, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, "status")
		if err == nil {
			//even though we imported pool, it disappeared (may be due to zrepl container crashing).
			// so we need to reimport.
			if PoolAddEventHandled && strings.Contains(string(out), StatusNoPoolsAvailable) {
				break
			}
			time.Sleep(ZreplRetryInterval)
			continue
		}
		glog.Errorf("zpool status returned error in zrepl healthcheck : %v, out: %s", err, out)
		break
	}
}

// LabelClear is to clear zpool label on block devices.
func LabelClear(blockDevices []string) error {
	var failLabelClear = false
	for _, bd := range blockDevices {
		labelClearStr := []string{"labelclear", "-f", bd}
		stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, labelClearStr...)
		if err != nil {
			glog.Errorf("Unable to clear label on blockdevice %v: %v, err = %v", bd,
				string(stdoutStderr), err)
			failLabelClear = true
		}
	}
	if failLabelClear {
		return fmt.Errorf("Unable to clear labels from all the blockdevices of the pool")
	}
	return nil
}

// GetDeviceIDs returns the list of device IDs for the csp.
func GetDeviceIDs(csp *apis.CStorPool) ([]string, error) {
	var bdDeviceID []string
	for _, group := range csp.Spec.Group {
		for _, blockDevice := range group.Item {
			bdDeviceID = append(bdDeviceID, blockDevice.DeviceID)
		}
	}
	if len(bdDeviceID) == 0 {
		return nil, errors.Errorf("No device IDs found on the csp %s", csp.Name)
	}
	return bdDeviceID, nil
}
