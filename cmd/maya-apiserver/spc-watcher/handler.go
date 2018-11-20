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
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/k8s"
	"github.com/openebs/maya/pkg/patch"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"strings"
)

const (
	NodePhaseOnline  = "Online"
	NodePhaseOffline = "Offline"
)

// PatchPayloadCSP struct is ussed to patch CSP object.
// Similarly, for other objects (if required to patch) we can have structs for them
// to have a implementation if patch function.
type PatchPayloadCSP struct {
	// 'Object' is the object which needs to be patched.
	Object *apis.CStorPool
	// PatchPayloadCSP is the payload to patch CSP.
	PatchPayload []patch.Patch
	//ClientSet    patch.ClientSet
	K8sClientSet *patch.ClientSet
}

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
		// It is a create event so resync should be false and pendingPoolcount is passed 0
		// pendingPoolcount is not used when resync is false.
		err := c.CreateStoragePool(spcGot, false, 0)
		if err != nil {
			glog.Error("Storagepool could not be created:", err)
			// To-Do
			// If Some error occur patch the spc object with appropriate reason
		}

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
		return syncEvent, err
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

func (c *Controller) syncSpc(spcGot *apis.StoragePoolClaim) error {
	// Get kubernetes clientset
	// namespaces is not required, hence passed empty.
	newK8sClient, err := k8s.NewK8sClient("")
	if err != nil {
		return err
	}
	// Get openebs clientset using a getter method (i.e. GetOECS() ) as
	// the openebs clientset is not exported.
	newOecsClient := newK8sClient.GetOECS()
	// Update CSP statuses, as part of the resync activity.
	c.updateCspStatus(spcGot)
	if err != nil {
		return fmt.Errorf("unable to update csp status in resync event:%s", err.Error())
	}

	if len(spcGot.Spec.Disks.DiskList) > 0 {
		// TODO : reconciliation for manual storagepool provisioning
		glog.V(1).Infof("No reconciliation needed for manual provisioned pool of storagepoolclaim %s", spcGot.Name)
		return nil
	}
	glog.V(1).Infof("Syncing storagepoolclaim %s", spcGot.Name)

	// Get the current count of provisioned pool for the storagepool claim
	spList, err := newOecsClient.OpenebsV1alpha1().StoragePools().List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spcGot.Name})
	if err != nil {
		return fmt.Errorf("unable to list storagepools: %v", err)
	}
	currentPoolCount := len(spList.Items)

	// If current pool count is less than maxpool count, try to converge to maxpool
	if currentPoolCount < int(spcGot.Spec.MaxPools) {
		glog.Infof("Converging storagepoolclaim %s to desired state:current pool count is %d,desired pool count is %d", spcGot.Name, currentPoolCount, spcGot.Spec.MaxPools)
		// pendingPoolCount holds the pending pool that should be provisioned to get the desired state.
		pendingPoolCount := int(spcGot.Spec.MaxPools) - currentPoolCount
		// Call the storage pool create logic to provision the pending pools.
		err := c.CreateStoragePool(spcGot, true, pendingPoolCount)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateCspStatus update statuses on csp by patching the csp object.
func (c *Controller) updateCspStatus(spcGot *apis.StoragePoolClaim) {
	// List all the CSPs for the given SPC.
	cspList, err := c.clientset.OpenebsV1alpha1().CStorPools().List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spcGot.Name})
	if err != nil {
		glog.Errorf("Unable to list cstor pool cr for spc '%s': %v", err, spcGot.Name)
	}
	// Iterate over CSP and patch it to update status.
	for _, csp := range cspList.Items {
		// nodePhase is the state of the node corresponding to CSP.
		var nodePhase string
		// nodeName is the name of node corresponding to CSP.
		var nodeName string
		// If there is no label on csp, node name cannot be extracted, hence throw proper error message.
		if csp.Labels == nil {
			glog.Errorf("Node name not found on CSP object %s as no labels are present on CSP", csp.Name)
		}
		// Extract the node name from the label of CSP object.
		nodeName = csp.Labels[string(apis.HostNameCPK)]
		// Check for empty node name.
		if strings.TrimSpace(nodeName) == "" {
			glog.Errorf("Node name not found on CSP object %s", csp.Name)
		}
		// Get the phase of the node.
		// ToDo: NodePhase has been deprecated in k8s upstream. Need to decide some node phases or states.
		// ToDo: The states/phases can then be mapped to the several node conditions present on node object.
		nodePhase = c.getNodePhase(nodeName)
		// Throw a warning if node phase is empty.
		if strings.TrimSpace(nodePhase) == "" {
			glog.Warningf("Got empty value for node phase for CSP %s", csp.Name)
		}
		// Form the object that will be patched to update the node status on CSP.
		nodePhasePayload := &apis.NodeStatus{
			nodeName,
			v1.NodePhase(nodePhase),
		}
		// Get the patch payload
		cspPatch, err := c.NewPatchPayloadCSP(nodePhasePayload, &csp)
		if err != nil {
			glog.Error("Unable to form payload to patch csp:", err)
		}
		_, err = cspPatch.Patch("", types.JSONPatchType)
		if err != nil {
			glog.Errorf("Unable to patch csp %s in resync event:%v", csp.Name, err)
		}
	}
}

// getNodePhase get the Network status of a node and if network is available it returns online else offline.
// If it does not get the network status from node object it will return empty.
func (c *Controller) getNodePhase(nodeName string) string {
	var nodePhase string
	getNode, err := c.kubeclientset.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		glog.Error("Error in getting node phase:", err)
	}
	for _, conditions := range getNode.Status.Conditions {
		if conditions.Type == v1.NodeNetworkUnavailable {
			if conditions.Status == v1.ConditionFalse {
				nodePhase = string(NodePhaseOnline)
			} else {
				nodePhase = string(NodePhaseOffline)
			}
			break
		}
	}
	return nodePhase
}

// NewPatchPayloadCSP constructs payload to patch csp.
func (c *Controller) NewPatchPayloadCSP(patchValue interface{}, csp *apis.CStorPool) (patch.Patcher, error) {
	var cspPatch patch.Patcher
	cspPatchPayload := patch.NewPatchPayload("add", "/status/nodeStatus", patchValue)
	cspPatch = &PatchPayloadCSP{
		Object:       csp,
		PatchPayload: cspPatchPayload,
		K8sClientSet: &patch.ClientSet{
			c.kubeclientset,
			c.clientset,
		},
	}
	return cspPatch, nil
}

// Patch is the specific implementation if Patch() interface for patching CSP objects.
// Similarly, we can have for other objects, if required.
func (payload *PatchPayloadCSP) Patch(namesapce string, patchType types.PatchType) (interface{}, error) {
	PatchJSON, err := json.Marshal(payload.PatchPayload)
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal patch payload for csp :%v", err)
	}
	cspObject, err := payload.K8sClientSet.PatchCsp(payload.Object.Name, patchType, PatchJSON)
	return cspObject, err
}
