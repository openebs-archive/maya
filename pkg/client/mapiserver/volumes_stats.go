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

package mapiserver

import (
	"encoding/json"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

const (
	statsVolumePath = "/latest/volumes/stats/"
)

// VolumeStats returns the VolumeMetrics fetched from apisever endpoint
func VolumeStats(volName, namespace string) (v1alpha1.VolumeMetricsList, error) {
	stats := v1alpha1.VolumeMetricsList{}
	body, err := getRequest(GetURL()+statsVolumePath+volName, namespace, false)
	if err != nil {
		return stats, err
	}
	err = json.Unmarshal(body, &stats)
	return stats, err
}
