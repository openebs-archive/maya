/*
Copyright 2019 The OpenEBS Authors

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

package cstorpoolspecs

import (
	"testing"

	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

func TestIsStripePoolSpec(t *testing.T) {
	tests := map[string]struct {
		poolSpec     *apisv1alpha1.PoolSpec
		expectStripe bool
	}{
		"When pool type stripe is mentioned on default pool configurations": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "",
					},
				},
				PoolConfig: apisv1alpha1.PoolConfig{
					DefaultRaidGroupType: "stripe",
				},
			},
			expectStripe: true,
		},
		"When pool type stripe is mentioned on raid group": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "stripe",
					},
				},
			},
			expectStripe: true,
		},
		"When pool type raidz is mentioned": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "",
					},
				},
				PoolConfig: apisv1alpha1.PoolConfig{
					DefaultRaidGroupType: "raidz",
				},
			},
			expectStripe: false,
		},
		"When pool type mirror is mentioned": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "mirror",
					},
				},
			},
			expectStripe: false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			isStripe := IsStripePoolSpec(test.poolSpec)
			if isStripe != test.expectStripe {
				t.Fatalf(
					"test: %s failed excepted output %t but got %t",
					name,
					test.expectStripe,
					isStripe,
				)
			}
		})
	}
}

func TestDoesPoolSpecHasCommonPoolType(t *testing.T) {
	tests := map[string]struct {
		poolSpec     *apisv1alpha1.PoolSpec
		expectOutput bool
	}{
		"When pool type stripe is mentioned on default pool configurations": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "",
					},
				},
				PoolConfig: apisv1alpha1.PoolConfig{
					DefaultRaidGroupType: "stripe",
				},
			},
			expectOutput: true,
		},
		"When pool type stripe is mentioned on raid group but not on default": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "stripe",
					},
				},
			},
			expectOutput: true,
		},
		"When pool type is not mentioned on raidGroup1 but specified on raidGroup2 and defaults": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "",
					},
					apisv1alpha1.RaidGroup{
						Type: "stripe",
					},
				},
				PoolConfig: apisv1alpha1.PoolConfig{
					DefaultRaidGroupType: "stripe",
				},
			},
			expectOutput: true,
		},
		"when pool type is not specified on raidgroup2 but specified on raidgroup1 and default": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "stripe",
					},
					apisv1alpha1.RaidGroup{
						Type: "",
					},
				},
				PoolConfig: apisv1alpha1.PoolConfig{
					DefaultRaidGroupType: "stripe",
				},
			},
			expectOutput: true,
		},
		"When pool type is common on raidgroup1 and raidgroup2": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "stripe",
					},
					apisv1alpha1.RaidGroup{
						Type: "stripe",
					},
				},
			},
			expectOutput: true,
		},
		"When pool type is mismatch on raidgroup1 and raidgroup2": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "mirror",
					},
					apisv1alpha1.RaidGroup{
						Type: "stripe",
					},
				},
			},
			expectOutput: false,
		},
		"when pool type missmatches1": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "stripe",
					},
					apisv1alpha1.RaidGroup{
						Type: "",
					},
				},
				PoolConfig: apisv1alpha1.PoolConfig{
					DefaultRaidGroupType: "mirror",
				},
			},
			expectOutput: false,
		},
		"when pool type missmatches2": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "",
					},
					apisv1alpha1.RaidGroup{
						Type: "stripe",
					},
				},
				PoolConfig: apisv1alpha1.PoolConfig{
					DefaultRaidGroupType: "mirror",
				},
			},
			expectOutput: false,
		},
		"when pool type is not specified": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "",
					},
					apisv1alpha1.RaidGroup{
						Type: "",
					},
				},
			},
			expectOutput: true,
		},
		"when default pool type alone is specified": {
			poolSpec: &apisv1alpha1.PoolSpec{
				RaidGroups: []apisv1alpha1.RaidGroup{
					apisv1alpha1.RaidGroup{
						Type: "",
					},
					apisv1alpha1.RaidGroup{
						Type: "",
					},
				},
				PoolConfig: apisv1alpha1.PoolConfig{
					DefaultRaidGroupType: "raidz",
				},
			},
			expectOutput: true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			hasCommonPoolType := DoesPoolSpecHasCommonPoolType(test.poolSpec)
			if hasCommonPoolType != test.expectOutput {
				t.Fatalf(
					"test: %s failed excepted output %t but got %t",
					name,
					test.expectOutput,
					hasCommonPoolType,
				)
			}
		})
	}
}
