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

// TODO
// Move to pkg/apis/openebs.io/v1alpha1

// TODO
// Rename InstallConfig to Install

// Option represent a install option that influences the installation
// workflow
type Option string

// InstallConfig is the config for installation workflow
type InstallConfig struct {
	Spec InstallConfigSpec `json:"spec"`
}

type InstallConfigSpec struct {
	// Options provides a ordered list of install related options
	Options []Option `json:"options"`
}
