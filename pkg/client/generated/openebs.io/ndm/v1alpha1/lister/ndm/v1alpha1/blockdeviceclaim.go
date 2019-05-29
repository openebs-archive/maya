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
	v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// BlockDeviceClaimLister helps list BlockDeviceClaims.
type BlockDeviceClaimLister interface {
	// List lists all BlockDeviceClaims in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.BlockDeviceClaim, err error)
	// Get retrieves the BlockDeviceClaim from the index for a given name.
	Get(name string) (*v1alpha1.BlockDeviceClaim, error)
	BlockDeviceClaimListerExpansion
}

// blockDeviceClaimLister implements the BlockDeviceClaimLister interface.
type blockDeviceClaimLister struct {
	indexer cache.Indexer
}

// NewBlockDeviceClaimLister returns a new BlockDeviceClaimLister.
func NewBlockDeviceClaimLister(indexer cache.Indexer) BlockDeviceClaimLister {
	return &blockDeviceClaimLister{indexer: indexer}
}

// List lists all BlockDeviceClaims in the indexer.
func (s *blockDeviceClaimLister) List(selector labels.Selector) (ret []*v1alpha1.BlockDeviceClaim, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.BlockDeviceClaim))
	})
	return ret, err
}

// Get retrieves the BlockDeviceClaim from the index for a given name.
func (s *blockDeviceClaimLister) Get(name string) (*v1alpha1.BlockDeviceClaim, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("blockdeviceclaim"), name)
	}
	return obj.(*v1alpha1.BlockDeviceClaim), nil
}
