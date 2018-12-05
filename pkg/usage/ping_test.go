package usage

import (
	"os"
	"testing"
	"time"
)

func TestGetPingPeriod(t *testing.T) {
	beforeFunc := func(value string) {
		if err := os.Setenv(string(OpenEBSPingPeriod), value); err != nil {
			t.Logf("Unable to set environment variable")
		}
	}
	afterFunc := func() {
		if err := os.Unsetenv(string(OpenEBSPingPeriod)); err != nil {
			t.Logf("Unable to unset environment variable")
		}
	}
	testSuite := map[string]struct {
		OpenEBSPingPeriodValue string
		ExpectedPeriodValue    time.Duration
	}{
		"24 seconds":          {"24s", 86400000000000},
		"24 minutes":          {"24m", 86400000000000},
		"24 hours":            {"24h", 86400000000000},
		"Negative 24 hours":   {"-24h", 86400000000000},
		"Random string input": {"Apache", 86400000000000},
		"Two hours":           {"2h", 7200000000000},
		"Three hundred hours": {"300h", 1080000000000000},
		"Fifty two seconds":   {"52000000000ns", 86400000000000},
		"Empty env value":     {"", 86400000000000},
	}
	for testKey, testData := range testSuite {
		beforeFunc(testData.OpenEBSPingPeriodValue)
		evaluatedValue := getPingPeriod()
		if evaluatedValue != testData.ExpectedPeriodValue {
			t.Fatalf("Tests failed for %s, expected=%d, got=%d", testKey, testData.ExpectedPeriodValue, evaluatedValue)
		}
		afterFunc()
	}
}
