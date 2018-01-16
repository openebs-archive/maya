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
	"github.com/openebs/maya/pkg/client/k8s"

	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
)

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

// TemplateMeta composes various properties that provides
// information about a MayaTemplate
type TemplateMeta struct {
	// Kind of the template contents
	Kind string `json:"kind"`
	// APIVersion of the template contents
	APIVersion string `json:"apiVersion"`
	// Namespace of the template contents
	Namespace string `json:"namespace"`
	// Action to be invoked on the template contents
	Action MayaRunAction `json:"action"`
	// MayaYamlV2 provides the templated yaml representation
	// of this instance
	MayaYamlV2
}

func NewTemplateMeta(yaml string, inputs map[string]string) *TemplateMeta {
	return &TemplateMeta{
		MayaYamlV2: MayaYamlV2{
			Yaml: yaml,
			CustomFuncsHolder: CustomFuncsHolder{
				Inputs: inputs,
				Stores: map[string]string{},
			},
		},
	}
}

// asTemplateMeta returns a new instance of TemplateMeta
// that corresponds to this instance's yaml
func (m *TemplateMeta) asTemplateMeta() (*TemplateMeta, error) {
	// unmarshall the yaml
	b, err := m.asTemplatedBytes()
	if err != nil {
		return nil, err
	}

	t := &TemplateMeta{}
	err = yaml.Unmarshal(b, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (m *TemplateMeta) isDeployment() bool {
	return m.Kind == string(k8s.DeploymentKK)
}

func (m *TemplateMeta) isService() bool {
	return m.Kind == string(k8s.ServiceKK)
}

func (m *TemplateMeta) isConfigMap() bool {
	return m.Kind == string(k8s.ConfigMapKK)
}

func (m *TemplateMeta) isGet() bool {
	return m.Action == GetMRA
}

func (m *TemplateMeta) isPut() bool {
	return m.Action == PutMRA
}

func (m *TemplateMeta) isExtnV1B1() bool {
	return m.APIVersion == string(k8s.ExtensionsV1Beta1KA)
}

func (m *TemplateMeta) isCoreV1() bool {
	return m.APIVersion == string(k8s.CoreV1KA)
}

func (m *TemplateMeta) isExtnV1B1Deploy() bool {
	return m.isExtnV1B1() && m.isDeployment()
}

func (m *TemplateMeta) isCoreV1Service() bool {
	return m.isCoreV1() && m.isService()
}

func (m *TemplateMeta) isPutExtnV1B1Deploy() bool {
	return m.isExtnV1B1Deploy() && m.isPut()
}

func (m *TemplateMeta) isPutCoreV1Service() bool {
	return m.isCoreV1Service() && m.isPut()
}

// MayaTemplate represents a MayaTemplate as seen in
// its YAML format
type MayaTemplate struct {
	// TemplateMeta provides the information about this
	// template
	TemplateMeta
	// MayaYamlV2 represents this template's embedded yaml
	MayaYamlV2 `json:"yaml"`
	// k8sClient is the client to invoke Kubernetes APIs
	k8sClient *k8s.K8sClient
}

func NewMayaTemplate(mayaTplYml, tplInfoYml string, inputs map[string]string) (*MayaTemplate, error) {

	t := NewTemplateMeta(tplInfoYml, inputs)
	t, err := t.asTemplateMeta()
	if err != nil {
		return nil, err
	}

	return &MayaTemplate{
		TemplateMeta: *t,
		MayaYamlV2: MayaYamlV2{
			Yaml: mayaTplYml,
			CustomFuncsHolder: CustomFuncsHolder{
				Inputs: inputs,
				Stores: map[string]string{},
			},
		},
	}, nil
}

// asExtnV1B1Deploy generates a K8s Deployment object
// out of the embedded yaml
func (m *MayaTemplate) asExtnV1B1Deploy() (*api_extn_v1beta1.Deployment, error) {
	if !m.isDeployment() {
		return nil, fmt.Errorf("Invalid kind: Deployment required")
	}

	if !m.isExtnV1B1() {
		return nil, fmt.Errorf("Invalid version: extensions/v1beta1 required")
	}

	b, err := m.asTemplatedBytes()
	if err != nil {
		return nil, err
	}

	d := NewMayaDeploymentV2ByByte(b)
	return d.AsExtnV1B1Deployment()
}

// asCoreV1Svc generates a K8s Service object
// out of the embedded yaml
func (m *MayaTemplate) asCoreV1Svc() (*api_core_v1.Service, error) {
	if !m.isService() {
		return nil, fmt.Errorf("Invalid kind: Service required")
	}

	if !m.isCoreV1() {
		return nil, fmt.Errorf("Invalid version: v1 required")
	}

	b, err := m.asTemplatedBytes()
	if err != nil {
		return nil, err
	}

	s := NewMayaServiceV2ByByte(b)
	return s.AsCoreV1Service()
}

// getK8sClient returns the K8sClient based on the namespace
// set in this instance
func (m *MayaTemplate) getK8sClient() (*k8s.K8sClient, error) {
	// Make use of the namespace that is set in TemplateMeta
	return k8s.NewK8sClient(m.Namespace)
}

// putExtnV1B1Deploy will put a Deployment as defined in
// the MayaTemplate
func (m *MayaTemplate) putExtnV1B1Deploy() (bool, error) {
	d, err := m.asExtnV1B1Deploy()
	if err != nil {
		return false, err
	}

	kc, err := m.getK8sClient()
	if err != nil {
		return false, err
	}

	d, err = kc.CreateExtnV1B1Deployment(d)
	if err != nil {
		return false, err
	}

	return true, nil
}

// putCoreV1Service will put a Service as defined in
// the MayaTemplate
func (m *MayaTemplate) putCoreV1Service(storeIdentifier string) (bool, error) {
	s, err := m.asCoreV1Svc()
	if err != nil {
		return false, err
	}

	kc, err := m.getK8sClient()
	if err != nil {
		return false, err
	}

	s, err = kc.CreateCoreV1Service(s)
	if err != nil {
		return false, err
	}

	// Set the resulting service ip
	m.mergeStoresIfEmpty(map[string]string{
		storeIdentifier + "-service-ip": s.Spec.ClusterIP,
	})

	return true, nil
}

// Execute will execute the MayaTemplate depending on informations
// available in TemplateMeta
func (m *MayaTemplate) execute(storeIdentifier string) error {
	var isExecuted bool
	var err error
	if m.isPutExtnV1B1Deploy() {
		isExecuted, err = m.putExtnV1B1Deploy()
	} else if m.isPutCoreV1Service() {
		isExecuted, err = m.putCoreV1Service(storeIdentifier)
	}

	if err != nil {
		return err
	}

	if isExecuted {
		return nil
	} else {
		return fmt.Errorf("Not supported operation '%v'", m.TemplateMeta)
	}
}
