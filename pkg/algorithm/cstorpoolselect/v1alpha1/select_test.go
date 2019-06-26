// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"errors"
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstor/pool/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	fakeValidHost                    string = "validHost"
	fakeinvalidHost                  string = "invalidHost"
	cstorPoolUIDLabelKey             string = "cstorpool.openebs.io/uid"
	fakeAntiAffinitySelector         string = "openebs.io/replica-anti-affinity=fake"
	fakePreferAntiAffinitySelector   string = "openebs.io/preferred-replica-anti-affinity=fake"
	fakePreferScheduleOnHostSelector string = "volume.kubernetes.io/selected-node=fake"
	fakeValue                        string = "fake"
)

type fakeLowPolicy struct{}

func (l fakeLowPolicy) name() policyName                               { return "" }
func (l fakeLowPolicy) priority() priority                             { return lowPriority }
func (l fakeLowPolicy) filter(pool *csp.CSPList) (*csp.CSPList, error) { return pool, nil }

type fakeMediumPolicy struct{}

func (l fakeMediumPolicy) name() policyName                               { return "" }
func (l fakeMediumPolicy) priority() priority                             { return mediumPriority }
func (l fakeMediumPolicy) filter(pool *csp.CSPList) (*csp.CSPList, error) { return pool, nil }

type fakeHighPolicy struct{}

func (l fakeHighPolicy) name() policyName                               { return "" }
func (l fakeHighPolicy) priority() priority                             { return highPriority }
func (l fakeHighPolicy) filter(pool *csp.CSPList) (*csp.CSPList, error) { return pool, nil }

type fakeFilterFirst struct{}

func (l fakeFilterFirst) name() policyName   { return "" }
func (l fakeFilterFirst) priority() priority { return highPriority }
func (l fakeFilterFirst) filter(pool *csp.CSPList) (*csp.CSPList, error) {
	return pool, nil
}

func fakeBuildOptionNoFilter() buildOption {
	return func(s *selection) {
		p := fakeLowPolicy{}
		s.policies.add(p)
	}
}

func fakePolicyListOk(policies []policy) *policyList {
	pl := &policyList{map[priority][]policy{}}
	for _, p := range policies {
		pl.add(p)
	}
	return pl
}

func fakeCSPListOk(uids ...string) *csp.CSPList {
	return csp.ListBuilder().WithUIDs(uids...).List()
}

func fakeCSPListScheduleOnHostOk(hostuid string, uids ...string) *csp.CSPList {
	fakeMap := map[string]string{}
	for _, uid := range uids {
		if hostuid == uid {
			fakeMap[hostuid] = fakeValidHost
		} else {
			fakeMap[uid] = fakeinvalidHost
		}
	}
	return csp.ListBuilder().WithUIDNode(fakeMap).List()
}

