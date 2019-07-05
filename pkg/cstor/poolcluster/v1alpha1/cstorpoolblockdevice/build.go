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

package cstorpoolclusterblockdevice

import (
	"github.com/pkg/errors"

	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// Builder is the builder object for CSPC blockdevices
type Builder struct {
	cspcbd *CSPCBlockDevice
	errs   []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		cspcbd: &CSPCBlockDevice{
			object: &apisv1alpha1.CStorPoolClusterBlockDevice{},
		},
	}
}

// WithBlockDeviceName sets the BlockDeviceName field of
// cstorpoolclusterblockdevice
func (b *Builder) WithBlockDeviceName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cspc blockdeviceobject: missing blockdevice name"),
		)
		return b
	}
	b.cspcbd.object.BlockDeviceName = name
	return b
}

// WithCapacity sets the capacity field of cstorpoolclusterblockdevice
func (b *Builder) WithCapacity(capacity string) *Builder {
	if len(capacity) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cspc blockdeviceobject: missing blockdevice capacity"),
		)
		return b
	}
	b.cspcbd.object.Capacity = capacity
	return b
}

// WithDevLink sets the devLink field of cstorpoolclusterblockdevice
func (b *Builder) WithDevLink(devLink string) *Builder {
	if len(devLink) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cspc blockdeviceobject: missing blockdevice devLink"),
		)
		return b
	}
	b.cspcbd.object.DevLink = devLink
	return b
}

// Build returns the CStorPoolClusterBlockDevice API instance
func (b *Builder) Build() (*CSPCBlockDevice, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.cspcbd, nil
}

// ToAPI  returns the cstorpoolclusterblockdevice api object from the builder object.
func (cspcbd *CSPCBlockDevice) ToAPI() *apisv1alpha1.CStorPoolClusterBlockDevice {
	return cspcbd.object
}
