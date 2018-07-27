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
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the spcPoolUpdated resource
// with the current status of the resource.
func (c *Controller) syncHandler(key, operation string) error {
	// getSpcResource will take a key as argument which conatins the namespace/name or simply name
	// of the object and will fetch the object.
	spcGot, err := c.getSpcResource(key)
	if err != nil {
		return err
	}
	// Call the spcEventHandler which will take spc object , key(namespace/name of object) and type of operation we need to to for storage pool
	// Type of operation for storage pool e.g. create, delete etc.
	events, err := c.spcEventHandler(operation, spcGot, key)
	if events == ignoreEvent {
		glog.Warning("None of the SPC handler was executed")
		return nil
	}
	// If this function returns a error then the object will be requeued.
	// No need to error out even if it occurs,
	return nil
}

// spcPoolEventHandler is to handle SPC related events.
func (c *Controller) spcEventHandler(operation string, spcGot *apis.StoragePoolClaim, key string) (string, error) {
	switch operation {
	case addEvent:
		// CreateStoragePool function will create the storage pool
		err := storagepoolactions.CreateStoragePool(spcGot)

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

	case deleteEvent:
		err := storagepoolactions.DeleteStoragePool(key)

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
func (c *Controller) enqueueSpc(obj interface{}, q QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	q.Key = key
	c.workqueue.AddRateLimited(q)
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

			runtime.HandleError(fmt.Errorf("spcGot '%s' in work queue no longer exists", key))
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
