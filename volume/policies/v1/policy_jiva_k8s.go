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

	// orch represents the K8s orchestrator
	orch orchprovider.OrchestratorInterface

	// capacity will hold the volume capacity
	capacity string

	// repIndex is the index which holds replica volume spec
	repIndex int

	// repCount will hold the replica count
	repCount *int32

	// repImage will hold jiva replica image
	repImage string

	// ctrlIndex is the index which holds jiva controller spec
	ctrlIndex int

	// ctrlCount will hold the jiva controller count
	ctrlCount *int32

	// ctrlImage will hold jiva controller image
	ctrlImage string
}

func NewJivaK8sPolicies() (PolicyInterface, error) {
	orch, err := k8s_v1.NewK8sOrchProvider()
	if err != nil {
		return nil, err
	}

	return &JivaK8sPolicies{
		orch: orch,
	}, nil
}

// Enforce will enforce k8s sc based policies against
// the volume instance
func (p *JivaK8sPolicies) Enforce(volume *v1.Volume) (*v1.Volume, error) {
	if volume == nil {
		return nil, fmt.Errorf("Nil volume provided for policy enforcement")
	}

	// This policy will be executed only if this is Jiva volume using K8s as
	// its volume orchestrator
	if volume.OrchProvider != v1.K8sOrchProvider || volume.VolumeType != v1.JivaVolumeType {
		// exit without error
		return volume, nil
	}

	// set it locally to be used in further operations
	p.volume = volume

	// init the policies
	err := p.init()
	if err != nil {
		return nil, err
	}

	// enforce the policies against the volume properties
	p.enforce()

	// run through validations after enforcement
	err = p.validate()
	if err != nil {
		return nil, err
	}

	return p.volume, nil
}

// init initializes this instance properties from volume
// properties
//
// NOTE:
//    The original volume property values should prevail
// over others. Hence, this should be the first invocation
// in the Enforce() method.
func (p *JivaK8sPolicies) init() error {
	// direct volume properties prevail over other methods of setting
	p.capacity = p.volume.Capacity

	for i, spec := range p.volume.Specs {
		if spec.Context == v1.ReplicaVolumeContext {
			p.repCount = spec.Replicas
			p.repImage = spec.Image
			p.repIndex = i
		} else if spec.Context == v1.ControllerVolumeContext {
			p.ctrlCount = spec.Replicas
			p.ctrlImage = spec.Image
			p.ctrlIndex = i
		}
	}

	// init using SC policies
	// will merge the un-set properties of this instance
	err := p.initWithSC()
	if err != nil {
		return err
	}

	// init using old volume labels
	// will merge the un-set properties of this instance
	p.initWithOldLabels()

	// init using ENV variables or Defaults
	// will merge the un-set properties of this instance
	p.initWithENVsAndDefs()

	return nil
}

// initWithSC fetches the k8s sc policies
// & sets them against this instance's properties
func (p *JivaK8sPolicies) initWithSC() error {
	// nothing to do if fetching via storageclass
	// is disabled
	if !p.volume.Labels.K8sStorageClassEnabled {
		return nil
	}

	// check if orchestrator is available for operations
	// w.r.t K8s StorageClass
	if p.orch == nil {
		return fmt.Errorf("Nil k8s orchestrator")
	}

	// fetch K8s SC based policies
	pOrch, supported, err := p.orch.PolicyOps(p.volume)
	if err != nil {
		return err
	}

	if !supported {
		return fmt.Errorf("K8s based policy operations is not supported")
	}

	// Fetch policies based on storage class name
	//
	// NOTE:
	//  StorageClass name would have set previously by
	// K8sPolicies against this volume
	policies, err := pOrch.FetchPolicies()
	if err != nil {
		return err
	}

	// Marshall these policies against this instance's properties
	p.marshall(policies)

	return nil
}

