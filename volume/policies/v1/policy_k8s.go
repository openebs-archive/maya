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

// K8sPolicies will enforce K8s specific policies on an
// OpenEBS volume.
//
// TIP:
//  Read this as Enforce Policies w.r.t K8s
type K8sPolicies struct {
	// volume is the instance on which policies will be
	// enforced
	volume *v1.Volume

	// scEnabled flags if fetching policies from K8s StorageClass
	// is enabled
	scEnabled bool

	// sc is the K8s StorageClass that will be referred to
	// during volume operation
	sc string

	// ns is the K8s namespace where volume operation
	// will be executed
	ns string

	// outCluster is the K8s cluster information that is
	// different from current K8s cluster where this volume
	// operation is being triggered
	outCluster string
}

// Enforce will enforce k8s based policies against
// the volume instance
func (p *K8sPolicies) Enforce(volume *v1.Volume) (*v1.Volume, error) {
	if volume == nil {
		return nil, fmt.Errorf("Nil volume provided for policy enforcement")
	}

	// This policy will be executed only if K8s is the volume's
	// orchestration provider
	if volume.OrchProvider != v1.K8sOrchProvider {
		// exit without error
		return volume, nil
	}

	// set it locally to be used in further operations
	p.volume = volume

	// initialize as per k8s requirements
	p.initSC()
	p.initNS()
	p.initOutCluster()

	// enforce policies
	p.enforce()

	err := p.validate()
	if err != nil {
		return nil, err
	}

	return p.volume, nil
}

// initSC intializes the storage class
func (p *K8sPolicies) initSC() {
	// There is no volume specific property for
	// storage class. Hence, Labels' based property
	// will prevail over others.
	p.scEnabled = p.volume.Labels.K8sStorageClassEnabled
	p.sc = p.volume.Labels.K8sStorageClass

	// return if confirmed that fetching via sc is not enabled
	if len(p.sc) == 0 && !p.scEnabled {
		return
	}

	// otherwise enable fetching via sc
	p.scEnabled = true

	// possible values for storageclass
	scVals := []string{
		v1.K8sStorageClassENV(),
	}

	// Ensure non-empty value is set
	for _, scval := range scVals {
		if len(p.sc) == 0 {
			p.sc = scval
		}
	}
}

// initSC intializes the storage class
func (p *K8sPolicies) initNS() {
	// The volume property will prevail over others
	p.ns = p.volume.Namespace

	// possible values for namespace
	nsVals := []string{
		p.volume.ObjectMeta.Namespace,
		p.volume.Labels.K8sNamespace,
		v1.NamespaceENV(),
		v1.DefaultNamespace,
	}

	// Ensure non-empty value is set
	for _, nval := range nsVals {
		if len(p.ns) == 0 {
			p.ns = nval
		}
	}
}

// initSC intializes the storage class
func (p *K8sPolicies) initOutCluster() {
	// There is no volume specific property for
	// out cluster. Hence, Labels' based property
	// will prevail over others.
	p.outCluster = p.volume.Labels.K8sOutCluster

	// possible values for outcluster
	oVals := []string{
		v1.K8sOutClusterENV(),
	}

	// Ensure non-empty value is set
	for _, oval := range oVals {
		if len(p.outCluster) == 0 {
			p.outCluster = oval
		}
	}
}

// enforce K8s based policies against the volume
func (p *K8sPolicies) enforce() {
	// enforce k8s storage class enabled flag
	p.volume.Labels.K8sStorageClassEnabled = p.scEnabled
	// enforce k8s storage class
	p.volume.Labels.K8sStorageClass = p.sc
	// enforce volume's namespace
	p.volume.Namespace = p.ns
	// enforce k8s out cluster info
	p.volume.Labels.K8sOutCluster = p.outCluster
}

// validate verifies the K8s related volume policies
func (p *K8sPolicies) validate() error {
	if p.volume.Labels.K8sStorageClassEnabled && len(p.volume.Labels.K8sStorageClass) == 0 {
		return fmt.Errorf("K8s storage class cannot be empty")
	}

	if len(p.volume.Namespace) == 0 {
		return fmt.Errorf("Volume namespace cannot be empty")
	}

	return nil
}
