package v1alpha2

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	fakeK8sResource = `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: icstcee
  name: icstcee
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: icstcee
  type: LoadBalancer
`
	fakeInvalidK8sResource = `
apiVersion: v1
kind: Service
metadata
  labels
    app: icstcee
  name: icstcee
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: icstcee
  type: LoadBalancer
	`
)

func fakeUnstructObject(name string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetName(name)
	return u
}

func TestUnstructBuilderForYAML(t *testing.T) {
	tests := map[string]struct {
		resourceYAML, expectedName string
		expectError                bool
	}{
		"Test 1": {fakeK8sResource, "icstcee", false},
		"Test 2": {fakeInvalidK8sResource, "", true},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := UnstructBuilderForYaml(mock.resourceYAML)
			if mock.expectError && len(b.errs) == 0 {
				t.Fatalf("Test %s failed, expected err but got 0", name)
			} else if b.unstruct.object.GetName() != mock.expectedName {
				t.Fatalf("Test %s failed, expected %v but got %v", name, mock.expectedName, b.unstruct.object.GetName())
			}
		})
	}
}

func TestUnstructBuilderForObject(t *testing.T) {
	tests := map[string]struct {
		resourceName, expectedName string
	}{
		"Test 1": {"icstcee", "icstcee"},
		"Test 2": {"icstcee1", "icstcee1"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mockObj := fakeUnstructObject(mock.resourceName)
			b := UnstructBuilderForObject(mockObj)
			if b.unstruct.object.GetName() != mock.expectedName {
				t.Fatalf("Test %s failed, expected %v but got %v", name, mock.expectedName, b.unstruct.object.GetName())
			}
		})
	}
}

func TestUnstructBuilderForYamlBuild(t *testing.T) {
	tests := map[string]struct {
		resourceYAML, expectedName string
		expectError                bool
	}{
		"Test 1": {fakeK8sResource, "icstcee", false},
		"Test 2": {fakeInvalidK8sResource, "", true},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b, err := UnstructBuilderForYaml(mock.resourceYAML).Build()
			if mock.expectError && err == nil {
				t.Fatalf("Test %s failed, expected err but got nil", name)
			} else if b != nil && b.object.GetName() != mock.expectedName {
				t.Fatalf("Test %s failed, expected %v but got %v", name, mock.expectedName, b.object.GetName())
			}
		})
	}
}
