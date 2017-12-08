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
	"strings"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/client/k8s"
	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// MayaAnyK8s is a wrapper over any kind of K8s object
type MayaAnyK8s struct {
	// Namespace represents the K8s namespace of this
	// K8s object
	//
	// NOTE:
	//  This will be the determining factor to execute
	// K8s APIs on its Kind(s) belonging to a
	// particular namespace
	Namespace string `json:"namespace,omitempty"`

	// Kind represents the kind of this K8s object
	Kind string `json:"kind"`

	// ApiVersion represents the api version of this
	// K8s object
	//
	// NOTE:
	//  This will be the determining factor to execute
	// K8s APIs on its Kind(s) belonging to a
	// particular version
	ApiVersion string `json:"apiVersion"`

	// MayaYaml represents this K8s object itself in yaml format
	MayaYaml `json:"yaml"`
}

func (m MayaAnyK8s) isDeployment() bool {
	return strings.ToLower(m.Kind) == "deployment"
}

func (m MayaAnyK8s) isService() bool {
	return strings.ToLower(m.Kind) == "service"
}

func (m MayaAnyK8s) GenerateDeployment() (*api_extn_v1beta1.Deployment, error) {
	if !m.isDeployment() {
		return nil, fmt.Errorf("Invalid operation")
	}

	d := NewMayaDeployment(m.Yaml)
	err := d.Load()
	if err != nil {
		return nil, err
	}

	return d.Deployment, nil
}

func (m MayaAnyK8s) GenerateService() (*api_core_v1.Service, error) {
	if !m.isService() {
		return nil, fmt.Errorf("Invalid operation")
	}

	s := NewMayaService(m.Yaml)
	err := s.Load()
	if err != nil {
		return nil, err
	}

	return s.Service, nil
}

type MayaConfigMap struct {
	// ConfigMap represents a K8s ConfigMap object
	// Maya expects ConfigMap to embed any K8s
	// object or K8s CR.
	ConfigMap *api_core_v1.ConfigMap

	// MayaYaml represents the above K8s ConfigMap in yaml format
	MayaYaml

	// EK8sObject represents the embedded K8s object or K8s CR
	EK8sObject *MayaAnyK8s

	// K8sClient represents the client to invoke K8s API
	// Note: This manner of composing is helpful during unit
	// testing
	K8sClient *k8s.K8sClient
}

// NewMayaConfigMap returns an instance of MayaConfigMap
// based on the provided yaml
func NewMayaConfigMap(yaml string) *MayaConfigMap {
	return &MayaConfigMap{
		MayaYaml: MayaYaml{
			Yaml: yaml,
		},
	}
}

// FetchMayaConfigMap returns an instance of MayaConfigMap
// based on the provided name of the ConfigMap and K8s namespace
func FetchMayaConfigMap(name string, ns string) (*MayaConfigMap, error) {
	kc, err := k8s.NewK8sClient(ns)
	if err != nil {
		return nil, err
	}

	cm, err := kc.GetConfigMap(name, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &MayaConfigMap{
		K8sClient: kc,
		ConfigMap: cm,
	}, nil
}

// Load initializes ConfigMap property of this instance
//
// NOTE:
//  This is alternative to using FetchMayaConfigMap call
func (m *MayaConfigMap) Load() error {
	// unmarshall the yaml
	b, err := m.Bytes()
	if err != nil {
		return err
	}

	cm := &api_core_v1.ConfigMap{}
	err = yaml.Unmarshal(b, cm)
	if err != nil {
		return err
	}

	// load the object
	m.ConfigMap = cm

	return nil
}

// LoadEmbeddedK8s initializes the embedded K8s object
// of this instance
func (m *MayaConfigMap) LoadEmbeddedK8s() error {
	if m.ConfigMap == nil {
		err := m.Load()
		if err != nil {
			return err
		}
	}

	// set the embedded K8s object details
	m.EK8sObject = &MayaAnyK8s{
		Namespace:  m.ConfigMap.Data["namespace"],
		Kind:       m.ConfigMap.Data["kind"],
		ApiVersion: m.ConfigMap.Data["apiVerson"],
		MayaYaml: MayaYaml{
			Yaml: m.ConfigMap.Data["yaml"],
		},
	}

	return nil
}

type MayaService struct {
	// Service represents a K8s Service object
	Service *api_core_v1.Service

	// MayaYaml represents the above K8s Service in yaml format
	MayaYaml
}

func NewMayaService(yaml string) *MayaService {
	return &MayaService{
		MayaYaml: MayaYaml{
			Yaml: yaml,
		},
	}
}

// Load initializes Service property of this instance
func (m *MayaService) Load() error {
	// unmarshall the yaml
	b, err := m.Bytes()
	if err != nil {
		return err
	}

	s := &api_core_v1.Service{}
	err = yaml.Unmarshal(b, s)
	if err != nil {
		return err
	}

	// load the object
	m.Service = s

	return nil
}

type MayaContainer struct {
	// Container represents a K8s Container object
	Container api_core_v1.Container

	// MayaYaml represents the above K8s Container in yaml format
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

type MayaDeployment struct {
	// Deployment represents a K8s Deployment object
	Deployment *api_extn_v1beta1.Deployment

	// MayaYaml represents the above K8s Deployment in yaml format
	MayaYaml
}

func NewMayaDeployment(yaml string) *MayaDeployment {
	return &MayaDeployment{
		MayaYaml: MayaYaml{
			Yaml: yaml,
		},
	}
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

	mc := NewMayaContainer(yaml)
	err := mc.Load()
	if err != nil {
		return err
	}

	cons := append(m.Deployment.Spec.Template.Spec.Containers, mc.Container)
	m.Deployment.Spec.Template.Spec.Containers = cons

	return nil
}
