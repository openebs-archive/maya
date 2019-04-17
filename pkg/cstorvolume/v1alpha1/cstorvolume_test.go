package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

func TestWithAPIList(t *testing.T) {
	inputURItems := []apis.CStorVolume{apis.CStorVolume{
		ObjectMeta: metav1.ObjectMeta{Name: "test1"}}}
	outputURItems := []*CStorVolume{&CStorVolume{object: &apis.CStorVolume{
		ObjectMeta: metav1.ObjectMeta{Name: "test1"}}}}
	tests := map[string]struct {
		inputURList    *apis.CStorVolumeList
		expectedOutput *CStorVolumeList
	}{
		"empty cstorvolume list": {&apis.CStorVolumeList{},
			&CStorVolumeList{}},
		"using nil input list": {nil, &CStorVolumeList{}},
		"non-empty cstorvolume list": {&apis.CStorVolumeList{Items: inputURItems},
			&CStorVolumeList{items: outputURItems}},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIList(mock.inputURList)
			if len(b.list.items) != len(mock.expectedOutput.items) {
				t.Fatalf("test %s failed, expected len: %d got: %d",
					name, len(mock.expectedOutput.items), len(b.list.items))
			}
			if !reflect.DeepEqual(b.list, mock.expectedOutput) {
				t.Fatalf("test %s failed, expected : %+v got : %+v",
					name, mock.expectedOutput, b.list)
			}
		})
	}
}

func TestList(t *testing.T) {
	inputURItems := []apis.CStorVolume{apis.CStorVolume{
		ObjectMeta: metav1.ObjectMeta{Name: "Test1"}}}
	outputURItems := []*CStorVolume{&CStorVolume{object: &apis.CStorVolume{
		ObjectMeta: metav1.ObjectMeta{Name: "Test1"}}}}
	tests := map[string]struct {
		inputURList    *apis.CStorVolumeList
		expectedOutput *CStorVolumeList
	}{
		"empty cstor volume list": {&apis.CStorVolumeList{},
			&CStorVolumeList{}},
		"using nil input list": {nil, &CStorVolumeList{}},
		"non-empty cstorvolume list": {&apis.CStorVolumeList{Items: inputURItems},
			&CStorVolumeList{items: outputURItems}},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIList(mock.inputURList).List()
			if len(b.items) != len(mock.expectedOutput.items) {
				t.Fatalf("test %s failed, expected len: %d got: %d",
					name, len(mock.expectedOutput.items), len(b.items))
			}
			if !reflect.DeepEqual(b, mock.expectedOutput) {
				t.Fatalf("test %s failed, expected : %+v got : %+v",
					name, mock.expectedOutput, b)
			}
		})
	}
}
