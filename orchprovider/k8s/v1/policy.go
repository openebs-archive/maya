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
	k8sClientApiV1Beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// DeployPolicy instance will enable transforming a
// volume specification to its K8s Deployment equivalent.
type DeployPolicy struct {
	// volName is the name of the volume whose specification
	// will be transformed. It is used as a contextual information.
	volName string

	// volSpec represents an OpenEBS volume structure that gets
	// transformed into a K8s Deployment
	volSpec v1.VolumeSpec

	// deploy represents the transformed K8s Deployment from the
	// volume specification
	deploy *k8sClientApiV1Beta1.Deployment
}

// NewDeployPolicy will create a new instance of DeployPolicy
//
// volSpec is the volume specification that gets transformed
//
// name of the volume which will be transformed
func NewDeployPolicy(volSpec v1.VolumeSpec, name string) (*DeployPolicy, error) {

	if name == "" {
		return nil, fmt.Errorf("Volume name is required to create a deploy policy")
	}

	if string(volSpec.Context) == "" {
		return nil, fmt.Errorf("Volume context is required to create a deploy policy")
	}

	// initialize the deployment
	deploy := &k8sClientApiV1Beta1.Deployment{}
	deploy.Name = name + string(volSpec.Context)

	return &DeployPolicy{
		volName: name,
		volSpec: volSpec,
		deploy:  deploy,
	}, nil
}

// Transform converts the volume specification object to
// its equivalent K8s Deploy object.
func (p *DeployPolicy) Transform() (*k8sClientApiV1Beta1.Deployment, error) {
	// fetch transformers
	// iterate & transform
	return nil, nil
}