func fakeCVRListOk(uids ...string) cvrListFn {
	return func(namespace string, opts metav1.ListOptions) (*apis.CStorVolumeReplicaList, error) {
		l := &apis.CStorVolumeReplicaList{}
		for _, uid := range uids {
			l.Items = append(l.Items,
				apis.CStorVolumeReplica{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							string(cstorPoolUIDLabelKey): uid,
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

func TestAntiAffinityFilter(t *testing.T) {
	tests := map[string]struct {
		cvrList                       cvrListFn
		availablePools, expectedPools []string
		isError                       bool
	}{
		"Test 1": {
			cvrList:        fakeCVRListOk(),
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 1", "pool 2", "pool 3"},
			isError:        false,
		},
		"Test 2": {
			cvrList:        fakeCVRListOk("pool 4", "pool 2", "pool 7"),
			availablePools: []string{"pool 6"},
			expectedPools:  []string{"pool 6"},
			isError:        false,
		},
		"Test 3": {
			cvrList:        fakeCVRListOk("pool 4", "pool 2", "pool 7"),
			availablePools: []string{"pool 6"},
			expectedPools:  []string{"pool 6"},
			isError:        false,
		},
		"Test 4": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2"),
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 3"},
			isError:        false,
		},
		"Test 5": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2"),
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 3"},
			isError:        false,
		},
		"Test 6": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2", "pool 3"),
			availablePools: []string{"pool 1", "pool 5", "pool 3"},
			expectedPools:  []string{"pool 5"},
			isError:        false,
		},
		"Test 7": {
			cvrList:        fakeCVRListOk(),
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 1", "pool 2", "pool 3"},
			isError:        false,
		},
		"Test 8": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2", "pool 3"),
			availablePools: []string{"pool 1", "pool 5", "pool 3"},
			expectedPools:  []string{"pool 5"},
			isError:        false,
		},
		"Test 9": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2", "pool 3", "pool 4"),
			availablePools: []string{"pool 1", "pool 2", "pool 3", "pool 5"},
			expectedPools:  []string{"pool 5"},
			isError:        false,
		},
		"Test 10": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2", "pool 3", "pool 4"),
			availablePools: []string{"pool 1", "pool 2", "pool 3", "pool 5"},
			expectedPools:  []string{"pool 5"},
			isError:        false,
		},
		"Test 11": {
			cvrList:        fakeCVRListErr(),
			availablePools: []string{"pool 1", "pool 2", "pool 3", "pool 5"},
			expectedPools:  []string{},
			isError:        true,
		},
		"Test 12": {
			cvrList:        fakeCVRListErr(),
			availablePools: []string{"pool 1"},
			expectedPools:  []string{},
			isError:        true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			a := antiAffinityLabel{
				labelSelector: "should not be empty",
				cvrList:       test.cvrList,
			}
			output, err := a.filter(fakeCSPListOk(test.availablePools...))
			if test.isError && err == nil {
				t.Fatalf("test %q failed: expected error not to be nil", name)
			} else if !test.isError && err != nil {
				t.Fatalf("test %q failed: expected error to be nil", name)
			} else if output != nil && len(test.expectedPools) != len(output.GetPoolUIDs()) {
				t.Fatalf("test %q failed: expected %v but got %v", name, test.expectedPools, output.GetPoolUIDs())
			} else if output != nil && len(output.GetPoolUIDs()) != 0 && !reflect.DeepEqual(test.expectedPools, output.GetPoolUIDs()) {
				t.Fatalf("test %q failed: expected %v but got %v", name, test.expectedPools, output.GetPoolUIDs())
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
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 1", "pool 2", "pool 3"},
			isError:        false,
		},
		"Test 2": {
			cvrList:        fakeCVRListOk("pool 4", "pool 2", "pool 7"),
			availablePools: []string{"pool 6"},
			expectedPools:  []string{"pool 6"},
			isError:        false,
		},
		"Test 3": {
			cvrList:        fakeCVRListOk("pool 4", "pool 2", "pool 7"),
			availablePools: []string{"pool 6"},
			expectedPools:  []string{"pool 6"},
			isError:        false,
		},
		"Test 4": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2"),
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 3"},
			isError:        false,
		},
		"Test 5": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2"),
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 3"},
		},
		"Test 6": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2", "pool 3"),
			availablePools: []string{"pool 1", "pool 5", "pool 3"},
			expectedPools:  []string{"pool 5"},
			isError:        false,
		},
		"Test 7": {
			cvrList:        fakeCVRListOk(),
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 1", "pool 2", "pool 3"},
			isError:        false,
		},
		"Test 8": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2", "pool 3"),
			availablePools: []string{"pool 1", "pool 5", "pool 3"},
			expectedPools:  []string{"pool 5"},
			isError:        false,
		},
		"Test 9": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2", "pool 3", "pool 4"),
			availablePools: []string{"pool 1", "pool 2", "pool 3", "pool 5"},
			expectedPools:  []string{"pool 5"},
			isError:        false,
		},
		"Test 10": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2", "pool 3", "pool 4"),
			availablePools: []string{"pool 1", "pool 2", "pool 3", "pool 5"},
			expectedPools:  []string{"pool 5"},
			isError:        false,
		},
		"Test 11": {
			cvrList:        fakeCVRListErr(),
			availablePools: []string{"pool 1", "pool 2", "pool 3", "pool 5"},
			expectedPools:  []string{},
			isError:        true,
		},
		"Test 12": {
			cvrList:        fakeCVRListErr(),
			availablePools: []string{"pool 1"},
			expectedPools:  []string{},
			isError:        true,
		},
		"Test 13": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2", "pool 3", "pool 4"),
			availablePools: []string{"pool 1"},
			expectedPools:  []string{"pool 1"},
			isError:        false,
		},
		"Test 14": {
			cvrList:        fakeCVRListOk("pool 1", "pool 2", "pool 3", "pool 4"),
			availablePools: []string{"pool 1", "pool 2"},
			expectedPools:  []string{"pool 1", "pool 2"},
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

			output, err := a.filter(fakeCSPListOk(test.availablePools...))
			if test.isError && err == nil {
				t.Fatalf("test %q failed: expected error not to be nil", name)
			} else if !test.isError && err != nil {
				t.Fatalf("test %q failed: expected error to be nil", name)
			} else if output != nil && len(test.expectedPools) != len(output.GetPoolUIDs()) {
				t.Fatalf("test %q failed: expected %v but got %v", name, test.expectedPools, output.GetPoolUIDs())
			}
		})
	}
}

