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
	"time"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
)

// PoolOperator is the name of the tool that makes pool-related operations.
const (
	PoolOperator = "zpool"
)

var RunnerVar util.Runner

// ImportPool imports cStor pool if already present.
func ImportPool(cStorPool *apis.CStorPool) error {
	importAttr := importPoolBuilder(cStorPool)

	stdoutStderr, err := RunnerVar.RunCombinedOutput(PoolOperator, importAttr...)
	if err != nil {
		glog.Error("Unable to import pool:", err.Error(), string(stdoutStderr))
		return err
	}

	glog.Info("Importing Pool Successful")
	return nil
}

// importPoolBuilder is to build pool import command.
func importPoolBuilder(cStorPool *apis.CStorPool) []string {
	// populate pool import attributes.
	var importAttr []string
	importAttr = append(importAttr, "import")
	if cStorPool.Spec.PoolSpec.CacheFile != "" {
		importAttr = append(importAttr, "-c", cStorPool.Spec.PoolSpec.CacheFile,
			"-o", cStorPool.Spec.PoolSpec.CacheFile)
	}

	importAttr = append(importAttr, "cstor-"+string(cStorPool.ObjectMeta.UID))

	return importAttr
}

// CreatePool creates a new cStor pool.
func CreatePool(cStorPool *apis.CStorPool) error {
	createAttr := createPoolBuilder(cStorPool)
	glog.V(4).Info("createAttr : ", createAttr)

	stdoutStderr, err := RunnerVar.RunCombinedOutput(PoolOperator, createAttr...)
	if err != nil {
		glog.Error("Unable to create pool:", err.Error(), string(stdoutStderr))
		return err
	}

	glog.Info("Creating Pool Successful")
	return nil
}

// createPoolBuilder is to build create pool command.
func createPoolBuilder(cStorPool *apis.CStorPool) []string {
	// populate pool creation attributes.
	var createAttr []string
	createAttr = append(createAttr, "create")
	if cStorPool.Spec.PoolSpec.CacheFile != "" {
		cachefile := "cachefile=" + cStorPool.Spec.PoolSpec.CacheFile
		createAttr = append(createAttr, "-o", cachefile)
	}

	openebsPoolname := "io.openebs:poolname=" + cStorPool.Name
	createAttr = append(createAttr, "-O", openebsPoolname)

	poolNameUID := "cstor-" + string(cStorPool.ObjectMeta.UID)
	createAttr = append(createAttr, poolNameUID)

	// To generate mirror disk0 disk1 mirror disk2 disk3 format.
	for i, disk := range cStorPool.Spec.Disks.DiskList {
		if cStorPool.Spec.PoolSpec.PoolType == "mirror" && i%3 == 0 {
			createAttr = append(createAttr, "mirror")
		}
		createAttr = append(createAttr, disk)
	}

	return createAttr

}

// CheckValidPool checks for validity of CStorPool resource.
func CheckValidPool(cStorPool *apis.CStorPool) error {
	if string(cStorPool.ObjectMeta.UID) == "" {
		return fmt.Errorf("Poolname cannot be empty")
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
func GetPoolName() (string, error) {
	GetPoolStr := []string{"get", "-Hp", "name", "-o", "name"}
	poolNameByte, err := RunnerVar.RunStdoutPipe(PoolOperator, GetPoolStr...)
	if err != nil {
		glog.Errorf("Unable to get pool:", poolNameByte)
	}
	poolName := string(poolNameByte)
	glog.Infof("poolname : ", poolName)
	return poolName, nil
}

// DeletePool destroys the pool created.
func DeletePool(poolName string) error {
	deletePoolStr := []string{"destroy", poolName}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(PoolOperator, deletePoolStr...)
	if err != nil {
		glog.Errorf("Unable to delete pool:", err.Error(), string(stdoutStderr))
		return err
	}
	return nil
}

// CheckForZrepl is blocking call for checking status of zrepl in cstor-pool container.
func CheckForZrepl() {
	for {
		_, err := RunnerVar.RunCombinedOutput(PoolOperator, "status")
		if err != nil {
			time.Sleep(3 * time.Second)
			glog.Infof("Waiting for zrepl...")
			continue
		}
		break
	}
}

// LabelClear is to clear zpool label on disks.
func LabelClear(disks []string) error {
	for _, disk := range disks {
		labelClearStr := []string{"labelclear", "-f", disk}
		stdoutStderr, err := RunnerVar.RunCombinedOutput(PoolOperator, labelClearStr...)
		if err != nil {
			glog.Errorf("Unable to clear label", err.Error(), string(stdoutStderr))
			return err
		}
	}
	return nil
}
