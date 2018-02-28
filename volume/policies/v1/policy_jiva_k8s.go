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
// Should this be changed to policy_ov_props.go ?
// Should the name of the struct be renamed accordingly ?
// ov expands to openebs volume
//
// Why ?
//    Jiva & CStor seems to have same properties even
// though their behavior will vary !!
// It will be ideal to reuse these files for both the
// volume types. Hence, remove jiva from variable name, method
// name, logging, etc
package v1

import (
	"fmt"

	oe_api_v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
)

// JivaK8sPolicies will enforce policies on an OpenEBS jiva
// volume
//
// TIP:
//  Read this as Enforce Jiva based K8s SC Policies
type JivaK8sPolicies struct {
	// volume is the instance on which policies will be enforced
	volume *v1.Volume

	// k8sPolOps is the instance that can fetch volume policies from various
	// K8s Kinds
	k8sPolOps *K8sPolicyOps

	// scPolicies are the set of policies from K8s StorageClass
	scPolicies map[string]string

	// spSpec are a set of policies from K8s StoragePool
	spSpec oe_api_v1alpha1.StoragePoolSpec
}

func NewJivaK8sPolicies() (PolicyInterface, error) {
	k8sPolOps, err := NewK8sPolicyOps()
	if err != nil {
		return nil, err
	}

	return &JivaK8sPolicies{
		k8sPolOps: k8sPolOps,
	}, nil
}

// Enforce will enforce k8s sc based policies against
// the volume instance
func (p *JivaK8sPolicies) Enforce(volume *v1.Volume) (*v1.Volume, error) {
	if volume == nil {
		return nil, fmt.Errorf("Nil volume")
	}

	if p.k8sPolOps == nil {
		return nil, fmt.Errorf("Nil k8s policy operation instance")
	}

	// This policy will be executed only if this is Jiva volume using K8s as
	// its volume orchestrator
	if volume.OrchProvider != v1.K8sOrchProvider || volume.VolumeType != v1.JivaVolumeType {
		// exit without error
		return volume, nil
	}

	// fetch StorageClass policies
	scp, err := p.k8sPolOps.SCPolicies(volume)
	if err != nil {
		return nil, err
	}
	p.scPolicies = scp

	// set it locally to be used in further operations
	p.volume = volume

	err = p.enforce()
	if err != nil {
		return nil, err
	}

	return p.volume, nil
}

func (p *JivaK8sPolicies) getSPPolicies() error {
	// fetch StoragePool policies
	spc, err := p.k8sPolOps.SPPolicies(p.volume)
	if err != nil {
		return err
	}

	p.spSpec = spc
	return nil
}