func TestScheduleOnHostFilter(t *testing.T) {
	tests := map[string]struct {
		hostedPool                    string
		availablePools, expectedPools []string
		isError                       bool
	}{
		"Test 1": {
			hostedPool:     "pool 1",
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 1"},
			isError:        false,
		},
		"Test 2": {
			hostedPool:     "pool 6",
			availablePools: []string{"pool 6"},
			expectedPools:  []string{"pool 6"},
			isError:        false,
		},
		"Test 3": {
			hostedPool:     "pool 6",
			availablePools: []string{"pool 6"},
			expectedPools:  []string{"pool 6"},
			isError:        false,
		},
		"Test 4": {
			hostedPool:     "pool 2",
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 2"},
			isError:        false,
		},
		"Test 5": {
			hostedPool:     "pool 3",
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 3"},
		},
		"Test 6": {
			hostedPool:     "pool 5",
			availablePools: []string{"pool 1", "pool 5", "pool 3"},
			expectedPools:  []string{"pool 5"},
			isError:        false,
		},
		"Test 7": {
			hostedPool:     "not valid pool",
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{},
			isError:        false,
		},
		"Test 8": {
			hostedPool:     "pool 1",
			availablePools: []string{"pool 1", "pool 5", "pool 3"},
			expectedPools:  []string{"pool 1"},
			isError:        false,
		},
		"Test 9": {
			hostedPool:     "pool 5",
			availablePools: []string{"pool 1", "pool 2", "pool 3", "pool 5"},
			expectedPools:  []string{"pool 5"},
			isError:        false,
		},
		"Test 10": {
			hostedPool:     "pool 4",
			availablePools: []string{"pool 1", "pool 2", "pool 3", "pool 5"},
			expectedPools:  []string{},
			isError:        false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			a := scheduleOnHost{hostName: fakeValidHost}

			output, err := a.filter(fakeCSPListScheduleOnHostOk(test.hostedPool, test.availablePools...))
			if test.isError && err == nil {
				t.Fatalf("test %q failed: expected error not to be nil", name)
			} else if !test.isError && err != nil {
				t.Fatalf("test %q failed: expected error to be nil", name)
			} else if output != nil && len(test.expectedPools) != len(output.GetPoolUIDs()) {
				t.Fatalf("test %q failed: expected %v but got %v", name, test.expectedPools, output.GetPoolUIDs())
			}
		})
	}
}

