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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CStorVolumeClaimLister helps list CStorVolumeClaims.
type CStorVolumeClaimLister interface {
	// List lists all CStorVolumeClaims in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.CStorVolumeClaim, err error)
	// CStorVolumeClaims returns an object that can list and get CStorVolumeClaims.
	CStorVolumeClaims(namespace string) CStorVolumeClaimNamespaceLister
	CStorVolumeClaimListerExpansion
}

// cStorVolumeClaimLister implements the CStorVolumeClaimLister interface.
type cStorVolumeClaimLister struct {
	indexer cache.Indexer
}

// NewCStorVolumeClaimLister returns a new CStorVolumeClaimLister.
func NewCStorVolumeClaimLister(indexer cache.Indexer) CStorVolumeClaimLister {
	return &cStorVolumeClaimLister{indexer: indexer}
}

// List lists all CStorVolumeClaims in the indexer.
func (s *cStorVolumeClaimLister) List(selector labels.Selector) (ret []*v1alpha1.CStorVolumeClaim, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CStorVolumeClaim))
	})
	return ret, err
}

// CStorVolumeClaims returns an object that can list and get CStorVolumeClaims.
func (s *cStorVolumeClaimLister) CStorVolumeClaims(namespace string) CStorVolumeClaimNamespaceLister {
	return cStorVolumeClaimNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CStorVolumeClaimNamespaceLister helps list and get CStorVolumeClaims.
type CStorVolumeClaimNamespaceLister interface {
	// List lists all CStorVolumeClaims in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.CStorVolumeClaim, err error)
	// Get retrieves the CStorVolumeClaim from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.CStorVolumeClaim, error)
	CStorVolumeClaimNamespaceListerExpansion
}

// cStorVolumeClaimNamespaceLister implements the CStorVolumeClaimNamespaceLister
// interface.
type cStorVolumeClaimNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all CStorVolumeClaims in the indexer for a given namespace.
func (s cStorVolumeClaimNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.CStorVolumeClaim, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CStorVolumeClaim))
	})
	return ret, err
}

// Get retrieves the CStorVolumeClaim from the indexer for a given namespace and name.
func (s cStorVolumeClaimNamespaceLister) Get(name string) (*v1alpha1.CStorVolumeClaim, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("cstorvolumeclaim"), name)
	}
	return obj.(*v1alpha1.CStorVolumeClaim), nil
}
