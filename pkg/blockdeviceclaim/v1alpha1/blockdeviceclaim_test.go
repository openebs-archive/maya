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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
)

func fakeAPIBDCList(bdcNames []string) *apis.BlockDeviceClaimList {
	if len(bdcNames) == 0 {
		return nil
	}
	list := &apis.BlockDeviceClaimList{}
	for _, name := range bdcNames {
		bdc := apis.BlockDeviceClaim{}
		bdc.SetName(name)
		list.Items = append(list.Items, bdc)
	}
	return list
}

func fakeAPIBDCListFromNameStatusMap(bdcs map[string]apis.DeviceClaimPhase) *apis.BlockDeviceClaimList {
	if len(bdcs) == 0 {
		return nil
	}
	list := &apis.BlockDeviceClaimList{}
	for k, v := range bdcs {
		bdc := apis.BlockDeviceClaim{}
		bdc.SetName(k)
		bdc.Status.Phase = v
		list.Items = append(list.Items, bdc)
	}
	return list
}
