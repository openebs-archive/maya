package v1alpha1

import (
	"errors"
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakeCVRListOk(uids ...string) cvrListFn {
	return func(namespace string, opts metav1.ListOptions) (*apis.CStorVolumeReplicaList, error) {
		l := &apis.CStorVolumeReplicaList{}
		for _, uid := range uids {
			l.Items = append(l.Items,
				apis.CStorVolumeReplica{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							string(cstorPoolUIDLabel): uid,
						},
					},
				},
			)
		}
		return l, nil
	}
}

func fakeCVRListErr() cvrListFn {
	return func(namespace string, opts metav1.ListOptions) (*apis.CStorVolumeReplicaList, error) {
		return nil, errors.New("fake error")
	}
}

func TestPreferAntiAffinityLabel(t *testing.T) {
	tests := map[string]struct {
		label          string
		expectedoutput policy
	}{
		"Mock Test 1": {"label1", preferAntiAffinityLabel{antiAffinityLabel{labelSelector: "label1"}}},
		"Mock Test 2": {"label2", preferAntiAffinityLabel{antiAffinityLabel{labelSelector: "label2"}}},
		"Mock Test 3": {"label3", preferAntiAffinityLabel{antiAffinityLabel{labelSelector: "label3"}}},
		"Mock Test 4": {"label4", preferAntiAffinityLabel{antiAffinityLabel{labelSelector: "label4"}}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bo := PreferAntiAffinityLabel(test.label)
			s := &selection{}
			bo(s)
			if s.policies[len(s.policies)-1].name() != test.expectedoutput.name() {
				t.Fatalf("test %q failed : expected %v but got %v", name, test.expectedoutput, s.policies[len(s.policies)-1])
			}
		})
	}
}

