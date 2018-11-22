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

package debug

import (
	"os"
	"reflect"
	"testing"

	env "github.com/openebs/maya/pkg/env/v1alpha1"
)

func TestEnableCPUProfiling(t *testing.T) {
	testCases := map[string]struct {
		envValue    env.ENVValue
		expectValue bool
	}{
		"Missing debug env variable": {
			envValue:    "",
			expectValue: false,
		},
		"Present debug env with cpu": {
			envValue:    "cpu",
			expectValue: true,
		},
		"Present debug env with cpu as one value": {
			envValue:    "cpu,memory",
			expectValue: true,
		},
		"Present debug env without cpu": {
			envValue:    "memory",
			expectValue: false,
		},
	}

	for k, v := range testCases {
		t.Run(k, func(t *testing.T) {
			os.Setenv(string(env.DebugProfileENVK), string(v.envValue))
			defer os.Unsetenv(string(env.DebugProfileENVK))
			actualValue := EnableCPUProfiling()
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %v got %v", v.expectValue, actualValue)
			}
		})
	}
}

func TestGetProfilePath(t *testing.T) {
	testCases := map[string]struct {
		envValue    env.ENVValue
		expectValue string
	}{
		"Missing debug path variable": {
			envValue:    "",
			expectValue: "/tmp",
		},
		"Present debug path variable": {
			envValue:    "/test",
			expectValue: "/test",
		},
	}

	for k, v := range testCases {
		t.Run(k, func(t *testing.T) {
			os.Setenv(string(env.DebugProfilePathENVK), string(v.envValue))
			defer os.Unsetenv(string(env.DebugProfilePathENVK))
			actualValue := GetProfilePath()
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %v got %v", v.expectValue, actualValue)
			}
		})
	}
}
