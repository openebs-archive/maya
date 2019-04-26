// Copyright © 2017-2019 The OpenEBS Authors
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

package collector

import (
	"testing"

	v1 "github.com/openebs/maya/pkg/stats/v1alpha1"
)

type fakeVol struct {
}

func (f *fakeVol) get() (v1.VolumeStats, error) {
	return v1.VolumeStats{}, nil
}
func (f *fakeVol) parse(volStats v1.VolumeStats, metrics *metrics) stats {
	return stats{}
}

func TestCasType(t *testing.T) {
	cases := map[string]struct {
		vol Volume
		cas string
	}{
		"cas type is jiva": {
			vol: new(jiva),
			cas: "jiva",
		},
		"cas type is cstor": {
			vol: new(cstor),
			cas: "cstor",
		},
		"cas type is fakeVol": {
			vol: new(fakeVol),
			cas: "",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			got := casType(tt.vol)
			if got != tt.cas {
				t.Fatalf("casType(%v): expected %v, got %v", tt.vol, tt.cas, got)
			}
		})
	}
}
