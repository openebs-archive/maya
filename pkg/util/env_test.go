package util

import (
	"testing"
	"os"
	"github.com/magiconair/properties/assert"
)

func TestLookEnv(t *testing.T) {
	// set required environment
	os.Setenv("_MY_PRESENT_TEST_KEY_", "value1")
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
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, lookEnv(v.key), v.expectValue)
		})
	}
}

func TestCASTemplateFeatureGate(t *testing.T) {
	// negative testcase, incorrect value
	os.Setenv(string(CASTemplateFeatureGateENVK), "on")
	_, err := CASTemplateFeatureGate()
	if err == nil {
		t.Errorf("expected error when feature gate %s is %s", CASTemplateFeatureGateENVK, "on")
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
