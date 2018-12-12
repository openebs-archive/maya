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

package backupcontroller

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the CStorReplicaUpdated resource
// with the current status of the resource.
func (c *CStorBackupController) syncHandler(key string, operation common.QueueOperation) error {
	csbGot, err := c.getCStorBackupResource(key)
	if err != nil {
		return err
	}
	if csbGot == nil {
		return fmt.Errorf("cannot retrieve CStorBackup %q", key)
	}
	status, err := c.csbEventHandler(operation, csbGot)
	if status == "" {
		return nil
	}
	if err != nil {
		glog.Errorf(err.Error())
		glog.Infof("csb:%v, %v; Status: %v", csbGot.Name,
			string(csbGot.GetUID()), csbGot.Status.Phase)
		_, err := c.clientset.OpenebsV1alpha1().CStorBackups(csbGot.Namespace).Update(csbGot)
		if err != nil {
			return err
		}
		return err
	}
	_, err = c.clientset.OpenebsV1alpha1().CStorBackups(csbGot.Namespace).Update(csbGot)
	if err != nil {
		return err
	}
	glog.Infof("cVR:%v, %v; Status: %v", csbGot.Name,
		string(csbGot.GetUID()), csbGot.Status.Phase)
	return nil

}

func (c *CStorBackupController) csbEventHandler(operation common.QueueOperation, csbGot *apis.CStorBackup) (string, error) {

	switch operation {
	case common.QOpAdd:
		glog.Infof("Processing csb added event: %v, %v", csbGot.ObjectMeta.Name, string(csbGot.GetUID()))
		status, err := c.csbAddEventHandler(csbGot)
		return status, err

	case common.QOpDestroy:
		glog.Infof("Processing csb deleted event %v, %v", csbGot.ObjectMeta.Name, string(csbGot.GetUID()))
		return "", nil
	case common.QOpSync:
		glog.Infof("Synchronizing CstorBackup status for volume %s", csbGot.ObjectMeta.Name)
		return "", nil
	}
	return string(apis.CSBStatusInvalid), nil
}

func (c *CStorBackupController) csbAddEventHandler(csb *apis.CStorBackup) (string, error) {
	// IsEmptyStatus is to check if initial status of cVR object is empty.
	if IsEmptyStatus(csb) || IsPendingStatus(csb) {
		c.recorder.Event(csb, corev1.EventTypeNormal, string(common.SuccessCreated), string(common.MessageResourceCreated))
		glog.Infof("csb creation successful: %v, %v", csb.ObjectMeta.Name, string(csb.GetUID()))
		return string(apis.CSBStatusOnline), nil
	}
	//TODO: Buisness logic for csb add event
	return "", nil
}

// IsEmptyStatus is to check if the status of cStorVolumeReplica object is empty.
func IsEmptyStatus(csb *apis.CStorBackup) bool {
	if string(csb.Status.Phase) == string(apis.CSBStatusEmpty) {
		glog.Infof("csb empty status: %v", string(csb.ObjectMeta.UID))
		return true
	}
	glog.Infof("Not empty status: %v", string(csb.ObjectMeta.UID))
	return false
}

// IsPendingStatus is to check if the status of cStorVolumeReplica object is pending.
func IsPendingStatus(csb *apis.CStorBackup) bool {
	if string(csb.Status.Phase) == string(apis.CSBStatusPending) {
		glog.Infof("csb pending: %v", string(csb.ObjectMeta.UID))
		return true
	}
	glog.V(4).Infof("Not pending status: %v", string(csb.ObjectMeta.UID))
	return false
}

// getVolumeReplicaResource returns object corresponding to the resource key
func (c *CStorBackupController) getCStorBackupResource(key string) (*apis.CStorBackup, error) {
	// Convert the key(namespace/name) string into a distinct name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil, nil
	}

	CStorBackupUpdated, err := c.clientset.OpenebsV1alpha1().CStorBackups(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorPool resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("CStorBackupUpdated '%s' in work queue no longer exists", key))
			return nil, nil
		}
		return nil, err
	}
	return CStorBackupUpdated, nil
}

// IsRightCStorPoolMgmt is to check if the pool request is for particular pod/application.
func IsRightCStorPoolMgmt(cStorPool *apis.CStorBackup) bool {
	if os.Getenv(string(common.OpenEBSIOCStorID)) == string(cStorPool.ObjectMeta.UID) {
		return true
	}
	return false
}

// IsDeletionFailedBefore is to make sure no other operation should happen if the
// status of CStorVolumeReplica is deletion-failed.
func IsDeletionFailedBefore(csb *apis.CStorBackup) bool {
	if csb.Status.Phase == apis.CSBStatusDeletionFailed {
		return true
	}
	return false
}

// IsErrorDuplicate is to check if the status of cStorVolumeReplica object is error-duplicate.
func IsErrorDuplicate(csb *apis.CStorBackup) bool {
	if string(csb.Status.Phase) == string(apis.CSBStatusErrorDuplicate) {
		glog.Infof("csb duplication error: %v", string(csb.ObjectMeta.UID))
		return true
	}
	glog.V(4).Infof("Not error duplicate status: %v", string(csb.ObjectMeta.UID))
	return false
}
