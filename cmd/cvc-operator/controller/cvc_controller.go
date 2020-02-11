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
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apispdb "github.com/openebs/maya/pkg/kubernetes/poddisruptionbudget"
	errors "github.com/pkg/errors"
	"k8s.io/klog"

	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	cvclaim "github.com/openebs/maya/pkg/cstorvolumeclaim/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	ref "k8s.io/client-go/tools/reference"
	"k8s.io/kubernetes/pkg/util/slice"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a
	// cstorvolumeclaim is synced
	SuccessSynced = "Synced"
	// Provisioning is used as part of the Event 'reason' when a
	// cstorvolumeclaim is in provisioning stage
	Provisioning = "Provisioning"
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
	// MessageResourceCreated msg used for cstor volume provisioning success event
	MessageResourceCreated = "cstorvolumeclaim created successfully"
	// MessageCVCPublished msg used for cstor volume provisioning publish events
	MessageCVCPublished = "cstorvolumeclaim %q must be published/attached on node"
	// CStorVolumeClaimFinalizer name of finalizer on CStorVolumeClaim that
	// are bound by CStorVolume
	CStorVolumeClaimFinalizer = "cvc.openebs.io/finalizer"
	// DeProvisioning is used as part of the event 'reason' during
	// cstorvolumeclaim deprovisioning stage
	DeProvisioning = "DeProvisioning"
)

var knownResizeConditions = map[apis.CStorVolumeClaimConditionType]bool{
	apis.CStorVolumeClaimResizing:      true,
	apis.CStorVolumeClaimResizePending: true,
}

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
	klog.V(4).Infof("Started syncing cstorvolumeclaim %q (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing cstorvolumeclaim %q (%v)", key, time.Since(startTime))
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
				klog.Errorf("failed to get key from object: %v, %v", err, cvc)
				return
			}
			klog.V(5).Infof("enqueued %q for sync", objName)
			c.workqueue.Add(objName)
		}
	*/
}

// synCVC is the function which tries to converge to a desired state for the
// CStorVolumeClaims
func (c *CVCController) syncCVC(cvc *apis.CStorVolumeClaim) error {
	var err error
	// CStor Volume Claim should be deleted. Check if deletion timestamp is set
	// and remove finalizer.
	if c.isClaimDeletionCandidate(cvc) {
		klog.Infof("syncClaim: remove finalizer for CStorVolumeClaimVolume [%s]", cvc.Name)
		err = c.removeClaimFinalizer(cvc)
		if err != nil {
			c.recorder.Eventf(cvc, corev1.EventTypeWarning, DeProvisioning, err.Error())
		}
		return nil
	}

	volName := cvc.Name
	if volName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%+v: cvc name must be specified", cvc))
		return nil
	}

	nodeID := cvc.Publish.NodeID
	if nodeID == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("cvc must be publish/attached to Node: %+v", cvc))
		c.recorder.Event(cvc, corev1.EventTypeWarning,
			Provisioning,
			fmt.Sprintf(MessageCVCPublished, cvc.Name),
		)
		return nil
	}

	if cvc.Status.Phase == apis.CStorVolumeClaimPhasePending {
		klog.V(2).Infof("provisioning cstor volume %+v", cvc)
		_, err = c.createVolumeOperation(cvc)
		if err != nil {
			//Record an event to indicate that any provisioning operation is failed.
			c.recorder.Eventf(cvc, corev1.EventTypeWarning, Provisioning, err.Error())
		}
	}
	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	if c.cvcNeedResize(cvc) {
		err = c.resizeCVC(cvc)
	}
	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	if c.isCVCScalePending(cvc) {
		// process scalingup/scalingdown of volume replicas only if there is
		// change in curent and desired state of replicas pool information
		err = c.scaleVolumeReplicas(cvc)
	}
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

	if err == nil {
		c.recorder.Event(cvc, corev1.EventTypeNormal,
			SuccessSynced,
			MessageResourceCreated,
		)
	}
	return err
}

