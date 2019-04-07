package v1alpha2

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// GetNestedString returns the string value of a nested field.
func GetNestedString(obj map[string]interface{}, fields ...string) string {
	val, found, err := unstructured.NestedString(obj, fields...)
	if !found || err != nil {
		return ""
	}
	return val
}
