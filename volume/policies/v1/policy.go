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
	"github.com/openebs/maya/types/v1"
)

// PolicyInterface provides the contract that can be implemented
// by various volume policy enforcement implementations.
type PolicyInterface interface {
	// Enforce the policies on the volume instance
	Enforce(volume *v1.Volume) (*v1.Volume, error)
}

// Policy will enforce policies on an OpenEBS volume. This will be
// the typical instance that will be invoked by volume endpoints.
type Policy struct {
	// volume represents the instance against which policies
	// will be enforced
	volume *v1.Volume

	// policies are a set of policies that will be enforced on
	// volume
	policies []PolicyInterface
}

// VolumeAddPolicy provides a policy instance that enforces
// policies during volume provisioning
func VolumeAddPolicy() (*Policy, error) {
	// these are the set of policies that will
	// be enforced
	jkPolicies, err := NewJivaK8sPolicies()
	if err != nil {
		return nil, err
	}
	policies := []PolicyInterface{
		&ReqdPolicies{},
		&JivaPolicies{},
		&K8sPolicies{},
		jkPolicies,
	}

	return &Policy{
		policies: policies,
	}, nil
}

// VolumeDeletePolicy provides a policy instance that enforces
// policies during some of the volume operations other than
// provisioning
func VolumeGenericPolicy() (*Policy, error) {
	// these are the set of policies that will
	// be enforced
	policies := []PolicyInterface{
		&ReqdPolicies{},
		&K8sPolicies{},
	}

	return &Policy{
		policies: policies,
	}, nil
}

func (p *Policy) Enforce(volume *v1.Volume) (*v1.Volume, error) {
	p.volume = volume

	// iterate & enforce the policies
	//
	// NOTE:
	//  Error in any of these policies will break the chain
	for _, pol := range p.policies {
		vol, err := pol.Enforce(p.volume)
		if err != nil {
			return nil, err
		}
		p.volume = vol
	}

	return p.volume, nil
}
