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
	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"strings"
)

const (
	// DEFAULTPOOlCOUNT will be set to maxPool field of spc if maxPool field is not provided.
	DEFAULTPOOlCOUNT = 3
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
			return errors.New("storagepoolclaim object not found for storage pool deletion")
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
		err := c.addEventHandler(spcGot, false)
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
		runtime.HandleError(errors.Wrapf(err, "Invalid resource key: %s", key))
		return nil, err
	}
	spcGot, err := c.clientset.OpenebsV1alpha1().StoragePoolClaims().Get(name, metav1.GetOptions{})
	if err != nil {
		// The SPC resource may no longer exist, in which case we stop
		// processing.
		if k8serror.IsNotFound(err) {
			runtime.HandleError(errors.Wrapf(err, "spcGot '%s' in work queue no longer exists", key))
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
	glog.V(1).Infof("Reconciling storagepoolclaim %s", spcGot.Name)
	err := c.addEventHandler(spcGot, true)
	return err
}

// addEventHandler is the event handler for the add event of spc.
func (c *Controller) addEventHandler(spc *apis.StoragePoolClaim, resync bool) error {
	mutateSpc, err := validate(spc)
	if err != nil {
		return err
	}
	// validate can mutate spc object -- for example if maxPool field is not present in case
	// of auto provisioning, maxPool will default to 3.
	// We need to immediately patch SPC object here.
	if mutateSpc && !resync {
		spc, err = c.clientset.OpenebsV1alpha1().StoragePoolClaims().Update(spc)
		if err != nil {
			return errors.Wrap(err, "spc patch for defaulting the field(s) failed")
		}
	}
	pendingPoolCount, err := c.getPendingPoolCount(spc)
	if err != nil {
		return err
	}
	err = c.create(pendingPoolCount, spc)
	return nil
}

// create is a wrapper function that calls the actual function to create pool as many time
// as the number of pools need to be created.
func (c *Controller) create(pendingPoolCount int, spc *apis.StoragePoolClaim) error {
	var newSpcLease Leaser
	newSpcLease = &Lease{spc, SpcLeaseKey, c.clientset, c.kubeclientset}
	err := newSpcLease.Hold()
	if err != nil {
		return errors.Wrapf(err, "Could not acquire lease on spc object")
	}
	glog.Infof("Lease acquired successfully on storagepoolclaim %s ", spc.Name)
	defer newSpcLease.Release()
	for poolCount := 1; poolCount <= pendingPoolCount; poolCount++ {
		glog.Infof("Provisioning pool %d/%d for storagepoolclaim %s", poolCount, pendingPoolCount, spc.Name)
		err = CreateStoragePool(spc)
		if err != nil {
			runtime.HandleError(errors.Wrapf(err, "Pool provisioning failed for %d/%d for storagepoolclaim %s", poolCount, pendingPoolCount, spc.Name))
		}
	}
	return nil
}

// validate validates the spc configuration before creation of pool.
func validate(spc *apis.StoragePoolClaim) (bool, error) {
	// ToDo: Move these to admission webhook plugins or CRD validations
	// Validations for poolType
	var mutateSpc bool
	if len(spc.Spec.Disks.DiskList) == 0 {
		maxPools := spc.Spec.MaxPools
		if maxPools == 0 {
			spc.Spec.MaxPools = DEFAULTPOOlCOUNT
			mutateSpc = true
		}
		if maxPools < 0 {
			return mutateSpc, errors.Errorf("aborting storagepool create operation for %s as invalid maxPool value %d", spc.Name, maxPools)
		}
	}
	poolType := spc.Spec.PoolSpec.PoolType
	if poolType == "" {
		return mutateSpc, errors.Errorf("aborting storagepool create operation for %s as no poolType is specified", spc.Name)
	}

	if !(poolType == string(apis.PoolTypeStripedCPV) || poolType == string(apis.PoolTypeMirroredCPV)) {
		return mutateSpc, errors.Errorf("aborting storagepool create operation as specified poolType is %s which is invalid", poolType)
	}

	diskType := spc.Spec.Type
	if !(diskType == string(apis.TypeSparseCPV) || diskType == string(apis.TypeDiskCPV)) {
		return mutateSpc, errors.Errorf("aborting storagepool create operation as specified type is %s which is invalid", diskType)
	}
	return mutateSpc, nil
}

// getCurrentPoolCount give the current pool count for the given auto provisioned spc.
func (c *Controller) getCurrentPoolCount(spc *apis.StoragePoolClaim) (int, error) {
	// Get the current count of provisioned pool for the storagepool claim
	spList, err := c.clientset.OpenebsV1alpha1().StoragePools().List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spc.Name})
	if err != nil {
		return 0, errors.Errorf("unable to get current pool count:unable to list storagepools: %v", err)
	}
	return len(spList.Items), nil
}

// getPendingPoolCount gives the count of pool that needs to be provisioned for a given spc.
func (c *Controller) getPendingPoolCount(spc *apis.StoragePoolClaim) (int, error) {
	var pendingPoolCount int
	if len(spc.Spec.Disks.DiskList) == 0 {
		currentPoolCount, err := c.getCurrentPoolCount(spc)
		if err != nil {
			return 0, err
		}
		maxPoolCount := int(spc.Spec.MaxPools)
		pendingPoolCount = maxPoolCount - currentPoolCount
	} else {
		// ToDo: -- Refactor using disk refernces to find the used disks
		// allNodeDiskMap holds the map : diskName --> hostName ; for all the disks.
		allNodeDiskMap := make(map[string]string)
		// Get all disk in one shot from kube-apiserver
		diskList, err := c.clientset.OpenebsV1alpha1().Disks().List(metav1.ListOptions{})
		if err != nil {
			return 0, err
		}
		newClientSet := &clientSet{
			c.clientset,
		}
		// Get used disk for the SPC
		// usedDiskMap holds the disk which are already used.
		usedDiskMap, err := newClientSet.getUsedDiskMap()
		for _, disk := range diskList.Items {
			if usedDiskMap[disk.Name] == 1 {
				continue
			}
			allNodeDiskMap[disk.Name] = disk.Labels[string(apis.HostNameCPK)]
		}
		// nodeCountMap holds the node names as the key over which pool should be provisioned.
		nodeCountMap := make(map[string]int)
		for _, spcDisk := range spc.Spec.Disks.DiskList {
			if !(len(strings.TrimSpace(allNodeDiskMap[spcDisk])) == 0) {
				nodeCountMap[allNodeDiskMap[spcDisk]]++
			}
		}
		pendingPoolCount = len(nodeCountMap)
	}
	if pendingPoolCount < 0 {
		return 0, errors.Errorf("Got invalid pending pool count %d", pendingPoolCount)
	}
	return pendingPoolCount, nil
}
