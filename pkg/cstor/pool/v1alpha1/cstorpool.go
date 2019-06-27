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
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
)

const (
	// StorageNodePredicateKey is the key for StorageNodePredicate function.
	filterHealthyPredicateKey = "filterHealthy"
)

// KubernetesClient is the kubernetes client which will implement cstorpool actions/behaviours.
type KubernetesClient struct {
	// kubeclientset is a standard kubernetes clientset
	Kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	Clientset clientset.Interface
}

type errs []error

type CStorPool struct {
	CStorPool *apis.CStorPool
	errs
}

type CStorPoolList struct {
	CStorPoolList         *apis.CStorPoolList
	FilteredCStorPoolList *apis.CStorPoolList
	errs
}

// buildPredicate is the typed predicate function to build cstorpool object.
type buildPredicate func(*CStorPool)

// checkPredicate is the typed predicate function to validate cstorpool object.
type checkPredicate func(*CStorPool) (message string, ok bool)

// filterPredicate is the typed predicate function to filter cstorpool objects.
type filterPredicate func(*CStorPoolList)

// cstorpoolInterface abstracts operations on cstorpool entity.
// Different orchestrators may need to implement this interface.
type CstorpoolInterface interface {
	Get(name string) (*CStorPool, error)
	List(opts v1.ListOptions) (*CStorPoolList, error)
	Create(*apis.CStorPool) (*CStorPool, error)
}

// checkPredicatesFuncs is an array of check predicate functions.
var checkPredicatesFuncs = [...]checkPredicate{
	checkName,
}

// filterPredicatesFuncMap is an array of filter predicate functions.
// filter predicates should be tunable by client.
var filterPredicatesFuncMap = map[string]filterPredicate{
	filterHealthyPredicateKey: filterHealthy,
}

// predicateFailedError returns the predicate error which is provided to this function as an argument. .
func predicateFailedError(message string) error {
	return errors.Errorf("predicatefailed: %s", message)
}

// New is a constructor returns a new instance of cstorpool
func New(opts ...buildPredicate) *CStorPool {
	r := &CStorPool{CStorPool: &apis.CStorPool{}}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Build returns the final instance of cstorpool
func (d *CStorPool) Build() (*apis.CStorPool, []error) {
	d.validate()
	if len(d.errs) > 0 {
		return nil, d.errs
	}
	return d.CStorPool, nil
}

// validate validates the cstorpool object against the check predicates.
func (d *CStorPool) validate() {
	for _, c := range checkPredicatesFuncs {
		msg, ok := c(d)
		if !ok {
			d.errs = append(d.errs, predicateFailedError(msg))
		}
	}
}

// WithName method fills the name field of cstorpool object.
func (d *CStorPool) WithName(name string) *CStorPool {
	WithName(name)(d)
	return d
}

// WithName function is used by WithName method as a util.
// Ideas is to give flexibility for building object by using dot operator as well as passing
// build predicated to the New constructor.
func WithName(name string) buildPredicate {
	return func(r *CStorPool) {
		r.CStorPool.Name = name
	}
}

// WithPhase method fills the name field of cstorpool object.
func (d *CStorPool) WithPhase(state string) *CStorPool {
	WithPhase(state)(d)
	return d
}

// WithPhase function is used by WithPhase method as a util.
// Ideas is to give flexibility for building object by using dot operator as well as passing
// build predicated to the New constructor.
func WithPhase(state string) buildPredicate {
	return func(r *CStorPool) {
		r.CStorPool.Status.Phase = apis.CStorPoolPhase(state)
	}
}

//checkName validate the name field of cstorpool object.
func checkName(db *CStorPool) (string, bool) {
	if db.CStorPool.Name == "" {
		//TODO: Think about having some good organization in putting error messages.
		return "checkName predicate failed:name field on the object may not be empty", false
	}
	return "", true
}

func (d *CStorPoolList) filter(predicateKeys ...string) {
	// Initialize filtered cstorpool list
	d.FilteredCStorPoolList = &apis.CStorPoolList{}
	for _, key := range predicateKeys {
		filterPredicatesFuncMap[key](d)
	}
}

//filterInactive filters out all the inactive cstorpools.
func filterHealthy(db *CStorPoolList) {
	for _, cstorpool := range db.CStorPoolList.Items {
		if cstorpool.Status.Phase == apis.CStorPoolStatusOnline {
			db.FilteredCStorPoolList.Items = append(db.FilteredCStorPoolList.Items, cstorpool)
		}
	}
}
