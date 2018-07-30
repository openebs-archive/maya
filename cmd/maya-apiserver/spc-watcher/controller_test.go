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
package spc

import (
	"testing"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func TestIsDeleteEvent(t *testing.T) {

	tests := map[string]struct {
		fakestoragepoolclaim *apis.StoragePoolClaim
		expectedResult       bool
	}{
		"DeletionTimestamp is nil": {
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: api_meta.ObjectMeta{
					DeletionTimestamp: nil,
				},
			},
			expectedResult: false},
		"DeletionTimestamp is not nil": {
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: api_meta.ObjectMeta{
					DeletionTimestamp: &api_meta.Time{time.Now(),
					},
				},
			},
			expectedResult: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsDeleteEvent(test.fakestoragepoolclaim)
			if result != test.expectedResult {
				t.Errorf("Test case failed: expected '%v' but got '%v' ", test.expectedResult, result)
			}
		})
	}
}
