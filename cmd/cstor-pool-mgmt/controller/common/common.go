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
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/volumereplica"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	"github.com/openebs/maya/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//EventReason is used as part of the Event reason when a resource goes through different phases
type EventReason string

const (
	// ToDo: Improve the messages and event reason. ( Put these in a similar k8s style)
	// SuccessSynced is used as part of the Event 'reason' when a resource is synced.
	SuccessSynced EventReason = "Synced"
	// FailedSynced is used as part of the Event 'reason' when resource sync fails.
	FailedSynced EventReason = "FailedSync"
	// MessageCreateSynced holds message for corresponding create request sync.
	MessageCreateSynced EventReason = "Received Resource create event"
	// MessageModifySynced holds message for corresponding modify request sync.
	MessageModifySynced EventReason = "Received Resource modify event"
	// MessageDestroySynced holds message for corresponding destroy request sync.
	MessageDestroySynced EventReason = "Received Resource destroy event"
	// StatusSynced holds message for corresponding status request sync.
	StatusSynced EventReason = "Resource status sync event"
	// SuccessCreated holds status for corresponding created resource.
	SuccessCreated EventReason = "Created"
	// MessageResourceCreated holds message for corresponding created resource.
	MessageResourceCreated EventReason = "Resource created successfully"
	// FailureCreate holds status for corresponding failed create resource.
	FailureCreate EventReason = "FailCreate"
	// MessageResourceFailCreate holds message for corresponding failed create resource.
	MessageResourceFailCreate EventReason = "Resource creation failed"
	// SuccessImported holds status for corresponding imported resource.
	SuccessImported EventReason = "Imported"
	// FailureImported holds status for corresponding imported resource.
	FailureImported EventReason = "Import failure"
	// FailureImportOperations holds status for corresponding imported resource.
	FailureImportOperations EventReason = "Failure Import operations"
	// MessageResourceImported holds message for corresponding imported resource.
	MessageResourceImported EventReason = "Resource imported successfully"
	// FailureStatusSync holds status for corresponding failed status sync of resource.
	FailureStatusSync EventReason = "FailStatusSync"
	// FailureCapacitySync holds status for corresponding failed capacity sync of resource.
	FailureCapacitySync EventReason = "FailCapacitySync"
	// MessageResourceFailStatusSync holds message for corresponding failed status sync of resource.
	MessageResourceFailStatusSync EventReason = "Resource status sync failed"
	// MessageResourceFailCapacitySync holds message for corresponding failed capacity sync of resource.
	MessageResourceFailCapacitySync EventReason = "Resource capacity sync failed"
	// MessageResourceSyncSuccess holds message for corresponding successful sync of resource.
	MessageResourceSyncSuccess EventReason = "Resource successfully synced"
	// MessageResourceSyncFailure holds message for corresponding failed sync of resource.
	MessageResourceSyncFailure EventReason = "Resource sync failed:"
	// FailureDestroy holds status for corresponding failed destroy resource.
	FailureDestroy EventReason = "FailDestroy"
	// MessageResourceFailDestroy holds message for corresponding failed destroy resource.
	MessageResourceFailDestroy EventReason = "Resource Destroy failed"
	// FailureValidate holds status for corresponding failed validate resource.
	FailureValidate EventReason = "FailValidate"
	// MessageResourceFailValidate holds message for corresponding failed validate resource.
	MessageResourceFailValidate EventReason = "Resource validation failed"
	// AlreadyPresent holds status for corresponding already present resource.
	AlreadyPresent EventReason = "AlreadyPresent"
	// MessageResourceAlreadyPresent holds message for corresponding already present resource.
	MessageResourceAlreadyPresent EventReason = "Resource already present"
	// MessageImproperPoolStatus holds message for corresponding failed validate resource.
	MessageImproperPoolStatus EventReason = "Improper pool status"
)

// Periodic interval duration.
const (
	// CRDRetryInterval is used if CRD is not present.
	CRDRetryInterval = 10 * time.Second
	// PoolNameHandlerInterval is used when expected pool is not present.
	PoolNameHandlerInterval = 5 * time.Second
	// SharedInformerInterval is used to sync watcher controller.
	SharedInformerInterval = 30 * time.Second
	// ResourceWorkerInterval is used for resource sync.
	ResourceWorkerInterval = time.Second
	// InitialZreplRetryInterval is used while initially starting controller.
	InitialZreplRetryInterval = 3 * time.Second
	// ContinuousZreplRetryInterval is used while controller has started running.
	ContinuousZreplRetryInterval = 1 * time.Second
)

