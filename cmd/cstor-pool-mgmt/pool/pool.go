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
	"github.com/openebs/maya/pkg/util"
)

// PoolOperator is the name of the tool that makes pool-related operations.
const (
	PoolOperator           = "zpool"
	StatusNoPoolsAvailable = "no pools available"
)

//PoolAddEventHandled is a flag representing if the pool has been initially imported or created
var PoolAddEventHandled = false

// PoolNamePrefix defines prefix name for the pool
type PoolNamePrefix string

// PoolPrefix is a prefix as cstor- for the PoolNamePrefix
const (
	PoolPrefix PoolNamePrefix = "cstor-"
)

// RunnerVar the runner variable for executing binaries.
var RunnerVar util.Runner

// ImportPool imports cStor pool if already present.
func ImportPool(cStorPool *apis.CStorPool, cachefileFlag bool) error {
	importAttr := importPoolBuilder(cStorPool, cachefileFlag)
	stdoutStderr, err := RunnerVar.RunCombinedOutput(PoolOperator, importAttr...)
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
func CreatePool(cStorPool *apis.CStorPool) error {
	createAttr := createPoolBuilder(cStorPool)
	glog.V(4).Info("createAttr : ", createAttr)

	stdoutStderr, err := RunnerVar.RunCombinedOutput(PoolOperator, createAttr...)
	if err != nil {
		glog.Errorf("Unable to create pool: %v", string(stdoutStderr))
		return err
	}
	return nil
}

// createPoolBuilder is to build create pool command.
func createPoolBuilder(cStorPool *apis.CStorPool) []string {
	// populate pool creation attributes.
	var createAttr []string
	// When disks of other file formats, say ext4, are used to create cstorpool,
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

	// To generate mirror disk0 disk1 mirror disk2 disk3 format.
	for i, disk := range cStorPool.Spec.Disks.DiskList {
		if cStorPool.Spec.PoolSpec.PoolType == "mirror" && i%2 == 0 {
			createAttr = append(createAttr, "mirror")
		}
		createAttr = append(createAttr, disk)
	}
	return createAttr

}

// CheckValidPool checks for validity of CStorPool resource.
func CheckValidPool(cStorPool *apis.CStorPool) error {
	if len(string(cStorPool.ObjectMeta.UID)) == 0 {
		return fmt.Errorf("Poolname/UID cannot be empty")
	}
	if len(cStorPool.Spec.Disks.DiskList) < 1 {
		return fmt.Errorf("Disk name(s) cannot be empty")
	}
	if cStorPool.Spec.PoolSpec.PoolType == "mirror" &&
		len(cStorPool.Spec.Disks.DiskList)%2 != 0 {
		return fmt.Errorf("Mirror poolType needs even number of disks")
	}
	return nil
}

// GetPoolName return the pool already created.
func GetPoolName() ([]string, error) {
	GetPoolStr := []string{"get", "-Hp", "name", "-o", "name"}
	poolNameByte, err := RunnerVar.RunStdoutPipe(PoolOperator, GetPoolStr...)
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
	stdoutStderr, err := RunnerVar.RunCombinedOutput(PoolOperator, deletePoolStr...)
	if err != nil {
		glog.Errorf("Unable to delete pool: %v", string(stdoutStderr))
		return err
	}
	return nil
}

// SetCachefile is to set the cachefile for pool.
func SetCachefile(cStorPool *apis.CStorPool) error {
	poolNameUID := string(PoolPrefix) + string(cStorPool.ObjectMeta.UID)
	setCachefileStr := []string{"set", "cachefile=" + cStorPool.Spec.PoolSpec.CacheFile,
		poolNameUID}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(PoolOperator, setCachefileStr...)
	if err != nil {
		glog.Errorf("Unable to set cachefile: %v", string(stdoutStderr))
		return err
	}
	return nil
}

// CheckForZreplInitial is blocking call for checking status of zrepl in cstor-pool container.
func CheckForZreplInitial(ZreplRetryInterval time.Duration) {
	for {
		_, err := RunnerVar.RunCombinedOutput(PoolOperator, "status")
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
		out, err := RunnerVar.RunCombinedOutput(PoolOperator, "status")
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

// LabelClear is to clear zpool label on disks.
func LabelClear(disks []string) error {
	var failLabelClear = false
	for _, disk := range disks {
		labelClearStr := []string{"labelclear", "-f", disk}
		stdoutStderr, err := RunnerVar.RunCombinedOutput(PoolOperator, labelClearStr...)
		if err != nil {
			glog.Errorf("Unable to clear label: %v, err = %v", string(stdoutStderr), err)
			failLabelClear = true
		}
	}
	if failLabelClear {
		return fmt.Errorf("Unable to clear labels from the disks of the pool")
	}
	return nil
}