func (p *JivaK8sPolicies) enforce() error {
	fns := p.getPropertyEnforcers()
	// enforce volume policies
	for _, fn := range fns {
		err := fn(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *JivaK8sPolicies) getPropertyEnforcers() []enforcePropertyFn {
	return []enforcePropertyFn{
		enforceCapacity,
		enforceStoragePool,
		enforceHostPath,
		enforceReplicaImage,
		enforceReplicaCount,
		enforceControllerImage,
		enforceControllerCount,
		enforceMonitoring,
	}
}

// enforcePropertyFn is a typed function that defines
// the signature for volume policy enforcement scoped
// to a volume property
type enforcePropertyFn func(*JivaK8sPolicies) error

// enforceCapacity enforces capacity volume
// policy
func enforceCapacity(p *JivaK8sPolicies) error {
	// merge from old label;
	// this comes from PVC through openebs provisioner
	if len(p.volume.Capacity) == 0 {
		p.volume.Capacity = p.volume.Labels.CapacityOld
	}

	// merge from sc policy
	if len(p.volume.Capacity) == 0 {
		p.volume.Capacity = p.scPolicies[string(v1.CapacityVK)]
	}

	// merge from env variable & then from default
	capVals := []string{
		v1.CapacityENV(),
		v1.DefaultCapacity,
	}

	// Ensure non-empty value is set
	for _, capVal := range capVals {
		if len(p.volume.Capacity) == 0 {
			p.volume.Capacity = capVal
		}
	}

	if len(p.volume.Capacity) == 0 {
		return fmt.Errorf("Nil volume capacity")
	}

	return nil
}

// enforceStoragePool enforces storage pool volume
// policy
func enforceStoragePool(p *JivaK8sPolicies) error {
	// merge from sc policy
	if len(p.volume.StoragePool) == 0 {
		p.volume.StoragePool = p.scPolicies[string(v1.StoragePoolVK)]
	}

	// merge from env variable & then from default
	spVals := []string{
		v1.StoragePoolENV(),
		v1.DefaultStoragePool,
	}

	// Ensure non-empty value is set
	for _, spVal := range spVals {
		if len(p.volume.StoragePool) == 0 {
			p.volume.StoragePool = spVal
		}
	}

	if len(p.volume.StoragePool) == 0 {
		return fmt.Errorf("Nil storage pool")
	}

	return nil
}

// enforceMonitoring enforces volume monitoring policy
func enforceMonitoring(p *JivaK8sPolicies) error {
	// merge from sc policy
	if len(p.volume.Monitor) == 0 {
		p.volume.Monitor = p.scPolicies[string(v1.MonitorVK)]
	}

	// merge from env variable & then from default
	mVals := []string{
		v1.MonitorENV(),
		v1.DefaultMonitor,
	}

	// Ensure non-empty value is set
	for _, mVal := range mVals {
		if len(p.volume.Monitor) == 0 {
			p.volume.Monitor = mVal
		}
	}

	if len(p.volume.Monitor) == 0 {
		return fmt.Errorf("Nil volume monitor")
	}

	return nil
}

// enforceHostPath enforces host path property
func enforceHostPath(p *JivaK8sPolicies) error {
	// nothing needs to be done
	if len(p.volume.HostPath) != 0 {
		return nil
	}

	// get storagepool
	err := p.getSPPolicies()
	if err != nil {
		return err
	}

	// path might still be blank if the storagepool
	// is not found in cluster
	p.volume.HostPath = p.spSpec.Path

	// err for specific storagepool, do not err for default storagepool
	if len(p.volume.HostPath) == 0 && p.volume.StoragePool != v1.DefaultStoragePool {
		return fmt.Errorf("StoragePool '%s' is not found", p.volume.StoragePool)
	}

	// merge if empty from env variable & then from default
	hps := []string{
		v1.HostPathENV(),
		v1.DefaultHostPath,
	}

	// Ensure non-empty value is set
	for _, hp := range hps {
		if len(p.volume.HostPath) == 0 {
			p.volume.HostPath = hp
		}
	}

	// Need to err at this place as all attempts to
	// fetch the host path has failed
	if len(p.volume.HostPath) == 0 {
		return fmt.Errorf("Nil host path")
	}

	return nil
}

// enforceReplicaImage enforces replica image volume
// policy
func enforceReplicaImage(p *JivaK8sPolicies) error {
	var rIndex int
	for i, spec := range p.volume.Specs {
		if spec.Context == v1.ReplicaVolumeContext {
			rIndex = i
			break
		}
	}

	// merge from sc policy
	if len(p.volume.Specs[rIndex].Image) == 0 {
		p.volume.Specs[rIndex].Image = p.scPolicies[string(v1.JivaReplicaImageVK)]
	}

	// merge from old replica image
	if len(p.volume.Specs[rIndex].Image) == 0 {
		p.volume.Specs[rIndex].Image = p.volume.Labels.ReplicaImageOld
	}

	// merge from env variable & then from default
	iVals := []string{
		v1.JivaReplicaImageENV(),
		v1.DefaultJivaReplicaImage,
	}

	// Ensure non-empty value is set
	for _, iVal := range iVals {
		if len(p.volume.Specs[rIndex].Image) == 0 {
			p.volume.Specs[rIndex].Image = iVal
		}
	}

	if len(p.volume.Specs[rIndex].Image) == 0 {
		return fmt.Errorf("Nil replica image")
	}

	return nil
}

// enforceReplicaCount enforces replica count volume
// policy
func enforceReplicaCount(p *JivaK8sPolicies) error {
	var rIndex int
	for i, spec := range p.volume.Specs {
		if spec.Context == v1.ReplicaVolumeContext {
			rIndex = i
			break
		}
	}

	// merge from sc policy
	if p.volume.Specs[rIndex].Replicas == nil {
		p.volume.Specs[rIndex].Replicas = util.StrToInt32(p.scPolicies[string(v1.JivaReplicasVK)])
	}

	// merge from old volume label
	if p.volume.Specs[rIndex].Replicas == nil {
		p.volume.Specs[rIndex].Replicas = p.volume.Labels.ReplicasOld
	}

	// merge from env variable & then from default
	rcVals := []*int32{
		v1.JivaReplicasENV(),
		v1.DefaultJivaReplicas,
	}

	// Ensure non-empty value is set
	for _, rcVal := range rcVals {
		if p.volume.Specs[rIndex].Replicas == nil {
			p.volume.Specs[rIndex].Replicas = rcVal
		}
	}

	if p.volume.Specs[rIndex].Replicas == nil {
		return fmt.Errorf("Nil or invalid replica count")
	}

	return nil
}

// enforceControllerImage enforces controller image volume
// policy
func enforceControllerImage(p *JivaK8sPolicies) error {
	var cIndex int
	for i, spec := range p.volume.Specs {
		if spec.Context == v1.ControllerVolumeContext {
			cIndex = i
			break
		}
	}

	// merge from sc policy
	if len(p.volume.Specs[cIndex].Image) == 0 {
		p.volume.Specs[cIndex].Image = p.scPolicies[string(v1.JivaControllerImageVK)]
	}

	// merge from old replica image
	if len(p.volume.Specs[cIndex].Image) == 0 {
		p.volume.Specs[cIndex].Image = p.volume.Labels.ControllerImageOld
	}

	// merge from env variable & then from default
	iVals := []string{
		v1.JivaControllerImageENV(),
		v1.DefaultJivaControllerImage,
	}

	for _, iVal := range iVals {
		if len(p.volume.Specs[cIndex].Image) == 0 {
			p.volume.Specs[cIndex].Image = iVal
		}
	}

	if len(p.volume.Specs[cIndex].Image) == 0 {
		return fmt.Errorf("Nil controller image")
	}

	return nil
}

// enforceControllerCount enforces controller count volume
// policy
func enforceControllerCount(p *JivaK8sPolicies) error {
	var cIndex int
	for i, spec := range p.volume.Specs {
		if spec.Context == v1.ControllerVolumeContext {
			cIndex = i
			break
		}
	}

	// merge from sc policy
	if p.volume.Specs[cIndex].Replicas == nil {
		p.volume.Specs[cIndex].Replicas = util.StrToInt32(p.scPolicies[string(v1.JivaControllersVK)])
	}

	// merge from old volume label
	if p.volume.Specs[cIndex].Replicas == nil {
		p.volume.Specs[cIndex].Replicas = p.volume.Labels.ControllersOld
	}

	// merge from env variable & then from default
	ccVals := []*int32{
		v1.JivaControllersENV(),
		v1.DefaultJivaControllers,
	}

	// Ensure non-empty value is set
	for _, ccVal := range ccVals {
		if p.volume.Specs[cIndex].Replicas == nil {
			p.volume.Specs[cIndex].Replicas = ccVal
		}
	}

	if p.volume.Specs[cIndex].Replicas == nil {
		return fmt.Errorf("Nil or invalid controller count")
	}

	return nil
}
