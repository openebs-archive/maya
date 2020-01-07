// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

const (
	// CSPCFinalizer represents finalizer value used by cspc
	CSPCFinalizer = "cstorpoolcluster.openebs.io/finalizer"
	// PoolProtectionFinalizer is used to make sure cspi and it's bdcs
	// are not deleted before destroying the zpool
	PoolProtectionFinalizer = "openebs.io/pool-protection"
)

// CSPC is a wrapper over cstorpoolcluster api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type CSPC struct {
	object *apisv1alpha1.CStorPoolCluster
	// kubeconfig path
	configPath string
}

// CSPCList is a wrapper over cstorpoolcluster api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type CSPCList struct {
	items []*CSPC
}

// Len returns the number of items present
// in the CSPCList
func (c *CSPCList) Len() int {
	if c == nil {
		return 0
	}
	return len(c.items)
}

// ToAPIList converts CSPCList to API CSPCList
func (c *CSPCList) ToAPIList() *apisv1alpha1.CStorPoolClusterList {
	clist := &apisv1alpha1.CStorPoolClusterList{}
	for _, cspc := range c.items {
		clist.Items = append(clist.Items, *cspc.object)
	}
	return clist
}

type cspcBuildOption func(*CSPC)

// NewForAPIObject returns a new instance of CSPC
func NewForAPIObject(obj *apisv1alpha1.CStorPoolCluster, opts ...cspcBuildOption) *CSPC {
	c := &CSPC{object: obj}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided cspc instance
type Predicate func(*CSPC) bool

// IsNil returns true if the CSPC instance
// is nil
func (c *CSPC) IsNil() bool {
	return c.object == nil
}

// IsNil is predicate to filter out nil CSPC
// instances
func IsNil() Predicate {
	return func(c *CSPC) bool {
		return c.IsNil()
	}
}

// PredicateList holds a list of predicate
type PredicateList []Predicate

// all returns true if all the predicates
// succeed against the provided CSPC
// instance
func (l PredicateList) all(p *CSPC) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// BuilderForAPIObject returns an instance of the Builder object based on cspc api object.
func BuilderForAPIObject(cspc *apisv1alpha1.CStorPoolCluster) *Builder {
	return &Builder{
		cspc: &CSPC{cspc, ""},
	}
}

// WithFilter adds filters on which the cspc's has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// HasFinalizer is a predicate to filter out based on provided
// finalizer being present on the object.
func HasFinalizer(finalizer string) Predicate {
	return func(cspc *CSPC) bool {
		return cspc.HasFinalizer(finalizer)
	}
}

// HasFinalizer returns true if the provided finalizer is present on the object.
func (c *CSPC) HasFinalizer(finalizer string) bool {
	finalizersList := c.object.GetFinalizers()
	return util.ContainsString(finalizersList, finalizer)
}

// RemoveFinalizer removes the given finalizer from the object.
func (c *CSPC) RemoveFinalizer(finalizer string) error {
	if len(c.object.Finalizers) == 0 {
		klog.V(2).Infof("no finalizer present on CSPC %s", c.object.Name)
		return nil
	}

	if !c.HasFinalizer(finalizer) {
		klog.V(2).Infof("finalizer %s is already removed on CSPC %s", finalizer, c.object.Name)
		return nil
	}

	c.object.Finalizers = util.RemoveString(c.object.Finalizers, finalizer)

	_, err := NewKubeClient(WithKubeConfigPath(c.configPath)).
		WithNamespace(c.object.Namespace).
		Update(c.object)
	if err != nil {
		return errors.Wrap(err, "failed to update object while removing finalizer")
	}
	klog.Infof("Finalizer %s removed successfully from CSPC %s", finalizer, c.object.Name)
	return nil
}

// AddFinalizer adds the given finalizer to the object.
func (c *CSPC) AddFinalizer(finalizer string) (*apisv1alpha1.CStorPoolCluster, error) {
	if c.HasFinalizer(finalizer) {
		klog.V(2).Infof("finalizer %s is already present on CSPC %s", finalizer, c.object.Name)
		return c.object, nil
	}

	c.object.Finalizers = append(c.object.Finalizers, finalizer)

	cspcAPIObj, err := NewKubeClient(WithKubeConfigPath(c.configPath)).
		WithNamespace(c.object.Namespace).
		Update(c.object)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to update CSPC %s while adding finalizer %s",
			c.object.Name, finalizer)
	}

	klog.Infof("Finalizer %s added on CSPC %s", finalizer, c.object.Name)
	return cspcAPIObj, nil
}