// createVolumeOperation trigers the all required resource create operation.
// 1. Create volume service.
// 2. Create cstorvolume resource with required iscsi information.
// 3. Create target deployment.
// 4. Create cstorvolumeclaim resource.
// 5. Create PDB provisioning volume is HA volume.
// 6. Update the cstorvolumeclaim with claimRef info, PDB label(only for HA
//    volumes) and bound with cstorvolume.
func (c *CVCController) createVolumeOperation(cvc *apis.CStorVolumeClaim) (*apis.CStorVolumeClaim, error) {

	policyName := cvc.Annotations[string(apis.VolumePolicyKey)]
	volumePolicy, err := c.getVolumePolicy(policyName, cvc)
	if err != nil {
		return nil, err
	}

	klog.V(2).Infof("creating cstorvolume service resource")
	svcObj, err := getOrCreateTargetService(cvc)
	if err != nil {
		return nil, err
	}

	klog.V(2).Infof("creating cstorvolume resource")
	cvObj, err := getOrCreateCStorVolumeResource(svcObj, cvc)
	if err != nil {
		return nil, err
	}

	klog.V(2).Infof("creating cstorvolume target deployment")
	_, err = getOrCreateCStorTargetDeployment(cvObj, volumePolicy)
	if err != nil {
		return nil, err
	}

	klog.V(2).Infof("creating cstorvolume replica resource")
	err = c.distributePendingCVRs(cvc, cvObj, svcObj, volumePolicy)
	if err != nil {
		return nil, err
	}

	// Fetch the volume replica pool names and use them in PDB and updating in
	// spec and status of CVC
	poolNames, err := cvr.GetVolumeReplicaPoolNames(cvc.Name, openebsNamespace)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to get volume replica pool names of volume %s", cvObj.Name)
	}

	if isHAVolume(cvc) {
		// TODO: When multiple threads or multiple CVC controllers are set then
		// we have to revist entier PDB code path
		var pdbObj *policy.PodDisruptionBudget
		pdbObj, err = getOrCreatePodDisruptionBudget(cvObj, getCSPC(cvc), poolNames)
		if err != nil {
			return nil, errors.Wrapf(err,
				"failed to create PDB for volume: %s", cvc.Name)
		}
		addPDBLabelOnCVC(cvc, pdbObj)
	}

	volumeRef, err := ref.GetReference(scheme.Scheme, cvObj)
	if err != nil {
		return nil, err
	}

	// update volume replica pool information on cvc spec and status
	addReplicaPoolInfo(cvc, poolNames)
	// update the cstorvolume reference, phase as "Bound" and desired
	// capacity
	cvc.Spec.CStorVolumeRef = volumeRef
	cvc.Spec.Policy = volumePolicy.Spec
	cvc.Status.Phase = apis.CStorVolumeClaimPhaseBound
	cvc.Status.Capacity = cvc.Spec.Capacity

	err = c.updateCVCObj(cvc, cvObj)
	if err != nil {
		return nil, err
	}
	return cvc, nil
}

func (c *CVCController) getVolumePolicy(
	policyName string,
	cvc *apis.CStorVolumeClaim,
) (*apis.CStorVolumePolicy, error) {

	volumePolicy := &apis.CStorVolumePolicy{}
	var err error

	if policyName != "" {
		klog.Infof("uses cstorvolume policy %q to configure volume %q", policyName, cvc.Name)
		volumePolicy, err = c.clientset.OpenebsV1alpha1().CStorVolumePolicies(getNamespace()).Get(policyName, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to get volume policy %q of volume %q",
				policyName,
				cvc.Name,
			)
		}
	}
	return volumePolicy, nil
}

// isReplicaAffinityEnabled checks if replicaAffinity has been enabled using
// cstor volume policy
func (c *CVCController) isReplicaAffinityEnabled(policy *apis.CStorVolumePolicy) bool {
	return policy.Spec.Provision.ReplicaAffinity
}