// marshall extracts the K8s sc based policies to
// corresponding properties of this instance
func (p *JivaK8sPolicies) marshall(policies map[string]string) {
	for k, v := range policies {
		// volume capacity
		if k == string(v1.CapacityVK) && len(p.capacity) == 0 {
			p.capacity = v
		}
		// jiva replica count
		if k == string(v1.JivaReplicasVK) && p.repCount == nil {
			p.repCount = util.StrToInt32(v)
		}
		// jiva replica image
		if k == string(v1.JivaReplicaImageVK) && len(p.repImage) == 0 {
			p.repImage = v
		}
		// jiva controller count
		if k == string(v1.JivaControllersVK) && p.ctrlCount == nil {
			p.ctrlCount = util.StrToInt32(v)
		}
		// jiva controller image
		if k == string(v1.JivaControllerImageVK) && len(p.ctrlImage) == 0 {
			p.ctrlImage = v
		}
	}
}

// initWithOldLabels fetch the volume policies from
// volume's labels property & sets them against this instance's
// properties.
//
// NOTE:
//  This is to maintain backward compatibility
func (p *JivaK8sPolicies) initWithOldLabels() {
	// volume capacity
	if len(p.capacity) == 0 {
		p.capacity = p.volume.Labels.CapacityOld
	}
	// jiva replica count
	if p.repCount == nil {
		p.repCount = p.volume.Labels.ReplicasOld
	}
	// jiva replica image
	if len(p.repImage) == 0 {
		p.repImage = p.volume.Labels.ReplicaImageOld
	}
	// jiva controller count
	if p.ctrlCount == nil {
		p.ctrlCount = p.volume.Labels.ControllersOld
	}
	// jiva controller image
	if len(p.ctrlImage) == 0 {
		p.ctrlImage = p.volume.Labels.ControllerImageOld
	}
}

// initENVsAndDefs will initialize this instance properties
// using ENV variables or Defaults
func (p *JivaK8sPolicies) initWithENVsAndDefs() {
	// possible values for capacity
	capVals := []string{
		v1.CapacityENV(),
		v1.DefaultCapacity,
	}

	// Ensure non-empty value is set
	for _, capVal := range capVals {
		if len(p.capacity) == 0 {
			p.capacity = capVal
		}
	}

	// possible values for jiva replica count
	repCVals := []*int32{
		v1.JivaReplicasENV(),
		v1.DefaultJivaReplicas,
	}

	// Ensure non-empty value is set
	for _, repCVal := range repCVals {
		if p.repCount == nil {
			p.repCount = repCVal
		}
	}

	// possible values for jiva replica image
	repIVals := []string{
		v1.JivaReplicaImageENV(),
		v1.DefaultJivaReplicaImage,
	}

	// Ensure non-empty value is set
	for _, repIVal := range repIVals {
		if len(p.repImage) == 0 {
			p.repImage = repIVal
		}
	}

	// possible values for jiva controller count
	ctrlCVals := []*int32{
		v1.JivaControllersENV(),
		v1.DefaultJivaControllers,
	}

	// Ensure non-empty value is set
	for _, ctrlCVal := range ctrlCVals {
		if p.ctrlCount == nil {
			p.ctrlCount = ctrlCVal
		}
	}

	// possible values for jiva controller image
	ctrlIVals := []string{
		v1.JivaControllerImageENV(),
		v1.DefaultJivaControllerImage,
	}

	// Ensure non-empty value is set
	for _, ctrlIVal := range ctrlIVals {
		if len(p.ctrlImage) == 0 {
			p.ctrlImage = ctrlIVal
		}
	}
}

// enforce essential policies against the volume properties
// from this instance's properties
func (p *JivaK8sPolicies) enforce() {
	p.volume.Capacity = p.capacity

	p.volume.Specs[p.repIndex].Replicas = p.repCount
	p.volume.Specs[p.repIndex].Image = p.repImage

	p.volume.Specs[p.ctrlIndex].Replicas = p.ctrlCount
	p.volume.Specs[p.ctrlIndex].Image = p.ctrlImage
}

// validate verifies the volume properties that were
// just enforced
func (p *JivaK8sPolicies) validate() error {
	if len(p.volume.Capacity) == 0 {
		return fmt.Errorf("Nil volume capacity was provided")
	}

	if p.volume.Specs[p.repIndex].Replicas == nil {
		return fmt.Errorf("Nil or Invalid jiva replica count was provided")
	}

	if p.volume.Specs[p.ctrlIndex].Replicas == nil {
		return fmt.Errorf("Nil or Invalid jiva controller count was provided")
	}

	return nil
}
