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

package raidgroups

import (
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
)

// Builder is the builder object for RG
type Builder struct {
	rg   *RG
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{rg: &RG{object: &apisv1alpha1.RaidGroup{}}}
}

// WithType sets the type field of raid group with provided value.
func (b *Builder) WithType(poolType string) *Builder {
	if len(poolType) == 0 {
		b.errs = append(b.errs, errors.New("failed to build raid group object: missing group type"))
		return b
	}
	b.rg.object.Type = poolType
	return b
}

// WithName sets the name field of raid group with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build raid group object: missing group name"))
		return b
	}
	b.rg.object.Name = name
	return b
}

// WithWriteCache flags the IsWriteCache field of raid group.
func (b *Builder) WithWriteCache(cacheFile string) *Builder {
	b.rg.object.IsWriteCache = true
	return b
}

// WithSpare flags IsSpare field of raid group.
func (b *Builder) WithSpare() *Builder {
	b.rg.object.IsSpare = true
	return b
}

// WithReadCache flags the IsReadCache field of raid group.
func (b *Builder) WithReadCache() *Builder {
	b.rg.object.IsReadCache = true
	return b
}

// Build returns the raid group
func (b *Builder) Build() (*RG, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.rg, nil
}

// ToAPI returns raid group api object from the Raid Group object.
func (rg *RG) ToAPI() *apisv1alpha1.RaidGroup {
	return rg.object
}
