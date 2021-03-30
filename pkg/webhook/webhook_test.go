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

package webhook

import (
	"context"
	"encoding/json"
	"testing"

	snapshotapi "github.com/openebs/maya/pkg/apis/openebs.io/snapshot/v1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	snapFakeClientset "github.com/openebs/maya/pkg/client/generated/openebs.io/snapshot/v1/clientset/internalclientset/fake"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sclientset "k8s.io/client-go/kubernetes/fake"
)

func TestAdmissionRequired(t *testing.T) {
	cases := []struct {
		Name, Namespace string
		want            bool
	}{
		{"default-policy", "test-namespace", true},
		{"default-policy", "default", true},
		{"no-policy-in-kube-system", "kube-system", false},
		{"no-policy-in-kube-public", "kube-public", false},
	}

	for _, c := range cases {
		meta := &metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		}

		if got := validationRequired(ignoredNamespaces, meta); got != c.want {
			t.Errorf("admissionRequired(%v)  got %v want %v", meta.Name, got, c.want)
		}
	}
}

func TestValidate(t *testing.T) {
	wh := webhook{}
	cases := map[string]struct {
		testAdmissionRev *v1.AdmissionReview
		expectedResponse bool
	}{
		"PVC update request": {
			testAdmissionRev: &v1.AdmissionReview{
				Request: &v1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Kind: "PersistentVolumeClaim",
					},
					Operation: v1.Update,
				},
			},
			expectedResponse: true,
		},
		"PVC connect request": {
			testAdmissionRev: &v1.AdmissionReview{
				Request: &v1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Kind: "PersistentVolumeClaim",
					},
					Operation: v1.Connect,
				},
			},
			expectedResponse: true,
		},
		"CSP create request": {
			testAdmissionRev: &v1.AdmissionReview{
				Request: &v1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Kind: "CStorPool",
					},
					Operation: v1.Create,
				},
			},
			expectedResponse: true,
		},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			resp := wh.validate(test.testAdmissionRev)
			if resp.Allowed != test.expectedResponse {
				t.Errorf("validate request failed got: '%v' expected: '%v'", resp.Allowed, test.expectedResponse)
			}
		})
	}
}

func serialize(v interface{}) []byte {
	bytes, _ := json.Marshal(v)
	return bytes
}

func TestValidatePVCCreateRequest(t *testing.T) {
	wh := webhook{}
	fakepvcAnnotation := make(map[string]string)
	fakepvcAnnotation["apiVersion"] = "v1"
	fakepvcAnnotation["kind"] = "PersistentVolumeClaim"
	cases := map[string]struct {
		fakePVC          corev1.PersistentVolumeClaim
		expectedResponse bool
	}{
		"Empty PVC Create request": {
			fakePVC:          corev1.PersistentVolumeClaim{},
			expectedResponse: true,
		},
		"Valid PVC Create Request": {
			fakePVC: corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: fakepvcAnnotation,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "pvc-1",
				},
			},
			expectedResponse: true,
		},
	}
	for _, test := range cases {
		webhookReq := &v1.AdmissionRequest{
			Operation: v1.Create,
			Object: runtime.RawExtension{
				Raw: serialize(test.fakePVC),
			},
		}
		resp := wh.validatePVCCreateRequest(webhookReq)
		if resp.Allowed != test.expectedResponse {
			t.Errorf("validate request failed got: '%v' expected: '%v'", resp.Allowed, test.expectedResponse)
		}
	}
}

