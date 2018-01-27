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

	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
)

// Deployment provides utility methods over K8s
// Deployment object
type Deployment struct {
	// YmlInBytes represents a K8s Deployment in
	// yaml format
	YmlInBytes []byte
}

func NewDeployment(b []byte) *Deployment {
	return &Deployment{
		YmlInBytes: b,
	}
}

// AsExtnV1B1Deployment returns a extensions/v1beta1 Deployment instance
func (m *Deployment) AsExtnV1B1Deployment() (*api_extn_v1beta1.Deployment, error) {
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

// Service provides utility methods over K8s Service
type Service struct {
	// YmlInBytes represents a K8s Service in
	// yaml format
	YmlInBytes []byte
}

func NewService(b []byte) *Service {
	return &Service{
		YmlInBytes: b,
	}
}

// AsCoreV1Service returns a v1 Service instance
func (m *Service) AsCoreV1Service() (*api_core_v1.Service, error) {
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
