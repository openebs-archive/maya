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
	"strings"
)

// ENVKey is a typed variable that holds all environment
// variables
type ENVKey string

const (
	// CASTemplateFeatureGateENVK is the ENV key to fetch cas template feature gate
	// i.e. if cas template based provisioning is enabled or disabled
	CASTemplateFeatureGateENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_FEATURE_GATE"
)

// CASTemplateFeatureGate returns true if cas template feature gate is
// enabled
func CASTemplateFeatureGate() bool {
	val := getEnv(CASTemplateFeatureGateENVK)
	return CheckTruthy(val)
}

// getEnv fetches the environment variable value from the runtime's environment
func getEnv(envKey ENVKey) string {
	return strings.TrimSpace(os.Getenv(string(envKey)))
}
