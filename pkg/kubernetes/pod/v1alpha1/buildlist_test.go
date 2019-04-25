package v1alpha1

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func fakeAPIPodList(podNames []string) *corev1.PodList {
	if len(podNames) == 0 {
		return nil
	}

	list := &corev1.PodList{}
	for _, name := range podNames {
		pod := corev1.Pod{}
		pod.SetName(name)
		list.Items = append(list.Items, pod)
	}
	return list
}

func fakeRunningPods(podNames []string) []*Pod {
	plist := []*Pod{}
	for _, podName := range podNames {
		pod := corev1.Pod{}
		pod.SetName(podName)
		pod.Status.Phase = "Running"
		plist = append(plist, &Pod{&pod})
	}
	return plist
}

func fakeAPIPodListFromStatusMap(pods map[string]string) []*Pod {
	plist := []*Pod{}
	for k, v := range pods {
		p := &corev1.Pod{}
		p.SetName(k)
		p.Status.Phase = corev1.PodPhase(v)
		plist = append(plist, &Pod{p})
	}
	return plist
}

func TestListBuilderForAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePods    []string
		expectedPodCount int
	}{
		"Pod set 1": {
			availablePods:    []string{},
			expectedPodCount: 0,
		},
		"Pod set 2": {
			availablePods:    []string{"pod1"},
			expectedPodCount: 1,
		},
		"Pod set 3": {
			availablePods:    []string{"pod1", "pod2"},
			expectedPodCount: 2,
		},
		"Pod set 4": {
			availablePods:    []string{"pod1", "pod2", "pod3"},
			expectedPodCount: 3,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			lb := ListBuilderForAPIList(fakeAPIPodList(mock.availablePods))
			if mock.expectedPodCount != len(lb.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPodCount, len(lb.list.items))
			}
		})
	}
}

func TestListBuilderForObjectList(t *testing.T) {
	tests := map[string]struct {
		availablePods    []string
		expectedPodCount int
	}{
		"Pod set 1": {
			availablePods:    []string{},
			expectedPodCount: 0,
		},
		"Pod set 2": {
			availablePods:    []string{"pod1"},
			expectedPodCount: 1,
		},
		"Pod set 3": {
			availablePods:    []string{"pod1", "pod2"},
			expectedPodCount: 2,
		},
		"Pod set 4": {
			availablePods:    []string{"pod1", "pod2", "pod3"},
			expectedPodCount: 3,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			lb := ListBuilderForObjectList(fakeRunningPods(mock.availablePods)...)
			if mock.expectedPodCount != len(lb.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPodCount, len(lb.list.items))
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
		"Pod Set 4": {
			availablePods: map[string]string{"Pod 1": "Running", "Pod 2": "Running"},
			filteredPods:  []string{},
			filters:       predicateList{IsNil()},
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			list := ListBuilderForObjectList(fakeAPIPodListFromStatusMap(mock.availablePods)...).WithFilter(mock.filters...).List()
			if len(list.items) != len(mock.filteredPods) {
				t.Fatalf("Test %v failed: expected %v got %v", name, len(mock.filteredPods), len(list.items))
			}
		})
	}
}
