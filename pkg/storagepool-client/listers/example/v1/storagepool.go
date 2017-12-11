/*
Copyright 2017 The OpenEBS Authors

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

// This file was automatically generated by lister-gen

package v1

import (
	v1 "github.com/openebs/maya/pkg/storagepool-apis/openebs.io/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// StoragepoolLister helps list Storagepools.
type StoragepoolLister interface {
	// List lists all Storagepools in the indexer.
	List(selector labels.Selector) (ret []*v1.Storagepool, err error)
	// Storagepools returns an object that can list and get Storagepools.
	Storagepools(namespace string) StoragepoolNamespaceLister
	StoragepoolListerExpansion
}

// storagepoolLister implements the StoragepoolLister interface.
type storagepoolLister struct {
	indexer cache.Indexer
}

// NewStoragepoolLister returns a new StoragepoolLister.
func NewStoragepoolLister(indexer cache.Indexer) StoragepoolLister {
	return &storagepoolLister{indexer: indexer}
}

// List lists all Storagepools in the indexer.
func (s *storagepoolLister) List(selector labels.Selector) (ret []*v1.Storagepool, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Storagepool))
	})
	return ret, err
}

// Storagepools returns an object that can list and get Storagepools.
func (s *storagepoolLister) Storagepools(namespace string) StoragepoolNamespaceLister {
	return storagepoolNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// StoragepoolNamespaceLister helps list and get Storagepools.
type StoragepoolNamespaceLister interface {
	// List lists all Storagepools in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.Storagepool, err error)
	// Get retrieves the Storagepool from the indexer for a given namespace and name.
	Get(name string) (*v1.Storagepool, error)
	StoragepoolNamespaceListerExpansion
}

// storagepoolNamespaceLister implements the StoragepoolNamespaceLister
// interface.
type storagepoolNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Storagepools in the indexer for a given namespace.
func (s storagepoolNamespaceLister) List(selector labels.Selector) (ret []*v1.Storagepool, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Storagepool))
	})
	return ret, err
}

// Get retrieves the Storagepool from the indexer for a given namespace and name.
func (s storagepoolNamespaceLister) Get(name string) (*v1.Storagepool, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("storagepool"), name)
	}
	return obj.(*v1.Storagepool), nil
}
