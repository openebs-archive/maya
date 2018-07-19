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
			"",
			"",
			"false",
		},
		"Present env variable with value": {
			"_MY_PRESENT_TEST_KEY_",
			"value1",
			"value1",
		},
		"Present env variable with empty value": {
			"_MY_PRESENT_TEST_KEY_W_EMPTY_VALUE",
			"",
			"",
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
	testCases := map[string]struct {
		expectedError error
		key, value    string
		expectedValue bool
	}{
		"Incorrect value 'on'": {
			&strconv.NumError{Func: "ParseBool", Err: errors.New("invalid syntax"), Num: "on"},
			string(CASTemplateFeatureGateENVK),
			"on", false},
		"Incorrect value empty string": {
			&strconv.NumError{Func: "ParseBool", Err: errors.New("invalid syntax"), Num: ""},
			string(CASTemplateFeatureGateENVK),
			"", false},
		"Missing key": {
			expectedError: nil,
			value:         "",
			key:           "",
			expectedValue: false,
		},
		"Correct key and value": {
			expectedError: nil,
			value:         "true",
			key:           string(CASTemplateFeatureGateENVK),
			expectedValue: true,
		},
	}

	for _, v := range testCases {
		os.Setenv(v.key, v.value)
		feature, err := CASTemplateFeatureGate()
		if !reflect.DeepEqual(v.expectedError, err) {
			t.Errorf("expected %s got %s", v.expectedError, err)
		}
		if !reflect.DeepEqual(v.expectedValue, feature) {
			t.Errorf("expected %s got %t", v.expectedValue, feature)
		}

		os.Unsetenv(v.key)
	}
}
