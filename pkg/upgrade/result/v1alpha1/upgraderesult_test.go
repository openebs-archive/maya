package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
)

func fakePredicate() Predicate {
	return func(p *upgradeResult) bool {
		return true
	}
}
func TestNewBuilder(t *testing.T) {
	tests := map[string]struct {
		expectUpgradeResult bool
		expectChecks        bool
	}{
		"call NewBuilder": {
			true, true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := NewBuilder()
			if (b.upgradeResult != nil) != mock.expectUpgradeResult {
				t.Fatalf("test %s failed, expect upgraderesult: %t but got: %t",
					name, mock.expectUpgradeResult, b.upgradeResult != nil)
			}
			if (b.checks != nil) != mock.expectChecks {
				t.Fatalf("test %s failed, expect checks: %t but got: %t",
					name, mock.expectChecks, b.checks != nil)
			}
		})
	}
}
func TestBuilderForRuntask(t *testing.T) {
	var tv map[string]interface{}
	validYaml :=
		`
apiVersion: openebs.io/v1alpha1
config:
kind: UpgradeResult
metadata:
   name: test-pr-abc12345
   namespace: default
status:
  actualCount: 1
  desiredCount: 2
  failedCount: 1
  resource:
    apiVersion: v1
    kind: Persistent Volume
    name: pv-1
    namespace: default
    postState:
      lastTransitionTime: 2019-03-12T06:59:46Z
      message: CStor volume Replica "cvr-1" is healthy after upgrade.
      status: Healthy
    preState:
      lastTransitionTime: 2019-03-12T06:59:46Z
      message: CStor volume Replica "cvr-1" is healthy.
      status: Healthy
  subResources:
  - apiVersion: extensions/v1beta1
    kind: Deployment
    name: target-deploy-abc
    namespace: openebs
    postState:
      lastTransitionTime: null
      message: ""
      status: ""
    preState:
      lastTransitionTime: null
      message: ""
      status: ""
tasks:
- endTime: null
  lastError: ""
  lastTransitionTime: 2019-03-12T07:50:41Z
  message: Deployment "target-deploy-abc" has been successfully patched.
  name: patch-target-deploy
  retries: 0
  startTime: null
  status: completed
- endTime: null
  lastError: ""
  lastTransitionTime: 2019-03-12T07:59:40Z
  message: ""
  name: patch-cvr
  retries: 0
  startTime: null
  status: CStor volume replica "cvr-1" has been successfully patched.
`
	invalidYaml :=
		`apiVersion: openebs.io/v1alpha1
    config:
    kind:
    UpgradeResult
    metadata:
       name: test-pr-abc12345
       namespace: default
    status:
      actualCount: 1
      desiredCount: 2
      failedCount: 1
      `

	tests := map[string]struct {
		context           string
		templateYaml      string
		templateValues    map[string]interface{}
		expectedErrLength int
	}{
		"When all the correct inputs are given": {
			"ur1", validYaml, tv, 0,
		},
		"When invalid yaml is given": {
			"ur1", invalidYaml, tv, 1,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := BuilderForRuntask(mock.context, mock.templateYaml, mock.templateValues)
			if len(b.errors) != mock.expectedErrLength {
				t.Fatalf("test %s failed, expected error length %+v, but got : %+v error:%+v",
					name, mock.expectedErrLength, len(b.errors), b.errors)
			}
		})
	}
}

func TestAddCheck(t *testing.T) {
	tests := map[string]struct {
		input                Predicate
		expectedChecksLength int
	}{
		"When a predicate is given": {
			fakePredicate(), 1,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().AddCheck(mock.input)
			if len(b.checks) != mock.expectedChecksLength {
				t.Fatalf("test %s failed, expected checks length %+v but got : %+v",
					name, mock.expectedChecksLength, len(b.checks))
			}
		})
	}
}

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
