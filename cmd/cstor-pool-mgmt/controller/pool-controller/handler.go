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

package poolcontroller

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/volumereplica"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the cStorPoolUpdated resource
// with the current status of the resource.
func (c *CStorPoolController) syncHandler(key string, operation common.QueueOperation) error {
	cStorPoolGot, err := c.getPoolResource(key)
	if err != nil {
		return err
	}
	status, err := c.cStorPoolEventHandler(operation, cStorPoolGot)
	if status == "" {
		return nil
	}
	cStorPoolGot.Status.Phase = apis.CStorPoolPhase(status)
	if err != nil {
		glog.Errorf(err.Error())
		_, err := c.clientset.OpenebsV1alpha1().CStorPools().Update(cStorPoolGot)
		if err != nil {
			return err
		}
		glog.Infof("cStorPool:%v, %v; Status: %v", cStorPoolGot.Name,
			string(cStorPoolGot.GetUID()), cStorPoolGot.Status.Phase)
		return err
	}
	_, err = c.clientset.OpenebsV1alpha1().CStorPools().Update(cStorPoolGot)
	if err != nil {
		return err
	}
	glog.Infof("cStorPool:%v, %v; Status: %v", cStorPoolGot.Name,
		string(cStorPoolGot.GetUID()), cStorPoolGot.Status.Phase)
	return nil
}

// cStorPoolEventHandler is to handle cstor pool related events.
func (c *CStorPoolController) cStorPoolEventHandler(operation common.QueueOperation, cStorPoolGot *apis.CStorPool) (string, error) {
	pool.RunnerVar = util.RealRunner{}
	switch operation {
	case common.QOpAdd:
		glog.Infof("Processing cStorPool added event: %v, %v", cStorPoolGot.ObjectMeta.Name, string(cStorPoolGot.GetUID()))

		// lock is to synchronize pool and volumereplica. Until certain pool related
		// operations are over, the volumereplica threads will be held.
		common.SyncResources.Mux.Lock()
		status, err := c.cStorPoolAddEventHandler(cStorPoolGot)
		common.SyncResources.Mux.Unlock()
		pool.PoolAddEventHandled = true
		return status, err

	case common.QOpDestroy:
		glog.Infof("Processing cStorPool Destroy event %v, %v", cStorPoolGot.ObjectMeta.Name, string(cStorPoolGot.GetUID()))

		status, err := c.cStorPoolDestroyEventHandler(cStorPoolGot)
		return status, err
	}
	return string(apis.CStorPoolStatusInvalid), nil
}

