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

// instead of pkg/maya/maya.go
package k8s

import (
	"fmt"
	"github.com/ghodss/yaml"

	"github.com/openebs/maya/pkg/template"
	api_apps_v1beta1 "k8s.io/api/apps/v1beta1"
	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
)

// DeploymentYml provides utility methods to generate K8s Deployment objects
type DeploymentYml struct {
	// YmlInBytes represents a K8s Deployment in
	// yaml format
	YmlInBytes []byte
}

func NewDeploymentYml(context, yml string, values map[string]interface{}) (*DeploymentYml, error) {
	b, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		return nil, err
	}

	return &DeploymentYml{
		YmlInBytes: b,
	}, nil
}

// AsExtnV1B1Deployment returns a extensions/v1beta1 Deployment instance
func (m *DeploymentYml) AsExtnV1B1Deployment() (*api_extn_v1beta1.Deployment, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("Missing yaml")
	}

	// unmarshall the byte into k8s Deployment object
	deploy := &api_extn_v1beta1.Deployment{}
	err := yaml.Unmarshal(m.YmlInBytes, deploy)
	if err != nil {
		return nil, err
	}

	return deploy, nil
}

// AsAppsV1B1Deployment returns a apps/v1 Deployment instance
func (m *DeploymentYml) AsAppsV1B1Deployment() (*api_apps_v1beta1.Deployment, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("Missing yaml")
	}

	// unmarshall the byte into k8s Deployment object
	deploy := &api_apps_v1beta1.Deployment{}
	err := yaml.Unmarshal(m.YmlInBytes, deploy)
	if err != nil {
		return nil, err
	}

	return deploy, nil
}

// Service provides utility methods to generate K8s Service objects
type ServiceYml struct {
	// YmlInBytes represents a K8s Service in
	// yaml format
	YmlInBytes []byte
}

func NewServiceYml(context, yml string, values map[string]interface{}) (*ServiceYml, error) {
	b, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		return nil, err
	}

	return &ServiceYml{
		YmlInBytes: b,
	}, nil
}

// AsCoreV1Service returns a v1 Service instance
func (m *ServiceYml) AsCoreV1Service() (*api_core_v1.Service, error) {
	if m.YmlInBytes == nil {
		return nil, fmt.Errorf("Missing yaml")
	}

	// unmarshall the byte into k8s Service object
	svc := &api_core_v1.Service{}
	err := yaml.Unmarshal(m.YmlInBytes, svc)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
