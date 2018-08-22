/*
Copyright 2017 The OpenEBS Authors

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
	"github.com/openebs/maya/cmd/maya-apiserver/spc-actions"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"github.com/openebs/maya/pkg/client/k8s"
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
		// CreateStoragePool function will create the storage pool
		// It is a create event so resync should be false and sparepoolcount is passed 0
		// sparepoolcount is not used when resync is false.
		err := storagepoolactions.CreateStoragePool(spcGot, false, 0)

		if err != nil {
			glog.Error("Storagepool could not be created:", err)
			// To-Do
			// If Some error occur patch the spc object with appropriate reason
		}

		return addEvent, err
		break

	case updateEvent:
		// TO-DO : Handle Business Logic
		// Hook Update Business Logic Here
		return updateEvent, nil
		break
	case syncEvent:
		err := syncSpc(spcGot)
		if err != nil {
			glog.Errorf("Storagepool %s could not be synced:%v", spcGot.Name, err)
		}
		return syncEvent, nil
		break
	case deleteEvent:
		err := storagepoolactions.DeleteStoragePool(spcGot)

		if err != nil {
			glog.Error("Storagepool could not be deleted:", err)
		}

		return deleteEvent, err
		break
	default:
		// opeartion with tag other than add,update and delete are ignored.
		break
	}
	return ignoreEvent, nil
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

func syncSpc(spcGot *apis.StoragePoolClaim) (error) {
	glog.Infof("Syncing storagepoolclaim %s", spcGot.Name)
	// Get kubernetes clientset
	// namespaces is not required, hence passed empty.
	newK8sClient, err := k8s.NewK8sClient("")
	if err != nil {
		return err
	}
	// Get openebs clientset using a getter method (i.e. GetOECS() ) as
	// the openebs clientset is not exported.
	newOecsClient := newK8sClient.GetOECS()

	// Get the current count of provisione pool for the storagepool claim
	cspList, err := newOecsClient.OpenebsV1alpha1().CStorPools().List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spcGot.Name})
	currentPoolCount := len(cspList.Items)

	// If current pool count is less than maxpool count, try to converge to maxpool
	if (currentPoolCount < int(spcGot.Spec.MaxPools)) {
		glog.Infof("Converging storagepoolclaim %s to desired state:current pool count is %d,desired pool count is %d", spcGot.Name, currentPoolCount, spcGot.Spec.MaxPools)
		// sparePoolCount holds the spared pool that should be provisioned to get the desired state.
		sparePoolCount := int(spcGot.Spec.MaxPools) - currentPoolCount
		// Call the storage pool create logic to proviison the spare pools.
		err := storagepoolactions.CreateStoragePool(spcGot, true, sparePoolCount)
		if err != nil {
			return err
		}
	}
	return nil
}
