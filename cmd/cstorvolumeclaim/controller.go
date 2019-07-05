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

package cstorvolumeclaim

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	merrors "github.com/openebs/maya/pkg/errors/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	ref "k8s.io/client-go/tools/reference"
	"k8s.io/kubernetes/pkg/util/slice"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a
	// cstorvolumeclaim is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a
	// cstorvolumeclaim fails to sync due to a cstorvolumeclaim of the same
	// name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a cstorvolumeclaim already existing
	MessageResourceExists = "Resource %q already exists and is not managed by CVC"
	// MessageResourceSynced is the message used for an Event fired when a
	// cstorvolumeclaim is synced successfully
	MessageResourceSynced = "cstorvolumeclaim synced successfully"

	// CStorVolumeClaimFinalizer name of finalizer on CStorVolumeClaim that
	// are bound by CStorVolume
	CStorVolumeClaimFinalizer = "cvc.openebs.io/finalizer"
)

// Patch struct represent the struct used to patch
// the cstorvolumeclaim object
type Patch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the spcPoolUpdated resource
// with the current status of the resource.
func (c *CVCController) syncHandler(key string) error {
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
func (c *CVCController) enqueueCVC(obj interface{}) {
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
func (c *CVCController) syncCVC(cvc *apis.CStorVolumeClaim) error {
	//	var newCVCLease Leaser
	//	newCVCLease = &Lease{cvc, cvcLeaseKey, c.clientset, c.kubeclientset}
	//	err := newCVCLease.Hold()
	//	if err != nil {
	//		return errors.Wrapf(err, "Could not acquire lease on cvc object")
	//	}
	//	glog.V(4).Infof("Lease acquired successfully on CStorVolumeClaims %s ", spc.Name)
	//	defer newCVCLease.Release()

	// CStor Volume Claim should be deleted. Check if deletion timestamp is set
	// and remove finalizer.
	if c.isClaimDeletionCandidate(cvc) {
		glog.Infof("syncClaim: remove finalizer for CStorVolumeClaimVolume [%s]", cvc.Name)
		return c.removeClaimFinalizer(cvc)
	}

	volName := cvc.Name
	if volName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%+v: cvc name must be specified", cvc))
		return nil
	}

	//NodeId indicates where the volume is needed to be mounted, i.e the node
	// where the app has been scheduled.
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

	// IsCVRPending checks for the pending cstorvolume replica and requeue the
	// create operation if count doesn't matches the desired count
	pending, err := c.IsCVRPending(cvc)
	if err != nil {
		return err
	}
	if pending {
		glog.Infof("create remaining volume replica %+v", cvc)
		_, err = c.createVolumeOperation(cvc)
	}

	if err != nil {
		return err
	}

	// Finally, we update the status block of the CVC resource to reflect the
	// current state of the world
	c.recorder.Event(cvc, corev1.EventTypeNormal,
		SuccessSynced,
		MessageResourceSynced,
	)
	return nil
}

// UpdateCVCObj updates the cstorvolumeclaim object resource to reflect the
// current state of the world
func (c *CVCController) updateCVCObj(
	cvc *apis.CStorVolumeClaim,
	cv *apis.CStorVolume,
) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	cvcCopy := cvc.DeepCopy()
	if cvc.Name != cv.Name {
		return fmt.
			Errorf("could not bind cstorvolumeclaim %s and cstorvolume %s, name does not match",
				cvc.Name,
				cv.Name)
	}

	_, err := c.clientset.OpenebsV1alpha1().CStorVolumeClaims(cvc.Namespace).Update(cvcCopy)
	return err
}

// createVolumeOperation trigers the all required resource create operation.
// 1. Create volume service.
// 2. Create cstorvolume resource with required iscsi information.
// 3. Create target deployment.
// 4. Create cstorvolumeclaim resource.
// 5. Update the cstorvolumeclaim with claimRef info and bound with cstorvolume.
func (c *CVCController) createVolumeOperation(cvc *apis.CStorVolumeClaim) (*apis.CStorVolumeClaim, error) {

	scName := cvc.Annotations[string(apis.StorageConfigClassKey)]
	scObj, err := getStorageClass(scName)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("creating cstorvolume service resource")
	svcObj, err := getOrCreateTargetService(scName, cvc)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("creating cstorvolume resource")
	cvObj, err := getOrCreateCStorVolumeResource(svcObj, cvc, scObj)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("creating cstorvolume target deployment")
	_, err = getOrCreateCStorTargetDeployment(cvObj)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("creating cstorvolume replica resource")
	err = c.distributePendingCVRs(cvc, cvObj, svcObj, scObj)
	if err != nil {
		return nil, err
	}

	volumeRef, err := ref.GetReference(scheme.Scheme, cvObj)
	if err != nil {
		return nil, err
	}
	cvc.Spec.CStorVolumeRef = volumeRef
	cvc.Status.Phase = "Bound"

	err = c.updateCVCObj(cvc, cvObj)
	if err != nil {
		return nil, err
	}
	return cvc, nil
}

