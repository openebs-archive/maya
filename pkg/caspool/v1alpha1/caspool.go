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
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// CasPool encapsulates CasPool object.
type CasPool struct {
	// actual CasPool object
	Object *apisv1alpha1.CasPool
}

// Builder is the builder object for CasPool.
type Builder struct {
	CasPoolObject *CasPool
}

// NewBuilder returns an empty instance of the Builder object.
func NewBuilder() *Builder {
	return &Builder{
		CasPoolObject: &CasPool{&apisv1alpha1.CasPool{}},
	}
}

// Build returns the CasPool object built by this builder.
func (cb *Builder) Build() *CasPool {
	return cb.CasPoolObject
}

// WithCasTemplateName builds with cas template name.
func (cb *Builder) WithCasTemplateName(casTemplateName string) *Builder {
	cb.CasPoolObject.Object.CasCreateTemplate = casTemplateName
	return cb
}

// WithCspcName builds with cspc name.
func (cb *Builder) WithCspcName(name string) *Builder {
	cb.CasPoolObject.Object.CStorPoolCluster = name
	return cb
}

// WithNodeName builds with node name.
func (cb *Builder) WithNodeName(nodeName string) *Builder {
	cb.CasPoolObject.Object.NodeName = nodeName
	return cb
}

// WithPoolType builds with pool type.
func (cb *Builder) WithPoolType(poolType string) *Builder {
	cb.CasPoolObject.Object.PoolType = poolType
	return cb
}

// WithMaxPool builds with max pool.
func (cb *Builder) WithMaxPool(cspc *apisv1alpha1.CStorPoolCluster) *Builder {
	cb.CasPoolObject.Object.MaxPools = *cspc.Spec.MaxPools
	return cb
}

// WithDiskType builds with disk type.
func (cb *Builder) WithDiskType(diskType string) *Builder {
	cb.CasPoolObject.Object.Type = diskType
	return cb
}

// WithAnnotations builds with annotations.
func (cb *Builder) WithAnnotations(annotations map[string]string) *Builder {
	cb.CasPoolObject.Object.Annotations = annotations
	return cb
}

// WithDiskGroup builds with disk group.
func (cb *Builder) WithDiskGroup(diskGroup []apisv1alpha1.CStorPoolClusterDiskGroups) *Builder {
	cb.CasPoolObject.Object.DiskGroups = diskGroup
	return cb
}

// WithDiskDeviceIDMap builds with device ID map.
func (cb *Builder) WithDiskDeviceIDMap(diskDeviceIDMap map[string]string) *Builder {
	cb.CasPoolObject.Object.DiskDeviceIDMap = diskDeviceIDMap
	return cb
}