func TestPreferScheduleOnHostFilter(t *testing.T) {
	tests := map[string]struct {
		hostedPool                    string
		availablePools, expectedPools []string
		isError                       bool
	}{
		"Test 1": {
			hostedPool:     "pool 1",
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 1"},
			isError:        false,
		},
		"Test 2": {
			hostedPool:     "pool 6",
			availablePools: []string{"pool 6"},
			expectedPools:  []string{"pool 6"},
			isError:        false,
		},
		"Test 3": {
			hostedPool:     "pool 6",
			availablePools: []string{"pool 6"},
			expectedPools:  []string{"pool 6"},
			isError:        false,
		},
		"Test 4": {
			hostedPool:     "pool 2",
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 2"},
			isError:        false,
		},
		"Test 5": {
			hostedPool:     "pool 3",
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 3"},
		},
		"Test 6": {
			hostedPool:     "pool 5",
			availablePools: []string{"pool 1", "pool 5", "pool 3"},
			expectedPools:  []string{"pool 5"},
			isError:        false,
		},
		"Test 7": {
			hostedPool:     "not valid pool",
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			expectedPools:  []string{"pool 1", "pool 2", "pool 3"},
			isError:        false,
		},
		"Test 8": {
			hostedPool:     "pool 1",
			availablePools: []string{"pool 1", "pool 5", "pool 3"},
			expectedPools:  []string{"pool 1"},
			isError:        false,
		},
		"Test 9": {
			hostedPool:     "pool 5",
			availablePools: []string{"pool 1", "pool 2", "pool 3", "pool 5"},
			expectedPools:  []string{"pool 5"},
			isError:        false,
		},
		"Test 10": {
			hostedPool:     "pool 4",
			availablePools: []string{"pool 1", "pool 2", "pool 3", "pool 5"},
			expectedPools:  []string{"pool 1", "pool 2", "pool 3", "pool 5"},
			isError:        false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			a := preferScheduleOnHost{scheduleOnHost{hostName: fakeValidHost}}

			output, err := a.filter(fakeCSPListScheduleOnHostOk(test.hostedPool, test.availablePools...))
			if test.isError && err == nil {
				t.Fatalf("test %q failed: expected error not to be nil", name)
			} else if !test.isError && err != nil {
				t.Fatalf("test %q failed: expected error to be nil", name)
			} else if output != nil && len(test.expectedPools) != len(output.GetPoolUIDs()) {
				t.Fatalf("test %q failed: expected %v but got %v", name, test.expectedPools, output.GetPoolUIDs())
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
			mockSelection := &selection{policies: &policyList{map[priority][]policy{}}}
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
		"Test 1": {5},
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
		"Test 3": {preferScheduleOnHost{}, "prefer-schedule-on-host"},
		"Test 4": {scheduleOnHost{}, "schedule-on-host"},
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
			s := &selection{policies: &policyList{map[priority][]policy{}}}
			bo(s)
			policyPriority := test.expectedoutput.priority()
			if s.policies.items[policyPriority][len(s.policies.items[policyPriority])-1].name() != test.expectedoutput.name() {
				t.Fatalf("test %q failed : expected %v but got %v", name, test.expectedoutput, s.policies.items[policyPriority][len(s.policies.items[policyPriority])-1])
			}
		})
	}
}

func TestPreferScheduleOnHostAnnotation(t *testing.T) {
	tests := map[string]struct {
		hostName       string
		expectedoutput policy
	}{
		"Mock Test 1": {"host 1", preferScheduleOnHost{scheduleOnHost{hostName: "host 1"}}},
		"Mock Test 2": {"host 2", preferScheduleOnHost{scheduleOnHost{hostName: "host 2"}}},
		"Mock Test 3": {"host 3", preferScheduleOnHost{scheduleOnHost{hostName: "host 3"}}},
		"Mock Test 4": {"host 4", preferScheduleOnHost{scheduleOnHost{hostName: "host 4"}}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bo := PreferScheduleOnHostAnnotation(test.hostName)
			s := &selection{policies: &policyList{map[priority][]policy{}}}
			bo(s)
			policyPriority := test.expectedoutput.priority()
			if s.policies.items[policyPriority][len(s.policies.items[policyPriority])-1].name() != test.expectedoutput.name() {
				t.Fatalf("test %q failed : expected %v but got %v", name, test.expectedoutput, s.policies.items[policyPriority][len(s.policies.items[policyPriority])-1])
			}
		})
	}
}

