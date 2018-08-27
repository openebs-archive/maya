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

// InstallENVKey is a typed string to represent various environment keys
// used for install
type InstallENVKey string

const (
	// EnvKeyForInstallConfigName is the environment variable to get the
	// the install config's name
	EnvKeyForInstallConfigName InstallENVKey = "OPENEBS_IO_INSTALL_CONFIG_NAME"
	// CASDefaultCstorPoolENVK is the ENV key that specifies wether default cstor pool
	// should be configured or not
	// If value is "true", default cstor pool will be configured else for "false"
	// it will not be configured.
	CASDefaultCstorPool InstallENVKey = "OPENEBS_IO_INSTALL_DEFAULT_CSTOR_SPARSE_POOL"
)
