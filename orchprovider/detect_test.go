package orchprovider

import (
	"os"
	"testing"
)

func TestDetectOrchProviderFromEnv(t *testing.T) {
	tt := []struct {
		description      string
		env              string
		expectedProvider string
	}{
		{
			description:      "kubernetes variable",
			env:              "KUBERNETES_SERVICE_HOST",
			expectedProvider: "KUBERNETES",
		},
		{
			description:      "nomad addr variable",
			env:              "NOMAD_ADDR",
			expectedProvider: "NOMAD",
		},
		{
			description:      "nomad addr as suffix",
			env:              "MY_CUSTOM_NOMAD_ADDR",
			expectedProvider: "NOMAD",
		},
		{
			description:      "nomad addr as prefix",
			env:              "NOMAD_ADDR_FOR_MY_SERVICE",
			expectedProvider: "NOMAD",
		},
		{
			description:      "nomad addr as part of another environment variable",
			env:              "CUSTOM_NOMAD_ADDR_VAR",
			expectedProvider: "NOMAD",
		},
		{
			description:      "unknown provider",
			expectedProvider: "Unknown",
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			if test.env != "" {
				os.Setenv(test.env, "")
				defer os.Unsetenv(test.env)
			}

			provider := DetectOrchProviderFromEnv()

			if provider != test.expectedProvider {
				t.Fatalf("expected %s provider; got %s", test.expectedProvider, provider)
			}
		})
	}
}