const (
	// NoOfPoolWaitAttempts is number of attempts to wait in case of pod/container restarts.
	NoOfPoolWaitAttempts = 30
	// PoolWaitInterval is the interval to wait for pod/container restarts.
	PoolWaitInterval = 2 * time.Second
)

// InitialImportedPoolVol is to store pool-volume names while pod restart.
var InitialImportedPoolVol []string

// QueueLoad is for storing the key and type of operation before entering workqueue
type QueueLoad struct {
	Key       string
	Operation QueueOperation
}

// Environment is for environment variables passed for cstor-pool-mgmt.
type Environment string

const (
	// OpenEBSIOCStorID is the environment variable specified in pod.
	OpenEBSIOCStorID Environment = "OPENEBS_IO_CSTOR_ID"
)

//QueueOperation represents the type of operation on resource
type QueueOperation string

//Different type of operations on the controller
const (
	QOpAdd     QueueOperation = "add"
	QOpDestroy QueueOperation = "destroy"
	QOpModify  QueueOperation = "modify"
	// QOpSync is the operation for syncing(reconciling) on cstor pool object.
	QOpSync QueueOperation = "Sync"
)

// namespace defines kubernetes namespace specified for cvr.
type namespace string

// Different types of k8s namespaces.
const (
	defaultNameSpace namespace = "default"
)

// SyncResources is to synchronize pool and volumereplica.
var SyncResources SyncCStorPoolCVR

// SyncCStorPoolCVR is to hold synchronization related variables.
type SyncCStorPoolCVR struct {
	// Mux is mutex variable to block cvr until certain pool operations are complete.
	Mux *sync.Mutex

	// IsImported is boolean flag to check at cvr until certain pool import operations are complete.
	IsImported bool
}

// PoolNameHandler tries to get pool name and blocks for
// particular number of attempts.
func PoolNameHandler(cVR *apis.CStorVolumeReplica, cnt int) bool {
	for i := 0; ; i++ {
		poolname, _ := pool.GetPoolName()
		if reflect.DeepEqual(poolname, []string{}) ||
			!CheckIfPresent(poolname, string(pool.PoolPrefix)+cVR.Labels["cstorpool.openebs.io/uid"]) {
			glog.Warningf("Attempt %v: No pool found", i+1)
			time.Sleep(PoolNameHandlerInterval)
			if i > cnt {
				return false
			}
		} else if CheckIfPresent(poolname, string(pool.PoolPrefix)+cVR.Labels["cstorpool.openebs.io/uid"]) {
			return true
		}
	}
}

// CheckForCStorPoolCRD is Blocking call for checking status of CStorPool CRD.
func CheckForCStorPoolCRD(clientset clientset.Interface) {
	for {
		_, err := clientset.OpenebsV1alpha1().CStorPools().List(metav1.ListOptions{})
		if err != nil {
			glog.Errorf("CStorPool CRD not found. Retrying after %v, error: %v", CRDRetryInterval, err)
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
		// Since this blocking function is restricted to check if CVR CRD is present
		// or not, we are trying to handle only the error of CVR CR List api indirectly.
		// CRD has only two types of scope, cluster and namespaced. If CR list api
		// for default namespace works fine, then CR list api works for all namespaces.
		_, err := clientset.OpenebsV1alpha1().CStorVolumeReplicas(string(defaultNameSpace)).List(metav1.ListOptions{})
		if err != nil {
			glog.Errorf("CStorVolumeReplica CRD not found. Retrying after %v, error: %v", CRDRetryInterval, err)
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

// CheckForCStorPool tries to get pool name and blocks forever because
// volumereplica can be created only if pool is present.
func CheckForCStorPool() {
	for {
		poolname, _ := pool.GetPoolName()
		if reflect.DeepEqual(poolname, []string{}) {
			glog.Warningf("CStorPool not found. Retrying after %v", PoolNameHandlerInterval)
			time.Sleep(PoolNameHandlerInterval)
			continue
		}
		glog.Info("CStorPool found")
		break
	}
}

// Init is to instantiate variable used between pool and volumereplica while
// starting controller.
func Init() {
	// Instantiate mutex variable.
	SyncResources.Mux = &sync.Mutex{}

	// Making RunnerVar to use RealRunner
	pool.RunnerVar = util.RealRunner{}
	volumereplica.RunnerVar = util.RealRunner{}
}
