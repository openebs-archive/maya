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
	"testing"
)

// test if configFromENV implements ConfigGetter interface
var _ ConfigGetter = &configFromENV{}

// test if configFromREST implements ConfigGetter interface
var _ ConfigGetter = &configFromREST{}

// test if ConfigGetters implements ConfigGetter interface
var _ ConfigGetter = ConfigGetters{}

func TestConfigFromENV(t *testing.T) {
	tests := map[string]struct {
		masterip   string
		kubeconfig string
		iserr      bool
	}{
		"101": {"", "", true},
		"102": {"", "/etc/config/kubeconfig", true},
		"103": {"1.1.1.1", "", false},
		"104": {"1.1.1.1", "/etc/config/config", true},
	}

	// Sub tests is not used here as env key is set & unset to test. Since env
	// is a global setting, the tests should run serially
	for name, mock := range tests {
		masterip := os.Getenv(string(K8sMasterIPEnvironmentKey))
		defer os.Setenv(string(K8sMasterIPEnvironmentKey), masterip)

		kubeconfig := os.Getenv(string(KubeConfigEnvironmentKey))
		defer os.Setenv(string(KubeConfigEnvironmentKey), kubeconfig)

		err := os.Setenv(string(K8sMasterIPEnvironmentKey), mock.masterip)
		if err != nil {
			t.Fatalf("Test '%s' failed: %s", name, err)
		}
		err = os.Setenv(string(KubeConfigEnvironmentKey), mock.kubeconfig)
		if err != nil {
			t.Fatalf("Test '%s' failed: %s", name, err)
		}

		c := &configFromENV{}
		config, err := c.Get()

		if !mock.iserr && config == nil {
			t.Fatalf("Test '%s' failed: expected config: actual nil config", name)
		}
		if !mock.iserr && err != nil {
			t.Fatalf("Test '%s' failed: expected no error: actual %s", name, err)
		}
	}
}