func TestAntiAffinityLabel(t *testing.T) {
	tests := map[string]struct {
		mocklabel      string
		expectedoutput policy
	}{
		"Mock Test 1": {"label1", antiAffinityLabel{labelSelector: "label1"}},
		"Mock Test 2": {"label2", antiAffinityLabel{labelSelector: "label2"}},
		"Mock Test 3": {"label3", antiAffinityLabel{labelSelector: "label3"}},
		"Mock Test 4": {"label4", antiAffinityLabel{labelSelector: "label4"}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bo := AntiAffinityLabel(test.mocklabel)
			s := &selection{}
			bo(s)
			if s.policies[len(s.policies)-1].name() != test.expectedoutput.name() {
				t.Fatalf("test %q failed : expected %v but got %v", name, test.expectedoutput, s.policies[len(s.policies)-1])
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		isPreferAntiAffinity, isAntiAffinity bool
		expectedError                        bool
	}{
		"Mock Test 1": {false, false, false},
		"Mock Test 2": {true, false, false},
		"Mock Test 3": {false, true, false},
		"Mock Test 4": {true, true, true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockSelection := &selection{}
			if test.isAntiAffinity {
				p := AntiAffinityLabel("antiAffinity")
				p(mockSelection)
			}

			if test.isPreferAntiAffinity {
				p := PreferAntiAffinityLabel("preferedAntiAffinity")
				p(mockSelection)
			}

			err := mockSelection.validate()
			if test.expectedError && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !test.expectedError && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestTemplateFunctionsCount(t *testing.T) {
	tests := map[string]struct {
		expectedLength int
	}{
		"Test 1": {4},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := TemplateFunctions()
			if len(p) != test.expectedLength {
				t.Fatalf("test %q failed: expected items %v but got %v", name, test.expectedLength, len(p))
			}
		})
	}
}

func TestName(t *testing.T) {
	tests := map[string]struct {
		Invoker      policy
		expectedName string
	}{
		"Test 1": {antiAffinityLabel{}, "anti-affinity-label"},
		"Test 2": {preferAntiAffinityLabel{}, "prefer-anti-affinity-label"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			n := test.Invoker.name()
			if !reflect.DeepEqual(string(n), test.expectedName) {
				t.Fatalf("test %q failed : expected %v but got %v", name, test.expectedName, string(n))
			}
		})
	}
}

func TestNewSelection(t *testing.T) {
	tests := map[string]struct {
		expectedUIDs, expectedBuildOptions                 int
		UIDs, preferAntiAffinityLabels, AntiAffinityLabels []string
	}{
		"Test 1": {
			expectedUIDs: 1, expectedBuildOptions: 2,
			UIDs:                     []string{"uid1"},
			preferAntiAffinityLabels: []string{"PAlabel1"},
			AntiAffinityLabels:       []string{"Alabel1"},
		},
		"Test 2": {
			expectedUIDs: 2, expectedBuildOptions: 3,
			UIDs:                     []string{"uid1", "uid2"},
			preferAntiAffinityLabels: []string{"PAlabel1", "PAlabel2"},
			AntiAffinityLabels:       []string{"Alabel1"},
		},
		"Test 3": {
			expectedUIDs: 3, expectedBuildOptions: 3,
			UIDs:                     []string{"uid1", "uid2", "uid3"},
			preferAntiAffinityLabels: []string{"PAlabel1"},
			AntiAffinityLabels:       []string{"Alabel1", "Alabel2"},
		},
		"Test 4": {
			expectedUIDs: 4, expectedBuildOptions: 4,
			UIDs:                     []string{"uid1", "uid2", "uid3", "uid4"},
			preferAntiAffinityLabels: []string{"PAlabel1", "PAlabel2", "PAlabel3"},
			AntiAffinityLabels:       []string{"Alabel1"},
		},
		"Test 5": {
			expectedUIDs: 5, expectedBuildOptions: 4,
			UIDs:                     []string{"uid1", "uid2", "uid3", "uid4", "uid5"},
			preferAntiAffinityLabels: []string{"PAlabel1"},
			AntiAffinityLabels:       []string{"Alabel1", "Alabel2", "Alabel3"},
		},
		"Test 6": {
			expectedUIDs: 6, expectedBuildOptions: 5,
			UIDs:                     []string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6"},
			preferAntiAffinityLabels: []string{"PAlabel1", "PAlabel2", "PAlabel3", "PAlabel4"},
			AntiAffinityLabels:       []string{"Alabel1"},
		},
		"Test 7": {
			expectedUIDs: 7, expectedBuildOptions: 5,
			UIDs:                     []string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7"},
			preferAntiAffinityLabels: []string{"PAlabel1"},
			AntiAffinityLabels:       []string{"Alabel1", "Alabel2", "Alabel3", "Alabel4"}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockBuildOptions := []buildOption{}
			for _, lab := range test.AntiAffinityLabels {
				mockBuildOptions = append(mockBuildOptions, AntiAffinityLabel(lab))
			}

			for _, lab := range test.preferAntiAffinityLabels {
				mockBuildOptions = append(mockBuildOptions, PreferAntiAffinityLabel(lab))
			}

			p := newSelection(test.UIDs, mockBuildOptions...)
			if len(p.policies) != test.expectedBuildOptions {
				t.Fatalf("test %q failed: expected %v but got %v", name, test.expectedBuildOptions, len(p.policies))
			}
			if len(p.poolUIDs) != test.expectedUIDs {
				t.Fatalf("test %q failed: expected %v but got %v", name, test.expectedUIDs, p.poolUIDs)
			}
		})
	}
}

func TestAntiAffinityFilter(t *testing.T) {
	tests := map[string]struct {
		cvrList                       cvrListFn
		availablePools, expectedPools []string
		isError                       bool
	}{
		"Test 1": {
			cvrList:        fakeCVRListOk(),
			availablePools: []string{"uid 1", "uid 2", "uid 3"},
			expectedPools:  []string{"uid 1", "uid 2", "uid 3"},
			isError:        false,
		},
		"Test 2": {
			cvrList:        fakeCVRListOk("uid 4", "uid 2", "uid 7"),
			availablePools: []string{"uid 6"},
			expectedPools:  []string{"uid 6"},
			isError:        false,
		},
		"Test 3": {
			cvrList:        fakeCVRListOk("uid 4", "uid 2", "uid 7"),
			availablePools: []string{"uid 6"},
			expectedPools:  []string{"uid 6"},
			isError:        false,
		},
		"Test 4": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2"),
			availablePools: []string{"uid 1", "uid 2", "uid 3"},
			expectedPools:  []string{"uid 3"},
			isError:        false,
		},
		"Test 5": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2"),
			availablePools: []string{"uid 1", "uid 2", "uid 3"},
			expectedPools:  []string{"uid 3"},
			isError:        false,
		},
		"Test 6": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2", "uid 3"),
			availablePools: []string{"uid 1", "uid 5", "uid 3"},
			expectedPools:  []string{"uid 5"},
			isError:        false,
		},
		"Test 7": {
			cvrList:        fakeCVRListOk(),
			availablePools: []string{"uid 1", "uid 2", "uid 3"},
			expectedPools:  []string{"uid 1", "uid 2", "uid 3"},
			isError:        false,
		},
		"Test 8": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2", "uid 3"),
			availablePools: []string{"uid 1", "uid 5", "uid 3"},
			expectedPools:  []string{"uid 5"},
			isError:        false,
		},
		"Test 9": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2", "uid 3", "uid 4"),
			availablePools: []string{"uid 1", "uid 2", "uid 3", "uid 5"},
			expectedPools:  []string{"uid 5"},
			isError:        false,
		},
		"Test 10": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2", "uid 3", "uid 4"),
			availablePools: []string{"uid 1", "uid 2", "uid 3", "uid 5"},
			expectedPools:  []string{"uid 5"},
			isError:        false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			a := antiAffinityLabel{
				labelSelector: "should not be empty",
				cvrList:       test.cvrList,
			}
			output, err := a.filter(test.availablePools)
			if test.isError && err == nil {
				t.Fatalf("test %q failed: expected error not to be nil", name)
			} else if !test.isError && err != nil {
				t.Fatalf("test %q failed: expected error to be nil", name)
			} else if len(test.expectedPools) != len(output) {
				t.Fatalf("test %q failed: expected %v but got %v", name, test.expectedPools, output)
			} else if len(output) != 0 && !reflect.DeepEqual(test.expectedPools, output) {
				t.Fatalf("test %q failed: expected %v but got %v", name, test.expectedPools, output)
			}
		})
	}
}

