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

func TestIsCommand(t *testing.T) {
	tests := map[string]struct {
		taskIdentifier taskIdentifier
		isCommand      bool
	}{
		"kindCommand#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Command",
				},
			},
			isCommand: true},
		"kindNotCommand#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "COMMAND",
				},
			},
			isCommand: false},
		"kindDeployment#2": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Deployment",
				},
			},
			isCommand: false},
		"kindNotCommand#2": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Commands",
				},
			},
			isCommand: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			isCommand := tt.taskIdentifier.isCommand()
			if tt.isCommand != isCommand {
				t.Fatalf("isCommand() => got %v, want %v", isCommand, tt.isCommand)
			}
		})
	}
}

func TestIsPod(t *testing.T) {
	tests := map[string]struct {
		taskIdentifier taskIdentifier
		isPod          bool
	}{
		"kindPod#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Pod",
				},
			},
			isPod: true},
		"kindNotPod#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "POD",
				},
			},
			isPod: false},
		"kindDeployment#2": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Deployment",
				},
			},
			isPod: false},
		"kindNotPod#2": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Pods",
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

func TestIsDeployment(t *testing.T) {
	tests := map[string]struct {
		taskIdentifier taskIdentifier
		isDeployment   bool
	}{
		"kindDeployment#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Deployment",
				},
			},
			isDeployment: true},
		"kindNotDeployment#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "DEPLOYMENT",
				},
			},
			isDeployment: false},
		"kindPod": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Pod",
				},
			},
			isDeployment: false},
		"kindNotDeployment#2": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Deployments",
				},
			},
			isDeployment: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			isDeployment := tt.taskIdentifier.isDeployment()
			if tt.isDeployment != isDeployment {
				t.Fatalf("isDeployment() => got %v, want %v", isDeployment, tt.isDeployment)
			}
		})
	}

}

func TestIsService(t *testing.T) {
	tests := map[string]struct {
		taskIdentifier taskIdentifier
		isService      bool
	}{
		"kindService#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Service",
				},
			},
			isService: true},
		"kindNotService#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "SERVICE",
				},
			},
			isService: false},
		"kindPod": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Pod",
				},
			},
			isService: false},
		"kindNotService#2": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Services",
				},
			},
			isService: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			isService := tt.taskIdentifier.isService()
			if tt.isService != isService {
				t.Fatalf("isService() => got %v, want %v", isService, tt.isService)
			}
		})
	}

}

func TestIsStoragePoolClaim(t *testing.T) {
	tests := map[string]struct {
		taskIdentifier     taskIdentifier
		isStoragePoolClaim bool
	}{
		"kindStoragePoolClaim#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "StoragePoolClaim",
				},
			},
			isStoragePoolClaim: true},
		"kindNotStoragePoolClaim#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "STORAGEPOOLCLAIM",
				},
			},
			isStoragePoolClaim: false},
		"kindPod": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Pod",
				},
			},
			isStoragePoolClaim: false},
		"kindNotStoragePooClaim#2": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "StoragePoolClaims",
				},
			},
			isStoragePoolClaim: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			isStoragePoolClaim := tt.taskIdentifier.isStoragePoolClaim()
			if tt.isStoragePoolClaim != isStoragePoolClaim {
				t.Fatalf("isStoragePoolClaim() => got %v, want %v", isStoragePoolClaim, tt.isStoragePoolClaim)
			}
		})
	}

}

func TestIsStoragePool(t *testing.T) {
	tests := map[string]struct {
		taskIdentifier taskIdentifier
		isStoragePool  bool
	}{
		"kindStoragePool#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "StoragePool",
				},
			},
			isStoragePool: true},
		"kindNotStoragePool#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "STORAGEPOOL",
				},
			},
			isStoragePool: false},
		"kindPod": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Pod",
				},
			},
			isStoragePool: false},
		"kindNotStoragePool#2": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "StoragePools",
				},
			},
			isStoragePool: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			isStoragePool := tt.taskIdentifier.isStoragePool()
			if tt.isStoragePool != isStoragePool {
				t.Fatalf("isStoragePool() => got %v, want %v", isStoragePool, tt.isStoragePool)
			}
		})
	}

}

func TestIsConfigMap(t *testing.T) {
	tests := map[string]struct {
		taskIdentifier taskIdentifier
		isConfigMap    bool
	}{
		"kindConfigMap#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "ConfigMap",
				},
			},
			isConfigMap: true},
		"kindNotConfigMap#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "ConfigMAP",
				},
			},
			isConfigMap: false},
		"kindPod": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Pod",
				},
			},
			isConfigMap: false},
		"kindNotConfigMap#2": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "ConfigMaps",
				},
			},
			isConfigMap: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			isConfigMap := tt.taskIdentifier.isConfigMap()
			if tt.isConfigMap != isConfigMap {
				t.Fatalf("isConfigMap() => got %v, want %v", isConfigMap, tt.isConfigMap)
			}
		})
	}

}

func TestIsPV(t *testing.T) {
	tests := map[string]struct {
		taskIdentifier taskIdentifier
		isPV           bool
	}{
		"kindPV#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "PersistentVolume",
				},
			},
			isPV: true},
		"kindNotPV#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "PersistentVolumE",
				},
			},
			isPV: false},
		"kindPod": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Pod",
				},
			},
			isPV: false},
		"kindNotPV#2": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "PersistentVolumes",
				},
			},
			isPV: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			isPV := tt.taskIdentifier.isPV()
			if tt.isPV != isPV {
				t.Fatalf("isPV() => got %v, want %v", isPV, tt.isPV)
			}
		})
	}

}

func TestIsPVC(t *testing.T) {
	tests := map[string]struct {
		taskIdentifier taskIdentifier
		isPVC          bool
	}{
		"kindPVC#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "PersistentVolumeClaim",
				},
			},
			isPVC: true},
		"kindNotPVC#1": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "PersistentVolumeClaiM",
				},
			},
			isPVC: false},
		"kindPod": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "Pod",
				},
			},
			isPVC: false},
		"kindNotPVC#2": {
			taskIdentifier: taskIdentifier{
				identity: MetaTaskIdentity{
					Kind: "PersistentVolumeClaims",
				},
			},
			isPVC: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			isPVC := tt.taskIdentifier.isPVC()
			if tt.isPVC != isPVC {
				t.Fatalf("isPVC() => got %v, want %v", isPVC, tt.isPVC)
			}
		})
	}

}

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
