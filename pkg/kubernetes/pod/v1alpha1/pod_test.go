package v1alpha1

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func fakeAPIPodList(podNames []string) *v1.PodList {
	if len(podNames) == 0 {
		return nil
	}

	list := &v1.PodList{}
	for _, name := range podNames {
		pod := v1.Pod{}
		pod.SetName(name)
		list.Items = append(list.Items, pod)
	}
	return list
}

func fakeRunningAPIPodObject(podNames []string) []v1.Pod {
	plist := []v1.Pod{}
	for _, podName := range podNames {
		pod := v1.Pod{}
		pod.SetName(podName)
		pod.Status.Phase = "Running"
		plist = append(plist, pod)
	}
	return plist
}

func fakeNonRunningPodList(podNames []string) []v1.Pod {
	plist := []v1.Pod{}
	for _, podName := range podNames {
		pod := v1.Pod{}
		pod.SetName(podName)
		plist = append(plist, pod)
	}
	return plist
}

func fakeAPIPodListFromNameStatusMap(pods map[string]string) []*pod {
	plist := []*pod{}
	for k, v := range pods {
		p := &v1.Pod{}
		p.SetName(k)
		p.Status.Phase = v1.PodPhase(v)
		plist = append(plist, &pod{p})
	}
	return plist
}

func TestListBuilderWithAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePods  []string
		expectedPodLen int
	}{
		"Pod set 1":  {[]string{}, 0},
		"Pod set 2":  {[]string{"pod1"}, 1},
		"Pod set 3":  {[]string{"pod1", "pod2"}, 2},
		"Pod set 4":  {[]string{"pod1", "pod2", "pod3"}, 3},
		"Pod set 5":  {[]string{"pod1", "pod2", "pod3", "pod4"}, 4},
		"Pod set 6":  {[]string{"pod1", "pod2", "pod3", "pod4", "pod5"}, 5},
		"Pod set 7":  {[]string{"pod1", "pod2", "pod3", "pod4", "pod5", "pod6"}, 6},
		"Pod set 8":  {[]string{"pod1", "pod2", "pod3", "pod4", "pod5", "pod6", "pod7"}, 7},
		"Pod set 9":  {[]string{"pod1", "pod2", "pod3", "pod4", "pod5", "pod6", "pod7", "pod8"}, 8},
		"Pod set 10": {[]string{"pod1", "pod2", "pod3", "pod4", "pod5", "pod6", "pod7", "pod8", "pod9"}, 9},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := ListBuilder().WithAPIList(fakeAPIPodList(mock.availablePods))
			if mock.expectedPodLen != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPodLen, len(b.list.items))
			}
		})
	}
}

func TestListBuilderWithAPIObjects(t *testing.T) {
	tests := map[string]struct {
		availablePods  []string
		expectedPodLen int
	}{
		"Pod set 2":  {[]string{"pod1"}, 1},
		"Pod set 3":  {[]string{"pod1", "pod2"}, 2},
		"Pod set 4":  {[]string{"pod1", "pod2", "pod3"}, 3},
		"Pod set 5":  {[]string{"pod1", "pod2", "pod3", "pod4"}, 4},
		"Pod set 6":  {[]string{"pod1", "pod2", "pod3", "pod4", "pod5"}, 5},
		"Pod set 7":  {[]string{"pod1", "pod2", "pod3", "pod4", "pod5", "pod6"}, 6},
		"Pod set 8":  {[]string{"pod1", "pod2", "pod3", "pod4", "pod5", "pod6", "pod7"}, 7},
		"Pod set 9":  {[]string{"pod1", "pod2", "pod3", "pod4", "pod5", "pod6", "pod7", "pod8"}, 8},
		"Pod set 10": {[]string{"pod1", "pod2", "pod3", "pod4", "pod5", "pod6", "pod7", "pod8", "pod9"}, 9},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			poditems := fakeAPIPodList(mock.availablePods).Items
			b := ListBuilder().WithAPIObject(poditems...)
			if mock.expectedPodLen != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPodLen, len(b.list.items))
			}

			for index, ob := range b.list.items {
				if ob.object.Name != poditems[index].Name {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, poditems[index].Name, ob.object.Name)
				}
			}
		})
	}
}

func TestListBuilderToAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePods  []string
		expectedPodLen int
	}{
		"Pod set 1":  {[]string{}, 0},
		"Pod set 2":  {[]string{"pod1"}, 1},
		"Pod set 3":  {[]string{"pod1", "pod2"}, 2},
		"Pod set 4":  {[]string{"pod1", "pod2", "pod3"}, 3},
		"Pod set 5":  {[]string{"pod1", "pod2", "pod3", "pod4"}, 4},
		"Pod set 6":  {[]string{"pod1", "pod2", "pod3", "pod4", "pod5"}, 5},
		"Pod set 7":  {[]string{"pod1", "pod2", "pod3", "pod4", "pod5", "pod6"}, 6},
		"Pod set 8":  {[]string{"pod1", "pod2", "pod3", "pod4", "pod5", "pod6", "pod7"}, 7},
		"Pod set 9":  {[]string{"pod1", "pod2", "pod3", "pod4", "pod5", "pod6", "pod7", "pod8"}, 8},
		"Pod set 10": {[]string{"pod1", "pod2", "pod3", "pod4", "pod5", "pod6", "pod7", "pod8", "pod9"}, 9},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := ListBuilder().WithAPIList(fakeAPIPodList(mock.availablePods)).List().ToAPIList()
			if mock.expectedPodLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPodLen, len(b.Items))
			}
		})
	}
}

func TestFilterList(t *testing.T) {
	tests := map[string]struct {
		availablePods map[string]string
		filteredPods  []string
		filters       predicateList
	}{
		"Pods Set 1": {
			availablePods: map[string]string{"Pod 1": "Running", "Pod 2": "CrashLoopBackOff"},
			filteredPods:  []string{"Pod 1"},
			filters:       predicateList{IsRunning()},
		},
		"Pods Set 2": {
			availablePods: map[string]string{"Pod 1": "Running", "Pod 2": "Running"},
			filteredPods:  []string{"Pod 1", "Pod 2"},
			filters:       predicateList{IsRunning()},
		},

		"Pods Set 3": {
			availablePods: map[string]string{"Pod 1": "CrashLoopBackOff", "Pod 2": "CrashLoopBackOff", "Pod 3": "CrashLoopBackOff"},
			filteredPods:  []string{},
			filters:       predicateList{IsRunning()},
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			list := ListBuilder().WithObject(fakeAPIPodListFromNameStatusMap(mock.availablePods)...).WithFilter(mock.filters...).List()
			if len(list.items) != len(mock.filteredPods) {
				t.Fatalf("Test %v failed: expected %v got %v", name, len(mock.filteredPods), len(list.items))
			}
		})
	}
}
