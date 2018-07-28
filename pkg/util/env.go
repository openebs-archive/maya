/*
Copyright 2018 The OpenEBS Authors.

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

package util

import (
	"os"
	"strconv"
	"strings"
)

// ENVKey is a typed variable that holds all environment
// variables
type ENVKey string

const (
	// CASTemplateFeatureGateENVK is the ENV key to fetch cas template feature gate
	// i.e. if cas template based provisioning is enabled or disabled
	CASTemplateFeatureGateENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_FEATURE_GATE"

	// CASTemplateToListVolumeENVK is the ENV key that specifies the CAS Template
	// to list cas volumes
	CASTemplateToListVolumeENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_LIST_VOLUME"
)

// CASTemplateFeatureGate returns true if cas template feature gate is
// enabled
func CASTemplateToListVolume() string {
	return getEnv(CASTemplateToListVolumeENVK)
}

// CASTemplateFeatureGate returns true if cas template feature gate is
// enabled
func CASTemplateFeatureGate() (bool, error) {
	return strconv.ParseBool(lookEnv(CASTemplateFeatureGateENVK))
}

// getEnv fetches the environment variable value from the runtime's environment
func getEnv(envKey ENVKey) string {
	return strings.TrimSpace(os.Getenv(string(envKey)))
}

// lookEnv fetches the environment variable value from the runtime's environment
// if not present it returns "false", value otherwise
func lookEnv(envKey ENVKey) string {
	val, present := os.LookupEnv(string(envKey))
	if !present {
		return "false"
	}
	return strings.TrimSpace(val)
}
