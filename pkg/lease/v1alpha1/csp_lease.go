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

package lease

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	patch "github.com/openebs/maya/pkg/patch/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
)

const (
	// CspLeaseKey is the key that will be used to acquire lease on csp object.
	// It will be present in csp annotations.
	// If key has an empty value, that means no one has acquired a lease on csp object.
	//TODO : Evaluate if openebs.io/lease be a better label.
	CspLeaseKey = "openebs.io/csp-lease"
	// PatchOperation is the strategy of patch operation.
	PatchOperation = "replace"
	// PatchPath is the path to the field on csp object which need to be patched.
	PatchPath = "/metadata/annotations/openebs.io~1csp-lease"
	// PodName is the name of the pool pod.
	PodName = "POD_NAME"
	// NameSpace is the namespace where pool pod is running.
	NameSpace = "NAMESPACE"
)

// Hold is the implenetation of method from interface Leases
// It will try to hold a lease on csp object.
func (sl *Lease) Hold() (interface{}, error) {
	// Get the lease value.
	cspObject, ok := sl.Object.(*apis.CStorPool)
	if !ok {
		return nil, fmt.Errorf("expected csp object for leasing but got %#v", cspObject)
	}
	leaseValue := cspObject.Annotations[sl.LeaseKey]
	var leaseValueObj LeaseContract
	var err error
	if !(strings.TrimSpace(leaseValue) == "") {
		leaseValueObj, err = parseLeaseValue(leaseValue)
		if err != nil {
			return nil, err
		}
	}
	// If the pod which has already acquired lease and want again to acquire,grant it.
	if leaseValueObj.Holder == env.Get(NameSpace)+"/"+env.Get(PodName) {
		return cspObject, nil
	}
	// If leaseValue is empty acquire lease.
	// If leaseValue is empty check whether it is expired.
	// If leaseValue is not empty and not expired check whether the holder is live.
	if strings.TrimSpace(leaseValue) == "" || isLeaseExpired(leaseValueObj) || !sl.isLeaderALive(leaseValueObj) {
		podName, err := sl.getPodName()
		if err != nil {
			return nil, err
		}
		csp, err := sl.Update(podName)
		if err != nil {
			return nil, err
		}
		return csp, nil
	}
	// If none of the above three conditions are met, lease can not be acquired.
	return nil, fmt.Errorf("lease on csp already acquired by a live pod")
}

// Update will update a lease on csp depending on type of update that is required.
// We have following type of update strategy:
// 1.putKeyValue
// 2.putValue
// 3.putUpdatedValue
// See the functions(below) for more details on update strategy
func (sl *Lease) Update(podName string) (interface{}, error) {
	newCspObject := sl.Object.(*apis.CStorPool)
	if newCspObject.Annotations == nil {
		sl.putKeyValue(podName, newCspObject)
	} else if newCspObject.Annotations[sl.LeaseKey] == "" {
		sl.putValue(podName, newCspObject)
	} else {
		sl.putUpdatedValue(podName, newCspObject)
	}
	csp, err := sl.Oecs.OpenebsV1alpha1().CStorPools().
		Update(context.TODO(), newCspObject, meta_v1.UpdateOptions{})
	return csp, err
}

// Release method is implementation of  to release lease on a given csp.
func (sl *Lease) Release() {
	err := sl.patchCspLeaseAnnotation()
	if err != nil {
		newErr := fmt.Errorf("Lease could not be removed:%v", err)
		runtime.HandleError(newErr)
	}
	klog.Info("Lease removed successfully on csp")
}

func (sl *Lease) getPodName() (string, error) {
	podName := env.Get(PodName)
	if strings.TrimSpace(podName) == "" {
		return "", fmt.Errorf("Pod name not found in env variable")
	}
	podNameSpace := env.Get(NameSpace)
	if strings.TrimSpace(podName) == "" {
		return "", fmt.Errorf("Pod namespace not found in env variable")
	}
	return podNameSpace + "/" + podName, nil
}

// patchCspLeaseAnnotation will patch the lease key annotation on csp object to release the lease
func (sl *Lease) patchCspLeaseAnnotation() error {
	cspObject, ok := sl.Object.(*apis.CStorPool)
	if !ok {
		return fmt.Errorf("expected csp object for leasing but got %#v", cspObject)
	}
	leaseValueObj, err := parseLeaseValue(cspObject.Annotations[CspLeaseKey])
	if err != nil {
		return err
	}
	leaseValueObj.Holder = ""
	newLeaseValue, err := json.Marshal(leaseValueObj)
	if err != nil {
		return err
	}

	cspPatch, err := patch.NewPatchPayload(PatchOperation, PatchPath, string(newLeaseValue))
	if err != nil {
		return fmt.Errorf("unable to form payload to patch csp:%s", err)
	}
	_, err = sl.Patch(cspObject.Name, "", types.JSONPatchType, cspPatch)
	if err != nil {
		return fmt.Errorf("unable to patch csp %s :%v", cspObject.Name, err)
	}
	return nil
}

