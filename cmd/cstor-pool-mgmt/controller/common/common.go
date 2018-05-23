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
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CStor controllers to be watched
const (
	CStorPool          = "cStorPool"
	CStorVolumeReplica = "cStorVolumeReplica"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a resource is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a resource fails
	// to sync due to a resource of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events which
	// fails to sync due to a resource already existing
	MessageResourceExists = "Resource %q already exists and cannot be handled"
	// MessageResourceSynced is the message used for an Event fired when a resource
	// is synced successfully
	MessageResourceSynced = "Resource synced successfully"
)

// Status written onto CStorPool and CStorVolumeReplica objects.
const (
	StatusInit           = "init"
	StatusOnline         = "online"
	StatusOffline        = "offline"
	StatusDeletionFailed = "deletion-failed"
	StatusInvalid        = "invalid"

	StatusIgnore = "ignore"
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
func PoolNameHandler(cnt int) (string, error) {
	for i := 0; ; i++ {
		poolname, err := pool.GetPoolName()
		if err != nil {
			glog.Infof("Attempt %v: Waiting for Pool..", i)
			time.Sleep(5 * time.Second)
			if i > cnt {
				return "", err
			}
		} else {
			return poolname, nil
		}
	}
}

// CheckForCStorPoolCRD is Blocking call for checking status of CStorPool CRD.
func CheckForCStorPoolCRD(clientset clientset.Interface) {
	for {
		_, err := clientset.OpenebsV1alpha1().CStorPools().List(metav1.ListOptions{})
		if err != nil {
			glog.Errorf("CStorPool CRD not found...")
			time.Sleep(10 * time.Second)
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
			glog.Errorf("CStorVolumeReplica CRD not found...")
			time.Sleep(10 * time.Second)
			continue
		}
		glog.Info("CStorVolumeReplica CRD found")
		break
	}
}
