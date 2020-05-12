/*
Copyright 2019 The OpenEBS Authors

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
	"strconv"

	menv "github.com/openebs/maya/pkg/env/v1alpha1"
)

// IsDefaultStorageConfigEnabled reads from env variable to check
// whether default storage configuration should be created or not.
func IsDefaultStorageConfigEnabled() (enabled bool) {
	enabled, _ = strconv.ParseBool(menv.Get(CreateDefaultStorageConfig))
	return
}

// IsCstorSparsePoolEnabled reads from env variable to check
// whether cstor sparse pool should be created by default or not.
// In addition, cstor sparse pool should be created only if the
// creation of default storage configuration is enabled.
func IsCstorSparsePoolEnabled() bool {
	enabled, _ := strconv.ParseBool(menv.Get(DefaultCstorSparsePool))
	return IsDefaultStorageConfigEnabled() && enabled
}

// IsInstallCRDEnabled reads from env variable to check
// whether CRDs should be created by default or not.
func IsInstallCRDEnabled() (enabled bool) {
	enabled, _ = strconv.ParseBool(menv.Get(InstallCRD))
	return
}
