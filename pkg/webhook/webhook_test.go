package webhook

import (
	"encoding/json"
	"testing"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
		testAdmissionRev *v1beta1.AdmissionReview
		expectedResponse bool
	}{
		"PVC update request": {
			testAdmissionRev: &v1beta1.AdmissionReview{
				Request: &v1beta1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Kind: "PersistentVolumeClaim",
					},
					Operation: v1beta1.Update,
				},
			},
			expectedResponse: true,
		},
		"PVC connect request": {
			testAdmissionRev: &v1beta1.AdmissionReview{
				Request: &v1beta1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Kind: "PersistentVolumeClaim",
					},
					Operation: v1beta1.Connect,
				},
			},
			expectedResponse: true,
		},
		"CSP create request": {
			testAdmissionRev: &v1beta1.AdmissionReview{
				Request: &v1beta1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Kind: "CStorPool",
					},
					Operation: v1beta1.Create,
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
		webhookReq := &v1beta1.AdmissionRequest{
			Operation: v1beta1.Create,
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
