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
	"os"
	"reflect"
	"testing"
)

func TestLookEnv(t *testing.T) {
	testCases := map[string]struct {
		key         ENVKey
		value       string
		expectValue string
	}{
		"Missing env variable": {
			key:         "",
			value:       "",
			expectValue: "false",
		},
		"Present env variable with value": {
			key:         "_MY_PRESENT_TEST_KEY_",
			value:       "value1",
			expectValue: "value1",
		},
		"Present env variable with empty value": {
			key:         "_MY_PRESENT_TEST_KEY_W_EMPTY_VALUE",
			value:       "",
			expectValue: "",
		},
	}

	for k, v := range testCases {
		t.Run(k, func(t *testing.T) {
			os.Setenv(string(v.key), v.value)
			actualValue := LookupOrFalse(v.key)
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
			os.Unsetenv(string(v.key))
		})
	}
}

func TestCASTemplateFeatureGate(t *testing.T) {

	cases := map[string]struct {
		key, value  string
		expectValue bool
	}{
		"Incorrect value on": {
			key:         string(CASTemplateFeatureGateENVK),
			value:       "on",
			expectValue: false,
		},
		"Key and value nil": {
			key:         "",
			value:       "",
			expectValue: false,
		},
		"Value is nil": {
			key:         string(CASTemplateFeatureGateENVK),
			value:       "",
			expectValue: false,
		},
		"Valid key and value": {
			key:         string(CASTemplateFeatureGateENVK),
			value:       "true",
			expectValue: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			os.Setenv(tc.key, tc.value)
			defer os.Unsetenv(tc.key)

			feature := Truthy(CASTemplateFeatureGateENVK)
			if !reflect.DeepEqual(feature, tc.expectValue) {
				t.Errorf("Expected %v, got %v", tc.expectValue, feature)
			}
		})
	}
}

func TestgetEnv(t *testing.T) {
	testCases := map[string]struct {
		key         string
		value       string
		expectValue string
	}{
		"Missing env variable": {
			key:         "",
			value:       "",
			expectValue: "",
		},
		"Present env variable with value": {
			key:         "_MY_PRESENT_TEST_KEY_",
			value:       "value1",
			expectValue: "value1",
		},
		"Present env variable with empty value": {
			key:         "_MY_PRESENT_TEST_KEY_W_EMPTY_VALUE",
			value:       "",
			expectValue: "",
		},
	}
	for k, v := range testCases {
		t.Run(k, func(t *testing.T) {
			os.Setenv(v.key, v.value)
			actualValue := Get(ENVKey(v.key))
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
			os.Unsetenv(v.key)
		})
	}
}
