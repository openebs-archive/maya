package webhook

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAdmissionRequired(t *testing.T) {
	cases := []struct {
		meta *metav1.ObjectMeta
		key  string
		want bool
	}{
		{
			meta: &metav1.ObjectMeta{
				Name:        "default-policy",
				Namespace:   "test-namespace",
				Annotations: map[string]string{},
			},
			want: true,
		},
		{
			meta: &metav1.ObjectMeta{
				Name:      "no-policy",
				Namespace: "default",
			},
			want: true,
		},
		{
			meta: &metav1.ObjectMeta{
				Name:        "no-policy",
				Namespace:   "default",
				Annotations: map[string]string{"foo": "bar"},
			},
			want: true,
		},
		{
			meta: &metav1.ObjectMeta{
				Name:        "force-off-policy",
				Namespace:   "test-namespace",
				Annotations: map[string]string{admissionWebhookAnnotationValidateKey: "off"},
			},
			want: false,
		},
		{
			meta: &metav1.ObjectMeta{
				Name:        "force-off-policy",
				Namespace:   "test-namespace",
				Annotations: map[string]string{admissionWebhookAnnotationValidateKey: "n"},
			},
			want: false,
		},
		{
			meta: &metav1.ObjectMeta{
				Name:        "no-policy-in-kube-system",
				Namespace:   "kube-system",
				Annotations: map[string]string{},
			},
			want: false,
		},
		{
			meta: &metav1.ObjectMeta{
				Name:        "no-policy-in-kube-public",
				Namespace:   "kube-system",
				Annotations: map[string]string{},
			},
			want: false,
		},
	}
	for _, c := range cases {
		if got := validationRequired(ignoredNamespaces, c.meta); got != c.want {
			t.Errorf("admissionRequired(%v)  got %v want %v", c.meta.Name, got, c.want)
		}
	}
}
