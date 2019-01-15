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

package restorecontroller

import (
	"fmt"
	"os"
	"reflect"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/volumereplica"
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
func (c *CStorRestoreController) syncHandler(key string, operation common.QueueOperation) error {
	rstGot, err := c.getCStorRestoreResource(key)
	if err != nil {
		return err
	}
	if rstGot == nil {
		return fmt.Errorf("cannot retrieve CStorRestore %q", key)
	}
	status, err := c.rstEventHandler(operation, rstGot)
	if status == "" {
		return nil
	}
	rstGot.Status.Phase = apis.CStorRestorePhase(status)
	if err != nil {
		glog.Errorf(err.Error())
		glog.Infof("rst:%v, %v; Status: %v", rstGot.Name,
			string(rstGot.GetUID()), rstGot.Status.Phase)
		_, err := c.clientset.OpenebsV1alpha1().CStorRestores(rstGot.Namespace).Update(rstGot)
		if err != nil {
			return err
		}
		return err
	}
	_, err = c.clientset.OpenebsV1alpha1().CStorRestores(rstGot.Namespace).Update(rstGot)
	if err != nil {
		return err
	}
	glog.Infof("RST:%v, %v; Status: %v", rstGot.Name,
		string(rstGot.GetUID()), rstGot.Status.Phase)
	return nil

}

func (c *CStorRestoreController) rstEventHandler(operation common.QueueOperation, rstGot *apis.CStorRestore) (string, error) {

	switch operation {
	case common.QOpAdd:
		status, err := c.rstAddEventHandler(rstGot)
		return status, err
	case common.QOpDestroy:
		/*
			status, err := c.rstDestroyEventHandler(rstGot)
			return status, err
			glog.Infof("Processing rst deleted event %v, %v", rstGot.ObjectMeta.Name, string(rstGot.GetUID()))
		*/
		return "", nil
	case common.QOpSync:
		return "", nil
	case common.QOpModify:
		return "", nil
		//status, err := c.rstSyncEventHandler(rstGot)
		//return status, err
	}
	return string(apis.RSTStatusInvalid), nil
}

func (c *CStorRestoreController) rstAddEventHandler(rst *apis.CStorRestore) (string, error) {
	// IsEmptyStatus is to check if initial status of RST object is empty.
	if IsEmptyStatus(rst) || IsPendingStatus(rst) {
		c.recorder.Event(rst, corev1.EventTypeNormal, string(common.SuccessCreated), string(common.MessageResourceCreated))
		glog.Infof("rst creation successful: %v, %v", rst.ObjectMeta.Name, string(rst.GetUID()))
		return string(apis.RSTStatusOnline), nil
	}
	err := volumereplica.CreateVolumeRestore(rst)
	if err != nil {
		glog.Errorf("rst creation failure: %v", err.Error())
		return string(apis.RSTStatusOffline), err
	}
	return "", nil
}

func (c *CStorRestoreController) rstSyncEventHandler(rst *apis.CStorRestore) (string, error) {
	// IsEmptyStatus is to check if initial status of RST object is empty.
	if IsEmptyStatus(rst) || IsPendingStatus(rst) {
		c.recorder.Event(rst, corev1.EventTypeNormal, string(common.SuccessCreated), string(common.MessageResourceCreated))
		glog.Infof("rst creation successful: %v, %v", rst.Name, string(rst.GetUID()))
		return string(apis.RSTStatusOnline), nil
	}
	err := volumereplica.CreateVolumeRestore(rst)
	if err != nil {
		glog.Errorf("rst creation failure: %v", err.Error())
		return string(apis.RSTStatusOffline), err
	}
	return "", nil
}

// IsEmptyStatus is to check if the status of cStorVolumeReplica object is empty.
func IsEmptyStatus(rst *apis.CStorRestore) bool {
	if string(rst.Status.Phase) == string(apis.RSTStatusEmpty) {
		glog.Infof("rst empty status: %v", string(rst.ObjectMeta.UID))
		return true
	}
	glog.Infof("Not empty status: %v", string(rst.ObjectMeta.UID))
	return false
}

// IsPendingStatus is to check if the status of cStorVolumeReplica object is pending.
func IsPendingStatus(rst *apis.CStorRestore) bool {
	if string(rst.Status.Phase) == string(apis.RSTStatusPending) {
		glog.Infof("rst pending: %v", string(rst.ObjectMeta.UID))
		return true
	}
	glog.V(4).Infof("Not pending status: %v", string(rst.ObjectMeta.UID))
	return false
}

// getVolumeReplicaResource returns object corresponding to the resource key
func (c *CStorRestoreController) getCStorRestoreResource(key string) (*apis.CStorRestore, error) {
	// Convert the key(namespace/name) string into a distinct name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil, nil
	}

	CStorRestoreUpdated, err := c.clientset.OpenebsV1alpha1().CStorRestores(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorPool resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("CStorRestoreUpdated '%s' in work queue no longer exists", key))
			return nil, nil
		}
		return nil, err
	}
	return CStorRestoreUpdated, nil
}

// IsRightCStorPoolMgmt is to check if the pool request is for particular pod/application.
func IsRightCStorPoolMgmt(cStorPool *apis.CStorRestore) bool {
	if os.Getenv(string(common.OpenEBSIOCStorID)) == cStorPool.ObjectMeta.Labels["cstorpool.openebs.io/uid"] {
		return true
	}
	return false
}

// IsDestroyEvent is to check if the call is for CStorVolumeReplica destroy.
func IsDestroyEvent(RST *apis.CStorRestore) bool {
	if RST.ObjectMeta.DeletionTimestamp != nil {
		return true
	}
	return false
}

// IsOnlyStatusChange is to check only status change of cStorVolumeReplica object.
func IsOnlyStatusChange(oldRST, newRST *apis.CStorRestore) bool {
	if reflect.DeepEqual(oldRST.Spec, newRST.Spec) &&
		!reflect.DeepEqual(oldRST.Status, newRST.Status) {
		return true
	}
	return false
}

// IsDeletionFailedBefore is to make sure no other operation should happen if the
// status of CStorVolumeReplica is deletion-failed.
func IsDeletionFailedBefore(rst *apis.CStorRestore) bool {
	if rst.Status.Phase == apis.RSTStatusDeletionFailed {
		return true
	}
	return false
}

// IsErrorDuplicate is to check if the status of cStorVolumeReplica object is error-duplicate.
func IsErrorDuplicate(rst *apis.CStorRestore) bool {
	if string(rst.Status.Phase) == string(apis.RSTStatusErrorDuplicate) {
		glog.Infof("rst duplication error: %v", rst.Name, string(rst.ObjectMeta.UID))
		return true
	}
	glog.V(4).Infof("Not error duplicate status: %v", string(rst.ObjectMeta.UID))
	return false
}
