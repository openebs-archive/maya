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
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	pool "github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/volumereplica"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	zpool "github.com/openebs/maya/pkg/apis/openebs.io/zpool/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	lease "github.com/openebs/maya/pkg/lease/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

type upgradeParams struct {
	csp    *apis.CStorPool
	client clientset.Interface
}

type upgradeFunc func(u *upgradeParams) (*apis.CStorPool, error)

var (
	upgradeMap   = map[string]upgradeFunc{}
	cspROUpdated bool
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the cStorPoolUpdated resource
// with the current status of the resource.
func (c *CStorPoolController) syncHandler(key string, operation common.QueueOperation) error {
	cStorPoolGot, err := c.getPoolResource(key)
	if err != nil {
		return err
	}
	var newCspLease lease.Leaser
	newCspLease = &lease.Lease{
		Object:        cStorPoolGot,
		LeaseKey:      lease.CspLeaseKey,
		Oecs:          c.clientset,
		Kubeclientset: c.kubeclientset,
	}
	csp, err := newCspLease.Hold()
	cspObject, ok := csp.(*apis.CStorPool)
	if !ok {
		return fmt.Errorf("expected csp object but got %#v", cspObject)
	}
	if err != nil {
		klog.Errorf("Could not acquire lease on csp object:%v", err)
		return err
	}
	klog.V(4).Infof("Lease acquired successfully on csp %s ", cspObject.Name)
	cspObject, err = c.populateVersion(cspObject)
	if err != nil {
		klog.Errorf("failed to add versionDetails to csp %s:%s", cspObject.Name, err.Error())
		c.recorder.Event(
			cspObject,
			corev1.EventTypeWarning,
			"FailedPopulate",
			fmt.Sprintf("Failed to add current version: %s", err.Error()),
		)
		return nil
	}
	cspObject, err = c.reconcileVersion(cspObject)
	if err != nil {
		klog.Errorf("failed to upgrade csp %s:%s", cspObject.Name, err.Error())
		c.recorder.Event(
			cspObject,
			corev1.EventTypeWarning,
			"FailedUpgrade",
			fmt.Sprintf("Failed to upgrade csp to %s version: %s",
				cspObject.VersionDetails.Desired,
				err.Error(),
			),
		)
		cspObject.VersionDetails.Status.SetErrorStatus(
			"Failed to reconcile csp version",
			err,
		)
		_, err = c.clientset.OpenebsV1alpha1().CStorPools().Update(cspObject)
		if err != nil {
			klog.Errorf("failed to update versionDetails status for csp %s:%s", cspObject.Name, err.Error())
		}
		return nil
	}
	status, err := c.cStorPoolEventHandler(operation, cspObject)
	if status == "" {
		klog.Warning("Empty status recieved for csp status in sync handler")
		return nil
	}
	cspObject.Status.LastUpdateTime = metav1.Now()
	if cspObject.Status.Phase != apis.CStorPoolPhase(status) {
		cspObject.Status.LastTransitionTime = cspObject.Status.LastUpdateTime
		cspObject.Status.Phase = apis.CStorPoolPhase(status)
	}

	if err != nil {
		klog.Errorf(err.Error())
		_, err := c.clientset.OpenebsV1alpha1().CStorPools().Update(cspObject)
		if err != nil {
			return err
		}
		klog.Infof("cStorPool:%v, %v; Status: %v", cspObject.Name,
			string(cspObject.GetUID()), cspObject.Status.Phase)
		return err
	}
	// Synchronize cstor pool used and free capacity fields on CSP object.
	// Also verify and handle pool ReadOnly threshold limit
	// Any kind of sync activity should be done from here.
	// ToDo: Move status sync (of csp) here from cStorPoolEventHandler function.
	// ToDo: Instead of having statusSync, capacitySync we can make it generic resource sync which syncs all the
	// ToDo: requried fields on CSP ( Some code re-organization will be required)
	c.syncCsp(cspObject)
	_, err = c.clientset.OpenebsV1alpha1().CStorPools().Update(cspObject)
	if err != nil {
		cspROUpdated = false
		c.recorder.Event(cspObject, corev1.EventTypeWarning, string(common.FailedSynced), string(common.MessageResourceSyncFailure)+err.Error())
		return err
	}
	cspROUpdated = true
	if string(cspObject.Status.Phase) == string(apis.CStorPoolStatusOnline) {
		klog.V(4).Infof("cStorPool:%v, %v; Status: Online", cspObject.Name, string(cspObject.GetUID()))
	} else {
		klog.Infof("cStorPool:%v, %v; Status: %v", cspObject.Name,
			string(cspObject.GetUID()), cspObject.Status.Phase)
	}
	return nil
}

// cStorPoolAddEventHandler calls cStorPoolAddEvent, and makes in-mem structures of successfully imported/created pool
func (c *CStorPoolController) cStorPoolAddEventHandler(cStorPoolGot *apis.CStorPool) (string, error) {
	var zpoolDumpErr error
	pool.RunnerVar = util.RealRunner{}
	uid := string(cStorPoolGot.GetUID())
	csp := pool.ImportedCStorPools[uid]

	common.SyncResources.Mux.Lock()
	if csp != nil {
		c.recorder.Event(cStorPoolGot, corev1.EventTypeWarning, string(common.AlreadyPresent), string(common.MessageResourceAlreadyPresent))
	}
	status, err := c.cStorPoolAddEvent(cStorPoolGot)
	if status == string(apis.CStorPoolStatusOnline) {
		pool.ImportedCStorPools[uid] = cStorPoolGot.DeepCopy()
		pool.CStorZpools[uid], zpoolDumpErr = zpool.Dump()
		if zpoolDumpErr != nil {
			klog.Errorf("failed in getting zpool dump %v", zpoolDumpErr)
			delete(pool.CStorZpools, uid)
		}
	}
	common.SyncResources.Mux.Unlock()
	pool.PoolAddEventHandled = true
	return status, err
}

// cStorPoolEventHandler is to handle cstor pool related events.
func (c *CStorPoolController) cStorPoolEventHandler(operation common.QueueOperation, cStorPoolGot *apis.CStorPool) (string, error) {
	switch operation {
	case common.QOpAdd:
		klog.Infof("Processing cStorPool added event: %v, %v", cStorPoolGot.ObjectMeta.Name, string(cStorPoolGot.GetUID()))

		if IsCStorPoolCreateStatuses(cStorPoolGot) {
			return c.cStorPoolCreate(cStorPoolGot)
		}
		// lock is to synchronize pool and volumereplica. Until certain pool related
		// operations are over, the volumereplica threads will be held.
		status, err := c.cStorPoolAddEventHandler(cStorPoolGot)
		return status, err

	case common.QOpDestroy:
		klog.Infof("Processing cStorPool Destroy event %v, %v", cStorPoolGot.ObjectMeta.Name, string(cStorPoolGot.GetUID()))
		status, err := c.cStorPoolDestroyEventHandler(cStorPoolGot)
		return status, err
	case common.QOpSync:
		// Check if pool is not imported/created earlier due to any failure or failure in getting lease
		// try to import/create pool here as part of reconcile.

		if IsCStorPoolCreateStatuses(cStorPoolGot) {
			return c.cStorPoolCreate(cStorPoolGot)
		}
		if IsPendingStatus(cStorPoolGot) {
			status, err := c.cStorPoolAddEventHandler(cStorPoolGot)
			return status, err
		}
		klog.V(4).Infof("Synchronizing cStor pool status for pool %s", cStorPoolGot.ObjectMeta.Name)
		status, readOnly, err := c.getPoolStatus(cStorPoolGot)
		if err == nil {
			cStorPoolGot.Status.ReadOnly = readOnly
		}
		return status, err
	}
	klog.Errorf("ignored event '%s' for cstor pool '%s'", string(operation), string(cStorPoolGot.ObjectMeta.Name))
	return string(apis.CStorPoolStatusInvalid), nil
}

// cStorPoolCreate does create of pool, and, if it fails, returns CreateFailed
func (c *CStorPoolController) cStorPoolCreate(cStorPoolGot *apis.CStorPool) (string, error) {
	if !IsCStorPoolCreateStatuses(cStorPoolGot) {
		c.recorder.Event(cStorPoolGot, corev1.EventTypeWarning, string(common.FailureCreate), string(common.MessageImproperPoolStatus))
		return string(cStorPoolGot.Status.Phase), fmt.Errorf("invalid status %s for create pool %s", cStorPoolGot.Status.Phase, cStorPoolGot.Name)
	}

	devIDList, err := c.getDeviceIDs(cStorPoolGot)
	if err != nil {
		return string(apis.CStorPoolStatusCreateFailed), errors.Wrapf(err, "failed to get device id of disks for csp %s", cStorPoolGot.Name)
	}

	// ValidatePool is to check if pool attributes are correct.
	err = pool.ValidatePool(cStorPoolGot, devIDList)
	if err != nil {
		c.recorder.Event(cStorPoolGot, corev1.EventTypeWarning, string(common.FailureValidate), string(common.MessageResourceFailValidate))
		return string(apis.CStorPoolStatusCreateFailed), err
	}

	if len(common.InitialImportedPoolVol) != 0 {
		klog.Errorf("failed to handle add event: invalid status %v for pool %v with existing volumes", string(cStorPoolGot.Status.Phase), string(cStorPoolGot.GetUID()))
		c.recorder.Event(cStorPoolGot, corev1.EventTypeWarning, string(common.FailureValidate), string(common.MessageImproperPoolStatus))
		return string(apis.CStorPoolStatusCreateFailed), err
	}
	// CreatePool is to create cstor pool.
	err = pool.CreatePool(cStorPoolGot, devIDList)
	if err != nil {
		klog.Errorf("Pool creation failure: %v", string(cStorPoolGot.GetUID()))
		c.recorder.Event(cStorPoolGot, corev1.EventTypeWarning, string(common.FailureCreate), string(common.MessageResourceFailCreate))
		return string(apis.CStorPoolStatusCreateFailed), err
	}
	klog.Infof("Pool creation successful: %v", string(cStorPoolGot.GetUID()))
	c.recorder.Event(cStorPoolGot, corev1.EventTypeNormal, string(common.SuccessCreated), string(common.MessageResourceCreated))
	return string(apis.CStorPoolStatusOnline), nil
}

// cStorPoolAddEvent does import of pool
func (c *CStorPoolController) cStorPoolAddEvent(cStorPoolGot *apis.CStorPool) (string, error) {
	if pool.ImportedCStorPools == nil {
		pool.ImportedCStorPools = map[string]*apis.CStorPool{}
	}

	if pool.CStorZpools == nil {
		pool.CStorZpools = map[string]zpool.Topology{}
	}

	// To import the pool this check is required
	if cStorPoolGot.ObjectMeta.UID == "" {
		c.recorder.Event(cStorPoolGot, corev1.EventTypeWarning, string(common.FailureValidate), string(common.MessageResourceFailValidate))
		return string(apis.CStorPoolStatusOffline), fmt.Errorf("Poolname/UID cannot be empty")
	}

	if cStorPoolGot.Spec.PoolSpec.CacheFile == "" {
		cStorPoolGot.Spec.PoolSpec.CacheFile = "/tmp/pool1.cache"
	}

	/* 	If pool is already present.
	Pool CR status is online. This means pool (main car) is running successfully,
	but watcher container got restarted.
	Pool CR status is init/online. If entire pod got restarted, both zrepl and watcher
	are started.
	a) Zrepl could have come up first, in this case, watcher will update after
	the specified interval of (2*30) = 60s.
	b) Watcher could have come up first, in this case, there is a possibility
	that zrepl goes down and comes up and the watcher sees that no pool is there,
	so it will break the loop and attempt to import the pool. */

	// cnt is no of attempts to wait and handle in case of already present pool.
	cnt := common.NoOfPoolWaitAttempts
	existingPool, _ := pool.GetPoolName()
	isPoolExists := len(existingPool) != 0

	// There is no need of loop here, if the GetPoolName returns poolname with cStorPoolGot.GetUID.
	// It is going to stay forever until zrepl restarts
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
				if IsPendingStatus(cStorPoolGot) || IsEmptyStatus(cStorPoolGot) {
					// Pool CR status is init. This means pool deployment was done
					// successfully, but before updating the CR to Online status,
					// the watcher container got restarted.
					klog.Infof("Pool %v is online", string(pool.PoolPrefix)+string(cStorPoolGot.GetUID()))
					c.recorder.Event(cStorPoolGot, corev1.EventTypeNormal, string(common.AlreadyPresent), string(common.MessageResourceAlreadyPresent))
					common.SyncResources.IsImported = true
					return string(apis.CStorPoolStatusOnline), nil
				}
				klog.Infof("Pool %v already present", string(pool.PoolPrefix)+string(cStorPoolGot.GetUID()))
				c.recorder.Event(cStorPoolGot, corev1.EventTypeNormal, string(common.AlreadyPresent), string(common.MessageResourceAlreadyPresent))
				common.SyncResources.IsImported = true
				return string(apis.CStorPoolStatusErrorDuplicate), fmt.Errorf("Duplicate resource request")
			}
			klog.Infof("Attempt %v: Waiting...", i+1)
			time.Sleep(common.PoolWaitInterval)
		} else {
			// If no pool is present while trying for getpoolname, set isPoolExists to false and
			// break the loop, to import the pool later.
			isPoolExists = false
		}
	}
	var importPoolErr error
	var status string
	var importOptions pool.ImportOptions
	cachefileFlags := []bool{true, false}
	for _, cachefileFlag := range cachefileFlags {
		importOptions.CachefileFlag = cachefileFlag
		status, importPoolErr = c.triggerImportPool(cStorPoolGot, &importOptions)
		if status == string(apis.CStorPoolStatusOnline) {
			return status, nil
		}
		if importPoolErr != nil {
			return status, importPoolErr
		}
	}

	devIDList, err := c.getDeviceIDs(cStorPoolGot)
	if err != nil {
		return string(apis.CStorPoolStatusOffline), errors.Wrapf(err, "failed to get device id of disks for csp %s", cStorPoolGot.Name)
	}

	devPath := pool.GetDevPathIfNotSlashDev(devIDList[0])
	// Trigger import with dev path
	importOptions.CachefileFlag = false
	importOptions.DevPath = devPath
	status, importPoolErr = c.triggerImportPool(cStorPoolGot, &importOptions)
	if status == string(apis.CStorPoolStatusOnline) {
		return status, nil
	}
	if importPoolErr != nil {
		return status, importPoolErr
	}

	// make a check if initialImportedPoolVol is not empty, then notify cvr controller
	// through channel.
	if len(common.InitialImportedPoolVol) != 0 {
		common.SyncResources.IsImported = true
	} else {
		common.SyncResources.IsImported = false
	}

	klog.Infof("Not init status %v: %v, %v", string(cStorPoolGot.Status.Phase), cStorPoolGot.ObjectMeta.Name, string(cStorPoolGot.GetUID()))

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
	devIDList, err := c.getDeviceIDs(cStorPoolGot)
	if err != nil {
		return string(apis.CStorPoolStatusOffline), errors.Wrapf(err, "failed to get device id of disks for csp %s", cStorPoolGot.Name)
	}
	err = pool.LabelClear(devIDList)
	if err != nil {
		klog.Errorf(err.Error(), cStorPoolGot.GetUID())
	} else {
		klog.Infof("Label clear successful: %v", string(cStorPoolGot.GetUID()))
	}

	// removeFinalizer is to remove finalizer of cStorPool resource.
	err = c.removeFinalizer(cStorPoolGot)
	if err != nil {
		return string(apis.CStorPoolStatusOffline), err
	}
	return "", nil

}

