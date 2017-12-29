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
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/client/k8s"
	api_core_v1 "k8s.io/api/core/v1"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Task refers to a MayaTemplate
type Task struct {
	TemplateName string `json:"template"`
}

// RunTemplate is a structure that represents a
// MayaRun template
type RunTemplate struct {
	// CustomFuncsHolder holds the inputs provided to this
	// template
	CustomFuncsHolder
	// Tasks are the references to MayaTemplates
	Tasks []Task `json:"tasks"`
	// MayaYamlV2 is the yaml representation of RunTemplate
	MayaYamlV2
}

// NewRunTemplate returns an instance of MayaConfigMapV2
// based on the provided yaml
func NewRunTemplate(yaml string) *RunTemplate {
	return &RunTemplate{
		MayaYamlV2: MayaYamlV2{
			Yaml: yaml,
		},
	}
}

// runTemplateAsByte returns a byte slice format of
// its yaml representation
func (m *RunTemplate) runTemplateAsByte() ([]byte, error) {
	yml, err := m.getYaml()
	if err != nil {
		return nil, err
	}

	tpl, err := template.New("runtemplate").Parse(yml)
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

// AsTemplateInfo returns a TemplateInfo structure
// from its corresponding yaml representation
func (m *RunTemplate) AsRunTemplate() (*RunTemplate, error) {
	// unmarshall the yaml
	b, err := m.runTemplateAsByte()
	if err != nil {
		return nil, err
	}

	t := &RunTemplate{}
	err = yaml.Unmarshal(b, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// MayaRunner does the low level operations of running a
// MayaRun template
type MayaRunner struct {
	// runTplName of the RunTemplate that is actually the name of
	// a K8s ConfigMap
	runTplName string
	// searchNamespace where the RunTemplate & MayaTemplate are
	// available
	searchNamespace string
	// CustomFuncsHolder is the placeholder to keep the
	// inputs as well as the stores
	CustomFuncsHolder
	// array of templates that are found in RunTemplate & are
	// meant to be run
	templates []*MayaTemplate
}

func NewMayaRunner(name, searchNamespace string) *MayaRunner {
	return &MayaRunner{
		runTplName:      name,
		searchNamespace: searchNamespace,
		CustomFuncsHolder: CustomFuncsHolder{
			Inputs: map[string]string{},
			Stores: map[string]string{},
		},
	}
}

// getKCToFetchTemplates gets the K8s Client that is pointing to the
// namespace where maya template(s) are available
func (m *MayaRunner) getKCToFetchTemplates() (*k8s.K8sClient, error) {
	return k8s.NewK8sClient(m.searchNamespace)
}

// fetchTemplate returns the corresponding ConfigMap
// NOTE:
//  A Template e.g. MayaTemplate or RunTemplate is actually
// a K8s ConfigMap
func (m *MayaRunner) fetchTemplate(tplName string) (*api_core_v1.ConfigMap, error) {
	kc, err := m.getKCToFetchTemplates()
	if err != nil {
		return nil, err
	}

	return kc.GetConfigMap(tplName, mach_apis_meta_v1.GetOptions{})
}

// getRunTemplate returns an instance of RunTemplate
// from ConfigMap
func (m *MayaRunner) getRunTemplate() (*RunTemplate, error) {
	cm, err := m.fetchTemplate(m.runTplName)
	if err != nil {
		return nil, err
	}

	return NewRunTemplate(cm.Data["yaml"]), nil
}

// getMayaTemplate returns an instance of MayaTemplate
// from ConfigMap
func (m *MayaRunner) getMayaTemplate(mayaTplName string, inputs map[string]string) (*MayaTemplate, error) {
	cm, err := m.fetchTemplate(mayaTplName)
	if err != nil {
		return nil, err
	}

	return NewMayaTemplate(cm.Data["yaml"], cm.Data["meta"], inputs)
}

// loadRunTemplate loads up the RunTemplate instance
func (m *MayaRunner) loadRunTemplate() (*RunTemplate, error) {
	// get the instance of RunTemplate
	rt, err := m.getRunTemplate()
	if err != nil {
		return nil, err
	}

	// load run template's yaml representation into corresponding
	// go struct
	return rt.AsRunTemplate()
}

// Run will run each MayaTemplate in the provided
// namespace
func (m *MayaRunner) Run(runNamespace string) error {
	rt, err := m.loadRunTemplate()
	if err != nil {
		return err
	}

	// merge the RunTemplate's inputs & stores with runner
	m.mergeInputsIfEmpty(rt.Inputs)
	m.mergeStoresIfEmpty(rt.Stores)
	// This is the namespace where RunTemplate will execute its
	// tasks i.e. MayaTemplates. However, this may not be the case
	// always if MayaTemplate hardcodes its namespace.
	m.setInputIfEmpty("namespace", runNamespace)

	// get an instance of MayaTemplate & store them
	// to be executed later
	for _, task := range rt.Tasks {
		mt, err := m.getMayaTemplate(task.TemplateName, m.Inputs)
		if err != nil {
			return err
		}
		m.templates = append(m.templates, mt)
	}

	for _, template := range m.templates {
		err := template.Execute(m.Stores)
		if err != nil {
			return err
		}
	}
	return nil
}
