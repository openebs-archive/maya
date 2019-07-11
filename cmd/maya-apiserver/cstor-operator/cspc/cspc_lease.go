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

package cspc

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"strings"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
)

const (
	// CSPCLeaseKey is the key that will be used to acquire lease on cspc object.
	// It will be present in cspc annotations.
	// If key has an empty value, that means no one has acquired a lease on cspc object.
	CSPCLeaseKey = "openebs.io/cspc-lease"
	// PatchOperation is the strategy of patch operation.
	PatchOperation = "replace"
	// PatchPath is the path to the field on cspc object which need to be patched.
	PatchPath = "/metadata/annotations/openebs.io~1cspc-lease"
)

// Patch struct represent the struct used to patch
// the cspc object

// Patch struct will used to patch the cspc object by a lease holder
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
// It will try to hold a lease on cspc object.
func (sl *Lease) Hold() error {
	// Get the lease value.
	cspcObject, ok := sl.Object.(*apis.CStorPoolCluster)
	if !ok {
		return fmt.Errorf("expected cspc object for leasing but got %#v", cspcObject)
	}
	leaseValue := cspcObject.Annotations[sl.leaseKey]
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
	return fmt.Errorf("lease on cspc already acquired by a live pod")
}

// Update will update a lease on cspc depending on type of update that is required.
// We have following type of update strategy:
// 1.putKeyValue
// 2.putValue
// 3.putUpdatedValue
// See the functions(below) for more details on update strategy
func (sl *Lease) Update(podName string) error {
	var err error
	newSpcObject := sl.Object.(*apis.CStorPoolCluster)
	if newSpcObject.Annotations == nil {
		_, err = sl.putKeyValue(podName, newSpcObject)
	} else if newSpcObject.Annotations[sl.leaseKey] == "" {
		_, err = sl.putValue(podName, newSpcObject)
	} else {
		_, err = sl.putUpdatedValue(podName, newSpcObject)
	}
	if err != nil {
		return errors.Wrapf(err, "failed to update while taking lease")
	}
	_, err = sl.oecs.OpenebsV1alpha1().CStorPoolClusters(env.Get(env.OpenEBSNamespace)).Update(newSpcObject)
	return err
}

// Release method is implementation of  to release lease on a given cspc.
func (sl *Lease) Release() {
	err := sl.patchSpcLeaseAnnotation()
	if err != nil {
		newErr := fmt.Errorf("Lease could not be removed:%v", err)
		runtime.HandleError(newErr)
	}
	glog.V(5).Info("Lease removed successfully on cstorpoolcluster")
}

func (sl *Lease) getPodName() string {
	podName := env.Get(env.OpenEBSMayaPodName)
	podNameSpace := env.Get(env.OpenEBSNamespace)
	return podNameSpace + "/" + podName
}

// patchSpcLeaseAnnotation will patch the lease key annotation on cspc object to release the lease
func (sl *Lease) patchSpcLeaseAnnotation() error {
	cspcObject, ok := sl.Object.(*apis.CStorPoolCluster)
	if !ok {
		return fmt.Errorf("expected cspc object for leasing but got %#v", cspcObject)
	}
	cspcPatch := make([]Patch, 1)
	// setting operation as remove
	cspcPatch[0].Op = PatchOperation
	// object to be removed is finalizers
	cspcPatch[0].Path = PatchPath
	leaseValueObj, err := parseLeaseValue(cspcObject.Annotations[CSPCLeaseKey])
	if err != nil {
		return err
	}
	leaseValueObj.Holder = ""
	newLeaseValue, err := json.Marshal(leaseValueObj)
	if err != nil {
		return err
	}
	cspcPatch[0].Value = string(newLeaseValue)
	cspcPatchJSON, err := json.Marshal(cspcPatch)
	if err != nil {
		return fmt.Errorf("error marshalling cspcPatch object: %s", err)
	}
	_, err = sl.oecs.OpenebsV1alpha1().CStorPoolClusters(env.Get(env.OpenEBSNamespace)).Patch(cspcObject.Name, types.JSONPatchType, cspcPatchJSON)
	return err
}

// isLeaderLive checks whether the holder of lease is live or not
// If the holder is not live or does not exists the function will return false.

// If the holder of lease is not live or does not exists the lease can be acquired
// by the other contestant(i.e. maya pod)
func (sl *Lease) isLeaderLive(leaseValueObj LeaseContract) bool {
	holderName := leaseValueObj.Holder
	podDetails := strings.Split(holderName, "/")
	// Check whether the holder is live or not
	pod, _ := sl.kubeclientset.CoreV1().Pods(podDetails[0]).Get(podDetails[1], meta_v1.GetOptions{})
	if pod == nil {
		return false
	}
	podStatus := pod.Status.Phase
	return string(podStatus) == string(corev1.PodRunning)
}

// A lease is expired if it has a empty holder name.
// The holder of lease can only expire the lease.
// isLeaseExpired return true if the lease is expired.
func isLeaseExpired(leaseValueObj LeaseContract) bool {
	return strings.TrimSpace(leaseValueObj.Holder) == ""
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
func (sl *Lease) putKeyValue(podName string, newSpcObject *apis.CStorPoolCluster) (*apis.CStorPoolCluster, error) {
	// make a map that should contain the lease key in cspc
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
func (sl *Lease) putValue(podName string, newSpcObject *apis.CStorPoolCluster) (*apis.CStorPoolCluster, error) {
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
func (sl *Lease) putUpdatedValue(podName string, newSpcObject *apis.CStorPoolCluster) (*apis.CStorPoolCluster, error) {
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