func TestValidatePVCDeleteRequest(t *testing.T) {
	wh := &webhook{}
	wh.clientset = openebsFakeClientset.NewSimpleClientset()
	wh.snapClientSet = snapFakeClientset.NewSimpleClientset()
	wh.kubeClient = k8sclientset.NewSimpleClientset()
	tests := map[string]struct {
		pvc                   *corev1.PersistentVolumeClaim
		snapshot              *snapshotapi.VolumeSnapshot
		snapshotData          *snapshotapi.VolumeSnapshotData
		isRequiresPVCCreation bool
		expectedResponse      bool
	}{
		"When PVC was bound and doesn't have any dependents": {
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "PVC1",
					Namespace: "test",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "PV1",
				},
				Status: corev1.PersistentVolumeClaimStatus{
					Phase: corev1.ClaimBound,
				},
			},
			isRequiresPVCCreation: true,
			expectedResponse:      true,
		},
		"When PVC was not bound and doesn't have any dependents": {
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "PVC2",
					Namespace: "test",
				},
			},
			isRequiresPVCCreation: true,
			expectedResponse:      true,
		},
		"When PVC was tried to delete when dependent snapshots exists": {
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "PVC3",
					Namespace: "test",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "PV1",
				},
			},
			snapshot: &snapshotapi.VolumeSnapshot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Snap1",
					Namespace: "test",
					Labels:    map[string]string{snapshotMetadataPVName: "PV1"},
				},
			},
			isRequiresPVCCreation: true,
			expectedResponse:      false,
		},
		"When PVC was tried to delete when dependent snapshotdata exists": {
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "PVC4",
					Namespace: "test",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "PV1",
				},
			},
			snapshotData: &snapshotapi.VolumeSnapshotData{
				ObjectMeta: metav1.ObjectMeta{
					Name: "SnapData1",
				},
				Spec: snapshotapi.VolumeSnapshotDataSpec{
					PersistentVolumeRef: &corev1.ObjectReference{
						Name: "PV1",
					},
				},
			},
			isRequiresPVCCreation: true,
			expectedResponse:      false,
		},
		"When non existing PVC tried to delete": {
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "PVC5",
					Namespace: "test",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "PV1",
				},
			},
			isRequiresPVCCreation: false,
			expectedResponse:      false,
		},
		"When PVC was tried to delete when there are no dependent snapshots exists": {
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "PVC6",
					Namespace: "test",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "PV1",
				},
			},
			snapshot: &snapshotapi.VolumeSnapshot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Snap1",
					Namespace: "test",
					Labels:    map[string]string{snapshotMetadataPVName: "PV2"},
				},
			},
			isRequiresPVCCreation: true,
			expectedResponse:      true,
		},
		"When PVC was tried to delete when there are no dependent snapshotdatas exists": {
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "PVC7",
					Namespace: "test",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "PV1",
				},
			},
			snapshotData: &snapshotapi.VolumeSnapshotData{
				ObjectMeta: metav1.ObjectMeta{
					Name: "SnapData1",
				},
				Spec: snapshotapi.VolumeSnapshotDataSpec{
					PersistentVolumeRef: &corev1.ObjectReference{
						Name: "PV2",
					},
				},
			},
			isRequiresPVCCreation: true,
			expectedResponse:      true,
		},
		"Skip PVC validations if skip-validaions annotations set even snapshotData exists": {
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "PVC7",
					Namespace:   "test",
					Annotations: map[string]string{skipValidation: "true"},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "PV1",
				},
			},
			snapshotData: &snapshotapi.VolumeSnapshotData{
				ObjectMeta: metav1.ObjectMeta{
					Name: "SnapData1",
				},
				Spec: snapshotapi.VolumeSnapshotDataSpec{
					PersistentVolumeRef: &corev1.ObjectReference{
						Name: "PV1",
					},
				},
			},
			isRequiresPVCCreation: true,
			expectedResponse:      true,
		},
		"validate pvc if skipValidation annotations set to false": {
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "PVC8",
					Namespace:   "test",
					Annotations: map[string]string{skipValidation: ""},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "PV1",
				},
			},
			snapshot: &snapshotapi.VolumeSnapshot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Snap1",
					Namespace: "test",
					Labels:    map[string]string{snapshotMetadataPVName: "PV1"},
				},
			},
			isRequiresPVCCreation: true,
			expectedResponse:      false,
		},

		"validate pvc if skipValidation annotations not properly set": {
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "PVC8",
					Namespace:   "test",
					Annotations: map[string]string{skipValidation: ""},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "PV1",
				},
			},
			snapshot: &snapshotapi.VolumeSnapshot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Snap1",
					Namespace: "test",
					Labels:    map[string]string{snapshotMetadataPVName: "PV1"},
				},
			},
			isRequiresPVCCreation: true,
			expectedResponse:      false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			request := &v1.AdmissionRequest{
				Operation: v1.Delete,
				Kind: metav1.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "PersistentVolumeClaim",
				},
				Name:      test.pvc.Name,
				Namespace: test.pvc.Namespace,
			}
			if test.pvc != nil && test.isRequiresPVCCreation {
				_, err := wh.kubeClient.CoreV1().PersistentVolumeClaims(test.pvc.Namespace).
					Create(context.TODO(), test.pvc, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("%q test failed to create fake PVC %s in namespace %s err: %v", name, test.pvc.Name, test.pvc.Namespace, err)
				}
			}
			if test.snapshot != nil {
				_, err := wh.snapClientSet.VolumesnapshotV1().VolumeSnapshots(test.snapshot.Namespace).
					Create(context.TODO(), test.snapshot, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("%q test failed to create fake snapshot %s in namespace %s error: %v", name, test.snapshot.Name, test.snapshot.Namespace, err)
				}
			}
			if test.snapshotData != nil {
				_, err := wh.snapClientSet.VolumesnapshotV1().VolumeSnapshotDatas().
					Create(context.TODO(), test.snapshotData, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("%q test failed to create fake snapshotdata %s error: %v", name, test.snapshotData.Name, err)
				}
			}
			resp := wh.validatePVCDeleteRequest(request)
			if resp.Allowed != test.expectedResponse {
				t.Errorf(
					"%s test case failed expected response: %t but got %t error: %s",
					name,
					test.expectedResponse,
					resp.Allowed,
					resp.Result.Message,
				)
			}
			// Cleanup objects
			if test.pvc != nil {
				wh.kubeClient.CoreV1().PersistentVolumeClaims(test.pvc.Namespace).
					Delete(context.TODO(), test.pvc.Name, metav1.DeleteOptions{})
			}
			if test.snapshot != nil {
				wh.snapClientSet.VolumesnapshotV1().VolumeSnapshots(test.snapshot.Namespace).
					Delete(context.TODO(), test.snapshot.Name, metav1.DeleteOptions{})
			}
			if test.snapshotData != nil {
				wh.snapClientSet.VolumesnapshotV1().VolumeSnapshotDatas().
					Delete(context.TODO(), test.snapshotData.Name, metav1.DeleteOptions{})
			}
		})
	}
}
