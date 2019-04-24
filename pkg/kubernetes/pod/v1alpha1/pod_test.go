package v1alpha1

import (
	"testing"
)

func TestListBuilderToAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePods  []string
		expectedPodLen int
	}{
		"Pod set 1": {[]string{}, 0},
		"Pod set 2": {[]string{"pod1"}, 1},
		"Pod set 3": {[]string{"pod1", "pod2"}, 2},
		"Pod set 4": {[]string{"pod1", "pod2", "pod3"}, 3},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := ListBuilderForAPIList(fakeAPIPodList(mock.availablePods)).List().ToAPIList()
			if mock.expectedPodLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPodLen, len(b.Items))
			}
		})
	}
}
