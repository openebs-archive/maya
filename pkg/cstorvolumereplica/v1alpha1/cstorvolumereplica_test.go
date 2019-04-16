package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

func TestGetPoolUIDs(t *testing.T) {
	tests := map[string]struct {
		cvrUIDs, expectedString []string
	}{
		//  UIDS are present
		"Present 1": {[]string{"uid1"}, []string{"uid1"}},
		"Present 2": {[]string{"uid1", "uid2"}, []string{"uid1", "uid2"}},
		"Present 3": {[]string{"uid1", "uid2", "uid3"}, []string{"uid1", "uid2", "uid3"}},
		"Present 4": {[]string{"uid1", "uid2", "uid3", "uid4"}, []string{"uid1", "uid2", "uid3", "uid4"}},
		// UIDS are not present
		"Not Present 1": {[]string{""}, []string{""}},
		"Not Present 2": {[]string{"", ""}, []string{"", ""}},
		"Not Present 3": {[]string{"", "", ""}, []string{"", "", ""}},
		"Not Present 4": {[]string{"", "", "", ""}, []string{"", "", "", ""}},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cvrItems := []cvr{}
			for _, p := range mock.cvrUIDs {
				cvrItems = append(cvrItems, cvr{
					apis.CStorVolumeReplica{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								string(cstorPoolUIDLabel): p,
							},
						},
					},
				},
				)
			}
			cvr := &cvrList{items: cvrItems}
			if output := cvr.GetPoolUIDs(); !reflect.DeepEqual(output, mock.expectedString) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedString, output)
			}
		})
	}
}

func TestWithListObject(t *testing.T) {
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
			cvrItems := []apis.CStorVolumeReplica{}
			for _, p := range mock.expectedUIDs {
				cvrItems = append(cvrItems, apis.CStorVolumeReplica{
					ObjectMeta: metav1.ObjectMeta{
						Name: p,
						Labels: map[string]string{
							string(cstorPoolUIDLabel): p,
						},
					},
				},
				)
			}

			b := ListBuilder().WithAPIList(&apis.CStorVolumeReplicaList{Items: cvrItems})
			for index, ob := range b.list.items {
				if !reflect.DeepEqual(ob.object, cvrItems[index]) {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, cvrItems[index], ob.object)
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
			cvrItems := []apis.CStorVolumeReplica{}
			for _, p := range mock.expectedUIDs {
				cvrItems = append(cvrItems, apis.CStorVolumeReplica{
					ObjectMeta: metav1.ObjectMeta{
						Name: p,
						Labels: map[string]string{
							string(cstorPoolUIDLabel): p,
						},
					},
				},
				)
			}

			b := ListBuilder().WithAPIList(&apis.CStorVolumeReplicaList{Items: cvrItems}).List()
			if len(b.items) != len(cvrItems) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, len(cvrItems), len(b.items))
			}

			for index, ob := range b.items {
				if !reflect.DeepEqual(ob.object, cvrItems[index]) {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, cvrItems[index], ob.object)
				}
			}
		})
	}
}

func TestListWithFilter(t *testing.T) {
	tests := map[string]struct {
		inputUIDs   []string
		phases      []apis.CStorVolumeReplicaPhase
		predicates  predicateList
		expectedLen int
	}{
		"UID set 1": {
			[]string{"uid1", "uid2"},
			[]apis.CStorVolumeReplicaPhase{apis.CVRStatusOnline, apis.CVRStatusOnline},
			predicateList{IsHealthy()},
			2,
		},
		"UID set 2": {
			[]string{"uid1", "uid2"},
			[]apis.CStorVolumeReplicaPhase{apis.CVRStatusOnline, apis.CVRStatusOffline},
			predicateList{IsHealthy()},
			1,
		},
		"UID set 3": {
			[]string{"uid1", "uid2"},
			[]apis.CStorVolumeReplicaPhase{apis.CVRStatusOffline, apis.CVRStatusOnline},
			predicateList{IsHealthy()},
			1,
		},
		"UID set 4": {
			[]string{"uid1", "uid2"},
			[]apis.CStorVolumeReplicaPhase{apis.CVRStatusOffline, apis.CVRStatusOffline},
			predicateList{IsHealthy()},
			0,
		},
	}
	for name, mock := range tests {
		name := name //pin it
		mock := mock //pin it
		t.Run(name, func(t *testing.T) {
			cvrItems := []apis.CStorVolumeReplica{}
			for i, p := range mock.inputUIDs {
				cvrItems = append(cvrItems, apis.CStorVolumeReplica{
					ObjectMeta: metav1.ObjectMeta{
						Name: p,
						Labels: map[string]string{
							string(cstorPoolUIDLabel): p,
						},
					},
					Status: apis.CStorVolumeReplicaStatus{
						Phase: mock.phases[i],
					},
				},
				)
			}

			b := ListBuilder().WithAPIList(&apis.CStorVolumeReplicaList{Items: cvrItems}).WithFilter(mock.predicates...).List()
			if len(b.items) != mock.expectedLen {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedLen, len(b.items))
			}

			for index, ob := range b.items {
				if ob.object.Status.Phase != "Healthy" {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, cvrItems[index], ob.object)
				}
			}

		})
	}
}
