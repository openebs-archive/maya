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

package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	oeapis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	//clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset"
)

const (
	// StorageNodePredicateKey is the key for StorageNodePredicate function.
	FilterInactive        = "filterInactive"
	FilterInactiveReverse = "filterInactiveReverse"
)

// KubernetesClient is the kubernetes client which will implement disk actions/behaviours.
type KubernetesClient struct {
	// kubeclientset is a standard kubernetes clientset
	Kubeclientset kubernetes.Interface

	// NDMclientset is a NDM custom resource package generated for custom API group.
	NDMClientset clientset.Interface
}

type SpcObjectClient struct {
	*KubernetesClient
	Spc *oeapis.StoragePoolClaim
}

type errs []error

type Disk struct {
	*apis.Disk
	errs
}

type DiskList struct {
	*apis.DiskList
	errs
}

// buildOptionFunc is the typed function to build disk object.
type buildOptionFunc func(*Disk)

// predicate is the typed predicate function to validate disk object.
type predicate func(*Disk) (message string, ok bool)

// filterOptionFunc is the typed function to filter disk objects.
type filterOptionFunc func(original *DiskList) *DiskList

// DiskInterface abstracts operations on disk entity.
// Different orchestrators may need to implement this interface.
type DiskInterface interface {
	Get(name string) (*Disk, error)
	List(opts v1.ListOptions) (*DiskList, error)
	Create(*apis.Disk) (*Disk, error)
}

// checkPredicatesFuncs is an array of check predicate functions.
var checkPredicatesFuncs = [...]predicate{
	checkName,
}

// filterPredicatesFuncMap is an array of filter predicate functions.
// filter predicates should be tunable by client.
var filterOptionFuncMap = map[string]filterOptionFunc{
	FilterInactive:        filterInactive,
	FilterInactiveReverse: filterInactiveReverse,
}

// predicateFailedError returns the predicate error which is provided to this function as an argument. .
func predicateFailedError(message string) error {
	return errors.Errorf("predicatefailed: %s", message)
}

// New is a constructor returns a new instance of disk
func New(opts ...buildOptionFunc) *Disk {
	r := &Disk{Disk: &apis.Disk{}}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Build returns the final instance of disk
func (d *Disk) Build() (*apis.Disk, []error) {
	d.validate()
	if len(d.errs) > 0 {
		return nil, d.errs
	}
	return d.Disk, nil
}

// validate validates the disk object against the check predicates.
func (d *Disk) validate() {
	for _, c := range checkPredicatesFuncs {
		msg, ok := c(d)
		if !ok {
			d.errs = append(d.errs, predicateFailedError(msg))
		}
	}
}

// WithName method fills the name field of disk object.
func (d *Disk) WithName(name string) *Disk {
	WithName(name)(d)
	return d
}

// WithName function is used by WithName method as a util.
// Ideas is to give flexibility for building object by using dot operator as well as passing
// build predicated to the New constructor.
func WithName(name string) buildOptionFunc {
	return func(r *Disk) {
		r.Disk.Name = name
	}
}

// WithState method fills the name field of disk object.
func (d *Disk) WithState(state string) *Disk {
	WithState(state)(d)
	return d
}

// WithState function is used by WithState method as a util.
// Ideas is to give flexibility for building object by using dot operator as well as passing
// build predicated to the New constructor.
func WithState(state string) buildOptionFunc {
	return func(r *Disk) {
		r.Disk.Status.State = state
	}
}

//checkName validate the name field of disk object.
func checkName(db *Disk) (string, bool) {
	if db.Disk.Name == "" {
		//TODO: Think about having some good organization in putting error messages.
		return "checkName predicate failed:name field on the object may not be empty", false
	}
	return "", true
}

func (d *DiskList) Filter(predicateKeys ...string) *DiskList {
	// Initialize filtered disk list
	filteredDiskList := &DiskList{
		DiskList: &apis.DiskList{},
		errs:     nil,
	}
	errMSg, ok := d.Hasitem()
	if !ok {
		filteredDiskList.errs = append(filteredDiskList.errs, errors.New(errMSg))
		return filteredDiskList
	}
	filteredDiskList = d
	for _, key := range predicateKeys {
		filteredDiskList = filterOptionFuncMap[key](filteredDiskList)
	}
	return filteredDiskList
}

func (d *DiskList) FilterAny(predicateKeys ...string) *DiskList {
	// Initialize filtered disk list
	filteredDiskList := &DiskList{
		DiskList: &apis.DiskList{},
		errs:     nil,
	}
	errMSg, ok := d.Hasitem()
	if !ok {
		filteredDiskList.errs = append(filteredDiskList.errs, errors.New(errMSg))
		return filteredDiskList
	}
	for _, key := range predicateKeys {
		resultList := filterOptionFuncMap[key](d)
		filteredDiskList.DiskList.Items = append(filteredDiskList.Items, resultList.Items...)
	}
	return filteredDiskList
}

//filterInactive filter and give out all the inactive disks.
func filterInactive(orignialList *DiskList) *DiskList {
	filteredList := &DiskList{
		DiskList: &apis.DiskList{},
		errs:     nil,
	}
	for _, disk := range orignialList.Items {
		if disk.Status.State == "Inactive" {
			filteredList.Items = append(filteredList.Items, disk)
		}
	}
	return filteredList
}

//filterInactiveReverse give out all the disks except inactive disk.
func filterInactiveReverse(orignialList *DiskList) *DiskList {
	filteredList := &DiskList{
		DiskList: &apis.DiskList{},
		errs:     nil,
	}
	for _, disk := range orignialList.Items {
		if !(disk.Status.State == "Inactive") {
			filteredList.Items = append(filteredList.Items, disk)
		}
	}
	return filteredList
}

func (d *DiskList) Hasitem() (string, bool) {
	if d == nil || d.DiskList == nil || d.Items == nil {
		return "No item found in disk list", false
	}
	return "", true
}
