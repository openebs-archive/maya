/*
Copyright 2017 The OpenEBS Authors.

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

package v1

import (
	"fmt"

	"github.com/openebs/maya/types/v1"
)

// ReqdPolicies will enforce policies on an OpenEBS volume. Required
// policies will be extracted from volume's Labels & will be enforced
// into specific properties of the same volume instance.
//
// Required Policies are:
// 1. VolumeType
// 2. OrchProvider
//
// TIP:
//  Read this as Enforce Required Policies
type ReqdPolicies struct {
	// volume is the instance on which policies will be enforced
	volume *v1.Volume

	// volType is an initial property
	volType v1.VolumeType

	// orch is an initial property
	orch v1.OrchProvider
}

// Enforce will enforce initial properties against
// the volume instance
func (p *ReqdPolicies) Enforce(volume *v1.Volume) (*v1.Volume, error) {
	if volume == nil {
		return nil, fmt.Errorf("Nil volume provided for policy enforcement")
	}

	p.volume = volume

	// init policies
	p.initVolType()
	p.initOrchProvider()

	// enforce policies
	p.enforce()

	// run through validations
	err := p.validate()
	if err != nil {
		return nil, err
	}

	return p.volume, nil
}

// initVolType initializes the volume type from labels
// ENVs or defaults
func (p *ReqdPolicies) initVolType() {
	// set volType from volume property which should
	// prevail over other values
	p.volType = p.volume.VolumeType

	// possible values for volume type
	volTypeVals := []v1.VolumeType{
		p.volume.Labels.VolumeType,
		v1.VolumeTypeENV(),
		v1.DefaultVolumeType,
	}

	// Ensure non-empty value is set
	for _, tval := range volTypeVals {
		if len(p.volType) == 0 {
			p.volType = tval
		}
	}
}

// initOrchProvider initializes the orchestrator from labels
// ENVs or defaults
func (p *ReqdPolicies) initOrchProvider() {
	// set orch from volume property which should
	// prevail over other values
	p.orch = p.volume.OrchProvider

	// possible values for orchestrator
	orchVals := []v1.OrchProvider{
		v1.OrchProviderENV(),
		v1.DefaultOrchProvider,
	}

	// Ensure non-empty value is set
	for _, oval := range orchVals {
		if len(p.orch) == 0 {
			p.orch = oval
		}
	}
}

// enforce essential policies against the volume
func (p *ReqdPolicies) enforce() {
	// Enforce volume type
	p.volume.VolumeType = p.volType
	// Enforce volume's orchestration provider
	p.volume.OrchProvider = p.orch
}

// validate verifies the required volume policies
func (p *ReqdPolicies) validate() error {
	if !v1.IsVolumeType(p.volume.VolumeType) {
		return fmt.Errorf("Invalid volume type '%s'", p.volume.VolumeType)
	}

	if !v1.IsOrchProvider(p.volume.OrchProvider) {
		return fmt.Errorf("Invalid volume orchestrator '%s'", p.volume.OrchProvider)
	}

	return nil
}
