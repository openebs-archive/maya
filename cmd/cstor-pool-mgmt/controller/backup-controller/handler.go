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
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/volumereplica"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the CStorReplicaUpdated resource
// with the current status of the resource.
func (c *BackupController) syncHandler(key string, operation common.QueueOperation) error {
	klog.Infof("Sync handler called for key:%s with op:%s", key, operation)
	bkp, err := c.getCStorBackupResource(key)
	if err != nil {
		return err
	}
	if bkp == nil {
		return fmt.Errorf("cannot retrieve CStorBackup %q", key)
	}
	if IsDoneStatus(bkp) || IsFailedStatus(bkp) {
		return nil
	}

	status, err := c.eventHandler(operation, bkp)
	if status == "" {
		return nil
	}

	if err != nil {
		klog.Errorf(err.Error())
		bkp.Status = apis.BKPCStorStatusFailed
	} else {
		bkp.Status = apis.CStorBackupStatus(status)
	}

	nbkp, err := c.clientset.OpenebsV1alpha1().CStorBackups(bkp.Namespace).
		Get(context.TODO(), bkp.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	nbkp.Status = bkp.Status

	_, err = c.clientset.OpenebsV1alpha1().CStorBackups(nbkp.Namespace).
		Update(context.TODO(), nbkp, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	klog.Infof("Completed operation:%v for backup:%v, status:%v", operation, nbkp.Name, nbkp.Status)
	return nil
}

// eventHandler will execute a function according to a given operation
func (c *BackupController) eventHandler(operation common.QueueOperation, bkp *apis.CStorBackup) (string, error) {
	switch operation {
	case common.QOpAdd:
		return c.addEventHandler(bkp)
	case common.QOpDestroy:
		/* TODO: Handle backup destroy event
		 */
		return "", nil
	case common.QOpSync:
		return c.syncEventHandler(bkp)
	case common.QOpModify:
		return "", nil
	}
	return string(apis.BKPCStorStatusInvalid), nil
}

// addEventHandler will change the state of backup to Init state.
func (c *BackupController) addEventHandler(bkp *apis.CStorBackup) (string, error) {
	if !IsPendingStatus(bkp) {
		return string(apis.BKPCStorStatusInvalid), nil
	}
	return string(apis.BKPCStorStatusInit), nil
}

// syncEventHandler will perform the backup if a given backup is in init state
func (c *BackupController) syncEventHandler(bkp *apis.CStorBackup) (string, error) {
	// If the backup is in init state then only we will complete the backup
	if IsInitStatus(bkp) {
		bkp.Status = apis.BKPCStorStatusInProgress
		_, err := c.clientset.OpenebsV1alpha1().CStorBackups(bkp.Namespace).
			Update(context.TODO(), bkp, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Failed to update backup:%s status : %v", bkp.Name, err.Error())
			return "", err
		}

		err = volumereplica.CreateVolumeBackup(bkp)
		if err != nil {
			c.recorder.Event(bkp, corev1.EventTypeNormal, string(common.SuccessCreated), string(common.MessageResourceCreated))
			klog.Errorf("Failed to create backup(%v): %v", bkp.ObjectMeta.Name, err.Error())
			return string(apis.BKPCStorStatusFailed), err
		}

		c.recorder.Event(bkp, corev1.EventTypeNormal, string(common.SuccessCreated), string(common.MessageResourceCreated))
		klog.Infof("backup creation successful: %v, %v", bkp.ObjectMeta.Name, string(bkp.GetUID()))
		err = c.updateCStorCompletedBackup(bkp)
		if err != nil {
			return string(apis.BKPCStorStatusFailed), err
		}
		return string(apis.BKPCStorStatusDone), nil
	}
	return "", nil
}

// getCStorBackupResource returns a backup object corresponding to the resource key
func (c *BackupController) getCStorBackupResource(key string) (*apis.CStorBackup, error) {
	// Convert the key(namespace/name) string into a distinct name
	klog.V(1).Infof("Finding backup for key:%s", key)
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil, nil
	}

	bkp, err := c.clientset.OpenebsV1alpha1().CStorBackups(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("bkp '%s' in work queue no longer exists", key))
			return nil, nil
		}
		return nil, err
	}
	return bkp, nil
}

