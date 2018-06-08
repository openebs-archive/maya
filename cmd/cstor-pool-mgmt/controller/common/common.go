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

package common

import (
	"reflect"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a resource is synced
	SuccessSynced = "Synced"
	// MessageCreateSynced holds message for corresponding create request sync.
	MessageCreateSynced = "Received Resource create event"
	// MessageModifySynced holds message for corresponding modify request sync.
	MessageModifySynced = "Received Resource modify event"
	// MessageDestroySynced holds message for corresponding destroy request sync.
	MessageDestroySynced = "Received Resource destroy event"

	// SuccessCreated holds status for corresponding created resource.
	SuccessCreated = "Created"
	// MessageResourceCreated holds message for corresponding created resource.
	MessageResourceCreated = "Resource created successfully"

	// FailureCreate holds status for corresponding failed create resource.
	FailureCreate = "FailCreate"
	// MessageResourceFailCreate holds message for corresponding failed create resource.
	MessageResourceFailCreate = "Resource creation failed"

	// SuccessImported holds status for corresponding imported resource.
	SuccessImported = "Imported"
	// MessageResourceImported holds message for corresponding imported resource.
	MessageResourceImported = "Resource imported successfully"

	// FailureImport holds status for corresponding failed import resource.
	FailureImport = "FailImport"
	// MessageResourceFailImport holds message for corresponding failed import resource.
	MessageResourceFailImport = "Resource import failed"

	// FailureDestroy holds status for corresponding failed destroy resource.
	FailureDestroy = "FailDestroy"
	// MessageResourceFailDestroy holds message for corresponding failed destroy resource.
	MessageResourceFailDestroy = "Resource Destroy failed"

	// FailureValidate holds status for corresponding failed validate resource.
	FailureValidate = "FailValidate"
	// MessageResourceFailValidate holds message for corresponding failed validate resource.
	MessageResourceFailValidate = "Resource validation failed"

	// AlreadyPresent holds status for corresponding already present resource.
	AlreadyPresent = "AlreadyPresent"
	// MessageResourceAlreadyPresent holds message for corresponding already present resource.
	MessageResourceAlreadyPresent = "Resource already present"
)

// Periodic interval duration.
const (
	CRDRetryInterval        = 10 * time.Second
	PoolNameHandlerInterval = 5 * time.Second
	SharedInformerInterval  = 5 * time.Minute
	ResourceWorkerInterval  = time.Second
)

// InitialImportedPoolVol is to store pool-volume names while pod restart.
var InitialImportedPoolVol []string

// QueueLoad is for storing the key and type of operation before entering workqueue
type QueueLoad struct {
	Key       string
	Operation string
}

// PoolNameHandler tries to get pool name and blocks for
// particular number of attempts.
func PoolNameHandler(cVR *apis.CStorVolumeReplica, cnt int) bool {
	for i := 0; ; i++ {
		poolname, _ := pool.GetPoolName()
		if reflect.DeepEqual(poolname, []string{}) || !CheckIfPresent(poolname, "cstor-"+cVR.Labels["cstorpool.openebs.io/uid"]) {
			glog.Infof("Attempt %v: No pool found", i)
			time.Sleep(PoolNameHandlerInterval)
			if i > cnt {
				return false
			}
		} else if CheckIfPresent(poolname, "cstor-"+cVR.Labels["cstorpool.openebs.io/uid"]) {
			return true
		}
	}
}

// CheckForCStorPoolCRD is Blocking call for checking status of CStorPool CRD.
func CheckForCStorPoolCRD(clientset clientset.Interface) {
	for {
		_, err := clientset.OpenebsV1alpha1().CStorPools().List(metav1.ListOptions{})
		if err != nil {
			glog.Errorf("CStorPool CRD not found. Retrying after %v", CRDRetryInterval)
			time.Sleep(CRDRetryInterval)
			continue
		}
		glog.Info("CStorPool CRD found")
		break
	}
}

// CheckForCStorVolumeReplicaCRD is Blocking call for checking status of CStorVolumeReplica CRD.
func CheckForCStorVolumeReplicaCRD(clientset clientset.Interface) {
	for {
		_, err := clientset.OpenebsV1alpha1().CStorVolumeReplicas().List(metav1.ListOptions{})
		if err != nil {
			glog.Errorf("CStorVolumeReplica CRD not found. Retrying after %v", CRDRetryInterval)
			time.Sleep(CRDRetryInterval)
			continue
		}
		glog.Info("CStorVolumeReplica CRD found")
		break
	}
}

// CheckForInitialImportedPoolVol is to check if volume is already
// imported with pool.
func CheckForInitialImportedPoolVol(InitialImportedPoolVol []string, fullvolname string) bool {
	for i, initialVol := range InitialImportedPoolVol {
		if initialVol == fullvolname {
			if i < len(InitialImportedPoolVol) {
				InitialImportedPoolVol = append(InitialImportedPoolVol[:i], InitialImportedPoolVol[i+1:]...)
			}
			return true
		}
	}
	return false
}

// CheckIfPresent is to check if search string is present in array of string.
func CheckIfPresent(arrStr []string, searchStr string) bool {
	for _, str := range arrStr {
		if str == searchStr {
			return true
		}
	}
	return false
}
