/*
Copyright 2018 The OpenEBS Authors

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

package spc

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/openebs/CITF/pkg/apis/openebs.io/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the spcPoolUpdated resource
// with the current status of the resource.
func (c *Controller) syncHandler(key, operation string, object interface{}) error {
	// getSpcResource will take a key as argument which contains the namespace/name or simply name
	// of the object and will fetch the object.
	spcGot, err := c.getSpcResource(key)
	if err != nil {
		return err
	}
	// Check if the event is for delete and use the spc object that was pushed in the queue
	// for utilising details from it e.g. delete cas template name for storagepool deletion.
	if operation == deleteEvent {
		// Need to typecast the interface object to storagepoolclaim object because
		// interface type of nil is different from nil but all other type of nil has the same type as that of nil.
		spcObject := object.(*apis.StoragePoolClaim)
		if spcObject == nil {
			return fmt.Errorf("storagepoolclaim object not found for storage pool deletion")
		}
		spcGot = spcObject
	}

	// Call the spcEventHandler which will take spc object , key(namespace/name of object) and type of operation we need to to for storage pool
	// Type of operation for storage pool e.g. create, delete etc.
	events, err := c.spcEventHandler(operation, spcGot)
	if events == ignoreEvent {
		glog.Warning("None of the SPC handler was executed")
		return nil
	}
	if err != nil {
		return err
	}
	// If this function returns a error then the object will be requeued.
	// No need to error out even if it occurs,
	return nil
}

// spcPoolEventHandler is to handle SPC related events.
func (c *Controller) spcEventHandler(operation string, spcGot *apis.StoragePoolClaim) (string, error) {
	switch operation {
	case addEvent:
		// Call addEventHandler in case of add event.
		err := c.addEventHandler(spcGot)
		return addEvent, err

	case updateEvent:
		// TO-DO : Handle Business Logic
		// Hook Update Business Logic Here
		return updateEvent, nil
	case syncEvent:
		err := c.syncSpc(spcGot)
		if err != nil {
			glog.Errorf("Storagepool %s could not be synced:%v", spcGot.Name, err)
		}
		return syncEvent, nil
	case deleteEvent:
		err := DeleteStoragePool(spcGot)

		if err != nil {
			glog.Error("Storagepool could not be deleted:", err)
		}

		return deleteEvent, err
	default:
		// operation with tag other than add,update and delete are ignored.
		return ignoreEvent, nil
	}
}

// enqueueSpc takes a SPC resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than SPC.
func (c *Controller) enqueueSpc(queueLoad *QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(queueLoad.Object); err != nil {
		runtime.HandleError(err)
		return
	}
	queueLoad.Key = key
	c.workqueue.AddRateLimited(queueLoad)
}

// getSpcResource returns object corresponding to the resource key
func (c *Controller) getSpcResource(key string) (*apis.StoragePoolClaim, error) {
	// Convert the key(namespace/name) string into a distinct name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("Invalid resource key: %s", key))
		return nil, err
	}
	spcGot, err := c.clientset.OpenebsV1alpha1().StoragePoolClaims().Get(name, metav1.GetOptions{})
	if err != nil {
		// The SPC resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("spcGot '%s' in work queue no longer exists:'%v'", key, err))
			// No need to return error to caller as we still want to fire the delete handler
			// using the spc key(name)
			// If error is returned the caller function will return without calling the spcEventHandler
			// function that invokes business logic for pool deletion
			return nil, nil
		}
		return nil, err
	}
	return spcGot, nil
}

// synSpc is the function which tries to converge to a desired state for the spc.
func (c *Controller) syncSpc(spcGot *apis.StoragePoolClaim) error {
	glog.V(1).Infof("Syncing storagepoolclaim %s", spcGot.Name)
	currentPoolCount, err := c.getCurrentPoolCount(spcGot)
	if err != nil {
		return err
	}
	maxPoolCount := int(spcGot.Spec.MaxPools)
	// If current pool count is less than maxPool count, try to converge to maxPool.
	if currentPoolCount < maxPoolCount {
		glog.Infof("Converging storagepoolclaim %s to desired state:current pool count is %d,desired pool count is %d", spcGot.Name, currentPoolCount, spcGot.Spec.MaxPools)
		// pendingPoolCount holds the pending pool that should be provisioned to get the desired state.
		pendingPoolCount := maxPoolCount - currentPoolCount
		// Call the storage pool create logic to provision the pending pools.
		err = c.storagePoolCreateWrapper(pendingPoolCount, spcGot)
		if err != nil {
			return err
		}
	}
	return nil
}

// addEventHandler is the event handler for the add event of spc.
func (c *Controller) addEventHandler(spc *apis.StoragePoolClaim) error {
	err := storagePoolValidator(spc)
	if err != nil {
		return err
	}
	currentPoolCount, err := c.getCurrentPoolCount(spc)
	if err != nil {
		return err
	}
	err = c.storagePoolCreateWrapper(spc.Spec.MaxPools-currentPoolCount, spc)
	if err != nil {
		return err
	}
	return nil
}

// storagePoolCreateWrapper is a wrapper function that calls the actual function to create pool as many time
// as the number of pools need to be created.
func (c *Controller) storagePoolCreateWrapper(maxPool int, spc *apis.StoragePoolClaim) error {
	var newSpcLease Leaser
	newSpcLease = &Lease{spc, SpcLeaseKey, c.clientset, c.kubeclientset}
	err := newSpcLease.Hold()
	if err != nil {
		return fmt.Errorf("Could not acquire lease on spc object:%v", err)
	}
	glog.Infof("Lease acquired successfully on storagepoolclaim %s ", spc.Name)
	defer newSpcLease.Release()
	for poolCount := 1; poolCount <= maxPool; poolCount++ {
		glog.Infof("Provisioning pool %d/%d for storagepoolclaim %s", poolCount, maxPool, spc.Name)
		err = CreateStoragePool(spc)
		if err != nil {
			glog.Errorf("Pool provisioning failed for %d/%d for storagepoolclaim %s", poolCount, maxPool, spc.Name)
		}
	}
	return nil
}

// storagePoolValidator validates the spc configuration before creation of pool.
func storagePoolValidator(spc *apis.StoragePoolClaim) error {
	// Validations for poolType
	if spc.Spec.MaxPools <= 0 {
		return fmt.Errorf("aborting storagepool create operation for %s as maxPool count is invalid ", spc.Name)
	}
	poolType := spc.Spec.PoolSpec.PoolType
	if poolType == "" {
		return fmt.Errorf("aborting storagepool create operation for %s as no poolType is specified", spc.Name)
	}

	if !(poolType == string(v1alpha1.PoolTypeStripedCPV) || poolType == string(v1alpha1.PoolTypeMirroredCPV)) {
		return fmt.Errorf("aborting storagepool create operation as specified poolType is %s which is invalid", poolType)
	}

	diskType := spc.Spec.Type
	if !(diskType == string(v1alpha1.TypeSparseCPV) || diskType == string(v1alpha1.TypeDiskCPV)) {
		return fmt.Errorf("aborting storagepool create operation as specified type is %s which is invalid", diskType)
	}
	return nil
}

// getCurrentPoolCount give the current pool count for the given spc.
func (c *Controller) getCurrentPoolCount(spc *apis.StoragePoolClaim) (int, error) {
	// Get the current count of provisioned pool for the storagepool claim
	spList, err := c.clientset.OpenebsV1alpha1().StoragePools().List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spc.Name})
	if err != nil {
		return 0, fmt.Errorf("unable to get current pool count:unable to list storagepools: %v", err)
	}
	return len(spList.Items), nil
}
