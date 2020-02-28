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
		bdr             *BlockDeviceReplacement
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
			bdr: &BlockDeviceReplacement{
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
