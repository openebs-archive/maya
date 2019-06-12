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

	apis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	ndm "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestGetBDList(t *testing.T) {
	tests := map[string]struct {
		bdcList     *BlockDeviceClaimList
		nodeCount   int
		expectedLen []int
	}{
		"blockDeviceClaimList1": {
			bdcList: &BlockDeviceClaimList{
				ObjectList: &apis.BlockDeviceClaimList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDeviceClaim{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceClaimSpec{
								HostName:        "openebs-1234",
								BlockDeviceName: "blockdevice1",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceClaimSpec{
								HostName:        "openebs-1234",
								BlockDeviceName: "blockdevice2",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceClaimSpec{
								HostName:        "openebs-1234",
								BlockDeviceName: "blockdevice3",
							},
						},
					},
				},
			},
			nodeCount:   1,
			expectedLen: []int{3},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			nodeBDList := test.bdcList.GetBlockDeviceNamesByNode()
			if len(nodeBDList) != test.nodeCount {
				t.Errorf("Test %q failed: expected block device object count %d but got %d", name, test.nodeCount, len(nodeBDList))
			}
		})
	}
}