func (c *CStorPoolController) cStorPoolAddEventHandler(cStorPoolGot *apis.CStorPool) (string, error) {
	// CheckValidPool is to check if pool attributes are correct.
	err := pool.CheckValidPool(cStorPoolGot)
	if err != nil {
		c.recorder.Event(cStorPoolGot, corev1.EventTypeWarning, string(common.FailureValidate), string(common.MessageResourceFailValidate))
		return string(apis.CStorPoolStatusOffline), err
	}

	/* 	If pool is already present.
	Pool CR status is online. This means pool (main car) is running successfully,
	but watcher container got restarted.
	Pool CR status is init/online. If entire pod got restarted, both zrepl and watcher
	are started.
	a) Zrepl could have come up first, in this case, watcher will update after
	the specified interval of 120s.
	b) Watcher could have come up first, in this case, there is a possibility
	that zrepl goes down and comes up and the watcher sees that no pool is there,
	so it will break the loop and attempt to import the pool. */

	// cnt is no of attempts to wait and handle in case of already present pool.
	cnt := common.NoOfPoolWaitAttempts
	existingPool, _ := pool.GetPoolName()
	isPoolExists := len(existingPool) != 0

	for i := 0; isPoolExists && i < cnt; i++ {
		// GetVolumes is called because, while importing a pool, volumes corresponding
		// to the pool are also imported. This needs to be handled and made visible
		// to cvr controller.
		common.InitialImportedPoolVol, _ = volumereplica.GetVolumes()
		// GetPoolName is to get pool name for particular no. of attempts.
		existingPool, _ := pool.GetPoolName()
		if common.CheckIfPresent(existingPool, string(pool.PoolPrefix)+string(cStorPoolGot.GetUID())) {
			// In the last attempt, ignore and update the status.
			if i == cnt-1 {
				isPoolExists = false
				if IsPendingStatus(cStorPoolGot) || IsEmptyStatus(cStorPoolGot) {
					// Pool CR status is init. This means pool deployment was done
					// successfully, but before updating the CR to Online status,
					// the watcher container got restarted.
					glog.Infof("Pool %v is online", string(pool.PoolPrefix)+string(cStorPoolGot.GetUID()))
					c.recorder.Event(cStorPoolGot, corev1.EventTypeNormal, string(common.AlreadyPresent), string(common.MessageResourceAlreadyPresent))
					common.SyncResources.IsImported = true
					return string(apis.CStorPoolStatusOnline), nil
				}
				glog.Infof("Pool %v already present", string(pool.PoolPrefix)+string(cStorPoolGot.GetUID()))
				c.recorder.Event(cStorPoolGot, corev1.EventTypeNormal, string(common.AlreadyPresent), string(common.MessageResourceAlreadyPresent))
				common.SyncResources.IsImported = true
				return string(apis.CStorPoolStatusErrorDuplicate), fmt.Errorf("Duplicate resource request")
			}
			glog.Infof("Attempt %v: Waiting...", i+1)
			time.Sleep(common.PoolWaitInterval)
		} else {
			// If no pool is present while trying for getpoolname, set isPoolExists to false and
			// break the loop, to import the pool later.
			isPoolExists = false
		}
	}
	var importPoolErr error
	var status string
	cachfileFlags := []bool{true, false}
	for _, cachefileFlag := range cachfileFlags {
		status, importPoolErr = c.importPool(cStorPoolGot, cachefileFlag)
		if status == string(apis.CStorPoolStatusOnline) {
			c.recorder.Event(cStorPoolGot, corev1.EventTypeNormal, string(common.SuccessImported), string(common.MessageResourceImported))
			common.SyncResources.IsImported = true
			return status, nil
		}
	}

	// make a check if initialImportedPoolVol is not empty, then notify cvr controller
	// through channel.
	if len(common.InitialImportedPoolVol) != 0 {
		common.SyncResources.IsImported = true
	} else {
		common.SyncResources.IsImported = false
	}

	// IsInitStatus is to check if initial status of cstorpool object is `init`.
	if IsEmptyStatus(cStorPoolGot) || IsPendingStatus(cStorPoolGot) {
		// LabelClear is to clear pool label
		err = pool.LabelClear(cStorPoolGot.Spec.Disks.DiskList)
		if err != nil {
			glog.Errorf(err.Error(), cStorPoolGot.GetUID())
		} else {
			glog.Infof("Label clear successful: %v", string(cStorPoolGot.GetUID()))
		}

		// CreatePool is to create cstor pool.
		err = pool.CreatePool(cStorPoolGot)
		if err != nil {
			glog.Errorf("Pool creation failure: %v", string(cStorPoolGot.GetUID()))
			c.recorder.Event(cStorPoolGot, corev1.EventTypeWarning, string(common.FailureCreate), string(common.MessageResourceFailCreate))
			return string(apis.CStorPoolStatusOffline), err
		}
		glog.Infof("Pool creation successful: %v", string(cStorPoolGot.GetUID()))
		c.recorder.Event(cStorPoolGot, corev1.EventTypeNormal, string(common.SuccessCreated), string(common.MessageResourceCreated))
		return string(apis.CStorPoolStatusOnline), nil
	}
	glog.Infof("Not init status: %v, %v", cStorPoolGot.ObjectMeta.Name, string(cStorPoolGot.GetUID()))

	return string(apis.CStorPoolStatusOffline), importPoolErr
}

func (c *CStorPoolController) cStorPoolDestroyEventHandler(cStorPoolGot *apis.CStorPool) (string, error) {
	// DeletePool is to delete cstor pool.
	err := pool.DeletePool(string(pool.PoolPrefix) + string(cStorPoolGot.ObjectMeta.UID))
	if err != nil {
		c.recorder.Event(cStorPoolGot, corev1.EventTypeWarning, string(common.FailureDestroy), string(common.MessageResourceFailDestroy))
		return string(apis.CStorPoolStatusDeletionFailed), err
	}

	// LabelClear is to clear pool label
	err = pool.LabelClear(cStorPoolGot.Spec.Disks.DiskList)
	if err != nil {
		glog.Errorf(err.Error(), cStorPoolGot.GetUID())
	} else {
		glog.Infof("Label clear successful: %v", string(cStorPoolGot.GetUID()))
	}

	// removeFinalizer is to remove finalizer of cStorPool resource.
	err = c.removeFinalizer(cStorPoolGot)
	if err != nil {
		return string(apis.CStorPoolStatusOffline), err
	}
	return "", nil

}

