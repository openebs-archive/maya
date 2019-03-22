package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
)

func TestWithAPIList(t *testing.T) {
	tests := map[string]struct {
		expectedURName []string
	}{
		"Name set 1": {[]string{}},
		"Name set 2": {[]string{"ur1"}},
		"Name set 3": {[]string{"ur1", "ur2"}},
		"Name set 4": {[]string{"ur1", "ur2", "ur3"}},
		"Name set 5": {[]string{"ur1", "ur2", "ur3", "ur4"}},
		"Name set 6": {[]string{"ur1", "ur2", "ur3", "ur4", "ur5"}},
		"Name set 7": {[]string{"ur1", "ur2", "ur3", "ur4", "ur5", "ur6"}},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			urItems := []apis.UpgradeResult{}
			for i := range mock.expectedURName {
				urItems = append(urItems, apis.UpgradeResult{ObjectMeta: metav1.ObjectMeta{Name: mock.expectedURName[i]}})
			}

			b := ListBuilder().WithAPIList(&apis.UpgradeResultList{Items: urItems})
			for i := range b.list.items {
				if !reflect.DeepEqual(*b.list.items[i].object, urItems[i]) {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, urItems[i], *b.list.items[i].object)
				}
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := map[string]struct {
		expectedURName []string
	}{
		"Name set 1": {[]string{}},
		"Name set 2": {[]string{"ur1"}},
		"Name set 3": {[]string{"ur1", "ur2"}},
		"Name set 4": {[]string{"ur1", "ur2", "ur3"}},
		"Name set 5": {[]string{"ur1", "ur2", "ur3", "ur4"}},
		"Name set 6": {[]string{"ur1", "ur2", "ur3", "ur4", "ur5"}},
		"Name set 7": {[]string{"ur1", "ur2", "ur3", "ur4", "ur5", "ur6"}},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			urItems := []apis.UpgradeResult{}
			for i := range mock.expectedURName {
				urItems = append(urItems, apis.UpgradeResult{ObjectMeta: metav1.ObjectMeta{Name: mock.expectedURName[i]}})
			}

			b := ListBuilder().WithAPIList(&apis.UpgradeResultList{Items: urItems}).List()
			if len(b.items) != len(urItems) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, len(urItems), len(b.items))
			}

			for i := range b.items {
				if !reflect.DeepEqual(*b.items[i].object, urItems[i]) {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, urItems[i], *b.items[i].object)
				}
			}
		})
	}
}
