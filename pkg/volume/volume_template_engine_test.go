/*
Copyright 2017 The OpenEBS Authors

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

package volume

import (
	"testing"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
)

func TestUnMarshallToConfig(t *testing.T) {
	tests := map[string]struct {
		config string
		isErr  bool
		count  int
	}{
		"unmarshall to config - +ve test case - blank config": {
			config: "",
			isErr:  false,
			count:  0,
		},
		"unmarshall to config - +ve test case - one config": {
			config: `
        - name: StoragePool
          value: "default"
      `,
			isErr: false,
			count: 1,
		},
		"unmarshall to config - +ve test case - two configs": {
			config: `
        - name: ReplicaCount
          value: 3
        - name: ReplicaImage
          value: "openebs/jiva:0.6.0"
      `,
			isErr: false,
			count: 2,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c, err := cast.UnMarshallToConfig(mock.config)

			if err != nil && !mock.isErr {
				t.Fatalf("failed to test unmarshall to config: expected 'no error': actual '%#v'", err)
			}

			if !mock.isErr && len(c) != mock.count {
				t.Fatalf("failed to test unmarshall to config: expected config count '%d': actual config count '%d'", mock.count, len(c))
			}
		})
	}
}

func TestMergeConfig(t *testing.T) {
	tests := map[string]struct {
		highPriorityConfig []v1alpha1.Config
		lowPriorityConfig  []v1alpha1.Config
		countAfterMerge    int
	}{
		"merge config - +ve test case - all elements are exclusive": {
			highPriorityConfig: []v1alpha1.Config{
				{
					Name:  "ReplicaCount",
					Value: "3",
				},
				{
					Name:  "ReplicaImage",
					Value: "openebs.io/jiva:0.5.4",
				},
			},
			lowPriorityConfig: []v1alpha1.Config{
				{
					Name:  "ControllerCount",
					Value: "3",
				},
				{
					Name:  "ControllerImage",
					Value: "openebs.io/jiva:0.5.4",
				},
			},
			countAfterMerge: 4,
		},
		"merge config - +ve test case - all elements are common": {
			highPriorityConfig: []v1alpha1.Config{
				{
					Name:  "ReplicaCount",
					Value: "3",
				},
				{
					Name:  "ReplicaImage",
					Value: "openebs.io/jiva:0.5.4",
				},
			},
			lowPriorityConfig: []v1alpha1.Config{
				{
					Name:  "ReplicaCount",
					Value: "2",
				},
				{
					Name:  "ReplicaImage",
					Value: "openebs.io/jiva:2.0.0",
				},
			},
			countAfterMerge: 2,
		},
		"merge config - +ve test case - some elements are common": {
			highPriorityConfig: []v1alpha1.Config{
				{
					Name:  "ReplicaCount",
					Value: "3",
				},
				{
					Name:  "ReplicaImage",
					Value: "openebs.io/jiva:0.5.4",
				},
			},
			lowPriorityConfig: []v1alpha1.Config{
				{
					Name:  "ReplicaCount",
					Value: "2",
				},
				{
					Name:  "ControllerCount",
					Value: "1",
				},
			},
			countAfterMerge: 3,
		},
		"merge config - +ve test case - empty high priority config": {
			highPriorityConfig: nil,
			lowPriorityConfig: []v1alpha1.Config{
				{
					Name:  "ReplicaCount",
					Value: "2",
				},
				{
					Name:  "ControllerCount",
					Value: "1",
				},
			},
			countAfterMerge: 2,
		},
		"merge config - +ve test case - empty low priority config": {
			highPriorityConfig: []v1alpha1.Config{
				{
					Name:  "ReplicaCount",
					Value: "3",
				},
				{
					Name:  "ReplicaImage",
					Value: "openebs.io/jiva:0.5.4",
				},
			},
			lowPriorityConfig: nil,
			countAfterMerge:   2,
		},
		"merge config - +ve test case - both configs are empty": {
			highPriorityConfig: nil,
			lowPriorityConfig:  nil,
			countAfterMerge:    0,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			fc := cast.MergeConfig(mock.highPriorityConfig, mock.lowPriorityConfig)

			if len(fc) != mock.countAfterMerge {
				t.Fatalf("failed to test merge config: expected count '%d': actual count '%d'", mock.countAfterMerge, len(fc))
			}
		})
	}
}

// TestPrepareFinalConfig will focus on testing if config priority is maintained.
// In other words CAS volume config from PVC overrides the CAS volume config from
// SC. CAS volume config from SC overrides the default CAS volume config of
// CASTemplate.
//
// PVC --overrides--> SC --overrides--> CASTemplate
func TestPrepareFinalConfig(t *testing.T) {
	tests := map[string]struct {
		configDefault   []v1alpha1.Config
		configSC        []v1alpha1.Config
		configPVC       []v1alpha1.Config
		expectedName    string
		expectedValue   string
		countAfterMerge int
	}{
		"prepare final config - +ve test case - all configs are exclusive": {
			configDefault: []v1alpha1.Config{
				{
					Name:  "ReplicaCount",
					Value: "3",
				},
			},
			configSC: []v1alpha1.Config{
				{
					Name:  "ReplicaImage",
					Value: "openebs.io/jiva:0.5.5",
				},
			},
			configPVC: []v1alpha1.Config{
				{
					Name:  "ControllerCount",
					Value: "1",
				},
			},
			expectedName:    "ControllerCount",
			expectedValue:   "1",
			countAfterMerge: 3,
		},
		"prepare final config - +ve test case - all configs are common": {
			configDefault: []v1alpha1.Config{
				{
					Name:  "ControllerCount",
					Value: "3",
				},
			},
			configSC: []v1alpha1.Config{
				{
					Name:  "ControllerCount",
					Value: "2",
				},
			},
			configPVC: []v1alpha1.Config{
				{
					Name:  "ControllerCount",
					Value: "1",
				},
			},
			expectedName:    "ControllerCount",
			expectedValue:   "1",
			countAfterMerge: 1,
		},
		"prepare final config - +ve test case - some configs are common": {
			configDefault: []v1alpha1.Config{
				{
					Name:  "ControllerCount",
					Value: "3",
				},
			},
			configSC: []v1alpha1.Config{
				{
					Name:  "ControllerCount",
					Value: "2",
				},
			},
			configPVC: []v1alpha1.Config{
				{
					Name:  "ReplicaCount",
					Value: "1",
				},
			},
			expectedName:    "ControllerCount",
			expectedValue:   "2",
			countAfterMerge: 2,
		},
		"prepare final config - +ve test case - pvc configs is empty": {
			configDefault: []v1alpha1.Config{
				{
					Name:  "ControllerCount",
					Value: "3",
				},
			},
			configSC: []v1alpha1.Config{
				{
					Name:  "ControllerCount",
					Value: "2",
				},
			},
			configPVC:       nil,
			expectedName:    "ControllerCount",
			expectedValue:   "2",
			countAfterMerge: 1,
		},
		"prepare final config - +ve test case - sc config is empty": {
			configDefault: []v1alpha1.Config{
				{
					Name:  "ControllerCount",
					Value: "3",
				},
			},
			configSC: nil,
			configPVC: []v1alpha1.Config{
				{
					Name:  "ReplicaCount",
					Value: "1",
				},
			},
			expectedName:    "ControllerCount",
			expectedValue:   "3",
			countAfterMerge: 2,
		},
		"prepare final config - +ve test case - pvc and sc configs are empty": {
			configDefault: []v1alpha1.Config{
				{
					Name:  "ControllerCount",
					Value: "3",
				},
			},
			configSC:        nil,
			configPVC:       nil,
			expectedName:    "ControllerCount",
			expectedValue:   "3",
			countAfterMerge: 1,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			e := &Engine{
				defaultConfig: mock.configDefault,
				casConfigSC:   mock.configSC,
				casConfigPVC:  mock.configPVC,
			}

			f := e.prepareFinalConfig()

			if len(f) != mock.countAfterMerge {
				t.Fatalf("failed to test prepare final config: expected count '%d': actual count '%d'", mock.countAfterMerge, len(f))
			}

			for _, c := range f {
				if c.Name == mock.expectedName && c.Value != mock.expectedValue {
					t.Fatalf("failed to test prepare final config for '%s': expected value '%s': actual value '%s'", c.Name, mock.expectedValue, c.Value)
				}
			}
		})
	}
}
