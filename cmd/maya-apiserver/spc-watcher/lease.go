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
	"fmt"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"strings"
)

const (
	// SpcLeaseKey is the key that will be used to acquire lease on spc object.
	// It will be present in spc annotations
	// If key has an empty value, that means no one has acquired a lease on spc object.
	SpcLeaseKey = "openebs.io/spc-create-lease"
	// PodNameEnvKey is the key to fetch name of the pod,which will be combined with PodNameSpaceEnvKey
	// to be used as a value of SpcLeaseKey.
	// e.g. "openebs.io/spc-lease":"openebs/maya-apiserver-6b4695c9f8-nbwl9"
	PodNameEnvKey = "OPENEBS_MAYA_POD_NAME"
	// PodNameSpaceEnvKey is the key to fetch the namespace of the pod
	PodNameSpaceEnvKey = "OPENEBS_NAMESPACE"
)

// Leases is an interface which assists in getting and releasing lease over an spc object
type Leases interface {
	// GetLease will try to get a lease on spc, in case of failure it will return error
	GetLease() (string, error)
	// UpdateLease will update the lease value of the spc
	UpdateLease(leaseValue string) (*apis.StoragePoolClaim, error)
	// RemoveLease will remove the acquired lease on the spc
	RemoveLease() *apis.StoragePoolClaim
}

// spcLease is the struct which will implement the Leases interface
type spcLease struct {
	// spcObject is the storagepoolclaim object over which lease is to be taken
	spcObject *apis.StoragePoolClaim
	// leaseKey is lease key on current storagepoolclaim object
	leaseKey string
	// oecs is the openebs clientset
	oecs clientset.Interface
}

func (sl *spcLease) GetLease() (string, error) {
	// Get the lease value.
	leaseValue := sl.spcObject.Annotations[sl.leaseKey]
	// If leaseValue is empty acquire lease.
	if strings.TrimSpace(leaseValue) == "" {
		spcObject, err := sl.UpdateLease(sl.getPodName())
		if err != nil {
			return "", err
		}
		return spcObject.Annotations[sl.leaseKey], nil
	}
	// If leaseValue is not empty, lease cannot be acquired.
	return "", fmt.Errorf("lease on spc already acquired")
}

func (sl *spcLease) UpdateLease(leaseValue string) (*apis.StoragePoolClaim, error) {
	newSpcObject := sl.spcObject
	if newSpcObject.Annotations == nil {
		// make a map that should contain the lease key in spc
		mapLease := make(map[string]string)

		// Fill the map lease key with lease value
		mapLease[sl.leaseKey] = leaseValue
		newSpcObject.Annotations = mapLease

	} else {
		newSpcObject.Annotations[sl.leaseKey] = leaseValue
	}
	spcObject, err := sl.oecs.OpenebsV1alpha1().StoragePoolClaims().Update(sl.spcObject)
	if err != nil {
		return nil, err
	}
	return spcObject, nil
}

// Can be used for reconcile loop use cases
// TODO Remove using patch instead of update
func (sl *spcLease) RemoveLease() *apis.StoragePoolClaim {
	spcObject, err := sl.UpdateLease("")
	if err != nil {
		runtime.HandleError(fmt.Errorf("Lease could not be removed:%v", err))
	}
	return spcObject
}

func (sl *spcLease) getPodName() string {
	podName := env.Get(PodNameEnvKey)
	podNameSpace := env.Get(PodNameSpaceEnvKey)
	return podNameSpace + "/" + podName
}
