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
	"context"
	"time"

	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// DefaultSharedInformerInterval is used to sync watcher controller.
const DefaultSharedInformerInterval = 30 * time.Second

const (
	// CStorVolume is the controller to be watched
	CStorVolume = "cStorVolume"
	// EventMsgFormatter is the format string for event message generation
	EventMsgFormatter = "Volume is in %s state"
)

// EventReason is used as part of the Event reason when a resource goes through different phases
type EventReason string

const (
	// SuccessSynced is used as part of the Event 'reason' when a resource is synced
	SuccessSynced EventReason = "Synced"
	// MessageCreateSynced holds message for corresponding create request sync.
	MessageCreateSynced EventReason = "Received Resource create event"
	// MessageModifySynced holds message for corresponding modify request sync.
	MessageModifySynced EventReason = "Received Resource modify event"
	// MessageDestroySynced holds message for corresponding destroy request sync.
	MessageDestroySynced EventReason = "Received Resource destroy event"

	// SuccessCreated holds status for corresponding created resource.
	SuccessCreated EventReason = "Created"
	// MessageResourceCreated holds message for corresponding created resource.
	MessageResourceCreated EventReason = "Resource created successfully"

	// FailureCreate holds status for corresponding failed create resource.
	FailureCreate EventReason = "FailCreate"
	// MessageResourceFailCreate holds message for corresponding failed create resource.
	MessageResourceFailCreate EventReason = "Resource creation failed"

	// FailureUpdate holds status for corresponding failed update resource.
	FailureUpdate EventReason = "FailUpdate"

	// SuccessImported holds status for corresponding imported resource.
	SuccessImported EventReason = "Imported"
	// MessageResourceImported holds message for corresponding imported resource.
	MessageResourceImported EventReason = "Resource imported successfully"

	// FailureImport holds status for corresponding failed import resource.
	FailureImport EventReason = "FailImport"
	// MessageResourceFailImport holds message for corresponding failed import resource.
	MessageResourceFailImport EventReason = "Resource import failed"

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

	// SuccessUpdated holds status for corresponding updated resource.
	SuccessUpdated EventReason = "Updated"
	// MessageResourceUpdated holds message for corresponding updated resource.
	MessageResourceUpdated EventReason = "Resource updated successfully"
)

const (
	// CRDRetryInterval is used if CRD is not present.
	CRDRetryInterval = 10 * time.Second
	// ResourceWorkerInterval is used for resource sync.
	ResourceWorkerInterval = time.Second
)

// CStorVolumeStatus represents the status of a CStorVolume object
type CStorVolumeStatus string

// Status written onto CStorVolume objects.
const (
	// volume is getting initialized
	CVStatusInit CStorVolumeStatus = "Init"
	// volume allows IOs and snapshot
	CVStatusHealthy CStorVolumeStatus = "Healthy"
	// volume only satisfies consistency factor
	CVStatusDegraded CStorVolumeStatus = "Degraded"
	// Volume is offline
	CVStatusOffline CStorVolumeStatus = "Offline"
	// Error in retrieving volume details
	CVStatusError CStorVolumeStatus = "Error"
	// volume controller config generation failed due to invalid parameters
	CVStatusInvalid CStorVolumeStatus = "Invalid"
	// CR event ignored
	CVStatusIgnore CStorVolumeStatus = "Ignore"
)

// QueueLoad is for storing the key and type of operation before entering workqueue
type QueueLoad struct {
	Key       string // Key is the name of cstor volume given in metadata name field in the yaml
	Operation QueueOperation
}

// Environment is for environment variables passed for cstor-volume-mgmt.
type Environment string

const (
	// OpenEBSIOCStorVolumeID is the environment variable specified in pod.
	OpenEBSIOCStorVolumeID Environment = "OPENEBS_IO_CSTOR_VOLUME_ID"
)

// QueueOperation represents the type of operation on resource
type QueueOperation string

// Different type of operations on the controller
const (
	QOpAdd          QueueOperation = "add"
	QOpDestroy      QueueOperation = "destroy"
	QOpModify       QueueOperation = "modify"
	QOpPeriodicSync QueueOperation = "sync"
)

// namespace defines kubernetes namespace specified for cvr.
type namespace string

// Different types of k8s namespaces.
const (
	DefaultNameSpace namespace = "openebs"
)

// CheckForCStorVolumeCRD is Blocking call for checking status of CStorVolume CRD.
func CheckForCStorVolumeCRD(clientset clientset.Interface) {
	for {
		// Since this blocking function is restricted to check if CVR CRD is present
		// or not, we are trying to handle only the error of CVR CR List api indirectly.
		// CRD has only two types of scope, cluster and namespaced. If CR list api
		// for default namespace works fine, then CR list api works for all namespaces.
		_, err := clientset.OpenebsV1alpha1().CStorVolumes(string(DefaultNameSpace)).
			List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Errorf("CStorVolume CRD not found. Retrying after %v, err : %v", CRDRetryInterval, err)
			time.Sleep(CRDRetryInterval)
			continue
		}
		klog.Info("CStorVolume CRD found")
		break
	}
}