func TestPreferredAntiAffinityFilter(t *testing.T) {
	tests := map[string]struct {
		cvrList                       cvrListFn
		availablePools, expectedPools []string
		isError                       bool
	}{
		"Test 1": {
			cvrList:        fakeCVRListOk(),
			availablePools: []string{"uid 1", "uid 2", "uid 3"},
			expectedPools:  []string{"uid 1", "uid 2", "uid 3"},
			isError:        false,
		},
		"Test 2": {
			cvrList:        fakeCVRListOk("uid 4", "uid 2", "uid 7"),
			availablePools: []string{"uid 6"},
			expectedPools:  []string{"uid 6"},
			isError:        false,
		},
		"Test 3": {
			cvrList:        fakeCVRListOk("uid 4", "uid 2", "uid 7"),
			availablePools: []string{"uid 6"},
			expectedPools:  []string{"uid 6"},
			isError:        false,
		},
		"Test 4": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2"),
			availablePools: []string{"uid 1", "uid 2", "uid 3"},
			expectedPools:  []string{"uid 3"},
			isError:        false,
		},
		"Test 5": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2"),
			availablePools: []string{"uid 1", "uid 2", "uid 3"},
			expectedPools:  []string{"uid 3"},
		},
		"Test 6": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2", "uid 3"),
			availablePools: []string{"uid 1", "uid 5", "uid 3"},
			expectedPools:  []string{"uid 5"},
			isError:        false,
		},
		"Test 7": {
			cvrList:        fakeCVRListOk(),
			availablePools: []string{"uid 1", "uid 2", "uid 3"},
			expectedPools:  []string{"uid 1", "uid 2", "uid 3"},
			isError:        false,
		},
		"Test 8": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2", "uid 3"),
			availablePools: []string{"uid 1", "uid 5", "uid 3"},
			expectedPools:  []string{"uid 5"},
			isError:        false,
		},
		"Test 9": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2", "uid 3", "uid 4"),
			availablePools: []string{"uid 1", "uid 2", "uid 3", "uid 5"},
			expectedPools:  []string{"uid 5"},
			isError:        false,
		},
		"Test 10": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2", "uid 3", "uid 4"),
			availablePools: []string{"uid 1", "uid 2", "uid 3", "uid 5"},
			expectedPools:  []string{"uid 5"},
			isError:        false,
		},
		"Test 11": {
			cvrList:        fakeCVRListErr(),
			availablePools: []string{"uid 1", "uid 2", "uid 3", "uid 5"},
			expectedPools:  []string{},
			isError:        true,
		},
		"Test 12": {
			cvrList:        fakeCVRListErr(),
			availablePools: []string{"uid 1"},
			expectedPools:  []string{},
			isError:        true,
		},
		"Test 13": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2", "uid 3", "uid 4"),
			availablePools: []string{"uid 1"},
			expectedPools:  []string{"uid 1"},
			isError:        false,
		},
		"Test 14": {
			cvrList:        fakeCVRListOk("uid 1", "uid 2", "uid 3", "uid 4"),
			availablePools: []string{"uid 1", "uid 2"},
			expectedPools:  []string{"uid 1", "uid 2"},
			isError:        false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			a := preferAntiAffinityLabel{
				antiAffinityLabel: antiAffinityLabel{
					labelSelector: "should not be empty",
					cvrList:       test.cvrList,
				},
			}

			output, err := a.filter(test.availablePools)
			if test.isError && err == nil {
				t.Fatalf("test %q failed: expected error not to be nil", name)
			} else if !test.isError && err != nil {
				t.Fatalf("test %q failed: expected error to be nil", name)
			} else if len(test.expectedPools) != len(output) {
				t.Fatalf("test %q failed: expected %v but got %v", name, test.expectedPools, output)
			} else if len(output) != 0 && !reflect.DeepEqual(test.expectedPools, output) {
				t.Fatalf("test %q failed: expected %v but got %v", name, test.expectedPools, output)
			}
		})
	}
}

func TestIsPolicy(t *testing.T) {
	tests := map[string]struct {
		policies     []policy
		expectpolicy policyName
		isPresent    bool
	}{
		"Test 1": {[]policy{&antiAffinityLabel{}}, antiAffinityLabelPolicyName, true},
		"Test 2": {[]policy{&antiAffinityLabel{}}, antiAffinityLabelPolicyName, true},
		"Test 3": {[]policy{&preferAntiAffinityLabel{}}, antiAffinityLabelPolicyName, false},
		"Test 4": {[]policy{&preferAntiAffinityLabel{}}, antiAffinityLabelPolicyName, false},
		"Test 5": {[]policy{&preferAntiAffinityLabel{}, &preferAntiAffinityLabel{}}, antiAffinityLabelPolicyName, false},
		"Test 6": {[]policy{&preferAntiAffinityLabel{}, &preferAntiAffinityLabel{}}, preferAntiAffinityLabelPolicyName, true},
	}

	for name, test := range tests {
		s := &selection{policies: test.policies}

		output := s.isPolicy(test.expectpolicy)
		if output != test.isPresent {
			t.Fatalf("test %q failed: expected %v but got %v", name, test.isPresent, output)
		}
	}
}