// getPoolResource returns object corresponding to the resource key
func (c *CStorPoolController) getPoolResource(key string) (*apis.CStorPool, error) {
	// Convert the key(namespace/name) string into a distinct name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil, nil
	}

	cStorPoolGot, err := c.clientset.OpenebsV1alpha1().CStorPools().Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorPool resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("cStorPoolGot '%s' in work queue no longer exists", key))
			return nil, nil
		}

		return nil, err
	}
	return cStorPoolGot, nil
}

// removeFinalizer is to remove finalizer of cstorpool resource.
func (c *CStorPoolController) removeFinalizer(cStorPoolGot *apis.CStorPool) error {
	if len(cStorPoolGot.Finalizers) > 0 {
		cStorPoolGot.Finalizers = []string{}
	}
	_, err := c.clientset.OpenebsV1alpha1().CStorPools().Update(cStorPoolGot)
	if err != nil {
		return err
	}
	glog.Infof("Removed Finalizer: %v, %v", cStorPoolGot.Name, string(cStorPoolGot.GetUID()))
	return nil
}

func (c *CStorPoolController) importPool(cStorPoolGot *apis.CStorPool, cachefileFlag bool) (string, error) {
	err := pool.ImportPool(cStorPoolGot, cachefileFlag)
	if err == nil {
		err = pool.SetCachefile(cStorPoolGot)
		if err != nil {
			common.SyncResources.IsImported = false
			return string(apis.CStorPoolStatusOffline), err
		}
		glog.Infof("Set cachefile successful: %v", string(cStorPoolGot.GetUID()))
		// GetVolumes is called because, while importing a pool, volumes corresponding
		// to the pool are also imported. This needs to be handled and made visible
		// to cvr controller.
		common.InitialImportedPoolVol, err = volumereplica.GetVolumes()
		if err != nil {
			common.SyncResources.IsImported = false
			return string(apis.CStorPoolStatusOffline), err
		}
		glog.Infof("Import Pool with cachefile successful: %v", string(cStorPoolGot.GetUID()))
		return string(apis.CStorPoolStatusOnline), nil
	}
	return "", nil
}

// IsRightCStorPoolMgmt is to check if the pool request is for particular pod/application.
func IsRightCStorPoolMgmt(cStorPool *apis.CStorPool) bool {
	if os.Getenv(string(common.OpenEBSIOCStorID)) == string(cStorPool.ObjectMeta.UID) {
		return true
	}
	return false
}

// IsDestroyEvent is to check if the call is for cStorPool destroy.
func IsDestroyEvent(cStorPool *apis.CStorPool) bool {
	if cStorPool.ObjectMeta.DeletionTimestamp != nil {
		return true
	}
	return false
}

// IsOnlyStatusChange is to check only status change of cStorPool object.
func IsOnlyStatusChange(oldCStorPool, newCStorPool *apis.CStorPool) bool {
	if reflect.DeepEqual(oldCStorPool.Spec, newCStorPool.Spec) &&
		!reflect.DeepEqual(oldCStorPool.Status, newCStorPool.Status) {
		return true
	}
	return false
}

// IsEmptyStatus is to check if the status of cStorPool object is empty.
func IsEmptyStatus(cStorPool *apis.CStorPool) bool {
	if string(cStorPool.Status.Phase) == string(apis.CStorPoolStatusEmpty) {
		glog.Infof("cStorPool empty status: %v", string(cStorPool.ObjectMeta.UID))
		return true
	}
	glog.Infof("Not empty status: %v", string(cStorPool.ObjectMeta.UID))
	return false
}

// IsPendingStatus is to check if the status of cStorPool object is pending.
func IsPendingStatus(cStorPool *apis.CStorPool) bool {
	if string(cStorPool.Status.Phase) == string(apis.CStorPoolStatusPending) {
		glog.Infof("cStorPool pending: %v", string(cStorPool.ObjectMeta.UID))
		return true
	}
	glog.V(4).Infof("Not pending status: %v", string(cStorPool.ObjectMeta.UID))
	return false
}

// IsErrorDuplicate is to check if the status of cStorPool object is error-duplicate.
func IsErrorDuplicate(cStorPool *apis.CStorPool) bool {
	if string(cStorPool.Status.Phase) == string(apis.CStorPoolStatusErrorDuplicate) {
		glog.Infof("cStorPool duplication error: %v", string(cStorPool.ObjectMeta.UID))
		return true
	}
	glog.V(4).Infof("Not error duplicate status: %v", string(cStorPool.ObjectMeta.UID))
	return false
}

// IsDeletionFailedBefore is to make sure no other operation should happen if the
// status of cStorPool is deletion-failed.
func IsDeletionFailedBefore(cStorPool *apis.CStorPool) bool {
	if cStorPool.Status.Phase == apis.CStorPoolStatusDeletionFailed {
		return true
	}
	return false
}
