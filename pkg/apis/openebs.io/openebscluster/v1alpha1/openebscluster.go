/*
Copyright 2018 The OpenEBS Authors

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=openebscluster

// OpenebsCluster forms the desired specification of
// openebs components that should be present in openebs
// cluster
type OpenebsCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenebsClusterSpec   `json:"spec"`
	Status OpenebsClusterStatus `json:"status"`
}

// OpenebsClusterSpec is the specifications of an
// openebs cluster
type OpenebsClusterSpec struct {
	// Version implies the version of openebs
	// components that should be deployed
	// in the cluster
	Version string `json:"version"`

	// Components represents a list of openebs
	// components that should be deployed
	// in the cluster
	Components ComponentList `json:"components"`
}

// ComponentList states the desired components of an
// openebs cluster
type ComponentList struct {
	// Labels to be applied against each component
	Labels map[string]string `json:"labels"`

	// Annotations to be applied against each component
	Annotations map[string]string `json:"annotations"`

	// Namespace in which each component gets deployed
	Namespace string `json:"namespace"`

	// Items represents the actual components to
	// be deployed in the cluster
	Items []Component `json:"items"`
}

// Component represents a desired component
// in an openebs cluster
type Component struct {
	// Name of the component
	Name string `json:"name"`

	// Enabled is the flag that determines if
	// this component should be deployed or
	// un-deployed if it was deployed previously
	Enabled bool `json:"enabled"`

	// Labels to be applied against this component
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations to be applied against this component
	Annotations map[string]string `json:"annotations,omitempty"`

	// Namespace of this component
	Namespace string `json:"namespace,omitempty"`

	// Template that has the specifications of this
	// component
	Template Template `json:"template"`
}

// TemplateType represents the supported template kinds
// that can define an openebs component. In other words
// it has the component specifications
type TemplateType string

const (
	// Catalog is a supported template that defines an
	// openebs component
	Catalog TemplateType = "Catalog"
)

// Template contains specifications that defines an
// openebs component
type Template struct {
	// Name of the template
	Name string `json:"name"`

	// Kind of the template
	Kind TemplateType `json:"kind"`

	// APIVersion of the template
	APIVersion string `json:"apiVersion"`

	// Namespace where this template is to
	// be found
	Namespace string `json:"namespace"`
}

// OpenebsClusterStatus represents the current
// state of OpenebsCluster
type OpenebsClusterStatus struct {
	Phase string `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=openebsclusters

// OpenebsClusterList is a list of OpenebsCluster resources
type OpenebsClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []OpenebsCluster `json:"items"`
}
