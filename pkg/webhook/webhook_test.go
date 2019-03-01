package webhook

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
