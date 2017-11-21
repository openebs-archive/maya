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

// TODO
// Should this be changed to policy_ov.go ?
// Should the name of the struct be renamed accordingly ?
// ov expands to openebs volume
//
// Why ?
//    Jiva & CStor seems to have same properties even
// though their behavior will vary !!
// It will be ideal to reuse these files for both the
// volume types. Hence, this name openebs volume !!
package v1

import (
	"fmt"

	"github.com/openebs/maya/types/v1"
)

// JivaPolicies will enforce policies on an OpenEBS volume.
//
// TIP:
//  Read this as Enforce Policies w.r.t Jiva
type JivaPolicies struct {
	// volume is the instance on which policies will be enforced
	volume *v1.Volume

	// isNoSpecs flags if no volume spec is available
	isNoSpecs bool

	// isReplicaSpecAvail flags if replica volume spec
	// is available
	isReplicaSpecAvail bool

	// isControllerSpecAvail flags if controller volume
	// spec is available
	isControllerSpecAvail bool
}

// Enforce will enforce jiva based policies against
// the volume instance
func (p *JivaPolicies) Enforce(volume *v1.Volume) (*v1.Volume, error) {
	if volume == nil {
		return nil, fmt.Errorf("Nil volume provided for policy enforcement")
	}

	// This policy will be executed only if Jiva is the volume type
	if volume.VolumeType != v1.JivaVolumeType {
		// exit without error
		return volume, nil
	}

	// set it locally to be used in further operations
	p.volume = volume

	// initialize as per jiva volume type
	p.init()

	// enforce policies
	p.enforce()

	err := p.validate()
	if err != nil {
		return nil, err
	}

	return p.volume, nil
}

// init intializes volume structure w.r.t jiva volume type
func (p *JivaPolicies) init() {
	if len(p.volume.Specs) == 0 {
		p.isNoSpecs = true
		return
	}

	for _, spec := range p.volume.Specs {
		if spec.Context == v1.ReplicaVolumeContext {
			p.isReplicaSpecAvail = true
		} else if spec.Context == v1.ControllerVolumeContext {
			p.isControllerSpecAvail = true
		}
	}

	if !p.isControllerSpecAvail && !p.isReplicaSpecAvail {
		p.isNoSpecs = true
	}

}

// enforce essential policies against the volume
func (p *JivaPolicies) enforce() {
	if p.isNoSpecs {
		p.volume.Specs = []v1.VolumeSpec{
			v1.VolumeSpec{
				Context: v1.ControllerVolumeContext,
			},
			v1.VolumeSpec{
				Context: v1.ReplicaVolumeContext,
			},
		}

		return
	}

	if !p.isReplicaSpecAvail {
		p.volume.Specs = append(p.volume.Specs, v1.VolumeSpec{
			Context: v1.ReplicaVolumeContext,
		})
	}

	if !p.isControllerSpecAvail {
		p.volume.Specs = append(p.volume.Specs, v1.VolumeSpec{
			Context: v1.ControllerVolumeContext,
		})
	}
}

// validate jiva policies
func (p *JivaPolicies) validate() error {
	if len(p.volume.Specs) > 2 {
		return fmt.Errorf("Invalid volume specifications were provided")
	}

	return nil
}
