/*
Copyright 2017 The OpenEBS Authors

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

package maya

import (
	"fmt"

	"github.com/ghodss/yaml"
	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
)

// MayaYaml represents the yaml definition
// that is typically embedded in various other
// Maya types
type MayaYaml struct {
	// Yaml represents the yaml in string format
	Yaml string
}

func (m *MayaYaml) Bytes() ([]byte, error) {
	if len(m.Yaml) == 0 {
		return nil, fmt.Errorf("Nil yaml provided")
	}

	return []byte(m.Yaml), nil
}

type MayaContainer struct {
	// Container represents the K8s Container object
	Container api_core_v1.Container

	// This represents the Container in yaml format
	MayaYaml
}

func NewMayaContainer(yaml string) *MayaContainer {
	return &MayaContainer{
		MayaYaml: MayaYaml{
			Yaml: yaml,
		},
	}
}

// Load initializes Container property of this instance
func (m *MayaContainer) Load() error {
	// unmarshall the yaml
	b, err := m.Bytes()
	if err != nil {
		return err
	}

	con := api_core_v1.Container{}
	err = yaml.Unmarshal(b, &con)
	if err != nil {
		return err
	}

	// load the object
	m.Container = con

	return nil
}

// Reload updates the Container property of this instance
func (m *MayaContainer) Reload(yaml string) error {
	// update the existing yaml
	m.Yaml = yaml
	return m.Load()
}

type MayaDeployment struct {
	// Deployment represents the K8s Deployment object
	Deployment *api_extn_v1beta1.Deployment

	// This represents the Deployment in yaml format
	MayaYaml

	// MayaContainer provides container related methods
	// Note: This manner of composing is helpful during unit
	// testing
	MayaContainer *MayaContainer
}

func NewMayaDeployment(yaml string) *MayaDeployment {
	return &MayaDeployment{
		MayaYaml: MayaYaml{
			Yaml: yaml,
		},
	}
}

// SetMayaContainer sets the MayaContainer property
// of this instance.
func (m *MayaDeployment) SetMayaContainer() *MayaDeployment {
	m.MayaContainer = &MayaContainer{}
	return m
}

// Load initializes Deployment property of this instance
func (m *MayaDeployment) Load() error {
	// unmarshall the yaml
	b, err := m.Bytes()
	if err != nil {
		return err
	}

	// unmarshall the buffer into k8s Deployment object
	deploy := &api_extn_v1beta1.Deployment{}
	err = yaml.Unmarshal(b, deploy)
	if err != nil {
		return err
	}

	// load the object
	m.Deployment = deploy

	return nil
}

// AddContainer adds a container object to this
// instance's Deployment object
func (m *MayaDeployment) AddContainer(yaml string) error {
	if m.Deployment == nil {
		return fmt.Errorf("Deployment is not loaded")
	}

	if m.MayaContainer == nil {
		return fmt.Errorf("Nil maya container")
	}

	err := m.MayaContainer.Reload(yaml)
	if err != nil {
		return err
	}

	cons := append(m.Deployment.Spec.Template.Spec.Containers, m.MayaContainer.Container)
	m.Deployment.Spec.Template.Spec.Containers = cons

	return nil
}
