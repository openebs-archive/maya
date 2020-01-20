/*
Copyright 2019 The OpenEBS Authors.

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

package webhook

import (
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

func TestValidateSpecChanges(t *testing.T) {
	tests := map[string]struct {
		commonPoolSpecs *poolspecs
		bdr             *PoolOperations
		expectedOutput  bool
	}{
		"No change in poolSpecs": {
			commonPoolSpecs: &poolspecs{
				oldSpec: []apis.PoolSpec{
					apis.PoolSpec{
						RaidGroups: []apis.RaidGroup{
							apis.RaidGroup{
								Type: "mirror",
								BlockDevices: []apis.CStorPoolClusterBlockDevice{
									apis.CStorPoolClusterBlockDevice{
										BlockDeviceName: "bd1",
									},
									apis.CStorPoolClusterBlockDevice{
										BlockDeviceName: "bd2",
									},
								},
							},
						},
					},
				},
				newSpec: []apis.PoolSpec{
					apis.PoolSpec{
						RaidGroups: []apis.RaidGroup{
							apis.RaidGroup{
								Type: "mirror",
								BlockDevices: []apis.CStorPoolClusterBlockDevice{
									apis.CStorPoolClusterBlockDevice{
										BlockDeviceName: "bd1",
									},
									apis.CStorPoolClusterBlockDevice{
										BlockDeviceName: "bd2",
									},
								},
							},
						},
					},
				},
			},
			bdr: &PoolOperations{
				OldCSPC: &apis.CStorPoolCluster{},
				NewCSPC: &apis.CStorPoolCluster{},
			},
			expectedOutput: true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			isValid, _ := ValidateSpecChanges(test.commonPoolSpecs, test.bdr)
			if isValid != test.expectedOutput {
				t.Errorf("test: %s failed expected output %t but got %t", name, isValid, test.expectedOutput)
			}
		})
	}
}

func TestGetDuplicateBlockDeviceList(t *testing.T) {
	tests := map[string]struct {
		cspc          *apis.CStorPoolCluster
		expectedCount int
	}{
		"When CSPC has multiple block devices": {
			cspc: &apis.CStorPoolCluster{
				Spec: apis.CStorPoolClusterSpec{
					Pools: []apis.PoolSpec{
						apis.PoolSpec{
							RaidGroups: []apis.RaidGroup{
								apis.RaidGroup{
									BlockDevices: []apis.CStorPoolClusterBlockDevice{
										{BlockDeviceName: "bd-1"},
										{BlockDeviceName: "bd-2"},
									},
								},
							},
						},
						apis.PoolSpec{
							RaidGroups: []apis.RaidGroup{
								apis.RaidGroup{
									BlockDevices: []apis.CStorPoolClusterBlockDevice{
										{BlockDeviceName: "bd-3"},
										{BlockDeviceName: "bd-1"},
									},
								},
							},
						},
					},
				},
			},
			expectedCount: 1,
		},
		"When CSPC doesn't have any repetation of blockdevices": {
			cspc: &apis.CStorPoolCluster{
				Spec: apis.CStorPoolClusterSpec{
					Pools: []apis.PoolSpec{
						apis.PoolSpec{
							RaidGroups: []apis.RaidGroup{
								apis.RaidGroup{
									BlockDevices: []apis.CStorPoolClusterBlockDevice{
										{BlockDeviceName: "bd-1"},
										{BlockDeviceName: "bd-2"},
									},
								},
							},
						},
					},
				},
			},
			expectedCount: 0,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			bdList := getDuplicateBlockDeviceList(test.cspc)
			if len(bdList) != test.expectedCount {
				t.Fatalf(
					"test: %s failed expected duplicate blockdevice count: %d but got %d",
					name,
					test.expectedCount,
					len(bdList),
				)
			}
		})
	}
}

func TestGetOldCommonRaidGroups(t *testing.T) {
	test := map[string]struct {
		oldPoolSpec *apis.PoolSpec
		newPoolSpec *apis.PoolSpec
		expectedErr bool
	}{
		"When there are common raid groups": {
			oldPoolSpec: &apis.PoolSpec{
				RaidGroups: []apis.RaidGroup{
					apis.RaidGroup{
						BlockDevices: []apis.CStorPoolClusterBlockDevice{
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd1",
							},
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd2",
							},
						},
					},
					apis.RaidGroup{
						BlockDevices: []apis.CStorPoolClusterBlockDevice{
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd3",
							},
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd4",
							},
						},
					},
				},
			},
			newPoolSpec: &apis.PoolSpec{
				RaidGroups: []apis.RaidGroup{
					apis.RaidGroup{
						BlockDevices: []apis.CStorPoolClusterBlockDevice{
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd5",
							},
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd6",
							},
						},
					},
					apis.RaidGroup{
						BlockDevices: []apis.CStorPoolClusterBlockDevice{
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd7",
							},
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd1",
							},
						},
					},
					apis.RaidGroup{
						BlockDevices: []apis.CStorPoolClusterBlockDevice{
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd3",
							},
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd4",
							},
						},
					},
				},
			},
			expectedErr: false,
		},
		"When raid groups alone deleted": {
			oldPoolSpec: &apis.PoolSpec{
				RaidGroups: []apis.RaidGroup{
					apis.RaidGroup{
						BlockDevices: []apis.CStorPoolClusterBlockDevice{
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd1",
							},
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd2",
							},
						},
					},
				},
			},
			newPoolSpec: &apis.PoolSpec{
				RaidGroups: []apis.RaidGroup{
					apis.RaidGroup{
						BlockDevices: []apis.CStorPoolClusterBlockDevice{
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd3",
							},
							apis.CStorPoolClusterBlockDevice{
								BlockDeviceName: "bd4",
							},
						},
					},
				},
			},
			expectedErr: true,
		},
	}
	for name, test := range test {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			_, err := getOldCommonRaidGroups(test.oldPoolSpec, test.newPoolSpec)
			if test.expectedErr && err == nil {
				t.Fatalf("test: %s failed expected err but got nil", name)
			}
			if !test.expectedErr && err != nil {
				t.Fatalf("test: %s failed expected nil but got err: %v", name, err)
			}
		})
	}
}
