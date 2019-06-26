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

package cstorpoolspecs

import (
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	raidgroup "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/raidgroups"
	"github.com/pkg/errors"
)

// Builder is the builder object for PS
type Builder struct {
	ps   *PS
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{ps: &PS{object: &apisv1alpha1.PoolSpec{}}}
}

// AppendErrorToBuilder appends error to the builder object
func (b *Builder) AppendErrorToBuilder(err error) *Builder {
	b.errs = append(b.errs, err)
	return b
}

// WithNodeSelector sets the node selector with provided value.
func (b *Builder) WithNodeSelector(nodeSelector map[string]string) *Builder {
	if len(nodeSelector) == 0 {
		b.errs = append(b.errs, errors.New("failed to build pool spec object: missing node selector"))
		return b
	}
	b.ps.object.NodeSelector = nodeSelector
	return b
}

// WithCacheFilePath sets the cacheFile field of pool spec with provided value.
func (b *Builder) WithCacheFilePath(cacheFile string) *Builder {
	if len(cacheFile) == 0 {
		b.errs = append(b.errs, errors.New("failed to build pool spec object: missing cache file path"))
		return b
	}
	b.ps.object.PoolConfig.CacheFile = cacheFile
	return b
}

// WithDefaultRaidGroupType sets the cacheFile field of pool spec with provided value.
func (b *Builder) WithDefaultRaidGroupType(cacheFile string) *Builder {
	if len(cacheFile) == 0 {
		b.errs = append(b.errs, errors.New("failed to build pool spec object: missing cache file path"))
		return b
	}
	b.ps.object.PoolConfig.CacheFile = cacheFile
	return b
}

// WithOverProvisioning flags OverProvisioning field of the pool spec.
func (b *Builder) WithOverProvisioning() *Builder {
	b.ps.object.PoolConfig.OverProvisioning = true
	return b
}

// WithCompression sets the Compression field of pool spec with provided value.
func (b *Builder) WithCompression(compressionType string) *Builder {
	if len(compressionType) == 0 {
		b.errs = append(b.errs, errors.New("failed to build pool spec object: missing compression type"))
		return b
	}
	b.ps.object.PoolConfig.Compression = compressionType
	return b
}

// WithRaidGroupBuilder adds a raid group to this pool spec object.
//
// NOTE:
//   raidGroup details are present in the provided raidGroup
// builder object
func (b *Builder) WithRaidGroupBuilder(raidGroupBuilder *raidgroup.Builder) *Builder {
	raidgroupObj, err := raidGroupBuilder.Build()
	if err != nil {
		b.errs = append(b.errs, errors.Wrap(err, "failed to build raid group"))
		return b
	}
	b.ps.object.RaidGroups = append(
		b.ps.object.RaidGroups,
		*raidgroupObj.ToAPI(),
	)
	return b
}

// Build returns the PoolSpec API instance
func (b *Builder) Build() (*PS, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.ps, nil
}

// ToAPI  returns the poolspec api object from the builder object.
func (ps *PS) ToAPI() *apisv1alpha1.PoolSpec {
	return ps.object
}
