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

	"k8s.io/apimachinery/pkg/types"
)

// taskPatchMock is the mock structure to test patch operations
type taskPatchMock struct {
	// patchType indicates the type of patch
	patchType TaskPatchType
	// patch is the yaml document to be applied as a patch
	patch string
	// isError flags if the patch operation will fail
	isError bool
	// expectedPatchType is the expected type of patch after
	// running the patch executor
	expectedPatchType types.PatchType
}

func TestPatch(t *testing.T) {
	tests := map[string]taskPatchMock{
		"Test 'strategic patch' in yaml": {
			patchType: "strategic",
			patch: `
        spec:
          template:
            spec:
              affinity:
                nodeAffinity:
                  requiredDuringSchedulingIgnoredDuringExecution:
                    nodeSelectorTerms:
                    - matchExpressions:
                      - key: kubernetes.io/hostname
                        operator: In
                        values:
                        - amit-thinkpad-l470
                podAntiAffinity: null
`,
			isError:           false,
			expectedPatchType: types.StrategicMergePatchType,
		},
		"Test 'invalid patch' in yaml": {
			patchType: "invalid",
			patch: `
        spec:
          template:
            spec:
              affinity:
                podAntiAffinity: null
`,
			isError: true,
		},
		"Test 'merge patch' in yaml": {
			patchType: "merge",
			patch: `
        spec:
          template:
            spec:
              containers:
              - name: patch-demo-ctr-3
                image: gcr.io/google-samples/node-hello:1.0
`,
			isError:           false,
			expectedPatchType: types.MergePatchType,
		},
		"Test 'json patch' in yaml": {
			patchType: "json",
			patch: `
        op: add
        path: "/spec/template/metadata/labels/openebs.io/storage-type"
        value: ssd
`,
			isError:           false,
			expectedPatchType: types.JSONPatchType,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			p := TaskPatch{
				Type:  TaskPatchType(mock.patchType),
				Specs: mock.patch,
			}

			pe, err := newTaskPatchExecutor(p)
			if err != nil && !mock.isError {
				t.Fatalf("Failed to create patch executor: Expected: 'no error' Actual: '%#v'", err)
			}

			if pe == nil {
				// no need to execute cases which depend on pe instance
				return
			}

			_, err = pe.build()
			if err != nil && !mock.isError {
				t.Fatalf("Failed to build patch: Expected: 'no error' Actual: '%#v'", err)
			}

			pt := pe.patchType()
			if pt != mock.expectedPatchType && !mock.isError {
				t.Fatalf("Invalid patch type was received: Expected: '%s' Actual: '%s'", mock.expectedPatchType, pt)
			}
		}) // end of run
	}
}