// distributePendingCVRs trigers create and distribute pending cstorvolumereplica
// resource among the available cstor pools
func (c *CVCController) distributePendingCVRs(
	cvc *apis.CStorVolumeClaim,
	cv *apis.CStorVolume,
	service *corev1.Service,
	policy *apis.CStorVolumePolicy,
) error {

	pendingReplicaCount, err := c.getPendingCVRCount(cvc)
	if err != nil {
		return err
	}
	err = c.distributeCVRs(pendingReplicaCount, cvc, service, cv, policy)
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
// TODO Avoid removing clone finalizer
func (c *CVCController) removeClaimFinalizer(
	cvc *apis.CStorVolumeClaim,
) error {
	if isHAVolume(cvc) {
		err := deletePDBIfNotInUse(cvc)
		if err != nil {
			return errors.Wrapf(err,
				"failed to verify whether PDB %s is in use by other volumes",
				getPDBName(cvc),
			)
		}
	}
	cvcPatch := []Patch{
		Patch{
			Op:   "remove",
			Path: "/metadata/finalizers",
		},
	}

	cvcPatchBytes, err := json.Marshal(cvcPatch)
	if err != nil {
		return errors.Wrapf(
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
		return errors.Wrapf(
			err,
			"failed to remove finalizers from cstorvolumeclaim {%s}",
			cvc.Name,
		)
	}
	klog.Infof("finalizers removed successfully from cstorvolumeclaim {%s}", cvc.Name)
	return nil
}

// getPendingCVRCount gets the pending replica count to be created
// in case of any failures
func (c *CVCController) getPendingCVRCount(
	cvc *apis.CStorVolumeClaim,
) (int, error) {

	currentReplicaCount, err := c.getCurrentReplicaCount(cvc)
	if err != nil {
		runtime.HandleError(err)
		return 0, err
	}
	return cvc.Spec.ReplicaCount - currentReplicaCount, nil
}

// getCurrentReplicaCount give the current cstorvolumereplicas count for the
// given volume.
func (c *CVCController) getCurrentReplicaCount(cvc *apis.CStorVolumeClaim) (int, error) {
	// TODO use lister
	//	CVRs, err := c.cvrLister.CStorVolumeReplicas(cvc.Namespace).
	//		List(klabels.Set(pvLabel).AsSelector())

	pvLabel := pvSelector + "=" + cvc.Name

	cvrList, err := c.clientset.
		OpenebsV1alpha1().
		CStorVolumeReplicas(cvc.Namespace).
		List(metav1.ListOptions{LabelSelector: pvLabel})

	if err != nil {
		return 0, errors.Errorf("unable to get current replica count: %v", err)
	}
	return len(cvrList.Items), nil
}

// IsCVRPending look for pending cstorvolume replicas compared to desired
// replica count. returns true if count doesn't matches.
func (c *CVCController) IsCVRPending(cvc *apis.CStorVolumeClaim) (bool, error) {

	selector := klabels.SelectorFromSet(BaseLabels(cvc))
	CVRs, err := c.cvrLister.CStorVolumeReplicas(cvc.Namespace).
		List(selector)
	if err != nil {
		return false, errors.Errorf("failed to list cvr : %v", err)
	}
	// TODO: check for greater values
	return cvc.Spec.ReplicaCount != len(CVRs), nil
}

// BaseLabels returns the base labels we apply to cstorvolumereplicas created
func BaseLabels(cvc *apis.CStorVolumeClaim) map[string]string {
	base := map[string]string{
		pvSelector: cvc.Name,
	}
	return base
}

// cvcNeedResize returns true if a cvc desired a resize operation.
func (c *CVCController) cvcNeedResize(cvc *apis.CStorVolumeClaim) bool {

	desiredCVCSize := cvc.Spec.Capacity[corev1.ResourceStorage]
	actualCVCSize := cvc.Status.Capacity[corev1.ResourceStorage]

	return desiredCVCSize.Cmp(actualCVCSize) > 0
}

// resizeCVC will:
// 1. Mark cvc as resizing.
// 2. Resize the cstorvolume object.
// 3. Mark cvc as resizing finished
func (c *CVCController) resizeCVC(cvc *apis.CStorVolumeClaim) error {
	var updatedCVC *apis.CStorVolumeClaim
	var err error
	cv, err := c.clientset.OpenebsV1alpha1().CStorVolumes(cvc.Namespace).
		Get(cvc.Name, metav1.GetOptions{})
	if err != nil {
		runtime.HandleError(fmt.Errorf("falied to get cv %s: %v", cvc.Name, err))
		return err
	}
	desiredCVCSize := cvc.Spec.Capacity[corev1.ResourceStorage]

	if (cv.Spec.Capacity).Cmp(cv.Status.Capacity) > 0 {
		c.recorder.Event(cvc, corev1.EventTypeNormal, string(apis.CStorVolumeClaimResizing),
			fmt.Sprintf("Resize already in progress %s", cvc.Name))

		klog.Warningf("Resize already in progress on %q from: %v to: %v",
			cvc.Name, cv.Status.Capacity.String(), cv.Spec.Capacity.String())
		return nil
	}

	// markCVC as resized finished
	if desiredCVCSize.Cmp(cv.Status.Capacity) == 0 {
		// Resize volume succeeded mark it as resizing finished.
		return c.markCVCResizeFinished(cvc)
	}

	//if desiredCVCSize.Cmp(cv.Spec.Capacity) > 0 {
	if updatedCVC, err = c.markCVCResizeInProgress(cvc); err != nil {
		klog.Errorf("failed to mark cvc %q as resizing: %v", cvc.Name, err)
		return err
	}
	cvc = updatedCVC
	// Record an event to indicate that cvc-controller is resizing this volume.
	c.recorder.Event(cvc, corev1.EventTypeNormal, string(apis.CStorVolumeClaimResizing),
		fmt.Sprintf("CVCController is resizing volume %s", cvc.Name))

	err = c.resizeCV(cv, desiredCVCSize)
	if err != nil {
		// Record an event to indicate that resize operation is failed.
		c.recorder.Eventf(cvc, corev1.EventTypeWarning, string(apis.CStorVolumeClaimResizeFailed), err.Error())
		return err
	}
	return nil
}

func (c *CVCController) markCVCResizeInProgress(cvc *apis.CStorVolumeClaim) (*apis.CStorVolumeClaim, error) {
	// Mark CVC as Resize Started
	progressCondition := apis.CStorVolumeClaimCondition{
		Type:               apis.CStorVolumeClaimResizing,
		LastTransitionTime: metav1.Now(),
	}
	newCVC := cvc.DeepCopy()
	newCVC.Status.Conditions = MergeResizeConditionsOfCVC(newCVC.Status.Conditions,
		[]apis.CStorVolumeClaimCondition{progressCondition})
	return c.PatchCVCStatus(cvc, newCVC)
}

type resizeProcessStatus struct {
	condition apis.CStorVolumeClaimCondition
	processed bool
}

// MergeResizeConditionsOfCVC updates cvc with desired resize conditions
// leaving other conditions untouched.
func MergeResizeConditionsOfCVC(oldConditions, resizeConditions []apis.CStorVolumeClaimCondition) []apis.CStorVolumeClaimCondition {

	resizeConditionMap := map[apis.CStorVolumeClaimConditionType]*resizeProcessStatus{}

	for _, condition := range resizeConditions {
		resizeConditionMap[condition.Type] = &resizeProcessStatus{condition, false}
	}

	newConditions := []apis.CStorVolumeClaimCondition{}
	for _, condition := range oldConditions {
		// If Condition is of not resize type, we keep it.
		if _, ok := knownResizeConditions[condition.Type]; !ok {
			newConditions = append(newConditions, condition)
			continue
		}

		if newCondition, ok := resizeConditionMap[condition.Type]; ok {
			newConditions = append(newConditions, newCondition.condition)
			newCondition.processed = true
		}
	}
	// append all unprocessed conditions
	for _, newCondition := range resizeConditionMap {
		if !newCondition.processed {
			newConditions = append(newConditions, newCondition.condition)
		}
	}
	return newConditions
}

func (c *CVCController) markCVCResizeFinished(cvc *apis.CStorVolumeClaim) error {
	newCVC := cvc.DeepCopy()
	newCVC.Status.Capacity = cvc.Spec.Capacity

	newCVC.Status.Conditions = MergeResizeConditionsOfCVC(cvc.Status.Conditions, []apis.CStorVolumeClaimCondition{})
	_, err := c.PatchCVCStatus(cvc, newCVC)
	if err != nil {
		klog.Errorf("Mark CVC %q as resize finished failed: %v", cvc.Name, err)
		return err
	}

	klog.V(4).Infof("Resize CVC %q finished", cvc.Name)
	c.recorder.Eventf(cvc, corev1.EventTypeNormal, string(apis.CStorVolumeClaimResizeSuccess), "Resize volume succeeded")

	return nil
}

// PatchCVCStatus updates CVC status using patch api
func (c *CVCController) PatchCVCStatus(oldCVC,
	newCVC *apis.CStorVolumeClaim,
) (*apis.CStorVolumeClaim, error) {
	patchBytes, err := getPatchData(oldCVC, newCVC)
	if err != nil {
		return nil, fmt.Errorf("can't patch status of CVC %s as generate path data failed: %v", oldCVC.Name, err)
	}
	updatedClaim, updateErr := c.clientset.OpenebsV1alpha1().CStorVolumeClaims(oldCVC.Namespace).
		Patch(oldCVC.Name, types.MergePatchType, patchBytes)

	if updateErr != nil {
		return nil, fmt.Errorf("can't patch status of CVC %s with %v", oldCVC.Name, updateErr)
	}
	return updatedClaim, nil
}

func getPatchData(oldObj, newObj interface{}) ([]byte, error) {
	oldData, err := json.Marshal(oldObj)
	if err != nil {
		return nil, fmt.Errorf("marshal old object failed: %v", err)
	}
	newData, err := json.Marshal(newObj)
	if err != nil {
		return nil, fmt.Errorf("mashal new object failed: %v", err)
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, oldObj)
	if err != nil {
		return nil, fmt.Errorf("CreateTwoWayMergePatch failed: %v", err)
	}
	return patchBytes, nil
}

// resizeCV resize the cstor volume to desired size, and update CV's capacity
func (c *CVCController) resizeCV(cv *apis.CStorVolume, newCapacity resource.Quantity) error {
	newCV := cv.DeepCopy()
	newCV.Spec.Capacity = newCapacity

	patchBytes, err := getPatchData(cv, newCV)
	if err != nil {
		return fmt.Errorf("can't update capacity of CV %s as generate patch data failed: %v", cv.Name, err)
	}
	_, updateErr := c.clientset.OpenebsV1alpha1().CStorVolumes(getNamespace()).
		Patch(cv.Name, types.MergePatchType, patchBytes)
	if updateErr != nil {
		return updateErr
	}
	return nil
}

// deletePDBIfNotInUse deletes the PDB if no volume is refering to the
// cStorvolumeclaim PDB
func deletePDBIfNotInUse(cvc *apis.CStorVolumeClaim) error {
	//TODO: If HALease is enabled active-active then below code needs to be
	//revist
	pdbName := getPDBName(cvc)
	cvcLabelSelector := string(apis.PodDisruptionBudgetKey) + "=" + pdbName
	cvcList, err := cvclaim.NewKubeclient().
		WithNamespace(cvc.Namespace).
		List(metav1.ListOptions{LabelSelector: cvcLabelSelector})
	if err != nil {
		return errors.Wrapf(err,
			"failed to list volumes refering to PDB %s", pdbName)
	}
	if len(cvcList.Items) == 1 {
		err = apispdb.KubeClient().
			WithNamespace(openebsNamespace).
			Delete(pdbName, &metav1.DeleteOptions{})
		if k8serror.IsNotFound(err) {
			klog.Infof("pdb %s of volume %s was already deleted", pdbName, cvc.Name)
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// scaleVolumeReplicas identifies whether it is scaleup or scaledown case of
// volume replicas. If user added entry of pool info under the spec then changes
// are treated as scaleup case. If user removed poolInfo entry from spec then
// changes are treated as scale down case. If user just modifies the pool entry
// info under the spec then it is a kind of migration which is not yet supported
func (c *CVCController) scaleVolumeReplicas(cvc *apis.CStorVolumeClaim) error {
	var err error
	if len(cvc.Spec.Policy.ReplicaPool.PoolInfo) > len(cvc.Status.PoolInfo) {
		err = scaleUpVolumeReplicas(cvc)
	} else if len(cvc.Spec.Policy.ReplicaPool.PoolInfo) < len(cvc.Status.PoolInfo) {
		err = scaleDownVolumeReplicas(cvc)
	} else {
		c.recorder.Event(cvc, corev1.EventTypeWarning, "Migration",
			"Migration of volume replicas is not yet supported")
		return nil
	}
	if err != nil {
		c.recorder.Eventf(cvc,
			corev1.EventTypeWarning,
			"ScalingVolumeReplicas",
			"%v", err)
	}
	return nil
}
