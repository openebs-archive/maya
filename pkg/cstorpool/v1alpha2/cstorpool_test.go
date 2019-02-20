package v1alpha2

import (
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func mockAlwaysTrue(*csp) bool  { return true }
func mockAlwaysFalse(*csp) bool { return false }

func TestAll(t *testing.T) {
	tests := map[string]struct {
		Predicates     predicateList
		expectedOutput bool
	}{
		// Positive predicates
		"Positive Predicate 1": {[]predicate{mockAlwaysTrue}, true},
		"Positive Predicate 2": {[]predicate{mockAlwaysTrue, mockAlwaysTrue}, true},
		"Positive Predicate 3": {[]predicate{mockAlwaysTrue, mockAlwaysTrue, mockAlwaysTrue}, true},
		// Negative Predicates
		"Negative Predicate 1": {[]predicate{mockAlwaysFalse}, false},
		"Negative Predicate 2": {[]predicate{mockAlwaysTrue, mockAlwaysFalse}, false},
		"Negative Predicate 3": {[]predicate{mockAlwaysFalse, mockAlwaysTrue}, false},
		"Negative Predicate 4": {[]predicate{mockAlwaysFalse, mockAlwaysFalse}, false},
		"Negative Predicate 5": {[]predicate{mockAlwaysFalse, mockAlwaysTrue, mockAlwaysTrue}, false},
		"Negative Predicate 6": {[]predicate{mockAlwaysTrue, mockAlwaysFalse, mockAlwaysTrue}, false},
		"Negative Predicate 7": {[]predicate{mockAlwaysTrue, mockAlwaysTrue, mockAlwaysFalse}, false},
		"Negative Predicate 8": {[]predicate{mockAlwaysTrue, mockAlwaysFalse, mockAlwaysFalse}, false},
		"Negative Predicate 9": {[]predicate{mockAlwaysFalse, mockAlwaysFalse, mockAlwaysFalse}, false},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			if output := mock.Predicates.all(&csp{}); output != mock.expectedOutput {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output)
			}
		})
	}
}

func TestIsNotUID(t *testing.T) {
	tests := map[string]struct {
		cspuid         types.UID
		uids           []string
		expectedOutput bool
	}{
		// Positive Test
		"Positive 1": {"uid6", []string{"uid1", "uid2", "uid3", "uid4"}, true},
		"Positive 2": {"uid7", []string{"uid1", "uid2", "uid3", "uid4"}, true},
		"Positive 3": {"uid8", []string{"uid1", "uid2", "uid3", "uid4"}, true},
		"Positive 4": {"uid9", []string{"uid1", "uid2", "uid3", "uid4"}, true},
		"Positive 5": {"uid10", []string{"uid1", "uid2", "uid3", "uid4"}, true},

		// Negative Test
		"Negative 1": {"uid1", []string{"uid1", "uid2", "uid3", "uid4", "uid5"}, false},
		"Negative 2": {"uid2", []string{"uid1", "uid2", "uid3", "uid4", "uid5"}, false},
		"Negative 3": {"uid3", []string{"uid1", "uid2", "uid3", "uid4", "uid5"}, false},
		"Negative 4": {"uid4", []string{"uid1", "uid2", "uid3", "uid4", "uid5"}, false},
		"Negative 5": {"uid5", []string{"uid1", "uid2", "uid3", "uid4", "uid5"}, false},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mockCSP := &csp{&apis.CStorPool{ObjectMeta: metav1.ObjectMeta{UID: mock.cspuid}}}
			p := IsNotUID(mock.uids...)
			if output := p(mockCSP); output != mock.expectedOutput {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output)
			}
		})
	}
}

func TestFilterUIDs(t *testing.T) {
	tests := map[string]struct {
		Predicates     predicateList
		UIDs           []types.UID
		expectedOutput []string
	}{
		// With all Positive predicates
		"Positive 1": {[]predicate{mockAlwaysTrue}, []types.UID{"uid1", "uid2", "uid3"}, []string{"uid1", "uid2", "uid3"}},
		"Positive 2": {[]predicate{mockAlwaysTrue, mockAlwaysTrue}, []types.UID{"uid1", "uid2", "uid3"}, []string{"uid1", "uid2", "uid3"}},
		"Positive 3": {[]predicate{mockAlwaysTrue, mockAlwaysTrue}, []types.UID{"uid1", "uid2"}, []string{"uid1", "uid2"}},
		//  With all negative predicates
		"Negative 1": {[]predicate{mockAlwaysFalse}, []types.UID{"uid1", "uid2", "uid3"}, []string{}},
		"Negative 2": {[]predicate{mockAlwaysFalse, mockAlwaysFalse}, []types.UID{"uid1", "uid2", "uid3"}, []string{}},
		"Negative 3": {[]predicate{mockAlwaysFalse, mockAlwaysFalse}, []types.UID{"uid1", "uid2", "uid3"}, []string{}},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cspL := &cspList{}
			for _, uid := range mock.UIDs {
				t := &csp{&apis.CStorPool{}}
				t.object.SetUID(uid)
				cspL.items = append(cspL.items, t)
			}
			output := cspL.FilterUIDs(mock.Predicates...)
			if len(mock.expectedOutput) != len(output) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output)
			}
			for index, val := range output {
				if val != mock.expectedOutput[index] {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output)
				}
			}
		})
	}
}

func TestWithUIDs(t *testing.T) {
	tests := map[string]struct {
		expectedUIDs []string
	}{
		"UID set 1":  {[]string{}},
		"UID set 2":  {[]string{"uid1"}},
		"UID set 3":  {[]string{"uid1", "uid2"}},
		"UID set 4":  {[]string{"uid1", "uid2", "uid3"}},
		"UID set 5":  {[]string{"uid1", "uid2", "uid3", "uid4"}},
		"UID set 6":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5"}},
		"UID set 7":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6"}},
		"UID set 8":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7"}},
		"UID set 9":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8"}},
		"UID set 10": {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8", "uid9"}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			lb := ListBuilder().WithUIDs(mock.expectedUIDs...)
			if len(lb.list.items) != len(mock.expectedUIDs) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedUIDs, lb.list.items)
			}
			for index, val := range lb.list.items {
				if string(val.object.GetUID()) != mock.expectedUIDs[index] {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedUIDs[index], string(val.object.GetUID()))
				}
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := map[string]struct {
		expectedUIDs []string
	}{
		"UID set 1":  {[]string{}},
		"UID set 2":  {[]string{"uid1"}},
		"UID set 3":  {[]string{"uid1", "uid2"}},
		"UID set 4":  {[]string{"uid1", "uid2", "uid3"}},
		"UID set 5":  {[]string{"uid1", "uid2", "uid3", "uid4"}},
		"UID set 6":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5"}},
		"UID set 7":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6"}},
		"UID set 8":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7"}},
		"UID set 9":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8"}},
		"UID set 10": {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8", "uid9"}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			lb := ListBuilder().WithUIDs(mock.expectedUIDs...).List()
			if len(lb.items) != len(mock.expectedUIDs) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedUIDs, lb.items)
			}
			for index, val := range lb.items {
				if string(val.object.GetUID()) != mock.expectedUIDs[index] {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedUIDs[index], string(val.object.GetUID()))
				}
			}
		})
	}
}
