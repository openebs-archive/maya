// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"testing"
)

func TestListBuilderToAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePods  []string
		expectedPodLen int
	}{
		"Pod set 1": {[]string{}, 0},
		"Pod set 2": {[]string{"pod1"}, 1},
		"Pod set 3": {[]string{"pod1", "pod2"}, 2},
		"Pod set 4": {[]string{"pod1", "pod2", "pod3"}, 3},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := ListBuilderForAPIList(fakeAPIPodList(mock.availablePods)).List().ToAPIList()
			if mock.expectedPodLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPodLen, len(b.Items))
			}
		})
	}
}