// IsPendingStatus is to check if the backup is in a pending state.
func IsPendingStatus(bkp *apis.CStorBackup) bool {
	if string(bkp.Status) == string(apis.BKPCStorStatusPending) {
		return true
	}
	return false
}

// IsInProgressStatus is to check if the backup is in in-progress state.
func IsInProgressStatus(bkp *apis.CStorBackup) bool {
	if string(bkp.Status) == string(apis.BKPCStorStatusInProgress) {
		return true
	}
	return false
}

// IsInitStatus is to check if the backup is in init state.
func IsInitStatus(bkp *apis.CStorBackup) bool {
	if string(bkp.Status) == string(apis.BKPCStorStatusInit) {
		return true
	}
	return false
}

// IsDoneStatus is to check if the backup is completed or not
func IsDoneStatus(bkp *apis.CStorBackup) bool {
	if string(bkp.Status) == string(apis.BKPCStorStatusDone) {
		return true
	}
	return false
}

// IsFailedStatus is to check if the backup is failed or not
func IsFailedStatus(bkp *apis.CStorBackup) bool {
	if string(bkp.Status) == string(apis.BKPCStorStatusFailed) {
		return true
	}
	return false
}

// IsRightCStorPoolMgmt is to check if the backup request is for this cstor-pool or not.
func IsRightCStorPoolMgmt(bkp *apis.CStorBackup) bool {
	if os.Getenv(string(common.OpenEBSIOCStorID)) == bkp.ObjectMeta.Labels["cstorpool.openebs.io/uid"] {
		return true
	}
	return false
}

// IsDestroyEvent is to check if the call is for backup destroy.
func IsDestroyEvent(bkp *apis.CStorBackup) bool {
	if bkp.ObjectMeta.DeletionTimestamp != nil {
		return true
	}
	return false
}

// IsOnlyStatusChange is to check the only status change of CStorBackup object.
func IsOnlyStatusChange(oldbkp, newbkp *apis.CStorBackup) bool {
	if reflect.DeepEqual(oldbkp.Spec, newbkp.Spec) &&
		!reflect.DeepEqual(oldbkp.Status, newbkp.Status) {
		return true
	}
	return false
}

// updateCStorCompletedBackup updates the CStorCompletedBackups resource for the given backup
// CStorCompletedBackups stores the information of last two completed backups
// For example, if schedule `b` has last two backups b-0 and b-1 (b-0 created first and after that b-1 was created) having snapshots
// b-0 and b-1 respectively then CStorCompletedBackups for the schedule `b` will have following information :
//	CStorCompletedBackups.Spec.PrevSnapName =  b-1
//  CStorCompletedBackups.Spec.SnapName = b-0
func (c *BackupController) updateCStorCompletedBackup(bkp *apis.CStorBackup) error {
	lastbkpname := bkp.Spec.BackupName + "-" + bkp.Spec.VolumeName
	bkplast, err := c.clientset.OpenebsV1alpha1().CStorCompletedBackups(bkp.Namespace).
		Get(context.TODO(), lastbkpname, v1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get last completed backup for %s vol:%v", bkp.Spec.BackupName, bkp.Spec.VolumeName)
		return nil
	}

	// SnapName store the name of 2nd last backed up snapshot
	bkplast.Spec.SnapName = bkplast.Spec.PrevSnapName

	// PrevSnapName store the name of last backed up snapshot
	bkplast.Spec.PrevSnapName = bkp.Spec.SnapName
	_, err = c.clientset.OpenebsV1alpha1().CStorCompletedBackups(bkp.Namespace).
		Update(context.TODO(), bkplast, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update lastbackup for %s", bkplast.Name)
		return err
	}

	return nil
}
