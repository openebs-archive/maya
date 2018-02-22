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

	"github.com/openebs/maya/orchprovider"
	k8s_v1 "github.com/openebs/maya/orchprovider/k8s/v1"
	oe_api_v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/types/v1"
)

// K8sPolicyOps provides volume policy related operations. These policies
// are extracted from K8s Kind objects.
type K8sPolicyOps struct {
	// orch represents the K8s orchestrator
	orch orchprovider.OrchestratorInterface
}

func NewK8sPolicyOps() (*K8sPolicyOps, error) {
	orch, err := k8s_v1.NewK8sOrchProvider()
	if err != nil {
		return nil, err
	}

	return &K8sPolicyOps{
		orch: orch,
	}, nil
}

// SCPolicies will merge k8s sc based policies against
// the volume instance
func (p *K8sPolicyOps) SCPolicies(volume *v1.Volume) (map[string]string, error) {
	if volume == nil {
		return nil, fmt.Errorf("Nil volume provided")
	}

	// This policy will be executed only if this is Jiva volume using K8s as
	// its volume orchestrator
	if volume.OrchProvider != v1.K8sOrchProvider || volume.VolumeType != v1.JivaVolumeType {
		// exit without error
		return nil, nil
	}

	// nothing to do if fetching via storageclass
	// is disabled
	if !volume.Labels.K8sStorageClassEnabled {
		return nil, nil
	}

	// check if orchestrator is available for operations
	// w.r.t K8s StorageClass
	if p.orch == nil {
		return nil, fmt.Errorf("Nil k8s orchestrator")
	}

	// fetch K8s SC based policies
	pOrch, supported, err := p.orch.PolicyOps(volume)
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, fmt.Errorf("K8s based policy operations is not supported")
	}

	// Fetch policies based on storage class name
	//
	// NOTE:
	//  StorageClass name would have set previously by
	// K8sPolicies against this volume
	policies, err := pOrch.SCPolicies()
	if err != nil {
		return nil, err
	}

	return policies, nil
}

// SPPolicies will merge k8s StoragePool based policies against
// the volume instance
//
// NOTE:
//  StoragePool is a K8s CRD extension implemented by openebs
func (p *K8sPolicyOps) SPPolicies(volume *v1.Volume) (oe_api_v1alpha1.StoragePoolSpec, error) {
	if volume == nil {
		return oe_api_v1alpha1.StoragePoolSpec{}, fmt.Errorf("Nil volume provided")
	}

	// This policy will be executed only if this is Jiva volume using K8s as
	// its volume orchestrator
	if volume.OrchProvider != v1.K8sOrchProvider || volume.VolumeType != v1.JivaVolumeType {
		// exit without error
		return oe_api_v1alpha1.StoragePoolSpec{}, nil
	}

	if len(volume.StoragePool) == 0 {
		return oe_api_v1alpha1.StoragePoolSpec{}, fmt.Errorf("Nil storage pool name")
	}

	// check if orchestrator is available for operations
	// w.r.t K8s StoragePool
	if p.orch == nil {
		return oe_api_v1alpha1.StoragePoolSpec{}, fmt.Errorf("Nil k8s orchestrator")
	}

	// fetch K8s StoragePool based policies
	pOrch, supported, err := p.orch.PolicyOps(volume)
	if err != nil {
		return oe_api_v1alpha1.StoragePoolSpec{}, err
	}

	if !supported {
		return oe_api_v1alpha1.StoragePoolSpec{}, fmt.Errorf("K8s based policy operations is not supported")
	}

	// Fetch policies based on storage pool name
	//
	// NOTE:
	//  StoragePool name would have set previously against this volume
	return pOrch.SPPolicies()
}
