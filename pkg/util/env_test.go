package util

import (
	"testing"
	"os"
	"reflect"
)

func TestLookEnv(t *testing.T) {
	// set required environment
	os.Setenv("_MY_PRESENT_TEST_KEY_", "value1")
	os.Setenv("_MY_PRESENT_TEST_KEY_W_EMPTY_VALUE", "")
	os.Unsetenv("_MY_MISSING_TEST_KEY_") // for safety

	cases := map[string]struct {
		key         ENVKey
		expectValue string
	}{
		"Missing env variable": {
			key:         "_MY_MISSING_TEST_KEY_",
			expectValue: "false",
		},
		"Present env variable": {
			key:         "_MY_PRESENT_TEST_KEY_",
			expectValue: "value1",
		},
		"Present env variable with empty value": {
			key:         "_MY_PRESENT_TEST_KEY_W_EMPTY_VALUE",
			expectValue: "",
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			actualValue := lookEnv(v.key)
			if !reflect.DeepEqual(actualValue, v.expectValue) {
				t.Errorf("expected %s got %s", v.expectValue, actualValue)
			}
		})
	}
}

func TestCASTemplateFeatureGate(t *testing.T) {
	// avoiding table driven parallel tests as tests depend
	// on hosts environment variable

	// negative testcase, incorrect value
	os.Setenv(string(CASTemplateFeatureGateENVK), "on")
	_, err := CASTemplateFeatureGate()
	if err == nil {
		t.Errorf("expected error when feature gate %s is %s", CASTemplateFeatureGateENVK, "on")
	}

	// negative testcase, empty value
	os.Setenv(string(CASTemplateFeatureGateENVK), "")
	_, err = CASTemplateFeatureGate()
	if err == nil {
		t.Errorf("expected error when feature gate %s is empty", CASTemplateFeatureGateENVK)
	}

	// Positive testcase, absent env variable
	os.Unsetenv(string(CASTemplateFeatureGateENVK))
	feature, err := CASTemplateFeatureGate()
	if err != nil {
		t.Errorf("expected no error when feature gate %s is unset", CASTemplateFeatureGateENVK)
	}
	if feature {
		t.Errorf("expected false when feature gate %s is unset", CASTemplateFeatureGateENVK)
	}

	// Positive testcase, env variable set to true
	os.Setenv(string(CASTemplateFeatureGateENVK), "true")
	feature, err = CASTemplateFeatureGate()
	if err != nil {
		t.Errorf("expected no error when feature gate %s is 'true'", CASTemplateFeatureGateENVK)
	}
	if !feature {
		t.Errorf("expected true when feature gate %s is 'true'", CASTemplateFeatureGateENVK)
	}
}
