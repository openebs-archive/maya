/*
Copyright 2018-2019 The OpenEBS Authors

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
	"testing"
)

var _ EnvLister = &envInstall{}

func TestEnvInstallCount(t *testing.T) {
	tests := map[string]struct {
		expectedCount int
	}{
		"101": {21},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			e := EnvInstall()
			l, _ := e.List()
			if len(l.Items) != mock.expectedCount {
				t.Fatalf("Test '%s' failed: expected env variables count '%d': actual '%d'", name, mock.expectedCount, len(l.Items))
			}
		})
	}
}