// isLeaderLive checks whether the holder of lease is live or not
// If the holder is not live or does not exists the function will return false.

// If the holder of lease is not live or does not exists the lease can be acquired
// by the other contestant(i.e. pool pod)
func (sl *Lease) isLeaderALive(leaseValueObj LeaseContract) bool {
	holderName := leaseValueObj.Holder
	podDetails := strings.Split(holderName, "/")
	// Check whether the holder is live or not.
	pod, err := sl.Kubeclientset.CoreV1().Pods(podDetails[0]).
		Get(context.TODO(), podDetails[1], meta_v1.GetOptions{})
	if err != nil {
		// If the pod does not exist, an error will be thrown and if it is a not found error
		// meaning pod does not exist, we should return false.
		if errors.IsNotFound(err) {
			return false
		}
		klog.Warningf("Could not fetch the pod which have acquired the lease on CSP:%s", err)
		return true
	}
	if pod == nil {
		return false
	}
	podStatus := pod.Status.Phase
	if string(podStatus) == string(corev1.PodUnknown) {
		klog.Warning("Could not get the pod status which have acquired the lease on CSP")
		return true
	}
	if string(podStatus) != string(corev1.PodRunning) {
		return false
	}
	return true
}

// A lease is expired if it has a empty holder name.
// The holder of lease can only expire the lease.
// isLeaseExpired return true if the lease is expired.
func isLeaseExpired(leaseValueObj LeaseContract) bool {
	if strings.TrimSpace(leaseValueObj.Holder) == "" {
		return true
	}
	return false
}

// parseLeaseValue will parse a leaseValue string to lease object
func parseLeaseValue(leaseValue string) (LeaseContract, error) {
	leaseValueObj := &LeaseContract{}
	err := json.Unmarshal([]byte(leaseValue), leaseValueObj)
	if err != nil {
		return LeaseContract{}, err
	}
	return *leaseValueObj, nil
}

// putKeyValue function will update lease on such CSP which was not acquired by any pod ever in its lifetime.
func (sl *Lease) putKeyValue(podName string, newCspObject *apis.CStorPool) (*apis.CStorPool, error) {
	// make a map that should contain the lease key in csp
	mapLease := make(map[string]string)
	leaseValueObj := &LeaseContract{
		podName,
		1,
	}
	leaseValue, err := json.Marshal(leaseValueObj)
	if err != nil {
		return nil, err
	}
	// Fill the map lease key with lease value
	mapLease[sl.LeaseKey] = string(leaseValue)
	newCspObject.Annotations = mapLease
	return newCspObject, nil
}

// putValue function will update lease on CSP if the holder of lease has released the lease successfully.
func (sl *Lease) putValue(podName string, newCspObject *apis.CStorPool) (*apis.CStorPool, error) {
	leaseValueObj := &LeaseContract{
		podName,
		1,
	}
	leaseValue, err := json.Marshal(leaseValueObj)
	if err != nil {
		return nil, err
	}
	newCspObject.Annotations[sl.LeaseKey] = string(leaseValue)
	return newCspObject, nil
}

// putUpdatedValue function will update lease on CSP if the holder of lease has died before releasing the lease.
func (sl *Lease) putUpdatedValue(podName string, newCspObject *apis.CStorPool) (*apis.CStorPool, error) {
	leaseValueObj, err := parseLeaseValue(newCspObject.Annotations[sl.LeaseKey])
	if err != nil {
		return nil, err
	}
	leaseValueObj.LeaderTransition++
	leaseValueObj.Holder = podName
	leaseValue, err := json.Marshal(leaseValueObj)
	if err != nil {
		return nil, err
	}
	newCspObject.Annotations[sl.LeaseKey] = string(leaseValue)
	return newCspObject, nil
}

// Patch is the specific implementation if Patch() interface for patching CSP objects.
// Similarly, we can have for other objects, if required.
func (sl *Lease) Patch(name string, nameSpace string, patchType types.PatchType, patches []byte) (*apis.CStorPool, error) {
	obj, err := sl.Oecs.OpenebsV1alpha1().CStorPools().
		Patch(context.TODO(), name, patchType, patches, meta_v1.PatchOptions{})
	return obj, err
}
