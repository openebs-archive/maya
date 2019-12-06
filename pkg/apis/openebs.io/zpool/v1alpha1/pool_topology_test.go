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

package zpool

import (
	"testing"
)

func TestGetVdevFromPath(t *testing.T) {
	tests := map[string]struct {
		devList            VdevList
		path               string
		expectedVdevExists bool
	}{
		"vdev exists in the list": {
			devList: VdevList([]Vdev{
				{
					VdevType: "root",
					Children: []Vdev{
						{
							VdevType: "replacing",
							Children: []Vdev{
								{
									VdevType: "disk",
									Path:     "/dev/by-id/path1",
								},
							},
						},
					},
				},
			},
			),
			path:               "/dev/by-id/path1",
			expectedVdevExists: true,
		},
		"vdev doesn't exist in the list": {
			devList: VdevList([]Vdev{
				{
					VdevType: "root",
					Children: []Vdev{
						{
							VdevType: "replacing",
							Children: []Vdev{
								{
									VdevType: "disk",
									Path:     "/dev/by-id/path2",
								},
								{
									VdevType: "disk",
									Path:     "/dev/by-id/path3",
								},
							},
						},
						{
							VdevType: "disk",
							Path:     "/dev/by-id/path4",
						},
					},
				},
			},
			),
			path:               "/dev/path1",
			expectedVdevExists: false,
		},
		"Without any Vdev list": {
			devList:            VdevList([]Vdev{}),
			path:               "/dev/path1",
			expectedVdevExists: false,
		},
	}
	for name, test := range tests {
		// pin it
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			_, isVdevExists := test.devList.GetVdevFromPath(test.path)
			if isVdevExists != test.expectedVdevExists {
				t.Errorf("test %s failed expected isVdev exists %t but got %t",
					name,
					test.expectedVdevExists,
					isVdevExists,
				)
			}
		})
	}

}
