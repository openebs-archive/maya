package orchprovider

import (
	"os"
	"strings"
)

// Detect the Container Orchestrator based on ENV variables
func DetectOrchProviderFromEnv() string {

	_, ok := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	if ok {
		return "KUBERNETES"
	}
	for _, e := range os.Environ() {
		ok := strings.Contains(e, "NOMAD_ADDR")
		if ok {
			return "NOMAD"
		}
	}
	return "Unknown"
}
