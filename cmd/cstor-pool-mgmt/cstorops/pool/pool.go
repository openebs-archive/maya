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
	"os/exec"
	"strings"
	"time"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// PoolOperator is the name of the tool that makes pool-related operations.
const (
	PoolOperator = "zpool"
)

// ImportPool imports cStor pool if already present.
func ImportPool(cStorPoolUpdated *apis.CStorPool) error {
	cmdimport := importPoolBuilder(cStorPoolUpdated)
	stdoutStderrImport, err := cmdimport.CombinedOutput()
	if err != nil {
		glog.Error("Pool import err: ", err)
		glog.Error("stdoutStderr: ", string(stdoutStderrImport))
		return err
	}

	glog.Info("Importing Pool Successful")
	return nil
}

// importPoolBuilder is to build pool import command.
func importPoolBuilder(cStorPoolUpdated *apis.CStorPool) *exec.Cmd {
	// populate pool import attributes.
	var importAttr []string
	importAttr = append(importAttr, "import")
	if cStorPoolUpdated.Spec.PoolSpec.CacheFile != "" {
		importAttr = append(importAttr, "-c", cStorPoolUpdated.Spec.PoolSpec.CacheFile)
	}

	importAttr = append(importAttr, cStorPoolUpdated.Spec.PoolSpec.PoolName)

	// execute import pool command.
	cmdimport := exec.Command(PoolOperator, importAttr...)
	return cmdimport
}

// CreatePool creates a new cStor pool.
func CreatePool(cStorPoolUpdated *apis.CStorPool) error {

	poolCreateCmd := createPoolBuilder(cStorPoolUpdated)

	glog.V(4).Info("poolCreateCmd : ", poolCreateCmd)
	stdoutStderr, err := poolCreateCmd.CombinedOutput()
	if err != nil {
		glog.Error("err: ", err)
		glog.Error("stdoutStderr: ", string(stdoutStderr))
		return err
	}
	glog.Info("Creating Pool Successful")
	return nil
}

// createPoolBuilder is to build create pool command.
func createPoolBuilder(cStorPoolUpdated *apis.CStorPool) *exec.Cmd {
	// populate pool creation attributes.
	var createAttr []string
	createAttr = append(createAttr, "create", "-f", "-o")
	if cStorPoolUpdated.Spec.PoolSpec.CacheFile != "" {
		cachefile := "cachefile=" + cStorPoolUpdated.Spec.PoolSpec.CacheFile
		createAttr = append(createAttr, cachefile)
	}

	createAttr = append(createAttr, cStorPoolUpdated.Spec.PoolSpec.PoolName)

	for _, disk := range cStorPoolUpdated.Spec.Disks.DiskList {
		createAttr = append(createAttr, disk)
	}

	//execute pool creation command.
	poolCreateCmd := exec.Command(PoolOperator, createAttr...)
	return poolCreateCmd
}

// CheckValidPool checks for validity of CStorPool resource.
func CheckValidPool(cStorPoolUpdated *apis.CStorPool) error {
	if cStorPoolUpdated.Spec.PoolSpec.PoolName == "" {
		return fmt.Errorf("Poolname cannot be empty")
	}
	if len(cStorPoolUpdated.Spec.Disks.DiskList) < 1 {
		return fmt.Errorf("Disk name(s) cannot be empty")
	}
	return nil
}

// GetPoolName return the pool already created.
func GetPoolName() (string, error) {
	poolnameStr := PoolOperator + " status | grep pool:"
	poolnamecmd := exec.Command("bash", "-c", poolnameStr)
	stderr, err := poolnamecmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Unable to get poolname :%v ", err)
	}
	noisyPoolname := string(stderr)
	poolname := strings.TrimPrefix(noisyPoolname, "  pool: ")
	poolname = strings.TrimSpace(poolname)
	glog.V(4).Infof("poolname : ", poolname)
	return poolname, nil
}

// DeletePool destroys the pool created.
func DeletePool(poolName string) error {
	deletePoolStr := PoolOperator + " destroy -f " + poolName
	deletePoolCmd := exec.Command("bash", "-c", deletePoolStr)
	stdoutStderr, err := deletePoolCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Unable to delete pool :%v ", err.Error(), string(stdoutStderr))
	}
	return nil
}

// CheckForZrepl is blocking call for checking status of zrepl in cstor-pool container.
func CheckForZrepl() {
	for {
		statuscmd := exec.Command(PoolOperator, "status")
		_, err := statuscmd.CombinedOutput()
		if err != nil {
			time.Sleep(3 * time.Second)
			glog.Infof("Waiting for zrepl...")
			continue
		}
		break
	}
}
