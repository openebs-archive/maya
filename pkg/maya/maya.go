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
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/client/k8s"
	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomFuncsHolder contains properties that are
// used to build custom text/template functions
type CustomFuncsHolder struct {
	// Inputs contains the k:v pairs required to
	// set the template's placeholders
	Inputs map[string]string `json:"inputs"`
	// Stores contains the k:v pairs out of resulting
	// actions on the template's embedded object
	Stores map[string]string
}

func customFuncVal(pairs map[string]string, context, key string) (string, error) {
	if len(pairs) == 0 {
		return "", fmt.Errorf("No %s found", context)
	}

	if len(key) == 0 {
		return "", fmt.Errorf("Missing %s key", context)
	}

	val := pairs[key]
	if len(val) == 0 {
		return "", fmt.Errorf("Nil value for %s key '%s'", context, key)
	}

	return val, nil
}

func (f *CustomFuncsHolder) inputVal(key string) (string, error) {
	return customFuncVal(f.Inputs, "inputs", key)
}

func (f *CustomFuncsHolder) storeVal(key string) (string, error) {
	return customFuncVal(f.Stores, "stores", key)
}

func (f *CustomFuncsHolder) setInputIfEmpty(key, value string) {
	if len(f.Inputs[key]) == 0 {
		f.Inputs[key] = value
	}
}

func (f *CustomFuncsHolder) setStore(key, value string) {
	f.Stores[key] = value
}

func (f *CustomFuncsHolder) setStoreIfEmpty(key, value string) {
	if len(f.Stores[key]) == 0 {
		f.Stores[key] = value
	}
}

func (f *CustomFuncsHolder) mergeInputsIfEmpty(inputs map[string]string) {
	if len(f.Inputs) == 0 {
		f.Inputs = inputs
		return
	}

	for k, v := range inputs {
		f.setInputIfEmpty(k, v)
	}
}

func (f *CustomFuncsHolder) mergeStoresIfEmpty(stores map[string]string) {
	if len(f.Stores) == 0 {
		f.Stores = stores
		return
	}

	for k, v := range stores {
		f.setStoreIfEmpty(k, v)
	}
}

func (f *CustomFuncsHolder) mergeStores(stores map[string]string) {
	if len(f.Stores) == 0 {
		f.Stores = stores
		return
	}

	for k, v := range stores {
		f.setStore(k, v)
	}
}

// MayaYamlV2 represents a yaml definition
//
// This yaml is expected to be marshalled into
// corresponding go struct
type MayaYamlV2 struct {
	// Yaml represents a templated yaml in string format
	Yaml string

	// YmlInBytes represents the templated yaml in
	// byte slice format
	YmlInBytes []byte
}

func (m *MayaYamlV2) asByteArr() ([]byte, error) {
	if m.YmlInBytes != nil {
		return m.YmlInBytes, nil
	}

	if len(m.Yaml) == 0 {
		return nil, fmt.Errorf("Yaml is not set")
	}

	return []byte(m.Yaml), nil
}

func (m *MayaYamlV2) getYaml() (string, error) {
	if len(m.Yaml) == 0 {
		return "", fmt.Errorf("Yaml is not set")
	}

	return m.Yaml, nil
}

// load sets the byte slice corresponding to the yaml
func (m *MayaYamlV2) load() error {
	if m.YmlInBytes == nil {
		// unmarshall the yaml
		b, err := m.asByteArr()
		if err != nil {
			return err
		}
		m.YmlInBytes = b
	}

	return nil
}

// MayaRunAction signifies the action to be taken
// against a MayaTemplate
type MayaRunAction string

const (
	// GetMRA flags a action as get. Typically used to fetch
	// an object from its name.
	GetMRA MayaRunAction = "get"

	// PutMRA flags a action as put. Typically used to put
	// an  object.
	PutMRA MayaRunAction = "put"
)

// TemplateInfo composes various properties that provides
// information about a MayaTemplate
type TemplateInfo struct {
	// Kind of the template contents
	Kind string `json:"kind"`
	// APIVersion of the template contents
	APIVersion string `json:"apiVersion"`
	// Namespace of the template contents
	Namespace string `json:"namespace"`
	// Action to be invoked on the template contents
	Action MayaRunAction `json:"action"`
	// CustomFuncsHolder exposes the functions
	// that are set as custom functions in text/template
	CustomFuncsHolder
	// MayaYamlV2 provides the templated yaml representation
	// of this instance
	MayaYamlV2
}

