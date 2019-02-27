package v1alpha1

import (
	"reflect"
	"testing"
)

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
			if !reflect.DeepEqual(s.policies[len(s.policies)-1], test.expectedoutput) {
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
			if !reflect.DeepEqual(s.policies[len(s.policies)-1], test.expectedoutput) {
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

func TestTemplateFunctions(t *testing.T) {
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
		"Test 1": {1, 2, []string{"uid1"}, []string{"PAlabel1"}, []string{"Alabel1"}},
		"Test 2": {2, 3, []string{"uid1", "uid2"}, []string{"PAlabel1", "PAlabel2"}, []string{"Alabel1"}},
		"Test 3": {3, 3, []string{"uid1", "uid2", "uid3"}, []string{"PAlabel1"}, []string{"Alabel1", "Alabel2"}},
		"Test 4": {4, 4, []string{"uid1", "uid2", "uid3", "uid4"}, []string{"PAlabel1", "PAlabel2", "PAlabel3"}, []string{"Alabel1"}},
		"Test 5": {5, 4, []string{"uid1", "uid2", "uid3", "uid4", "uid5"}, []string{"PAlabel1"}, []string{"Alabel1", "Alabel2", "Alabel3"}},
		"Test 6": {6, 5, []string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6"}, []string{"PAlabel1", "PAlabel2", "PAlabel3", "PAlabel4"}, []string{"Alabel1"}},
		"Test 7": {7, 5, []string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7"}, []string{"PAlabel1"}, []string{"Alabel1", "Alabel2", "Alabel3", "Alabel4"}},
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
