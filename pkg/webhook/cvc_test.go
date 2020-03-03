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
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	"github.com/pkg/errors"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type fixture struct {
	wh             *webhook
	openebsObjects []runtime.Object
}

func newFixture() *fixture {
	return &fixture{
		wh: &webhook{},
	}
}

func (f *fixture) withOpenebsObjects(objects ...runtime.Object) *fixture {
	f.openebsObjects = objects
	f.wh.clientset = openebsFakeClientset.NewSimpleClientset(objects...)
	return f
}

func fakeGetCVCError(name, namespace string, clientset clientset.Interface) (*apis.CStorVolumeClaim, error) {
	return nil, errors.Errorf("fake error")
}

func TestValidateCVCUpdateRequest(t *testing.T) {
	f := newFixture().withOpenebsObjects()
	tests := map[string]struct {
		// existingObj is object existing in etcd via fake client
		existingObj  *apis.CStorVolumeClaim
		requestedObj *apis.CStorVolumeClaim
		expectedRsp  bool
		getCVCObj    getCVC
	}{
		"When Failed to Get Object From etcd": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc1",
					Namespace: "openebs",
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc1",
					Namespace: "openebs",
				},
			},
			expectedRsp: false,
			getCVCObj:   fakeGetCVCError,
		},
		"When ReplicaCount Updated": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc2",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc2",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 4,
				},
			},
			expectedRsp: false,
			getCVCObj:   getCVCObject,
		},
		"When Volume Boud Status Updated With Pool Info": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc3",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc3",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool3"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2", "pool3"},
				},
			},
			expectedRsp: true,
			getCVCObj:   getCVCObject,
		},
		"When Volume Replcias were Scaled by modifying exisitng pool names": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc4",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool3"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2", "pool3"},
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc4",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool5"},
							apis.ReplicaPoolInfo{PoolName: "pool4"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2", "pool3"},
				},
			},
			expectedRsp: false,
			getCVCObj:   getCVCObject,
		},
		"When Volume Replcias were migrated": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc5",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool3"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2", "pool3"},
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc5",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool0"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool5"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2", "pool3"},
				},
			},
			expectedRsp: false,
			getCVCObj:   getCVCObject,
		},
		"When CVC Scaling Up InProgress Performing Scaling Again": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc6",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool3"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2"},
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc6",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool3"},
							apis.ReplicaPoolInfo{PoolName: "pool4"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2"},
				},
			},
			expectedRsp: false,
			getCVCObj:   getCVCObject,
		},
		"When More Than One Replica Were Scale Down": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc7",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool3"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2", "pool3"},
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc7",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2", "pool3"},
				},
			},
			expectedRsp: false,
			getCVCObj:   getCVCObject,
		},
		"When Status Was Updated With Non Spec Pool Names": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc8",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool3"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2", "pool3"},
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc8",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool3"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2", "pool3", "pool4"},
				},
			},
			expectedRsp: false,
			getCVCObj:   getCVCObject,
		},
		"When Spec & Status Pool Names Was Updated By Controller With Invalid Pool Names Under Status": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc9",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc9",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 3,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool3"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2", "pool3", "pool4"},
				},
			},
			expectedRsp: false,
			getCVCObj:   getCVCObject,
		},
		"When Scale Up Alone Performed": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc10",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 1,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1"},
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc10",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 1,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool3"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1"},
				},
			},
			expectedRsp: true,
			getCVCObj:   getCVCObject,
		},
		"When Scale Down Alone Performed": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc11",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 1,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2"},
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc11",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 1,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2"},
				},
			},
			expectedRsp: true,
			getCVCObj:   getCVCObject,
		},
		"When Scale Up Status Was Updated Success": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc12",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 1,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1"},
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc12",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 1,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2"},
				},
			},
			expectedRsp: true,
			getCVCObj:   getCVCObject,
		},
		"When Scale Down Status Was Updated Success": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc13",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 1,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2"},
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc13",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 1,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1"},
				},
			},
			expectedRsp: true,
			getCVCObj:   getCVCObject,
		},
		"When CVC Spec Pool Names Were Repeated": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc14",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 1,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2"},
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc14",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 1,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool1"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2"},
				},
			},
			expectedRsp: false,
			getCVCObj:   getCVCObject,
		},
		"When CVC Status Pool Names Were Repeated": {
			existingObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc15",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 1,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool3"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2"},
				},
			},
			requestedObj: &apis.CStorVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cvc15",
					Namespace: "openebs",
				},
				Spec: apis.CStorVolumeClaimSpec{
					ReplicaCount: 1,
					Policy: apis.CStorVolumePolicySpec{
						ReplicaPoolInfo: []apis.ReplicaPoolInfo{
							apis.ReplicaPoolInfo{PoolName: "pool1"},
							apis.ReplicaPoolInfo{PoolName: "pool2"},
							apis.ReplicaPoolInfo{PoolName: "pool3"},
						},
					},
				},
				Status: apis.CStorVolumeClaimStatus{
					PoolInfo: []string{"pool1", "pool2", "pool2"},
				},
			},
			expectedRsp: false,
			getCVCObj:   getCVCObject,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			ar := &v1beta1.AdmissionRequest{
				Operation: v1beta1.Create,
				Object: runtime.RawExtension{
					Raw: serialize(test.requestedObj),
				},
			}
			// Create fake object in etcd
			_, err := f.wh.clientset.OpenebsV1alpha1().
				CStorVolumeClaims(test.existingObj.Namespace).
				Create(test.existingObj)
			if err != nil {
				t.Fatalf(
					"failed to create fake CVC %s Object in Namespace %s error: %v",
					test.existingObj.Name,
					test.existingObj.Namespace,
					err,
				)
			}
			resp := f.wh.validateCVCUpdateRequest(ar, test.getCVCObj)
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
