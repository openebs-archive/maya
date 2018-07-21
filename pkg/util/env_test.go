package util

import (
	"testing"
	"os"
	"reflect"
	"strconv"
	"errors"
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
			actualValue := lookEnv(v.key)
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
		expectErr   error
	}{
		"Incorrect value on": {
			key:         string(CASTemplateFeatureGateENVK),
			value:       "on",
			expectValue: false,
			expectErr:   errors.New("invalid syntax"),
		},
		"Key and value nil": {
			key:         "",
			value:       "",
			expectValue: false,
			expectErr:   nil,
		},
		"Value is nil": {
			key:         string(CASTemplateFeatureGateENVK),
			value:       "",
			expectValue: false,
			expectErr:   errors.New("invalid syntax"),
		},
		"Valid key and value": {
			key:         string(CASTemplateFeatureGateENVK),
			value:       "true",
			expectValue: true,
			expectErr:   nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			os.Setenv(tc.key, tc.value)
			defer os.Unsetenv(tc.key)

			feature, err := CASTemplateFeatureGate()
			if tc.expectErr != nil {
				if !reflect.DeepEqual(tc.expectErr, err.(*strconv.NumError).Err) {
					t.Errorf("Expected %s, got %s", tc.expectErr, err)
				}
			}
			if !reflect.DeepEqual(feature, tc.expectValue) {
				t.Errorf("Expected %v, got %v", tc.expectValue, feature)
			}
		})
	}
}
