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

package task

import (
	"testing"
)

// TODO
func TestNewTaskIdentifier(t *testing.T) {}

// TODO
func TestIsPod(t *testing.T) {
	tests := map[string]struct {
		taskIdentifier taskIdentifier
		isPod          bool
	}{
		"True": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Pod",
				},
			},
			isPod: true},
		"False": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "NotPod",
				},
			},
			isPod: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			isPod := tt.taskIdentifier.isPod()
			if tt.isPod != isPod {
				t.Fatalf("isPod() => got %v, want %v", isPod, tt.isPod)
			}
		})
	}
}

// TODO
func TestIsDeployment(t *testing.T) {}

// TODO
func TestIsService(t *testing.T) {}

// TODO
func TestIsStoragePool(t *testing.T) {}

// TODO
func TestIsConfigMap(t *testing.T) {}

// TODO
func TestIsPVC(t *testing.T) {}

// TODO
func TestIsExtnV1Beta1(t *testing.T) {}

// TODO
func TestIsAppsV1Beta1(t *testing.T) {}

// TODO
func TestIsCoreV1(t *testing.T) {}

// TODO
func TestIsOEV1alpha1(t *testing.T) {}

// TODO
func TestIsExtnV1Beta1Deploy(t *testing.T) {}

// TODO
func TestIsAppsV1Beta1Deploy(t *testing.T) {}

// TODO
func TestIsCoreV1Pod(t *testing.T) {}

// TODO
func TestIsCoreV1Service(t *testing.T) {}

// TODO
func TestIsCoreV1PVC(t *testing.T) {}

// TODO
func TestIsOEV1alpha1SP(t *testing.T) {}
