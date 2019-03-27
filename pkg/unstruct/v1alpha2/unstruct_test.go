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

func fakeK8sResourceList(resource string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += resource + "---"
	}
	return result
}

func fakeUnstructObjectList(name string, count int) []*unstructured.Unstructured {
	result := []*unstructured.Unstructured{}
	for i := 0; i < count; i++ {
		result = append(result, fakeUnstructObject(name))
	}
	return result
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

func TestListUnstructBuilderForYaml(t *testing.T) {
	tests := map[string]struct {
		resourceYAML          string
		expectedResourceCount int
		expectErr             bool
	}{
		"Test 1": {fakeK8sResourceList(fakeK8sResource, 1), 1, false},
		"Test 2": {fakeK8sResourceList(fakeK8sResource, 2), 2, false},
		"Test 3": {fakeK8sResourceList(fakeInvalidK8sResource, 1), 1, true},
		"Test 4": {fakeK8sResourceList(fakeInvalidK8sResource, 2), 2, true},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			lb := ListUnstructBuilderForYaml(mock.resourceYAML)
			if mock.expectErr && len(lb.errs) == 0 {
				t.Fatalf("Test %s failed, expected err but got nil", name)
			} else if len(lb.items) != mock.expectedResourceCount {
				t.Fatalf("Test %s failed, expected resource count %v but got %v", name, mock.expectedResourceCount, len(lb.items))
			}
		})
	}
}

func TestListUnstructBuilderForObject(t *testing.T) {
	tests := map[string]struct {
		availableResourceCount, expectedResourceCount int
		expectErr                                     bool
	}{
		"Test 1": {1, 1, false},
		"Test 2": {2, 2, false},
		"Test 3": {3, 3, false},
		"Test 4": {4, 4, false},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			lb := ListUnstructBuilderForObject(fakeUnstructObjectList("Test", mock.availableResourceCount)...)
			if mock.expectErr && len(lb.errs) == 0 {
				t.Fatalf("Test %s failed, expected err but got nil", name)
			} else if len(lb.items) != mock.expectedResourceCount {
				t.Fatalf("Test %s failed, expected resource count %v but got %v", name, mock.expectedResourceCount, len(lb.items))
			}
		})
	}
}
