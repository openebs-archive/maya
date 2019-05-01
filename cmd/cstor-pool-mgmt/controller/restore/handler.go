/*
Copyright 2019 The OpenEBS Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the CStorReplicaUpdated resource
// with the current status of the resource.
func (c *RestoreController) syncHandler(key string, operation common.QueueOperation) error {
	glog.Infof("Sync handler called for key:%s with op:%s", key, operation)
	rst, err := c.getCStorRestoreResource(key)
	if err != nil {
		return err
	}
	if rst == nil {
		return fmt.Errorf("can not retrieve CStorRestore %q", key)
	}

	status, err := c.rstEventHandler(operation, rst)
	if status == "" {
		return nil
	}

	if err != nil {
		glog.Errorf(err.Error())
		rst.Status = apis.RSTCStorStatusFailed
	} else {
		rst.Status = apis.CStorRestoreStatus(status)
	}

	nrst, err := c.clientset.OpenebsV1alpha1().CStorRestores(rst.Namespace).Get(rst.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	nrst.Status = rst.Status

	_, err = c.clientset.OpenebsV1alpha1().CStorRestores(nrst.Namespace).Update(nrst)
	if err != nil {
		return err
	}

	glog.Infof("Completed operation:%v for restore:%v, status:%v", operation, nrst.Name, nrst.Status)
	return nil
}

// eventHandler will execute a function according to a given operation
func (c *RestoreController) rstEventHandler(operation common.QueueOperation, rst *apis.CStorRestore) (string, error) {
	switch operation {
	case common.QOpAdd:
		return c.addEventHandler(rst)
	case common.QOpDestroy:
		/*
			status, err := c.rstDestroyEventHandler(rstGot)
			return status, err
			glog.Infof("Processing restore delete event %v, %v", rstGot.ObjectMeta.Name, string(rstGot.GetUID()))
		*/
		return "", nil
	case common.QOpSync:
		return c.syncEventHandler(rst)
	case common.QOpModify:
		return "", nil
	}
	return string(apis.RSTCStorStatusInvalid), nil
}

// addEventHandler will change the state of restore to Init state.
func (c *RestoreController) addEventHandler(rst *apis.CStorRestore) (string, error) {
	if !IsPendingStatus(rst) {
		return string(apis.RSTCStorStatusInvalid), nil
	}
	return string(apis.RSTCStorStatusInit), nil
}

// syncEventHandler will perform the restore if a given restore is in init state
func (c *RestoreController) syncEventHandler(rst *apis.CStorRestore) (string, error) {
	// If the restore is in init state then only we will complete the restore
	if IsInitStatus(rst) {
		rst.Status = apis.RSTCStorStatusInProgress
		_, err := c.clientset.OpenebsV1alpha1().CStorRestores(rst.Namespace).Update(rst)
		if err != nil {
			glog.Errorf("Failed to update restore:%s status : %v", rst.Name, err.Error())
			return "", err
		}

		err = volumereplica.CreateVolumeRestore(rst)
		if err != nil {
			glog.Errorf("restore creation failure: %v", err.Error())
			return string(apis.RSTCStorStatusFailed), err
		}
		c.recorder.Event(rst, corev1.EventTypeNormal, string(common.SuccessCreated), string(common.MessageResourceCreated))
		glog.Infof("restore creation successful: %v, %v", rst.ObjectMeta.Name, string(rst.GetUID()))
		return string(apis.RSTCStorStatusDone), nil
	} else if IsPendingStatus(rst) {
		glog.Infof("Updating restore:%s status to %v", rst.Name, apis.RSTCStorStatusInit)
		return string(apis.RSTCStorStatusInit), nil
	}
	return "", nil
}

// getCStorRestoreResource returns a restore object corresponding to the resource key
func (c *RestoreController) getCStorRestoreResource(key string) (*apis.CStorRestore, error) {
	// Convert the key(namespace/name) string into a distinct name
	glog.Infof("Finding restore for key:%s", key)
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil, nil
	}

	rst, err := c.clientset.OpenebsV1alpha1().CStorRestores(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("Restore resource for key:%s is missing", name)
		return nil, err
	}
	return rst, nil
}

// IsPendingStatus is to check if the restore is in a pending state.
func IsPendingStatus(rst *apis.CStorRestore) bool {
	if string(rst.Status) == string(apis.RSTCStorStatusPending) {
		return true
	}
	return false
}

// IsInProgressStatus is to check if the restore is in in-progress state.
func IsInProgressStatus(rst *apis.CStorRestore) bool {
	if string(rst.Status) == string(apis.RSTCStorStatusInProgress) {
		return true
	}
	return false
}

// IsInitStatus is to check if the restore is in init state.
func IsInitStatus(rst *apis.CStorRestore) bool {
	if string(rst.Status) == string(apis.RSTCStorStatusInit) {
		return true
	}
	return false
}

// IsDoneStatus is to check if the restore is completed or not
func IsDoneStatus(rst *apis.CStorRestore) bool {
	if string(rst.Status) == string(apis.RSTCStorStatusDone) {
		return true
	}
	return false
}

// IsFailedStatus is to check if the restore is failed or not
func IsFailedStatus(rst *apis.CStorRestore) bool {
	if string(rst.Status) == string(apis.RSTCStorStatusFailed) {
		return true
	}
	return false
}

// IsRightCStorPoolMgmt is to check if the restore request is for particular pod/application.
func IsRightCStorPoolMgmt(rst *apis.CStorRestore) bool {
	if os.Getenv(string(common.OpenEBSIOCStorID)) == rst.ObjectMeta.Labels["cstorpool.openebs.io/uid"] {
		return true
	}
	return false
}

// IsDestroyEvent is to check if the call is for restore destroy.
func IsDestroyEvent(rst *apis.CStorRestore) bool {
	if rst.ObjectMeta.DeletionTimestamp != nil {
		return true
	}
	return false
}

// IsOnlyStatusChange is to check only status change of restore object.
func IsOnlyStatusChange(oldrst, newrst *apis.CStorRestore) bool {
	if reflect.DeepEqual(oldrst.Spec, newrst.Spec) &&
		!reflect.DeepEqual(oldrst.Status, newrst.Status) {
		return true
	}
	return false
}
