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
	poolspec "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/cstorpoolspecs"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
)

// Builder is the builder object for CSPC
type Builder struct {
	cspc *CSPC
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{cspc: &CSPC{object: &apisv1alpha1.CStorPoolCluster{}}}
}

// WithName sets the Name field of CSPC with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build CSPC object: missing CSPC name"))
		return b
	}
	b.cspc.object.Name = name
	return b
}

// WithPoolSpecBuilder adds a pool to this cspc object.
//
// NOTE:
//   poolspec details are present in the provided pool spec
// builder object
func (b *Builder) WithPoolSpecBuilder(poolSpecBuilder *poolspec.Builder) *Builder {
	poolspecObj, err := poolSpecBuilder.Build()
	if err != nil {
		b.errs = append(b.errs, errors.Wrap(err, "failed to build pool spec"))
		return b
	}
	b.cspc.object.Spec.Pools = append(
		b.cspc.object.Spec.Pools,
		*poolspecObj.ToAPI(),
	)
	return b
}

// Build returns the CSPC  instance
func (b *Builder) Build() (*CSPC, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.cspc, nil
}
