/*
Copyright 2018 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package task

import (
	"reflect"
	"testing"
	"time"

	"github.com/openebs/maya/pkg/template"
	"sigs.k8s.io/yaml"
)

func TestMetaTaskPropsSelectOverride(t *testing.T) {
	tests := map[string]struct {
		origMetaTaskProps   MetaTaskProps
		targetMetaTaskProps MetaTaskProps
	}{
		//
		// start of test case
		//
		"Negative test: orig & target meta props are empty": {
			origMetaTaskProps:   MetaTaskProps{},
			targetMetaTaskProps: MetaTaskProps{},
		},
		//
		// start of test case
		//
		"Negative test: target meta props is empty": {
			origMetaTaskProps: MetaTaskProps{
				RunNamespace: "openebs",
				Options:      "app=storage",
			},
			targetMetaTaskProps: MetaTaskProps{},
		},
		//
		// start of test case
		//
		"Negative test: orig meta props is empty": {
			origMetaTaskProps: MetaTaskProps{},
			targetMetaTaskProps: MetaTaskProps{
				RunNamespace: "openebs",
				Options:      "app=storage",
			},
		},
		//
		// start of test case
		//
		"Positive test: override orig meta props with target": {
			origMetaTaskProps: MetaTaskProps{
				RunNamespace: "default",
				Owner:        "maya",
			},
			targetMetaTaskProps: MetaTaskProps{
				RunNamespace: "openebs",
				Options:      "app=storage",
			},
		},
		//
		// start of test case
		//
		"Positive test: override orig meta props with target entirely": {
			origMetaTaskProps: MetaTaskProps{
				RunNamespace: "default",
				Owner:        "maya",
				Options:      "app=pv",
				ObjectName:   "abc-123",
				Retry:        "5,10s",
			},
			targetMetaTaskProps: MetaTaskProps{
				RunNamespace: "openebs",
				Owner:        "user",
				Options:      "app=storage",
				ObjectName:   "def-223",
				Retry:        "5,20s",
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mock.origMetaTaskProps = mock.origMetaTaskProps.selectOverride(mock.targetMetaTaskProps)

			if len(mock.targetMetaTaskProps.Owner) != 0 && mock.targetMetaTaskProps.Owner != mock.origMetaTaskProps.Owner {
				t.Fatalf("failed to test meta task props select override: expected owner '%s': actual owner '%s'", mock.targetMetaTaskProps.Owner, mock.origMetaTaskProps.Owner)
			}

			if len(mock.targetMetaTaskProps.Options) != 0 && mock.targetMetaTaskProps.Options != mock.origMetaTaskProps.Options {
				t.Fatalf("failed to test meta task props select override: expected options '%s': actual options '%s'", mock.targetMetaTaskProps.Options, mock.origMetaTaskProps.Options)
			}

			if len(mock.targetMetaTaskProps.ObjectName) != 0 && mock.targetMetaTaskProps.ObjectName != mock.origMetaTaskProps.ObjectName {
				t.Fatalf("failed to test meta task props select override: expected ObjectName '%s': actual ObjectName '%s'", mock.targetMetaTaskProps.ObjectName, mock.origMetaTaskProps.ObjectName)
			}

			if len(mock.targetMetaTaskProps.RunNamespace) != 0 && mock.targetMetaTaskProps.RunNamespace != mock.origMetaTaskProps.RunNamespace {
				t.Fatalf("failed to test meta task props select override: expected RunNamespace '%s': actual RunNamespace '%s'", mock.targetMetaTaskProps.RunNamespace, mock.origMetaTaskProps.RunNamespace)
			}

			if len(mock.targetMetaTaskProps.Retry) != 0 && mock.targetMetaTaskProps.Retry != mock.origMetaTaskProps.Retry {
				t.Fatalf("failed to test meta task props select override: expected Retry '%s': actual Retry '%s'", mock.targetMetaTaskProps.Retry, mock.origMetaTaskProps.Retry)
			}
		})
	}
}

func TestNewMetaTaskExecutor(t *testing.T) {
	tests := map[string]struct {
		id     string
		yaml   string
		values map[string]interface{}
		isErr  bool
	}{
		"new meta task - -ve test case - invalid yaml": {
			id: "121",
			// an invalid yaml that can not unmarshall into MetaTask
			yaml: `Hi {{.there}}`,
			values: map[string]interface{}{
				"there": "openebs",
			},
			isErr: true,
		},
		"new meta task - -ve test case - invalid template": {
			id: "121",
			// an invalid template
			yaml: `Hi {{.there}`,
			values: map[string]interface{}{
				"there": "openebs",
			},
			isErr: true,
		},
	}

	for name, mock := range tests {
		name := name //pin it
		mock := mock //pin it
		t.Run(name, func(t *testing.T) {
			mte, err := NewMetaExecutor(mock.yaml, mock.values)

			if err != nil && !mock.isErr {
				t.Fatalf("failed to test new meta executor: expected 'no error': actual '%s'", err.Error())
			}

			if mte != nil && mte.getMetaInfo().Identity != mock.id {
				t.Fatalf("failed to test new meta executor: expected identity '%s': actual identity '%s'", mock.id, mte.getMetaInfo().Identity)
			}
		})
	}
}

func TestGetRunNamespace(t *testing.T) {
	tests := map[string]struct {
		yaml         string
		values       map[string]interface{}
		runNamespace string
	}{
		"get run namespace - +ve test case - valid yaml & values": {
			// valid yaml that can unmarshall into MetaTask
			yaml: `
runNamespace: {{ .volume.runNamespace }}
id: validid
apiVersion: v1
kind: Service
action: put
`,
			values: map[string]interface{}{
				"volume": map[string]interface{}{
					"runNamespace": "xyz",
				},
			},
			runNamespace: "xyz",
		},
		"get run namespace - -ve test case - valid meta task yaml with invalid templating": {
			// valid meta task yaml with invalid templating
			yaml: `
runNamespace: {{ .volume.namespace }}
id: valididd
apiVersion: v1
kind: Service
action: put
`,
			values: map[string]interface{}{
				"volume": map[string]interface{}{
					"runNamespace": "default",
				},
			},
			runNamespace: "<no value>",
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			mte := &MetaExecutor{
				metaTask: m,
			}

			if mte.getRunNamespace() != mock.runNamespace {
				t.Fatalf("failed to test get run namespace: expected namespace '%s': actual namespace '%s'", mock.runNamespace, mte.getRunNamespace())
			}
		})
	}
}

func TestGetRetry(t *testing.T) {
	tests := map[string]struct {
		yaml             string
		values           map[string]interface{}
		expectedAttempts int
		expectedInterval string
	}{
		"get retry - +ve test case - valid meta task yaml with valid retry value": {
			yaml: `
id: okid
apiVersion: v1
kind: Service
retry: "10,2s"
`,
			values:           map[string]interface{}{},
			expectedAttempts: 10,
			expectedInterval: "2s",
		},
		"get retry - -ve test case - valid meta task yaml with invalid retry interval": {
			yaml: `
id: hiid
apiVersion: v1
kind: Service
retry: "5,2z"
`,
			values:           map[string]interface{}{},
			expectedAttempts: 5,
			expectedInterval: "0s",
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			mte := &MetaExecutor{
				metaTask: m,
			}

			// actuals
			a, i := mte.getRetry()
			// expected
			expectedInterval, _ := time.ParseDuration(mock.expectedInterval)
			if a != mock.expectedAttempts && !reflect.DeepEqual(i, expectedInterval) {
				t.Fatalf("failed to test get retry: expected attempts '%d' interval '%#v': actual attempts '%d' interval '%#v'", mock.expectedAttempts, expectedInterval, a, i)
			}

		})
	}
}

func TestGetObjectName(t *testing.T) {
	tests := map[string]struct {
		yaml       string
		values     map[string]interface{}
		objectName string
	}{
		"get object name - +ve test case - valid meta task yaml with valid object name": {
			yaml: `
id: okid
apiVersion: v1
kind: Deployment
objectName: vol-ctrl
`,
			values:     map[string]interface{}{},
			objectName: "vol-ctrl",
		},
		"get object name - +ve test case - valid meta task yaml with valid templated object name": {
			yaml: `
id: okid
apiVersion: v1
kind: Deployment
objectName: {{ .objectName }}
`,
			values: map[string]interface{}{
				"objectName": "vol-rep",
			},
			objectName: "vol-rep",
		},
	}

	for name, mock := range tests {
		name := name //pin it
		mock := mock //pin it
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			mte := &MetaExecutor{
				metaTask: m,
			}

			if mte.getObjectName() != mock.objectName {
				t.Errorf("failed to get object name: expected object name '%s': actual object name '%s'", mock.objectName, mte.getObjectName())
			}
		})
	}
}

func TestGetListOptions(t *testing.T) {
	tests := map[string]struct {
		yaml          string
		values        map[string]interface{}
		isErr         bool
		labelSelector string
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with valid list options": {
			yaml: `
apiVersion: v1
kind: Pod
action: list
options: |-
  labelSelector: openebs.io/replica=jiva-replica,openebs.io/pv={{ .Volume.owner }}
`,
			values: map[string]interface{}{
				"Volume": map[string]string{
					"owner": "vol-ctrl",
				},
			},
			isErr:         false,
			labelSelector: "openebs.io/replica=jiva-replica,openebs.io/pv=vol-ctrl",
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			mte := &MetaExecutor{
				metaTask: m,
			}

			lo, err := mte.getListOptions()
			if err != nil && !mock.isErr {
				t.Errorf("failed to get list options: expected 'no error': actual '%s'", err.Error())
			}

			if !mock.isErr && lo.LabelSelector != mock.labelSelector {
				t.Errorf("failed to get list options: expected label selector '%s': actual label selector '%s'", mock.labelSelector, lo.LabelSelector)
			}
		})
	}
}

func TestIsList(t *testing.T) {
	tests := map[string]struct {
		yaml   string
		values map[string]interface{}
		isList bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with list action": {
			yaml: `
id: validid
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "list",
			},
			isList: true,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with non list action": {
			yaml: `
id: okid
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "get",
			},
			isList: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			mte := &MetaExecutor{
				metaTask: m,
			}

			if mte.isList() != mock.isList {
				t.Fatalf("failed to is list: expected is list '%t': actual is list '%t'", mock.isList, mte.isList())
			}
		})
	}
}

func TestIsGet(t *testing.T) {
	tests := map[string]struct {
		yaml   string
		values map[string]interface{}
		isGet  bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with get action": {
			yaml: `
id: validid
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "get",
			},
			isGet: true,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with non get action": {
			yaml: `
id: validid
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "patch",
			},
			isGet: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			mte := &MetaExecutor{
				metaTask: m,
			}

			if mte.isGet() != mock.isGet {
				t.Fatalf("failed to is get: expected is get '%t': actual is get '%t'", mock.isGet, mte.isGet())
			}
		})
	}
}

func TestIsPut(t *testing.T) {
	tests := map[string]struct {
		yaml   string
		values map[string]interface{}
		isPut  bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with put action": {
			yaml: `
id: validid
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "put",
			},
			isPut: true,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with non put action": {
			yaml: `
id: validid
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "list",
			},
			isPut: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			mte := &MetaExecutor{
				metaTask: m,
			}

			if mte.isPut() != mock.isPut {
				t.Fatalf("failed to is put: expected is put '%t': actual is put '%t'", mock.isPut, mte.isPut())
			}
		})
	}
}

func TestIsDelete(t *testing.T) {
	tests := map[string]struct {
		yaml     string
		values   map[string]interface{}
		isDelete bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with delete action": {
			yaml: `
id: validid
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "delete",
			},
			isDelete: true,
		},
		//
		// start of test
		//
		"Negative test  - valid meta task yaml with non delete action": {
			yaml: `
id: validid
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "list",
			},
			isDelete: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			mte := &MetaExecutor{
				metaTask: m,
			}

			if mte.isDelete() != mock.isDelete {
				t.Fatalf("failed to is delete: expected is delete '%t': actual is delete '%t'", mock.isDelete, mte.isDelete())
			}
		})
	}
}

func TestIsPatch(t *testing.T) {
	tests := map[string]struct {
		yaml    string
		values  map[string]interface{}
		isPatch bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with patch action": {
			yaml: `
id: validid
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "patch",
			},
			isPatch: true,
		},
		//
		// start of test
		//
		"Negative test case - valid meta task yaml with non patch action": {
			yaml: `
id: validid
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "list",
			},
			isPatch: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			mte := &MetaExecutor{
				metaTask: m,
			}

			if mte.isPatch() != mock.isPatch {
				t.Fatalf("failed to is patch: expected is patch '%t': actual is patch '%t'", mock.isPatch, mte.isPatch())
			}
		})
	}
}

func TestIsPutExtnV1B1Deploy(t *testing.T) {
	tests := map[string]struct {
		yaml                string
		values              map[string]interface{}
		isPutExtnV1B1Deploy bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with put extensions/v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "put",
			},
			isPutExtnV1B1Deploy: true,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with put apps/v1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1",
				"action":     "put",
			},
			isPutExtnV1B1Deploy: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with put extensions/v1beta1 pod": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "put",
			},
			isPutExtnV1B1Deploy: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with non put extensions/v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "get",
			},
			isPutExtnV1B1Deploy: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			i, _ := newTaskIdentifier(m.MetaTaskIdentity)

			mte := &MetaExecutor{
				metaTask:   m,
				identifier: i,
			}

			if mte.isPutExtnV1B1Deploy() != mock.isPutExtnV1B1Deploy {
				t.Fatalf("failed to is put extn v1beta1 deploy: expected '%t': actual '%t': actual meta task '%+v'", mock.isPutExtnV1B1Deploy, mte.isPutExtnV1B1Deploy(), mte.metaTask)
			}
		})
	}
}

func TestIsPatchExtnV1B1Deploy(t *testing.T) {
	tests := map[string]struct {
		yaml                  string
		values                map[string]interface{}
		isPatchExtnV1B1Deploy bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with patch extensions/v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "patch",
			},
			isPatchExtnV1B1Deploy: true,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with patch apps/v1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1",
				"action":     "patch",
			},
			isPatchExtnV1B1Deploy: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with patch extensions/v1beta1 pod": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "patch",
			},
			isPatchExtnV1B1Deploy: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with non patch extensions/v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "get",
			},
			isPatchExtnV1B1Deploy: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			i, _ := newTaskIdentifier(m.MetaTaskIdentity)

			mte := &MetaExecutor{
				metaTask:   m,
				identifier: i,
			}

			if mte.isPatchExtnV1B1Deploy() != mock.isPatchExtnV1B1Deploy {
				t.Fatalf("failed to is patch extn v1beta1 deploy: expected '%t': actual '%t'", mock.isPatchExtnV1B1Deploy, mte.isPatchExtnV1B1Deploy())
			}
		})
	}
}

func TestIsPutAppsV1B1Deploy(t *testing.T) {
	tests := map[string]struct {
		yaml                string
		values              map[string]interface{}
		isPutAppsV1B1Deploy bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with put apps/v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "put",
			},
			isPutAppsV1B1Deploy: true,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with put apps/v1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1",
				"action":     "put",
			},
			isPutAppsV1B1Deploy: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with put apps/v1beta1 pod": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "put",
			},
			isPutAppsV1B1Deploy: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with non put apps/v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "get",
			},
			isPutAppsV1B1Deploy: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			i, _ := newTaskIdentifier(m.MetaTaskIdentity)

			mte := &MetaExecutor{
				metaTask:   m,
				identifier: i,
			}

			if mte.isPutAppsV1B1Deploy() != mock.isPutAppsV1B1Deploy {
				t.Fatalf("failed to is put apps v1beta1 deploy: expected '%t': actual '%t'", mock.isPutAppsV1B1Deploy, mte.isPutAppsV1B1Deploy())
			}
		})
	}
}

func TestIsPatchAppsV1B1Deploy(t *testing.T) {
	tests := map[string]struct {
		id                    string
		yaml                  string
		values                map[string]interface{}
		isErr                 bool
		isPatchAppsV1B1Deploy bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with patch apps/v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "patch",
			},
			isPatchAppsV1B1Deploy: true,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with patch apps/v1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1",
				"action":     "patch",
			},
			isPatchAppsV1B1Deploy: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with patch apps/v1beta1 pod": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "patch",
			},
			isPatchAppsV1B1Deploy: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with non patch apps/v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "get",
			},
			isPatchAppsV1B1Deploy: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			i, _ := newTaskIdentifier(m.MetaTaskIdentity)

			mte := &MetaExecutor{
				metaTask:   m,
				identifier: i,
			}

			if mte.isPatchAppsV1B1Deploy() != mock.isPatchAppsV1B1Deploy {
				t.Fatalf("failed to is patch apps v1beta1 deploy: expected '%t': actual '%t'", mock.isPatchAppsV1B1Deploy, mte.isPatchAppsV1B1Deploy())
			}
		})
	}
}

func TestIsPutCoreV1Service(t *testing.T) {
	tests := map[string]struct {
		id                 string
		yaml               string
		values             map[string]interface{}
		isErr              bool
		isPutCoreV1Service bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with put v1 service": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Service
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "put",
			},
			isPutCoreV1Service: true,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with put v2 service": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Service
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v2",
				"action":     "put",
			},
			isPutCoreV1Service: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with put v1 pod": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "put",
			},
			isPutCoreV1Service: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with non put v1 service": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Service
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "get",
			},
			isPutCoreV1Service: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			i, _ := newTaskIdentifier(m.MetaTaskIdentity)

			mte := &MetaExecutor{
				metaTask:   m,
				identifier: i,
			}

			if mte.isPutCoreV1Service() != mock.isPutCoreV1Service {
				t.Fatalf("failed to is put v1 service: expected '%t': actual '%t'", mock.isPutCoreV1Service, mte.isPutCoreV1Service())
			}
		})
	}
}

func TestIsDeleteExtnV1B1Deploy(t *testing.T) {
	tests := map[string]struct {
		id                     string
		yaml                   string
		values                 map[string]interface{}
		isErr                  bool
		isDeleteExtnV1B1Deploy bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with delete extensions/v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "delete",
			},
			isDeleteExtnV1B1Deploy: true,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with non delete extensions/v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "put",
			},
			isDeleteExtnV1B1Deploy: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with delete extensions/v1beta1 pod": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "delete",
			},
			isDeleteExtnV1B1Deploy: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with delete v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1beta1",
				"action":     "delete",
			},
			isDeleteExtnV1B1Deploy: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			i, _ := newTaskIdentifier(m.MetaTaskIdentity)

			mte := &MetaExecutor{
				metaTask:   m,
				identifier: i,
			}

			if mte.isDeleteExtnV1B1Deploy() != mock.isDeleteExtnV1B1Deploy {
				t.Fatalf("failed to is delete extensions v1beta1 deploy: expected '%t': actual '%t'", mock.isDeleteExtnV1B1Deploy, mte.isDeleteExtnV1B1Deploy())
			}
		})
	}
}

func TestIsDeleteAppsV1B1Deploy(t *testing.T) {
	tests := map[string]struct {
		id                     string
		yaml                   string
		values                 map[string]interface{}
		isErr                  bool
		isDeleteAppsV1B1Deploy bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with delete apps/v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "delete",
			},
			isDeleteAppsV1B1Deploy: true,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with delete v2 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v2",
				"action":     "delete",
			},
			isDeleteAppsV1B1Deploy: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with delete apps/v1beta1 pod": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "delete",
			},
			isDeleteAppsV1B1Deploy: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with non delete apps/v1beta1 deploy": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "get",
			},
			isDeleteAppsV1B1Deploy: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			i, _ := newTaskIdentifier(m.MetaTaskIdentity)

			mte := &MetaExecutor{
				metaTask:   m,
				identifier: i,
			}

			if mte.isDeleteAppsV1B1Deploy() != mock.isDeleteAppsV1B1Deploy {
				t.Fatalf("failed to is delete apps v1beta1 deploy: expected '%t': actual '%t'", mock.isDeleteAppsV1B1Deploy, mte.isDeleteAppsV1B1Deploy())
			}
		})
	}
}

func TestIsDeleteCoreV1Service(t *testing.T) {
	tests := map[string]struct {
		id                    string
		yaml                  string
		values                map[string]interface{}
		isErr                 bool
		isDeleteCoreV1Service bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with delete v1 service": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Service
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "delete",
			},
			isDeleteCoreV1Service: true,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with delete v2 service": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Service
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v2",
				"action":     "delete",
			},
			isDeleteCoreV1Service: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with delete v1 pod": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "delete",
			},
			isDeleteCoreV1Service: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with non delete v1 service": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "get",
			},
			isDeleteCoreV1Service: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			i, _ := newTaskIdentifier(m.MetaTaskIdentity)

			mte := &MetaExecutor{
				metaTask:   m,
				identifier: i,
			}

			if mte.isDeleteCoreV1Service() != mock.isDeleteCoreV1Service {
				t.Fatalf("failed to is delete core v1 service: expected '%t': actual '%t'", mock.isDeleteCoreV1Service, mte.isDeleteCoreV1Service())
			}
		})
	}
}

func TestIsListCoreV1Pod(t *testing.T) {
	tests := map[string]struct {
		id              string
		yaml            string
		values          map[string]interface{}
		isErr           bool
		isListCoreV1Pod bool
	}{
		//
		// start of test
		//
		"Positive test - valid meta task yaml with list v1 pod": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "list",
			},
			isListCoreV1Pod: true,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with list v2 pod": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v2",
				"action":     "list",
			},
			isListCoreV1Pod: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with list v1 service": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Service
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "list",
			},
			isListCoreV1Pod: false,
		},
		//
		// start of test
		//
		"Negative test - valid meta task yaml with non list v1 pod": {
			yaml: `
id: validid
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "get",
			},
			isListCoreV1Pod: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// transform the yaml with provided values
			b, _ := template.AsTemplatedBytes("MetaTaskSpec", mock.yaml, mock.values)

			// unmarshall the yaml bytes into this instance
			var m MetaTaskSpec
			yaml.Unmarshal(b, &m)

			i, _ := newTaskIdentifier(m.MetaTaskIdentity)

			mte := &MetaExecutor{
				metaTask:   m,
				identifier: i,
			}

			if mte.isListCoreV1Pod() != mock.isListCoreV1Pod {
				t.Fatalf("failed to is list core v1 pod: expected '%t': actual '%t'", mock.isListCoreV1Pod, mte.isListCoreV1Pod())
			}
		})
	}
}
