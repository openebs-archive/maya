/*
Copyright 2020 The OpenEBS Authors.

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

package webhook

import (
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (f *fixture) createCVRsFromCVRList(cvrList *apis.CStorVolumeReplicaList) error {
	for _, cvrObj := range cvrList.Items {
		_, err := f.wh.clientset.OpenebsV1alpha1().CStorVolumeReplicas(cvrObj.Namespace).Create(&cvrObj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *fixture) createCSPsFromCSPList(cspList *apis.CStorPoolList) error {
	for _, cspObj := range cspList.Items {
		_, err := f.wh.clientset.OpenebsV1alpha1().CStorPools().Create(&cspObj)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestValidateSPCDeleteRequest(t *testing.T) {
	f := newFixture().withOpenebsObjects()
	tests := map[string]struct {
		spcObj      *apis.StoragePoolClaim
		cspList     *apis.CStorPoolList
		cvrList     *apis.CStorVolumeReplicaList
		expectedRsp bool
	}{
		"When CVR exists for given SPC": {
			spcObj: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "spc1",
				},
			},
			cspList: &apis.CStorPoolList{
				Items: []apis.CStorPool{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "spc1-csp1",
							Labels: map[string]string{
								string(apis.StoragePoolClaimCPK): "spc1",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "spc1-csp2",
							Labels: map[string]string{
								string(apis.StoragePoolClaimCPK): "spc1",
							},
						},
					},
				},
			},
			cvrList: &apis.CStorVolumeReplicaList{
				Items: []apis.CStorVolumeReplica{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "spc1-csp2-cvr1",
							Namespace: "openebs",
							Labels: map[string]string{
								string(apis.CStorPoolKey): "spc1-csp2",
							},
						},
					},
				},
			},
			expectedRsp: false,
		},
		"When CSP alone exist for deleting SPC": {
			spcObj: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "spc2",
				},
			},
			cspList: &apis.CStorPoolList{
				Items: []apis.CStorPool{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "spc2-csp1",
							Labels: map[string]string{
								string(apis.StoragePoolClaimCPK): "spc2",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "spc2-csp2",
							Labels: map[string]string{
								string(apis.StoragePoolClaimCPK): "spc2",
							},
						},
					},
				},
			},
			expectedRsp: true,
		},
		"When CSP doesn't exist for deleting SPC": {
			spcObj: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "spc3",
				},
			},
			expectedRsp: true,
		},
		"When other CSP and CVR exists in cluster": {
			spcObj: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "spc4",
				},
			},
			cspList: &apis.CStorPoolList{
				Items: []apis.CStorPool{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "unknown-spc4-csp1",
							Labels: map[string]string{
								string(apis.StoragePoolClaimCPK): "unknown-spc4",
							},
						},
					},
				},
			},
			cvrList: &apis.CStorVolumeReplicaList{
				Items: []apis.CStorVolumeReplica{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "spc4-csp1-cvr1",
							Namespace: "openebs",
							Labels: map[string]string{
								string(apis.CStorPoolKey): "unknown-spc4-csp1",
							},
						},
					},
				},
			},
			expectedRsp: true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			ar := &v1beta1.AdmissionRequest{
				Operation: v1beta1.Delete,
				Name:      test.spcObj.Name,
				Object: runtime.RawExtension{
					Raw: serialize(test.spcObj),
				},
			}
			// Create fake object in etcd
			if test.cspList != nil {
				err := f.createCSPsFromCSPList(test.cspList)
				if err != nil {
					t.Errorf("failed to create csp error: %s", err.Error())
				}
			}
			if test.cvrList != nil {
				err := f.createCVRsFromCVRList(test.cvrList)
				if err != nil {
					t.Errorf("failed to create cvr error: %s", err.Error())
				}
			}
			resp := f.wh.validateSPCDeleteRequest(ar)
			if resp.Allowed != test.expectedRsp {
				t.Errorf(
					"%s test case failed expected response: %t but got %t error: %s",
					name,
					test.expectedRsp,
					resp.Allowed,
					resp.Result.Message,
				)
			}
		})
	}
}
