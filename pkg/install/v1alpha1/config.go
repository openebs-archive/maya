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

// InstallConfig contains installation configuration specifications
type InstallConfig struct {
	Spec InstallConfigSpec `json:"spec"`
}

// InstallConfigSpec caters to installation and un-installation of
// artifacts understood by installer
type InstallConfigSpec struct {
	// Install caters to installation of artifacts that are understood by
	// installer
	//
	// NOTE:
	//  Only specific artifacts can be installed
	Install []Install `json:"install"`
	// Uninstall caters to un-installation of artifacts that are understood
	// by installer
	//
	// NOTE:
	//  Only specific resources can be un-installed
	Uninstall []Uninstall `json:"uninstall"`
}

// Install provides metadata information about one or more artifacts that
// need to be installed
type Install struct {
	// Version to be considered to install
	Version string `json:"version"`
	// SetOptions will override the defaults of this install version
	SetOptions SetOptions `json:"set"`
}

// SetOptions will override this install version resource(s) with these values
type SetOptions struct {
	// Namespace to be set against the artifacts before install
	Namespace string `json:"namespace"`
	// Labels to be set against the artifacts before install
	Labels map[string]string `json:"labels"`
	// Annotations to be set against the artifacts before install
	Annotations map[string]string `json:"annotations"`
}

// Uninstall provides metadata information about one or more artifacts that
// need to be un-installed
type Uninstall struct {
	// Version to be considered to un-install
	Version string `json:"version"`
	// FilterOptions will filter the resources to be un-installed
	FilterOptions FilterOptions `json:"filter"`
}

// FilterOptions is used to filter the resources based on the values specified
// in this structure
type FilterOptions struct {
	// Namespace to be considered
	Namespace string `json:"namespace"`
	// LabelSelector to be considered
	LabelSelector string `json:"labelSelector"`
}
