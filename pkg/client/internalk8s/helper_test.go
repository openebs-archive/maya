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

package internalk8s

import (
	"testing"

	"k8s.io/api/core/v1"
)

func TestIsEbsPod(t *testing.T) {
	tests := map[string]struct {
		claimNames []string
		pod        v1.Pod
		expected   bool
	}{
		"ebsPod": {
			claimNames: []string{"ebs-claim"},
			pod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "ebs-claim",
								},
							},
						},
					},
				},
			},

			expected: true,
		},
		"non-ebsPod": {
			claimNames: []string{},
			pod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "ebs-claim",
								},
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for testName, test := range tests {
		if output := isEBSPod(test.claimNames, test.pod); output != test.expected {
			t.Fatalf("%s test expected %v but got %v", testName, test.expected, output)
		}
	}
}