func NewTemplateInfo(yaml string, inputs map[string]string) TemplateInfo {
	return TemplateInfo{
		MayaYamlV2: MayaYamlV2{
			Yaml: yaml,
		},
		CustomFuncsHolder: CustomFuncsHolder{
			Inputs: inputs,
			Stores: map[string]string{},
		},
	}
}

// templateInfoAsByte returns a byte slice format of
// its yaml representation
func (m *TemplateInfo) templateInfoAsByte() ([]byte, error) {
	yml, err := m.getYaml()
	if err != nil {
		return nil, err
	}

	tpl := template.New("templateinfo")
	tpl.Funcs(template.FuncMap{
		"inputs": m.inputVal,
		"stores": m.storeVal,
	})

	tpl, err = tpl.Parse(yml)
	if err != nil {
		return nil, err
	}

	// this has implementation of io.Writer
	// that is required by the template
	var buf bytes.Buffer

	// execute the parsed yaml against this instance
	// & write the result into the buffer
	err = tpl.Execute(&buf, m)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// asTemplateInfo returns a TemplateInfo structure
// from its corresponding yaml representation
func (m *TemplateInfo) asTemplateInfo() (TemplateInfo, error) {
	t := TemplateInfo{}

	// unmarshall the yaml
	b, err := m.templateInfoAsByte()
	if err != nil {
		return t, err
	}

	err = yaml.Unmarshal(b, t)
	if err != nil {
		return t, err
	}

	return t, nil
}

func (m *TemplateInfo) isDeployment() bool {
	return m.Kind == string(k8s.DeploymentKK)
}

func (m *TemplateInfo) isService() bool {
	return m.Kind == string(k8s.ServiceKK)
}

func (m *TemplateInfo) isConfigMap() bool {
	return m.Kind == string(k8s.ConfigMapKK)
}

func (m *TemplateInfo) isGet() bool {
	return m.Action == GetMRA
}

func (m *TemplateInfo) isPut() bool {
	return m.Action == PutMRA
}

func (m *TemplateInfo) isExtnV1B1() bool {
	return m.APIVersion == string(k8s.ExtensionsV1Beta1KA)
}

func (m *TemplateInfo) isCoreV1() bool {
	return m.APIVersion == string(k8s.CoreV1KA)
}

func (m *TemplateInfo) isExtnV1B1Deploy() bool {
	return m.isExtnV1B1() && m.isDeployment()
}

func (m *TemplateInfo) isCoreV1Service() bool {
	return m.isCoreV1() && m.isService()
}

func (m *TemplateInfo) isPutExtnV1B1Deploy() bool {
	return m.isExtnV1B1Deploy() && m.isPut()
}

func (m *TemplateInfo) isPutCoreV1Service() bool {
	return m.isCoreV1Service() && m.isPut()
}

// MayaTemplate represents a MayaTemplate as seen in
// its YAML format
type MayaTemplate struct {
	// TemplateInfo provides the information about this
	// template
	TemplateInfo
	// CustomFuncsHolder exposes the functions
	// that are set as custom functions in text/template
	CustomFuncsHolder
	// MayaYamlV2 represents this template's embedded yaml
	MayaYamlV2 `json:"yaml"`
	// k8sClient is the client to invoke Kubernetes APIs
	k8sClient *k8s.K8sClient
}

func NewMayaTemplate(mayaTplYml, tplInfoYml string, inputs map[string]string) (*MayaTemplate, error) {

	t := NewTemplateInfo(tplInfoYml, inputs)
	t, err := t.asTemplateInfo()
	if err != nil {
		return nil, err
	}

	return &MayaTemplate{
		TemplateInfo: t,
		MayaYamlV2: MayaYamlV2{
			Yaml: mayaTplYml,
		},
		CustomFuncsHolder: CustomFuncsHolder{
			Inputs: inputs,
			Stores: map[string]string{},
		},
	}, nil
}

// mayaTemplateAsByte returns a byte slice format of its yaml
// representation
func (m *MayaTemplate) mayaTemplateAsByte() ([]byte, error) {
	yml, err := m.getYaml()
	if err != nil {
		return nil, err
	}

	tpl := template.New("mayatemplate")
	tpl.Funcs(template.FuncMap{
		"inputs": m.inputVal,
		"stores": m.storeVal,
	})

	tpl, err = tpl.Parse(yml)
	if err != nil {
		return nil, err
	}

	// this has implementation of io.Writer
	// that is required by the template
	var buf bytes.Buffer

	// execute the parsed yaml against this instance
	// & write the result into the buffer
	err = tpl.Execute(&buf, m)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Execute will execute the MayaTemplate depending on informations
// set in TemplateInfo
func (m *MayaTemplate) Execute(stores map[string]string) error {
	if m.k8sClient == nil {
		// Make use of the namespace that is set in TemplateInfo
		kc, err := k8s.NewK8sClient(m.Namespace)
		if err != nil {
			return err
		}
		m.k8sClient = kc
	}

	if m.isPutExtnV1B1Deploy() {
		d, err := m.asExtnV1B1Deploy(stores)
		if err != nil {
			return err
		}

		d, err = m.k8sClient.CreateExtnV1B1Deployment(d)
		if err != nil {
			return err
		}
	}

	if m.isPutCoreV1Service() {
		s, err := m.asCoreV1Svc(stores)
		if err != nil {
			return err
		}

		s, err = m.k8sClient.CreateCoreV1Service(s)
		if err != nil {
			return err
		}
		stores["service-ip"] = s.Spec.ClusterIP
	}

	return nil
}

// asExtnV1B1Deploy generates a K8s Deployment object
// out of the embedded yaml
func (m *MayaTemplate) asExtnV1B1Deploy(stores map[string]string) (*api_extn_v1beta1.Deployment, error) {
	if !m.isDeployment() {
		return nil, fmt.Errorf("Invalid kind: Deployment required")
	}

	if !m.isExtnV1B1() {
		return nil, fmt.Errorf("Invalid version: extensions/v1beta1 required")
	}

	m.mergeStores(stores)

	b, err := m.mayaTemplateAsByte()
	if err != nil {
		return nil, err
	}

	d := NewMayaDeploymentV2ByByte(b)
	return d.AsExtnV1B1Deployment()
}

// asCoreV1Svc generates a K8s Service object
// out of the embedded yaml
func (m *MayaTemplate) asCoreV1Svc(stores map[string]string) (*api_core_v1.Service, error) {
	if !m.isService() {
		return nil, fmt.Errorf("Invalid kind: Service required")
	}

	if !m.isCoreV1() {
		return nil, fmt.Errorf("Invalid version: v1 required")
	}

	m.mergeStores(stores)

	b, err := m.mayaTemplateAsByte()
	if err != nil {
		return nil, err
	}

	s := NewMayaServiceV2ByByte(b)
	return s.AsCoreV1Service()
}

// MayaDeploymentV2 provides utility methods over K8s
// Deployment object
type MayaDeploymentV2 struct {
	// MayaYamlV2 represents a K8s Deployment in
	// yaml format
	MayaYamlV2
}

func NewMayaDeploymentV2(yaml string) *MayaDeploymentV2 {
	return &MayaDeploymentV2{
		MayaYamlV2: MayaYamlV2{
			Yaml: yaml,
		},
	}
}

func NewMayaDeploymentV2ByByte(b []byte) *MayaDeploymentV2 {
	return &MayaDeploymentV2{
		MayaYamlV2: MayaYamlV2{
			YmlInBytes: b,
		},
	}
}

// AsExtnV1B1Deployment returns a extensions/v1beta1 Deployment instance
func (m *MayaDeploymentV2) AsExtnV1B1Deployment() (*api_extn_v1beta1.Deployment, error) {
	err := m.load()
	if err != nil {
		return nil, err
	}

	// unmarshall the byte into k8s Deployment object
	deploy := &api_extn_v1beta1.Deployment{}
	err = yaml.Unmarshal(m.YmlInBytes, deploy)
	if err != nil {
		return nil, err
	}

	return deploy, nil
}

// MayaServiceV2 provides utility methods over K8s Service
type MayaServiceV2 struct {
	// MayaYamlV2 represents a K8s Service in
	// yaml format
	MayaYamlV2
}

func NewMayaServiceV2(yaml string) *MayaServiceV2 {
	return &MayaServiceV2{
		MayaYamlV2: MayaYamlV2{
			Yaml: yaml,
		},
	}
}

func NewMayaServiceV2ByByte(b []byte) *MayaServiceV2 {
	return &MayaServiceV2{
		MayaYamlV2: MayaYamlV2{
			YmlInBytes: b,
		},
	}
}

// AsCoreV1Service returns a v1 Service instance
func (m *MayaServiceV2) AsCoreV1Service() (*api_core_v1.Service, error) {
	err := m.load()
	if err != nil {
		return nil, err
	}

	// unmarshall the byte into k8s Service object
	svc := &api_core_v1.Service{}
	err = yaml.Unmarshal(m.YmlInBytes, svc)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

// TODO
// Move these to pkg/maya/policy.go
//
// TemplateInfoFilter flags if the provided template
// suits the invoking function's requirement
//type TemplateInfoFilter func(TemplateInfo) bool

// OpenEBSKind represents various kinds of openebs objects
//type OpenEBSKind string

//const (
// MayaRunOEK represents a MayaRun object
//MayaRunOEK OpenEBSKind = "MayaRun"
// MayaTemplateOEK represents a MayaTemplate object
//MayaTemplateOEK OpenEBSKind = "MayaTemplate"
//)

// TemplateSelector helps in selecting or identifying
// required template
//type TemplateSelector struct {
// Kind of template that needs to be selected
//Kind OpenEBSKind
// Select is the boolean selection function
//Select TemplateInfoFilter
// Path is the required path found in the selected
// template. It represents any nested path location.
//Path ...string
//}

// TODO
// The structures & code written below will undergo changes.
// Some of these may be deprecated & removed in favour of the
// structures implemented above.

// MayaPlaceholders is a structure that is composed
// of various properties. These properties are set as
// placeholders in a template file.
type MayaPlaceholders struct {
	// Owner represents the name of the owner
	Owner string `json:"owner"`
}

// MayaYaml represents the yaml definition
// that is typically embedded in various other
// Maya types
//
// MayaYaml is expected to be marshalled into corresponding
// go structure.
type MayaYaml struct {
	// Yaml represents a yaml in string format. This string
	// formatted yaml acts as a template.
	Yaml string

	// MayaPlaceholders represents various placeholders
	// that might have been set in above Yaml property.
	MayaPlaceholders

	// YmlBytes represents above yaml in byte slice format
	//
	// NOTE:
	//  This is generated by using the templated yaml
	// (i.e. Yaml property) & placeholders (i.e.
	// MayaPlaceholders property)
	YmlBytes []byte
}

func (m *MayaYaml) Bytes() ([]byte, error) {
	if m.YmlBytes != nil {
		return m.YmlBytes, nil
	}

	if len(m.Yaml) == 0 {
		return nil, fmt.Errorf("Invalid instance")
	}

	return []byte(m.Yaml), nil
}

func (m *MayaYaml) GetYaml() (string, error) {
	if len(m.Yaml) == 0 {
		return "", fmt.Errorf("Yaml is not set")
	}

	return m.Yaml, nil
}

type MayaK8sAction string

const (
	// GetMKA flags a action as get. Typically used to fetch
	// a particular K8s object from its name.
	GetMKA MayaK8sAction = "get"

	// PutMKA flags a action as put. Typically used to put
	// a K8s object.
	PutMKA MayaK8sAction = "put"
)

func SetMayaK8sAction(action string) MayaK8sAction {
	return MayaK8sAction(strings.ToLower(action))
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

	// APIVersion represents the api version of this
	// K8s object
	//
	// NOTE:
	//  This will be the determining factor to execute
	// K8s APIs on its Kind(s) belonging to a
	// particular version
	APIVersion string `json:"apiVersion"`

	// Kind represents the kind of this K8s object
	Kind string `json:"kind"`

	// Action represents the operation to be undertaken.
	Action MayaK8sAction `json:"action"`

	// Name is typically used to fetch a K8s object having
	// this name. It is used along with other properties
	// of this instance.
	//
	// NOTE:
	//  Typically used during get action
	Name string `json:"name"`

	// MayaYaml represents this K8s object itself in yaml format
	//
	// NOTE:
	// Typically used during put action
	MayaYaml `json:"yaml"`
}

func (m MayaAnyK8s) isDeployment() bool {
	return m.Kind == "Deployment"
}

func (m MayaAnyK8s) isService() bool {
	return m.Kind == "Service"
}

func (m MayaAnyK8s) isConfigMap() bool {
	return m.Kind == "ConfigMap"
}

func (m MayaAnyK8s) isGet() bool {
	return m.Action == GetMKA
}

func (m MayaAnyK8s) isPut() bool {
	return m.Action == PutMKA
}

// ExecuteTemplate loads the placeholders' values against the
// templated yaml
func (m MayaAnyK8s) ExecuteTemplate() ([]byte, error) {
	// this has implementation of io.Writer
	// that is required by the template
	var buf bytes.Buffer

	yml, err := m.GetYaml()
	if err != nil {
		return nil, err
	}

	tpl, err := template.New("mayayaml").Parse(yml)
	if err != nil {
		return nil, err
	}

	// this applies its parsed yaml against this instance
	// & writes the result into the buffer
	err = tpl.Execute(&buf, m)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GenerateDeployment generates a K8s Deployment object
// out of the embedded yaml
func (m MayaAnyK8s) GenerateDeployment() (*api_extn_v1beta1.Deployment, error) {
	if !m.isDeployment() {
		return nil, fmt.Errorf("Invalid operation; Deployment kind is required")
	}

	b, err := m.ExecuteTemplate()
	if err != nil {
		return nil, err
	}

	d := NewMayaDeploymentByByte(b)
	err = d.Load()
	if err != nil {
		return nil, err
	}

	return d.Deployment, nil
}

// GenerateService generates a K8s Service object
// out of the embedded yaml
func (m MayaAnyK8s) GenerateService() (*api_core_v1.Service, error) {
	if !m.isService() {
		return nil, fmt.Errorf("Invalid operation; Service kind is required")
	}

	b, err := m.ExecuteTemplate()
	if err != nil {
		return nil, err
	}

	s := NewMayaServiceByByte(b)
	err = s.Load()
	if err != nil {
		return nil, err
	}

	return s.Service, nil
}

// GenerateConfigMap generates a K8s ConfigMap object
// out of the embedded yaml
func (m MayaAnyK8s) GenerateConfigMap() (*api_core_v1.ConfigMap, error) {
	if !m.isConfigMap() {
		return nil, fmt.Errorf("Invalid operation; ConfigMap kind is required")
	}

	b, err := m.ExecuteTemplate()
	if err != nil {
		return nil, err
	}

	cm := NewMayaConfigMapByByte(b)
	err = cm.Load()
	if err != nil {
		return nil, err
	}

	return cm.ConfigMap, nil
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

// NewMayaConfigMapByByte returns an instance of MayaConfigMap
// based on the provided yaml in byte slice format
func NewMayaConfigMapByByte(b []byte) *MayaConfigMap {
	return &MayaConfigMap{
		MayaYaml: MayaYaml{
			YmlBytes: b,
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
	if m.YmlBytes == nil {
		// unmarshall the yaml
		b, err := m.Bytes()
		if err != nil {
			return err
		}
		m.YmlBytes = b
	}

	cm := &api_core_v1.ConfigMap{}
	err := yaml.Unmarshal(m.YmlBytes, cm)
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
		APIVersion: m.ConfigMap.Data["apiVerson"],
		Name:       m.ConfigMap.Data["name"],
		Action:     SetMayaK8sAction(m.ConfigMap.Data["action"]),
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

// NewMayaService returns an instance of MayaService
// based on the provided yaml
func NewMayaService(yaml string) *MayaService {
	return &MayaService{
		MayaYaml: MayaYaml{
			Yaml: yaml,
		},
	}
}

// NewMayaServiceByByte returns an instance of MayaService
// based on the provided yaml in byte slice format
func NewMayaServiceByByte(b []byte) *MayaService {
	return &MayaService{
		MayaYaml: MayaYaml{
			YmlBytes: b,
		},
	}
}

// Load initializes Service property of this instance
func (m *MayaService) Load() error {
	if m.YmlBytes == nil {
		// unmarshall the yaml
		b, err := m.Bytes()
		if err != nil {
			return err
		}
		m.YmlBytes = b
	}

	s := &api_core_v1.Service{}
	err := yaml.Unmarshal(m.YmlBytes, s)
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
	// MayaYaml represents the K8s Deployment in
	// yaml format
	MayaYaml

	// Deployment represents a K8s Deployment object
	Deployment *api_extn_v1beta1.Deployment
}

func NewMayaDeployment(yaml string) *MayaDeployment {
	return &MayaDeployment{
		MayaYaml: MayaYaml{
			Yaml: yaml,
		},
	}
}

func NewMayaDeploymentByByte(b []byte) *MayaDeployment {
	return &MayaDeployment{
		MayaYaml: MayaYaml{
			YmlBytes: b,
		},
	}
}

// Load returns a extensions/v1beta1 Deployment instance
func (m *MayaDeployment) Load() error {
	if m.YmlBytes == nil {
		// unmarshall the yaml
		b, err := m.Bytes()
		if err != nil {
			return err
		}
		m.YmlBytes = b
	}

	// unmarshall the byte into k8s Deployment object
	deploy := &api_extn_v1beta1.Deployment{}
	err := yaml.Unmarshal(m.YmlBytes, deploy)
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
