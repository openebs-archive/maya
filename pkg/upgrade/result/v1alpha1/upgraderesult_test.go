package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
)

func TestWithAPIList(t *testing.T) {
	inputURItems := []apis.UpgradeResult{apis.UpgradeResult{
		ObjectMeta: metav1.ObjectMeta{Name: "upgradeResultList1"}}}
	outputURItems := []*upgradeResult{&upgradeResult{object: &apis.UpgradeResult{
		ObjectMeta: metav1.ObjectMeta{Name: "upgradeResultList1"}}}}
	tests := map[string]struct {
		inputURList    *apis.UpgradeResultList
		expectedOutput *UpgradeResultList
	}{
		"empty upgrade result list": {&apis.UpgradeResultList{},
			&UpgradeResultList{}},
		"using nil input": {nil, &UpgradeResultList{}},
		"non-empty upgrade result list": {&apis.UpgradeResultList{Items: inputURItems},
			&UpgradeResultList{items: outputURItems}},
	}

	for name, mock := range tests {
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
	inputURItems := []apis.UpgradeResult{apis.UpgradeResult{
		ObjectMeta: metav1.ObjectMeta{Name: "upgradeResultList1"}}}
	outputURItems := []*upgradeResult{&upgradeResult{object: &apis.UpgradeResult{
		ObjectMeta: metav1.ObjectMeta{Name: "upgradeResultList1"}}}}
	tests := map[string]struct {
		inputURList    *apis.UpgradeResultList
		expectedOutput *UpgradeResultList
	}{
		"empty upgrade result list": {&apis.UpgradeResultList{},
			&UpgradeResultList{}},
		"using nil input": {nil, &UpgradeResultList{}},
		"non-empty upgrade result list": {&apis.UpgradeResultList{Items: inputURItems},
			&UpgradeResultList{items: outputURItems}},
	}

	for name, mock := range tests {
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
