/*
Copyright 2019 The OpenEBS Authors

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

package cvc

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"

	openebs "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	ref "k8s.io/client-go/tools/reference"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a CVC already existing
	MessageResourceExists = "Resource %q already exists and is not managed by CVC"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "CVC synced successfully"
)

type clientSet struct {
	oecs openebs.Interface
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the spcPoolUpdated resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	startTime := time.Now()
	glog.V(4).Infof("Started syncing cstorvolumeclaim %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing cstorvolumeclaim %q (%v)", key, time.Since(startTime))
	}()

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the cvc resource with this namespace/name
	cvc, err := c.cvcLister.CStorVolumeClaims(namespace).Get(name)
	if k8serror.IsNotFound(err) {
		runtime.HandleError(fmt.Errorf("cstorvolumeclaim '%s' has been deleted", key))
		return nil
	}
	if err != nil {
		return err
	}
	cvcCopy := cvc.DeepCopy()
	err = c.syncCVC(cvcCopy)
	return err
}

// enqueueCVC takes a CVC resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CStorVolumeClaims.
func (c *Controller) enqueueCVC(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(key)

	/*	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
			obj = unknown.Obj
		}
		if cvc, ok := obj.(*apis.CStorVolumeClaim); ok {
			objName, err := cache.DeletionHandlingMetaNamespaceKeyFunc(cvc)
			if err != nil {
				glog.Errorf("failed to get key from object: %v, %v", err, cvc)
				return
			}
			glog.V(5).Infof("enqueued %q for sync", objName)
			c.workqueue.Add(objName)
		}
	*/
}

// synCVC is the function which tries to converge to a desired state for the
// CStorVolumeClaims
func (c *Controller) syncCVC(cvc *apis.CStorVolumeClaim) error {
	err := c.create(cvc)
	if err != nil {
		return err
	}
	return nil
}

// create is a wrapper function that calls the actual function to create pool as many time
// as the number of pools need to be created.
func (c *Controller) create(cvc *apis.CStorVolumeClaim) error {
	//	var newCVCLease Leaser
	//	newCVCLease = &Lease{cvc, cvcLeaseKey, c.clientset, c.kubeclientset}
	//	err := newCVCLease.Hold()
	//	if err != nil {
	//		return errors.Wrapf(err, "Could not acquire lease on cvc object")
	//	}
	//	glog.V(4).Infof("Lease acquired successfully on CStorVolumeClaims %s ", spc.Name)
	//	defer newCVCLease.Release()

	volName := cvc.Name
	if volName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%+v: cvc name must be specified", cvc))
		return nil
	}

	nodeID := cvc.Publish.NodeId
	if nodeID == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%v: cvc must be publish/attached to Node", cvc))
		return nil
	}

	// Get the cstorvolume with the name specified, if not found create all the
	// required resources
	_, err := c.cvLister.CStorVolumes(cvc.Namespace).Get(volName)
	if k8serror.IsNotFound(err) {
		glog.Infof("create cstor based volume using cvc %+v", cvc)
		_, err = c.createVolumeOperation(cvc)
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the CVC resource to reflect the
	// current state of the world
	//err = c.updateCVCStatus(cvc)
	if err != nil {
		return err
	}

	c.recorder.Event(cvc, corev1.EventTypeNormal,
		SuccessSynced,
		MessageResourceSynced,
	)
	return nil
}

// UpdateCVCStatus updates the status block of the CVC resource to reflect the
// current state of the world
func (c *Controller) updateCVCStatus(cvc *apis.CStorVolumeClaim,
	cv *apis.CStorVolume,
) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	cvcCopy := cvc.DeepCopy()
	if cvc.Name != cv.Name {
		return fmt.
			Errorf("could not bind cstorvolumeclaim %s and cstorvolume %s, name does not match", cvc.Name, cv.Name)
	}
	cvcCopy.Status.Phase = "Bound"
	_, err := c.clientset.Openebs().CStorVolumeClaims(cvc.Namespace).Update(cvcCopy)
	return err
}

//createVolumeOperation ...
func (c *Controller) createVolumeOperation(cvc *apis.CStorVolumeClaim) (*apis.CStorVolumeClaim, error) {

	glog.Infof("creating cstorvolume service resource")
	scName := cvc.Annotations[string(apis.StorageConfigClassKey)]
	svcObj, err := createTargetService(scName, cvc)
	if err != nil {
		return nil, err
	}

	glog.Infof("creating cstorvolume resource")
	cvObj, err := createCStorVolumecr(svcObj, cvc, scName)
	if err != nil {
		glog.Infof("creating cstorvolume : %s", err)
		return nil, err
	}

	glog.Infof("creating cstorvolume target deployment")
	_, err = createCStorTargetDeployment(cvObj)
	if err != nil {
		return nil, err
	}

	glog.Infof("creating cstorvolume replica resource")
	_, err = createCStorVolumeReplica(svcObj, cvObj, scName)
	if err != nil {
		return nil, err
	}

	volumeRef, err := ref.GetReference(scheme.Scheme, cvObj)
	if err != nil {
		return nil, err
	}
	cvc.Spec.CStorVolumeRef = volumeRef

	err = c.updateCVCStatus(cvc, cvObj)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// isCVCPending get the state if cstorvolume claim.
func (c *Controller) isCVCPending(cvc *apis.CStorVolumeClaim) bool {
	//status, err := c.getCVCStatus(cvc)
	//if err != nil {
	//	glog.Errorf("Unable to get status of cvc %s:%s", cvc.Name, err)
	//	return false
	//}
	if cvc.Status.Phase == "Pending" {
		return false
	}
	return true
}
