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
	Version    string        `json:"version"`
	Components ComponentList `json:"components"`
}

// ComponentList states the desired components of an
// openebs cluster
type ComponentList struct {
	Labels      map[string]string `json:"labels"`      // desired labels for each component in the list
	Annotations map[string]string `json:"annotations"` // desired annotations for each component in the list
	Namespace   string            `json:"namespace"`   // desired namespace for each component in the list
	Items       []Component       `json:"items"`       // desired components in an openebs cluster
}

// Component represents a desired component
// in an openebs cluster
type Component struct {
	Name        string            `json:"name"`                  // name of the component
	Enabled     bool              `json:"enabled"`               // flag to install if enabled or un-install if disabled
	Labels      map[string]string `json:"labels,omitempty"`      // desired labels for this component
	Annotations map[string]string `json:"annotations,omitempty"` // desired annotations for this component
	Namespace   string            `json:"namespace,omitempty"`   // desired namespace for this component
	Template    Template          `json:"template"`              // refers to a template that defines this component
}

// TemplateType represents the supported template kinds
// that can define an openebs component
type TemplateType int

const (
	// Catalog is a supported template that defines an
	// openebs component
	Catalog TemplateType = iota + 1
)

// Template contains specifications that defines an
// openebs component
type Template struct {
	Name       string       `json:"name"`       // name of the refered template
	Kind       TemplateType `json:"kind"`       // supported template kind
	APIVersion string       `json:"apiVersion"` // api version of this template
	Namespace  string       `json:"namespace"`  // namespace where this template is to be found
}

// OpenebsClusterStatus represents the current state of OpenebsCluster
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
