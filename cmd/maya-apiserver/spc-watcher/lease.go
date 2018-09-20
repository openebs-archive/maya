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
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"strings"
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

// SPCPatch struct represent the struct used to patch
// the spc object

// This struct will used to patch the spc object by a lease holder
// to release the lease once done.
type SPCPatch struct {
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

// This struct will be used as a value of lease key that will
// give information about an acquired lease on spc

// The struct object will be parsed to string which will be then
// put as a value to the lease key of spc annotation.
type lease struct {
	// HolderIdentity is the namespace/name of the pod who acquires the lease
	HolderIdentity   string `json:"holderIdentity"`
	LeaderTransition int    `json:"leaderTransition"`
	// More specific details can be added here that will describe the
	// current state of lease in more details.
	// e.g. acquiredTimeStamp, self-release etc
	// acquiredTimeStamp will tell when the lease was acquired
	// self-release will tell whether the lease was removed by the acquirer maya-pod
	// or by other maya-pod
}

// Leases is an interface which assists in getting and releasing lease over an spc object
type Leases interface {
	// GetLease will try to get a lease on spc, in case of failure it will return error
	GetLease() (string, error)
	// UpdateLease will update the lease value of the spc
	UpdateLease(leaseValue string) (*apis.StoragePoolClaim, error)
	// RemoveLease will remove the acquired lease on the spc
	RemoveLease()
}

// spcLease is the struct which will implement the Leases interface
type spcLease struct {
	// spcObject is the storagepoolclaim object over which lease is to be taken
	spcObject *apis.StoragePoolClaim
	// leaseKey is lease key on current storagepoolclaim object
	leaseKey string
	// oecs is the openebs clientset
	oecs clientset.Interface
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
}

func (sl *spcLease) GetLease() (string, error) {
	// Get the lease value.
	leaseValue := sl.spcObject.Annotations[sl.leaseKey]
	var leaseValueObj lease
	var err error
	if !(strings.TrimSpace(leaseValue) == "") {
		leaseValueObj, err = parseLeaseValue(leaseValue)
		if err != nil {
			return "", err
		}
	}
	// If leaseValue is empty acquire lease.
	// If leaseValue is empty check whether it is expired.
	// If leaseValue is not emtpy and not expired check wether the holder is live.
	if strings.TrimSpace(leaseValue) == "" || isLeaseExpired(leaseValueObj) || sl.checkLeaderLiveness(leaseValueObj) {
		spcObject, err := sl.UpdateLease(sl.getPodName())
		if err != nil {
			return "", err
		}
		return spcObject.Annotations[sl.leaseKey], nil
	}
	// If none of the above three conditions are met, lease can not be acquired.
	return "", fmt.Errorf("lease on spc already acquired by a live pod")
}

func (sl *spcLease) UpdateLease(podName string) (*apis.StoragePoolClaim, error) {
	newSpcObject := sl.spcObject
	if newSpcObject.Annotations == nil {
		// make a map that should contain the lease key in spc
		mapLease := make(map[string]string)
		leaseValueObj := &lease{
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
	} else {
		if newSpcObject.Annotations[sl.leaseKey] == "" {
			leaseValueObj := &lease{
				podName,
				1,
			}
			leaseValue, err := json.Marshal(leaseValueObj)
			if err != nil {
				return nil, err
			}
			newSpcObject.Annotations[sl.leaseKey] = string(leaseValue)
		} else {
			leaseValueObj, err := parseLeaseValue(newSpcObject.Annotations[sl.leaseKey])
			if err != nil {
				return nil, err
			}
			leaseValueObj.LeaderTransition++
			leaseValueObj.HolderIdentity = podName
			leaseValue, err := json.Marshal(leaseValueObj)
			if err != nil {
				return nil, err
			}
			newSpcObject.Annotations[sl.leaseKey] = string(leaseValue)
		}

	}
	spcObject, err := sl.oecs.OpenebsV1alpha1().StoragePoolClaims().Update(sl.spcObject)
	if err != nil {
		return nil, err
	}
	return spcObject, nil
}

func (sl *spcLease) RemoveLease() {
	_, err := sl.patchSpc()
	if err != nil {
		newErr := fmt.Errorf("Lease could not be removed:%v", err)
		runtime.HandleError(newErr)
	}
	glog.Info("Lease removed successfully on storagepoolclaim")
}

func (sl *spcLease) getPodName() string {
	podName := env.Get(env.OpenEBSMayaPodName)
	podNameSpace := env.Get(env.OpenEBSNamespace)
	return podNameSpace + "/" + podName
}

// patchSpc will patch the spc object to release the lease
func (sl *spcLease) patchSpc() (*apis.StoragePoolClaim, error) {
	spcPatch := make([]SPCPatch, 1)
	// setting operation as remove
	spcPatch[0].Op = PatchOperation
	// object to be removed is finalizers
	spcPatch[0].Path = PatchPath
	leaseValueObj, err := parseLeaseValue(sl.spcObject.Annotations[SpcLeaseKey])
	leaseValueObj.HolderIdentity = ""
	newLeaseValue, err := json.Marshal(leaseValueObj)
	if err != nil {
		return nil, err
	}
	spcPatch[0].Value = string(newLeaseValue)
	spcPatchJSON, err := json.Marshal(spcPatch)
	if err != nil {
		glog.Errorf("Error marshalling spcPatch object: %s", err)
	}
	obj, err := sl.oecs.OpenebsV1alpha1().StoragePoolClaims().Patch(sl.spcObject.Name, types.JSONPatchType, spcPatchJSON)
	return obj, err
}

// checkLeaderLiveness checks whether the holder of lease is live or not
// If the holder is not live or does not exists the function will return true.

// If the holder of lease is not live or does not exists the lease can be acquired
// by the other contestant(i.e. maya pod)
func (sl *spcLease) checkLeaderLiveness(leaseValueObj lease) bool {
	holderName := leaseValueObj.HolderIdentity
	podDetails := strings.Split(holderName, "/")
	// Check whether the holder is live or not
	pod, _ := sl.kubeclientset.CoreV1().Pods(podDetails[0]).Get(podDetails[1], meta_v1.GetOptions{})
	if pod == nil {
		return true
	}
	podStatus := pod.Status.Phase
	if string(podStatus) != string(v1.PodRunning) {
		return true
	}
	return false
}

// A lease is expired if it has a empty holder name.
// The holder of lease can only expire the lease.
// isLeaseExpired return true if the lease is expired.
func isLeaseExpired(leaseValueObj lease) bool {
	if strings.TrimSpace(leaseValueObj.HolderIdentity) == "" {
		return true
	}
	return false
}

// parseLeaseValue will parse a leaseValue string to lease object
func parseLeaseValue(leaseValue string) (lease, error) {
	leaseValueObj := &lease{}
	err := json.Unmarshal([]byte(leaseValue), leaseValueObj)
	if err != nil {
		return lease{}, err
	}
	return *leaseValueObj, nil
}