//  getPoolStatus is a wrapper that fetches the status of cstor pool.
func (c *CStorPoolController) getPoolStatus(cStorPoolGot *apis.CStorPool) (string, bool, error) {
	poolStatus, readOnly, err := pool.Status(string(pool.PoolPrefix) + string(cStorPoolGot.ObjectMeta.UID))
	if err != nil {
		// ToDO : Put error in event recorder
		c.recorder.Event(cStorPoolGot, corev1.EventTypeWarning, string(common.FailureStatusSync), string(common.MessageResourceFailStatusSync))
		return "", false, err
	}
	return poolStatus, readOnly, nil
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
		if k8serror.IsNotFound(err) {
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
	klog.Infof("Removed Finalizer: %v, %v", cStorPoolGot.Name, string(cStorPoolGot.GetUID()))
	return nil
}

func (c *CStorPoolController) triggerImportPool(cStorPoolGot *apis.CStorPool, importOptions *pool.ImportOptions) (string, error) {
	status, importPoolErr := c.importPool(cStorPoolGot, importOptions)
	if status == string(apis.CStorPoolStatusOnline) {
		c.recorder.Event(cStorPoolGot, corev1.EventTypeNormal, string(common.SuccessImported), string(common.MessageResourceImported))
		common.SyncResources.IsImported = true
		return status, nil
	}
	if importPoolErr != nil {
		klog.Errorf("failed to handle add event: import succeded with failed operations %v", importPoolErr)
		c.recorder.Event(cStorPoolGot, corev1.EventTypeWarning, string(common.FailureImported), string(common.FailureImportOperations))
		return status, importPoolErr
	}
	return status, importPoolErr
}

func (c *CStorPoolController) importPool(cStorPoolGot *apis.CStorPool, importOptions *pool.ImportOptions) (string, error) {
	_, err := pool.ImportPool(cStorPoolGot, importOptions)
	if err == nil {
		err = pool.SetCachefile(cStorPoolGot)
		if err != nil {
			common.SyncResources.IsImported = false
			return string(apis.CStorPoolStatusOffline), err
		}
		klog.Infof("Set cachefile successful: %v", string(cStorPoolGot.GetUID()))
		// GetVolumes is called because, while importing a pool, volumes corresponding
		// to the pool are also imported. This needs to be handled and made visible
		// to cvr controller.
		common.InitialImportedPoolVol, err = volumereplica.GetVolumes()
		if err != nil {
			common.SyncResources.IsImported = false
			return string(apis.CStorPoolStatusOffline), err
		}
		klog.Infof("Import Pool with cachefile successful: %v", string(cStorPoolGot.GetUID()))
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

// IsCStorPoolCreateStatuses is to check if the status of cStorPool object is to create pools.
func IsCStorPoolCreateStatuses(cstorPool *apis.CStorPool) bool {
	status := string(cstorPool.Status.Phase)
	if strings.EqualFold(status, string(apis.CStorPoolStatusInit)) || strings.EqualFold(status, string(apis.CStorPoolStatusCreateFailed)) {
		klog.V(4).Infof("cStorPool status %s for %v", status, string(cstorPool.ObjectMeta.UID))
		return true
	}
	klog.V(4).Infof("cstor pool '%s': uid '%s': phase '%s': is_empty_status: false", string(cstorPool.ObjectMeta.Name), string(cstorPool.ObjectMeta.UID), cstorPool.Status.Phase)
	return false
}

// IsEmptyStatus is to check if the status of cStorPool object is empty.
func IsEmptyStatus(cStorPool *apis.CStorPool) bool {
	if string(cStorPool.Status.Phase) == string(apis.CStorPoolStatusEmpty) {
		klog.Infof("cStorPool empty status: %v", string(cStorPool.ObjectMeta.UID))
		return true
	}
	klog.Infof("cstor pool '%s': uid '%s': phase '%s': is_empty_status: false", string(cStorPool.ObjectMeta.Name), string(cStorPool.ObjectMeta.UID), cStorPool.Status.Phase)
	return false
}

// IsPendingStatus is to check if the status of cStorPool object is pending.
func IsPendingStatus(cStorPool *apis.CStorPool) bool {
	if string(cStorPool.Status.Phase) == string(apis.CStorPoolStatusPending) {
		klog.Infof("cStorPool pending: %v", string(cStorPool.ObjectMeta.UID))
		return true
	}
	klog.V(4).Infof("cstor pool '%s': uid '%s': phase '%s': is_pending_status: false", string(cStorPool.ObjectMeta.Name), string(cStorPool.ObjectMeta.UID), cStorPool.Status.Phase)
	return false
}

// IsErrorDuplicate is to check if the status of cStorPool object is error-duplicate.
func IsErrorDuplicate(cStorPool *apis.CStorPool) bool {
	if string(cStorPool.Status.Phase) == string(apis.CStorPoolStatusErrorDuplicate) {
		klog.Infof("cStorPool duplication error: %v", string(cStorPool.ObjectMeta.UID))
		return true
	}
	klog.V(4).Infof("Not error duplicate status: %v", string(cStorPool.ObjectMeta.UID))
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

// syncCsp updates field on CSP object after fetching the values from zpool utility.
func (c *CStorPoolController) syncCsp(cStorPool *apis.CStorPool) {
	// Get capacity of the pool.
	capacity, err := pool.Capacity(string(pool.PoolPrefix) + string(cStorPool.ObjectMeta.UID))
	if err != nil {
		klog.Errorf("Unable to sync CSP capacity: %v", err)
		c.recorder.Event(cStorPool, corev1.EventTypeWarning, string(common.FailureCapacitySync), string(common.MessageResourceFailCapacitySync))
	} else {
		cStorPool.Status.Capacity = *capacity
		c.updateROMode(cStorPool)
	}
}

func (c *CStorPoolController) updateROMode(cStorPool *apis.CStorPool) {
	capacity := cStorPool.Status.Capacity
	rOThresholdLimit := cStorPool.Spec.PoolSpec.ROThresholdLimit

	qn, err := convertToBytes([]string{capacity.Total, capacity.Free, capacity.Used})
	if err != nil {
		klog.Errorf("Failed to parse capacity.. err=%s", err)
		return
	}

	total, _, used := qn[0], qn[1], qn[2]
	usedCapacity := (used * 100) / total

	if (int(usedCapacity) >= rOThresholdLimit) &&
		(rOThresholdLimit != 0 &&
			rOThresholdLimit != 100) {
		if !cStorPool.Status.ReadOnly {
			if err = pool.SetPoolRDMode(cStorPool, true); err != nil {
				klog.Errorf("Failed to set pool readOnly mode : %v", err)
			} else {
				cStorPool.Status.ReadOnly = true
				c.recorder.Event(cStorPool,
					corev1.EventTypeWarning,
					string(common.PoolROThreshold),
					string(common.MessagePoolROThreshold),
				)
			}
		}
	} else {
		if cStorPool.Status.ReadOnly || !cspROUpdated {
			if err = pool.SetPoolRDMode(cStorPool, false); err != nil {
				klog.Errorf("Failed to unset pool readOnly mode : %v", err)
			} else {
				cStorPool.Status.ReadOnly = false
			}
		}
	}
	return
}

func convertToBytes(a []string) (number []int64, err error) {
	if len(a) == 0 {
		err = errors.New("empty input")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.New("unable to parse")
		}
	}()

	parser := func(s string) int64 {
		d := resource.MustParse(s + "i")
		return d.Value()
	}

	for _, v := range a {
		number = append(number, parser(v))
	}
	return
}

func (c *CStorPoolController) getDeviceIDs(csp *apis.CStorPool) ([]string, error) {
	// TODO: Add error handling
	return pool.GetDeviceIDs(csp)
}

func (c *CStorPoolController) reconcileVersion(csp *apis.CStorPool) (*apis.CStorPool, error) {
	var err error
	// the below code uses deep copy to have the state of object just before
	// any update call is done so that on failure the last state object can be returned
	if csp.VersionDetails.Status.Current != csp.VersionDetails.Desired {
		if !apis.IsCurrentVersionValid(csp.VersionDetails.Status.Current) {
			return csp, errors.Errorf("invalid current version %s", csp.VersionDetails.Status.Current)
		}
		if !apis.IsDesiredVersionValid(csp.VersionDetails.Desired) {
			return csp, errors.Errorf("invalid desired version %s", csp.VersionDetails.Desired)
		}
		cspObj := csp.DeepCopy()
		if csp.VersionDetails.Status.State != apis.ReconcileInProgress {
			cspObj.VersionDetails.Status.SetInProgressStatus()
			cspObj, err = c.clientset.OpenebsV1alpha1().CStorPools().Update(cspObj)
			if err != nil {
				return csp, err
			}
		}
		path := strings.Split(cspObj.VersionDetails.Status.Current, "-")[0]
		u := &upgradeParams{
			csp:    cspObj,
			client: c.clientset,
		}
		// Get upgrade function for corresponding path, if path does not
		// exits then no upgrade is required and funcValue will be nil.
		funcValue := upgradeMap[path]
		if funcValue != nil {
			cspObj, err = funcValue(u)
			if err != nil {
				return cspObj, err
			}
		}
		csp = cspObj.DeepCopy()
		cspObj.VersionDetails.SetSuccessStatus()
		cspObj, err = c.clientset.OpenebsV1alpha1().CStorPools().Update(cspObj)
		if err != nil {
			return csp, errors.Wrap(err, "failed to update csp with versionDetails")
		}
		return cspObj, nil
	}
	return csp, nil
}

// populateVersion assigns VersionDetails for old csp object
func (c *CStorPoolController) populateVersion(csp *apis.CStorPool) (
	*apis.CStorPool, error,
) {
	var err error
	v := csp.Labels[string(apis.OpenEBSVersionKey)]
	// 1.3.0 onwards new CSP will have the field populated during creation
	if v < "1.3.0" && csp.VersionDetails.Status.Current == "" {
		cspObj := csp.DeepCopy()
		cspObj.VersionDetails.Status.Current = v
		cspObj.VersionDetails.Desired = v
		cspObj, err = c.clientset.OpenebsV1alpha1().CStorPools().
			Update(cspObj)

		if err != nil {
			return csp, errors.Wrap(err, "failed to update csp while adding versiondetails")
		}
		klog.Infof("Version %s added on csp %s", v, cspObj.Name)
		return cspObj, nil
	}
	return csp, nil
}
