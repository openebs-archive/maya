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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	//clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
)

const (
	// filterStripedPredicateKey is the key for filterStriped function.
	filterStripedPredicateKey = "filterStriped"
)

// KubernetesClient is the kubernetes client which will implement storagepool actions/behaviours.
type KubernetesClient struct {
	// kubeclientset is a standard kubernetes clientset
	Kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	Clientset clientset.Interface
}

type errs []error

type StoragePool struct {
	StoragePool *apis.StoragePool
	errs
}

type StoragePoolList struct {
	StoragePoolList         *apis.StoragePoolList
	FilteredStoragePoolList *apis.StoragePoolList
	errs
}

// buildPredicate is the typed predicate function to build storagepool object.
type buildPredicate func(*StoragePool)

// checkPredicate is the typed predicate function to validate storagepool object.
type checkPredicate func(*StoragePool) (message string, ok bool)

// filterPredicate is the typed predicate function to filter storagepool objects.
type filterPredicate func(*StoragePoolList)

// StoragepoolInterface abstracts operations on storagepool entity.
// Different orchestrators may need to implement this interface.
type StoragepoolInterface interface {
	Get(name string) (*StoragePool, error)
	List(opts v1.ListOptions) (*StoragePoolList, error)
	Create(storagepool *apis.StoragePool) (*StoragePool, error)
}

// checkPredicatesFuncs is an array of check predicate functions.
var checkPredicatesFuncs = [...]checkPredicate{
	checkName,
}

// filterPredicatesFuncMap is an array of filter predicate functions.
// filter predicates should be tunable by client.
var filterPredicatesFuncMap = map[string]filterPredicate{
	filterStripedPredicateKey: filterStriped,
}

// predicateFailedError returns the predicate error which is provided to this function as an argument. .
func predicateFailedError(message string) error {
	return errors.Errorf("predicatefailed: %s", message)
}

// New is a constructor returns a new instance of storagepool
func New(opts ...buildPredicate) *StoragePool {
	r := &StoragePool{StoragePool: &apis.StoragePool{}}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Build returns the final instance of storagepool
func (d *StoragePool) Build() (*apis.StoragePool, []error) {
	d.validate()
	if len(d.errs) > 0 {
		return nil, d.errs
	}
	return d.StoragePool, nil
}

// validate validates the storagepool object against the check predicates.
func (d *StoragePool) validate() {
	for _, c := range checkPredicatesFuncs {
		msg, ok := c(d)
		if !ok {
			d.errs = append(d.errs, predicateFailedError(msg))
		}
	}
}

// WithName method fills the name field of storagepool object.
func (d *StoragePool) WithName(name string) *StoragePool {
	WithName(name)(d)
	return d
}

// WithName function is used by WithName method as a util.
// Ideas is to give flexibility for building object by using dot operator as well as passing
// build predicated to the New constructor.
func WithName(name string) buildPredicate {
	return func(r *StoragePool) {
		r.StoragePool.Name = name
	}
}

// WithPoolType method fills the name field of storagepool object.
func (d *StoragePool) WithPoolType(state string) *StoragePool {
	WithPoolType(state)(d)
	return d
}

// WithPoolType function is used by WithPoolType method as a util.
// Ideas is to give flexibility for building object by using dot operator as well as passing
// build predicates to the New constructor.
func WithPoolType(state string) buildPredicate {
	return func(r *StoragePool) {
		r.StoragePool.Spec.PoolSpec.PoolType = state
	}
}

//checkName validate the name field of StoragePool object.
func checkName(db *StoragePool) (string, bool) {
	if db.StoragePool.Name == "" {
		//TODO: Think about having some good organization in putting error messages.
		return "checkName predicate failed:name field on the object may not be empty", false
	}
	return "", true
}

func (d *StoragePoolList) filter(predicateKeys ...string) {
	// Initialize filtered storagepool list
	d.FilteredStoragePoolList = &apis.StoragePoolList{}
	for _, key := range predicateKeys {
		filterPredicatesFuncMap[key](d)
	}
}

//filterStriped filters out all the inactive storagepools.
func filterStriped(db *StoragePoolList) {
	for _, storagepool := range db.StoragePoolList.Items {
		if storagepool.Spec.PoolSpec.PoolType == string(apis.PoolTypeStripedCPV) {
			db.FilteredStoragePoolList.Items = append(db.FilteredStoragePoolList.Items, storagepool)
		}
	}
}
