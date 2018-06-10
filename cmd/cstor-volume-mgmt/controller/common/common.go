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
	"time"

	"github.com/golang/glog"
	clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CStor controllers to be watched
const (
	CStorVolume = "cStorVolume"
)

//EventReason is used as part of the Event reason when a resource goes through different phases
type EventReason string

const (
	// SuccessSyncedER is used as part of the Event 'reason' when a resource is synced
	SuccessSyncedER EventReason = "Synced"
	// ErrResourceExistsER is used as part of the Event 'reason' when a resource fails
	// to sync due to a resource of the same name already existing.
	ErrResourceExistsER EventReason = "ErrResourceExists"

	// MessageResourceExistsER is the message used for Events which
	// fails to sync due to a resource already existing
	MessageResourceExistsER EventReason = "Resource %q already exists and cannot be handled"
	// MessageResourceSyncedER is the message used for an Event fired when a resource
	// is synced successfully
	MessageResourceSyncedER EventReason = "Resource synced successfully"
)

//CStorVolumeStatus represents the status of a CStorVolume object
type CStorVolumeStatus string

// Status written onto CStorVolume objects.
const (
	CVStatusInit           CStorVolumeStatus = "init"
	CVStatusOnline         CStorVolumeStatus = "online"
	CVStatusOffline        CStorVolumeStatus = "offline"
	CVStatusDeletionFailed CStorVolumeStatus = "deletion-failed"
	CVStatusInvalid        CStorVolumeStatus = "invalid"
	CVStatusFailed         CStorVolumeStatus = "failed"

	CVStatusIgnore CStorVolumeStatus = "ignore"
)

//QueueOperation represents the type of operation on the controller work queue
type QueueOperation string

//Different type of operations on the controller work queue
const (
	QOpAdd     QueueOperation = "add"
	QOpDestroy QueueOperation = "destroy"
	QOpModify  QueueOperation = "modify"
)

// QueueLoad is for storing the key and type of operation before entering workqueue
type QueueLoad struct {
	Key       string // Key is the name of cstor volume given in metadata name field in the yaml
	Operation QueueOperation
}

// CheckForCStorVolumeCR is Blocking call for checking status of CStorVolume CR.
func CheckForCStorVolumeCR(clientset clientset.Interface) {
	for {
		_, err := clientset.OpenebsV1alpha1().CStorVolumes().List(metav1.ListOptions{})
		if err != nil {
			glog.Errorf("CStorVolume CR not found...")
			time.Sleep(10 * time.Second)
			continue
		}
		glog.Info("CStorVolume CR found")
		break
	}
}
