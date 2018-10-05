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
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	"k8s.io/client-go/kubernetes"
)

// ToDo : Move this file to pkg
// This struct will be used as a value of lease key that will
// give information about an acquired lease on object
// The struct object will be parsed to string which will be then
// put as a value to the lease key of object annotation.
type lease struct {
	// Holder is the namespace/name of the pod who acquires the lease
	Holder           string `json:"holder"`
	LeaderTransition int    `json:"leaderTransition"`
	// More specific details can be added here that will describe the
	// current state of lease in more details.
	// e.g. acquiredTimeStamp, self-release etc
	// acquiredTimeStamp will tell when the lease was acquired
	// self-release will tell whether the lease was removed by the acquirer or not
}

// Leases is an interface which assists in getting and releasing lease on an object
type Leases interface {
	// Hold will try to get a lease, in case of failure it will return error
	Hold() error
	// Update will update the lease value on the object
	Update(leaseValue string) error
	// Release will remove the acquired lease on the object
	Release()
}

// Lease is the struct which will implement the Leases interface
type Lease struct {
	// Object is the object over which lease is to be taken
	Object interface{}
	// leaseKey is lease key on object
	leaseKey string
	// oecs is the openebs clientset
	oecs clientset.Interface
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
}
