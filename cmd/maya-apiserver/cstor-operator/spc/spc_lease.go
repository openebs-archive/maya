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
	"context"
	"encoding/json"
	"fmt"
	"strings"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
)

const (
	// SpcLeaseKey is the key that will be used to acquire lease on spc object.
	// It will be present in spc annotations.
	// If key has an empty value, that means no one has acquired a lease on spc object.
	SpcLeaseKey = "openebs.io/spc-lease"
	// PatchOperation is the strategy of patch operation.
	PatchOperation = "replace"
	// PatchPath is the path to the field on spc object which need to be patched.
	PatchPath = "/metadata/annotations/openebs.io~1spc-lease"
)

// Patch struct represent the struct used to patch
// the spc object

// Patch struct will used to patch the spc object by a lease holder
// to release the lease once done.
type Patch struct {
	// Op defines the operation
	Op string `json:"op"`
	// Path defines the key path
	// eg. for
	// {
	//  	"Name": "openebs"
	//	    Category: {
	//		  "Inclusive": "v1",
	//		  "Rank": "A"
	//	     }
	// }
	// The path of 'Inclusive' would be
	// "/Name/Category/Inclusive"
	Path  string `json:"path"`
	Value string `json:"value"`
}

// Hold is the implenetation of method from interface Leases
// It will try to hold a lease on spc object.
func (sl *Lease) Hold() error {
	// Get the lease value.
	spcObject, ok := sl.Object.(*apis.StoragePoolClaim)
	if !ok {
		return fmt.Errorf("expected spc object for leasing but got %#v", spcObject)
	}
	leaseValue := spcObject.Annotations[sl.leaseKey]
	var leaseValueObj LeaseContract
	var err error
	if !(strings.TrimSpace(leaseValue) == "") {
		leaseValueObj, err = parseLeaseValue(leaseValue)
		if err != nil {
			return err
		}
	}
	// If leaseValue is empty acquire lease.
	// If leaseValue is empty check whether it is expired.
	// If leaseValue is not empty and not expired check whether the holder is live.
	if strings.TrimSpace(leaseValue) == "" || isLeaseExpired(leaseValueObj) || !sl.isLeaderLive(leaseValueObj) {
		err := sl.Update(sl.getPodName())
		if err != nil {
			return err
		}
		return nil
	}
	// If none of the above three conditions are met, lease can not be acquired.
	return fmt.Errorf("lease on spc already acquired by a live pod")
}

// Update will update a lease on spc depending on type of update that is required.
// We have following type of update strategy:
// 1.putKeyValue
// 2.putValue
// 3.putUpdatedValue
// See the functions(below) for more details on update strategy
func (sl *Lease) Update(podName string) error {
	newSpcObject := sl.Object.(*apis.StoragePoolClaim)
	if newSpcObject.Annotations == nil {
		sl.putKeyValue(podName, newSpcObject)
	} else if newSpcObject.Annotations[sl.leaseKey] == "" {
		sl.putValue(podName, newSpcObject)
	} else {
		sl.putUpdatedValue(podName, newSpcObject)
	}
	_, err := sl.oecs.OpenebsV1alpha1().StoragePoolClaims().
		Update(context.TODO(), newSpcObject, meta_v1.UpdateOptions{})
	return err
}

// Release method is implementation of  to release lease on a given spc.
func (sl *Lease) Release() {
	err := sl.patchSpcLeaseAnnotation()
	if err != nil {
		newErr := fmt.Errorf("Lease could not be removed:%v", err)
		runtime.HandleError(newErr)
	}
	klog.Info("Lease removed successfully on storagepoolclaim")
}

func (sl *Lease) getPodName() string {
	podName := env.Get(env.OpenEBSMayaPodName)
	podNameSpace := env.Get(env.OpenEBSNamespace)
	return podNameSpace + "/" + podName
}

// patchSpcLeaseAnnotation will patch the lease key annotation on spc object to release the lease
func (sl *Lease) patchSpcLeaseAnnotation() error {
	spcObject, ok := sl.Object.(*apis.StoragePoolClaim)
	if !ok {
		return fmt.Errorf("expected spc object for leasing but got %#v", spcObject)
	}
	spcPatch := make([]Patch, 1)
	// setting operation as remove
	spcPatch[0].Op = PatchOperation
	// object to be removed is finalizers
	spcPatch[0].Path = PatchPath
	leaseValueObj, err := parseLeaseValue(spcObject.Annotations[SpcLeaseKey])
	if err != nil {
		return err
	}
	leaseValueObj.Holder = ""
	newLeaseValue, err := json.Marshal(leaseValueObj)
	if err != nil {
		return err
	}
	spcPatch[0].Value = string(newLeaseValue)
	spcPatchJSON, err := json.Marshal(spcPatch)
	if err != nil {
		return fmt.Errorf("error marshalling spcPatch object: %s", err)
	}
	_, err = sl.oecs.OpenebsV1alpha1().StoragePoolClaims().
		Patch(context.TODO(), spcObject.Name, types.JSONPatchType, spcPatchJSON, meta_v1.PatchOptions{})
	return err
}

// isLeaderLive checks whether the holder of lease is live or not
// If the holder is not live or does not exists the function will return true.

// If the holder of lease is not live or does not exists the lease can be acquired
// by the other contestant(i.e. maya pod)
func (sl *Lease) isLeaderLive(leaseValueObj LeaseContract) bool {
	holderName := leaseValueObj.Holder
	podDetails := strings.Split(holderName, "/")
	// Check whether the holder is live or not
	pod, _ := sl.kubeclientset.CoreV1().Pods(podDetails[0]).
		Get(context.TODO(), podDetails[1], meta_v1.GetOptions{})
	if pod == nil {
		return false
	}
	podStatus := pod.Status.Phase
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

// putKeyValue function will update lease on such SPC which was not acquired by any pod ever in its lifetime.
func (sl *Lease) putKeyValue(podName string, newSpcObject *apis.StoragePoolClaim) (*apis.StoragePoolClaim, error) {
	// make a map that should contain the lease key in spc
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
	mapLease[sl.leaseKey] = string(leaseValue)
	newSpcObject.Annotations = mapLease
	return newSpcObject, nil
}

// putValue function will update lease on SPC if the holder of lease has released the lease successfully.
func (sl *Lease) putValue(podName string, newSpcObject *apis.StoragePoolClaim) (*apis.StoragePoolClaim, error) {
	leaseValueObj := &LeaseContract{
		podName,
		1,
	}
	leaseValue, err := json.Marshal(leaseValueObj)
	if err != nil {
		return nil, err
	}
	newSpcObject.Annotations[sl.leaseKey] = string(leaseValue)
	return newSpcObject, nil
}

// putUpdatedValue function will update lease on SPC if the holder of lease has died before releasing the lease.
func (sl *Lease) putUpdatedValue(podName string, newSpcObject *apis.StoragePoolClaim) (*apis.StoragePoolClaim, error) {
	leaseValueObj, err := parseLeaseValue(newSpcObject.Annotations[sl.leaseKey])
	if err != nil {
		return nil, err
	}
	leaseValueObj.LeaderTransition++
	leaseValueObj.Holder = podName
	leaseValue, err := json.Marshal(leaseValueObj)
	if err != nil {
		return nil, err
	}
	newSpcObject.Annotations[sl.leaseKey] = string(leaseValue)
	return newSpcObject, nil
}