// distributePendingCVRs trigers create and distribute pending cstorvolumereplica
// resource among the available cstor pools
func (c *CVCController) distributePendingCVRs(
	cvc *apis.CStorVolumeClaim,
	cv *apis.CStorVolume,
	service *corev1.Service,
	class *storagev1.StorageClass,
) error {

	desiredReplicaCount, err := c.getPendingCVRCount(cvc, class)
	if err != nil {
		return err
	}
	err = distributeCVRs(desiredReplicaCount, service, cv, class)
	if err != nil {
		return err
	}
	return nil
}

// isClaimDeletionCandidate checks if a cstorvolumeclaim is a deletion candidate.
func (c *CVCController) isClaimDeletionCandidate(cvc *apis.CStorVolumeClaim) bool {
	return cvc.ObjectMeta.DeletionTimestamp != nil &&
		slice.ContainsString(cvc.ObjectMeta.Finalizers, CStorVolumeClaimFinalizer, nil)
}

// removeFinalizer removes finalizers present in CStorVolumeClaim resource
func (c *CVCController) removeClaimFinalizer(
	cvc *apis.CStorVolumeClaim,
) error {
	cvcPatch := []Patch{
		Patch{
			Op:   "remove",
			Path: "/metadata/finalizers",
		},
	}

	cvcPatchBytes, err := json.Marshal(cvcPatch)
	if err != nil {
		return merrors.Wrapf(
			err,
			"failed to remove finalizers from cstorvolumeclaim {%s}",
			cvc.Name,
		)
	}

	_, err = c.clientset.
		OpenebsV1alpha1().
		CStorVolumeClaims(cvc.Namespace).
		Patch(cvc.Name, types.JSONPatchType, cvcPatchBytes)
	if err != nil {
		return merrors.Wrapf(
			err,
			"failed to remove finalizers from cstorvolumeclaim {%s}",
			cvc.Name,
		)
	}
	glog.Infof("finalizers removed successfully from cstorvolumeclaim {%s}", cvc.Name)
	return nil
}

// getPendingCVRCount gets the pending replica count to be created
// in case of any failures
func (c *CVCController) getPendingCVRCount(
	cvc *apis.CStorVolumeClaim,
	class *storagev1.StorageClass,
) (int, error) {

	desiredReplicaCount, err := getReplicationFactor(class)
	if err != nil {
		return 0, err
	}

	currentReplicaCount, err := c.getCurrentReplicaCount(cvc)
	if err != nil {
		runtime.HandleError(err)
		return 0, err
	}
	return desiredReplicaCount - currentReplicaCount, nil
}

// getCurrentReplicaCount give the current cstorvolumereplicas count for the
// given volume.
func (c *CVCController) getCurrentReplicaCount(cvc *apis.CStorVolumeClaim) (int, error) {
	// TODO use lister
	//	CVRs, err := c.cvrLister.CStorVolumeReplicas(cvc.Namespace).
	//		List(klabels.Set(pvLabel).AsSelector())

	pvLabel := pvAnnotaion + cvc.Name

	cvrList, err := c.clientset.
		OpenebsV1alpha1().
		CStorVolumeReplicas(cvc.Namespace).
		List(metav1.ListOptions{LabelSelector: pvLabel})

	if err != nil {
		return 0, merrors.Errorf("unable to get current replica count: %v", err)
	}
	return len(cvrList.Items), nil
}

// IsCVRPending look for pending cstorvolume replicas compared to desired
// replica count. returns true if count doesn't matches.
func (c *CVCController) IsCVRPending(cvc *apis.CStorVolumeClaim) (bool, error) {
	rCount := cvc.Annotations["openebs.io/replicaCount"]
	desiredReplicaCount, err := strconv.Atoi(rCount)
	if err != nil {
		return false, err
	}
	selector := klabels.SelectorFromSet(BaseLabels(cvc))
	CVRs, err := c.cvrLister.CStorVolumeReplicas(cvc.Namespace).
		List(selector)
	if err != nil {
		return false, merrors.Errorf("failed to list cvr : %v", err)
	}
	// TODO: check for greater values
	return desiredReplicaCount != len(CVRs), nil
}

// BaseLabels returns the base labels we apply to cstorvolumereplicas created
func BaseLabels(cvc *apis.CStorVolumeClaim) map[string]string {
	base := map[string]string{
		pvAnnotaion: cvc.Name,
	}
	return base
}