func TestGetTopPriority(t *testing.T) {
	tests := map[string]struct {
		policies          []policy
		prioritisedPolicy policy
	}{
		"Test 1": {policies: []policy{}, prioritisedPolicy: nil},
		"Test 2": {policies: []policy{fakeLowPolicy{}, fakeLowPolicy{}}, prioritisedPolicy: fakeLowPolicy{}},
		"Test 3": {policies: []policy{fakeHighPolicy{}, fakeMediumPolicy{}}, prioritisedPolicy: fakeHighPolicy{}},
		"Test 4": {policies: []policy{fakeMediumPolicy{}, fakeLowPolicy{}}, prioritisedPolicy: fakeMediumPolicy{}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fakePolicies := fakePolicyListOk(test.policies)
			topPolicy := fakePolicies.getTopPriority()
			if topPolicy != nil && test.prioritisedPolicy != nil && topPolicy.priority() != test.prioritisedPolicy.priority() {
				t.Fatalf("Test %v failed: expected %v but got %v", name, test.prioritisedPolicy, topPolicy)
			}
		})
	}
}

func TestGetPolicies(t *testing.T) {
	tests := map[string]struct {
		selectors            []string
		expectedBuildoptions []buildOption
	}{
		"Test 1": {selectors: []string{}, expectedBuildoptions: []buildOption{}},
		"Test 2": {selectors: []string{fakeAntiAffinitySelector}, expectedBuildoptions: []buildOption{AntiAffinityLabel(fakeAntiAffinitySelector)}},
		"Test 3": {selectors: []string{fakePreferAntiAffinitySelector}, expectedBuildoptions: []buildOption{PreferAntiAffinityLabel(fakePreferAntiAffinitySelector)}},
		"Test 4": {selectors: []string{fakePreferScheduleOnHostSelector}, expectedBuildoptions: []buildOption{PreferScheduleOnHostAnnotation(fakePreferScheduleOnHostSelector)}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			output := GetPolicies(test.selectors...)
			if len(output) != len(test.expectedBuildoptions) {
				t.Fatalf("Test %v failed: Expected %+v but got %+v", name, test.expectedBuildoptions, output)
			}
			for i, b := range output {
				if reflect.ValueOf(b).Pointer() != reflect.ValueOf(test.expectedBuildoptions[i]).Pointer() {
					t.Fatalf("Test %v failed: Expected %v but got %v", name, test.expectedBuildoptions[i], b)
				}
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tests := map[string]struct {
		availablePools []string
		buildOptions   []buildOption
		expectedPools  []string
	}{
		"Test 1": {
			availablePools: []string{},
			buildOptions:   []buildOption{},
			expectedPools:  []string{},
		},
		"Test 2": {
			availablePools: []string{},
			buildOptions:   []buildOption{ExecutionMode(multiExecution)},
			expectedPools:  []string{},
		},
		"Test 3": {
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			buildOptions:   []buildOption{fakeBuildOptionNoFilter()},
			expectedPools:  []string{"pool 1", "pool 2", "pool 3"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			sl := newSelection(fakeCSPListOk(test.availablePools...), test.buildOptions...)
			filtered, _ := sl.filter()
			if len(test.expectedPools) != len(filtered.GetPoolUIDs()) {
				t.Fatalf("Test %v failed: Expected %v but got %v", name, test.expectedPools, filtered.GetPoolUIDs())
			}
		})
	}
}

func TestFilterPoolIDs(t *testing.T) {
	tests := map[string]struct {
		availablePools []string
		buildOptions   []buildOption
		expectedPools  []string
	}{
		"Test 1": {
			availablePools: []string{},
			buildOptions:   []buildOption{},
			expectedPools:  []string{},
		},
		"Test 2": {
			availablePools: []string{},
			buildOptions:   []buildOption{ExecutionMode(multiExecution)},
			expectedPools:  []string{},
		},
		"Test 3": {
			availablePools: []string{"pool 1", "pool 2", "pool 3"},
			buildOptions:   []buildOption{fakeBuildOptionNoFilter()},
			expectedPools:  []string{"pool 1", "pool 2", "pool 3"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			output, _ := FilterPoolIDs(fakeCSPListOk(test.availablePools...), test.buildOptions)
			if len(test.expectedPools) != len(output) {
				t.Fatalf("Test %v failed: Expected %v but got %v", name, test.expectedPools, output)
			}
		})
	}
}
