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
)

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
		"new meta task - +ve test case - minimal meta task yaml & empty values": {
			id: "121",
			// valid yaml that can unmarshall into MetaTask
			yaml:   "kind: Pod\napiVersion: v1",
			values: map[string]interface{}{},
			isErr:  false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)

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
		id           string
		yaml         string
		values       map[string]interface{}
		runNamespace string
		isErr        bool
	}{
		"get run namespace - +ve test case - valid id, yaml & values": {
			id: "121",
			// valid yaml that can unmarshall into MetaTask
			yaml: `
runNamespace: {{ .volume.runNamespace }}
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
			isErr:        false,
		},
		"get run namespace - -ve test case - valid meta task yaml with invalid templating": {
			id: "121",
			// valid meta task yaml with invalid templating
			yaml: `
runNamespace: {{ .volume.namespace }}
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
			isErr:        false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)

			if err != nil && !mock.isErr {
				t.Fatalf("failed to test get run namespace: expected 'no error': actual '%s'", err.Error())
			}

			if mte != nil && mte.getRunNamespace() != mock.runNamespace {
				t.Fatalf("failed to test get run namespace: expected namespace '%s': actual namespace '%s'", mock.runNamespace, mte.getRunNamespace())
			}
		})
	}
}

func TestGetRetry(t *testing.T) {
	tests := map[string]struct {
		id               string
		yaml             string
		values           map[string]interface{}
		expectedAttempts int
		expectedInterval string
		isErr            bool
	}{
		"get retry - +ve test case - valid meta task yaml with valid retry value": {
			id: "121",
			yaml: `
apiVersion: v1
kind: Service
retry: "10,2s"
`,
			values:           map[string]interface{}{},
			expectedAttempts: 10,
			expectedInterval: "2s",
			isErr:            false,
		},
		"get retry - -ve test case - valid meta task yaml with invalid retry interval": {
			id: "121",
			yaml: `
apiVersion: v1
kind: Service
retry: "5,2z"
`,
			values:           map[string]interface{}{},
			expectedAttempts: 5,
			expectedInterval: "0s",
			isErr:            false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)

			if err != nil && !mock.isErr {
				t.Fatalf("failed to test get retry: expected 'no error': actual '%s'", err.Error())
			}

			if mte != nil {
				// actuals
				a, i := mte.getRetry()
				// expected
				expectedInterval, _ := time.ParseDuration(mock.expectedInterval)
				if a != mock.expectedAttempts && !reflect.DeepEqual(i, expectedInterval) {
					t.Fatalf("failed to test get retry: expected attempts '%d' interval '%#v': actual attempts '%d' interval '%#v'", mock.expectedAttempts, expectedInterval, a, i)
				}
			}
		})
	}
}

func TestGetTaskResultQueries(t *testing.T) {
	tests := map[string]struct {
		id                   string
		yaml                 string
		values               map[string]interface{}
		isErr                bool
		taskResultQueryCount int
	}{
		"get task result queries - +ve test case - valid meta task yaml with valid queries": {
			id: "test",
			yaml: `
apiVersion: v1
kind: PersistentVolumeClaim
queries:
- alias: objectName
- alias: affinity
  path: |-
    {.metadata.annotations.controller\.openebs\.io/affinity}
- alias: affinityTopology
  path: |-
    {.metadata.annotations.controller\.openebs\.io/affinity-topology}
- alias: affinityType
  path: |-
    {.metadata.annotations.controller\.openebs\.io/affinity-type}
`,
			values:               map[string]interface{}{},
			isErr:                false,
			taskResultQueryCount: 4,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to get task result queries: expected 'no error': actual '%s'", err.Error())
			}

			if len(mte.getTaskResultQueries()) != mock.taskResultQueryCount {
				t.Fatalf("failed to get task result queries: expected task result query count '%d': actual task result query count '%d'", mock.taskResultQueryCount, len(mte.getTaskResultQueries()))
			}
		})
	}
}

func TestGetObjectName(t *testing.T) {
	tests := map[string]struct {
		id         string
		yaml       string
		values     map[string]interface{}
		objectName string
		isErr      bool
	}{
		"get object name - +ve test case - valid meta task yaml with valid object name": {
			id: "test",
			yaml: `
apiVersion: v1
kind: Deployment
objectName: vol-ctrl
`,
			values:     map[string]interface{}{},
			objectName: "vol-ctrl",
			isErr:      false,
		},
		"get object name - +ve test case - valid meta task yaml with valid templated object name": {
			id: "test",
			yaml: `
apiVersion: v1
kind: Deployment
objectName: {{ .objectName }}
`,
			values: map[string]interface{}{
				"objectName": "vol-rep",
			},
			objectName: "vol-rep",
			isErr:      false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Errorf("failed to get object name: expected 'no error': actual '%s'", err.Error())
			}

			if mte.getObjectName() != mock.objectName && !mock.isErr {
				t.Errorf("failed to get object name: expected object name '%s': actual object name '%s'", mock.objectName, mte.getObjectName())
			}
		})
	}
}

func TestGetListOptions(t *testing.T) {
	tests := map[string]struct {
		id            string
		yaml          string
		values        map[string]interface{}
		isErr         bool
		labelSelector string
	}{
		"get list options - +ve test case - valid meta task yaml with valid list options": {
			id: "test",
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
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Errorf("failed to get list options: expected 'no error': actual '%s'", err.Error())
			}

			lo, err := mte.getListOptions()
			if err != nil && !mock.isErr {
				t.Errorf("failed to get list options: expected list options: actual '%s'", err.Error())
			}

			if !mock.isErr && lo.LabelSelector != mock.labelSelector {
				t.Errorf("failed to get list options: expected label selector '%s': actual label selector '%s'", mock.labelSelector, lo.LabelSelector)
			}
		})
	}
}

func TestIsList(t *testing.T) {
	tests := map[string]struct {
		id     string
		yaml   string
		values map[string]interface{}
		isList bool
		isErr  bool
	}{
		"is list - +vs test case - valid meta task yaml with list action": {
			id: "test",
			yaml: `
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "list",
			},
			isList: true,
			isErr:  false,
		},
		"is list - -ve test case - valid meta task yaml with get action": {
			id: "test",
			yaml: `
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				// action is not list
				"action": "get",
			},
			// false as action is not list
			isList: false,
			isErr:  false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is list: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isList()
			if !mock.isErr && is != mock.isList {
				t.Fatalf("failed to is list: expected is list '%t': actual is list '%t'", mock.isList, is)
			}
		})
	}
}

func TestIsGet(t *testing.T) {
	tests := map[string]struct {
		id     string
		yaml   string
		values map[string]interface{}
		isErr  bool
		isGet  bool
	}{
		"is get - +ve test case - valid meta task yaml with get action": {
			id: "test",
			yaml: `
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "get",
			},
			isGet: true,
			isErr: false,
		},
		"is get - -ve test case - valid meta task yaml with patch action": {
			id: "test",
			yaml: `
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				// action is not get
				"action": "patch",
			},
			// false as action is not get
			isGet: false,
			isErr: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is get: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isGet()
			if !mock.isErr && is != mock.isGet {
				t.Fatalf("failed to is get: expected is get '%t': actual is get '%t'", mock.isGet, is)
			}
		})
	}
}

func TestIsPut(t *testing.T) {
	tests := map[string]struct {
		id     string
		yaml   string
		values map[string]interface{}
		isErr  bool
		isPut  bool
	}{
		"is put - +ve test case - valid meta task yaml with put action": {
			id: "test",
			yaml: `
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "put",
			},
			isPut: true,
			isErr: false,
		},
		"is put - -ve test case - valid meta task yaml with list action": {
			id: "test",
			yaml: `
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				// action is not put
				"action": "list",
			},
			// false as action is not put
			isPut: false,
			isErr: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is put: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isPut()
			if !mock.isErr && is != mock.isPut {
				t.Fatalf("failed to is put: expected is put '%t': actual is put '%t'", mock.isPut, is)
			}
		})
	}
}

func TestIsDelete(t *testing.T) {
	tests := map[string]struct {
		id       string
		yaml     string
		values   map[string]interface{}
		isErr    bool
		isDelete bool
	}{
		"is delete - +ve test case - valid meta task yaml with delete action": {
			id: "test",
			yaml: `
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "delete",
			},
			isDelete: true,
			isErr:    false,
		},
		"is delete - -ve test case - valid meta task yaml with list action": {
			id: "test",
			yaml: `
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				// action is not delete
				"action": "list",
			},
			// false as action is not delete
			isDelete: false,
			isErr:    false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is delete: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isDelete()
			if !mock.isErr && is != mock.isDelete {
				t.Fatalf("failed to is delete: expected is delete '%t': actual is delete '%t'", mock.isDelete, is)
			}
		})
	}
}

func TestIsPatch(t *testing.T) {
	tests := map[string]struct {
		id      string
		yaml    string
		values  map[string]interface{}
		isErr   bool
		isPatch bool
	}{
		"is patch - +ve test case - valid meta task yaml with patch action": {
			id: "test",
			yaml: `
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"action": "patch",
			},
			isPatch: true,
			isErr:   false,
		},
		"is patch - -ve test case - valid meta task yaml with list action": {
			id: "test",
			yaml: `
apiVersion: v1
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				// action is not patch
				"action": "list",
			},
			// false as action is not patch
			isPatch: false,
			isErr:   false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is patch: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isPatch()
			if !mock.isErr && is != mock.isPatch {
				t.Fatalf("failed to is patch: expected is patch '%t': actual is patch '%t'", mock.isPatch, is)
			}
		})
	}
}

func TestIsPutExtnV1B1Deploy(t *testing.T) {
	tests := map[string]struct {
		id                  string
		yaml                string
		values              map[string]interface{}
		isErr               bool
		isPutExtnV1B1Deploy bool
	}{
		"is put extn v1beta1 deploy - +ve test case - valid meta task yaml with put action & extn/v1beta1 api version & deploy resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "put",
			},
			isErr:               false,
			isPutExtnV1B1Deploy: true,
		},
		"is put extn v1beta1 deploy - -ve test case - valid meta task yaml with put action & apps/v1 api version & deploy resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				// api version is not extensions/v1beta1
				"apiVersion": "apps/v1",
				"action":     "put",
			},
			isErr: false,
			// false as wrong api version is provided
			isPutExtnV1B1Deploy: false,
		},
		"is put extn v1beta1 deploy - -ve test case - valid meta task yaml with put action & extn/v1beta1 api version & pod resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "put",
			},
			isErr: false,
			// false as the kind is not a Deployment
			isPutExtnV1B1Deploy: false,
		},
		"is put extn v1beta1 deploy - -ve test case - valid meta task yaml with get action & extn/v1beta1 api version & deploy resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "get",
			},
			isErr: false,
			// false as action is not put
			isPutExtnV1B1Deploy: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is put extn v1beta1 deploy: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isPutExtnV1B1Deploy()
			if !mock.isErr && is != mock.isPutExtnV1B1Deploy {
				t.Fatalf("failed to is put extn v1beta1 deploy: expected '%t': actual '%t'", mock.isPutExtnV1B1Deploy, is)
			}
		})
	}
}

func TestIsPatchExtnV1B1Deploy(t *testing.T) {
	tests := map[string]struct {
		id                    string
		yaml                  string
		values                map[string]interface{}
		isErr                 bool
		isPatchExtnV1B1Deploy bool
	}{
		"is patch extn v1beta1 deploy - +ve test case - valid meta task yaml with patch action & extn/v1beta1 api version & deploy resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "patch",
			},
			isErr: false,
			isPatchExtnV1B1Deploy: true,
		},
		"is patch extn v1beta1 deploy - -ve test case - valid meta task yaml with patch action & apps/v1 api version & deploy resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				// api version is not extensions/v1beta1
				"apiVersion": "apps/v1",
				"action":     "patch",
			},
			isErr: false,
			// false as wrong api version is provided
			isPatchExtnV1B1Deploy: false,
		},
		"is patch extn v1beta1 deploy - -ve test case - valid meta task yaml with patch action & extn/v1beta1 api version & pod resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "patch",
			},
			isErr: false,
			// false as the kind is not a Deployment
			isPatchExtnV1B1Deploy: false,
		},
		"is patch extn v1beta1 deploy - -ve test case - valid meta task yaml with get action & extn/v1beta1 api version & deploy resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "get",
			},
			isErr: false,
			// false as action is not put
			isPatchExtnV1B1Deploy: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is patch extn v1beta1 deploy: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isPatchExtnV1B1Deploy()
			if !mock.isErr && is != mock.isPatchExtnV1B1Deploy {
				t.Fatalf("failed to is patch extn v1beta1 deploy: expected '%t': actual '%t'", mock.isPatchExtnV1B1Deploy, is)
			}
		})
	}
}

func TestIsPutAppsV1B1Deploy(t *testing.T) {
	tests := map[string]struct {
		id                  string
		yaml                string
		values              map[string]interface{}
		isErr               bool
		isPutAppsV1B1Deploy bool
	}{
		"is put apps v1beta1 deploy - +ve test case - valid meta task yaml with put action & apps/v1beta1 api version & deploy resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "put",
			},
			isErr:               false,
			isPutAppsV1B1Deploy: true,
		},
		"is put apps v1beta1 deploy - -ve test case - valid meta task yaml with put action & apps/v1 api version & deploy resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				// api version is not extensions/v1beta1
				"apiVersion": "apps/v1",
				"action":     "put",
			},
			isErr: false,
			// false as wrong api version is provided
			isPutAppsV1B1Deploy: false,
		},
		"is put apps v1beta1 deploy - -ve test case - valid meta task yaml with put action & apps/v1beta1 api version & pod resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "put",
			},
			isErr: false,
			// false as the kind is not a Deployment
			isPutAppsV1B1Deploy: false,
		},
		"is put apps v1beta1 deploy - -ve test case - valid meta task yaml with get action & apps/v1beta1 api version & deploy resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "get",
			},
			isErr: false,
			// false as action is not put
			isPutAppsV1B1Deploy: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is put apps v1beta1 deploy: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isPutAppsV1B1Deploy()
			if !mock.isErr && is != mock.isPutAppsV1B1Deploy {
				t.Fatalf("failed to is put apps v1beta1 deploy: expected '%t': actual '%t'", mock.isPutAppsV1B1Deploy, is)
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
		"is patch apps v1beta1 deploy - +ve test case - valid meta task yaml with patch action & apps/v1beta1 api version & deploy resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "patch",
			},
			isErr: false,
			isPatchAppsV1B1Deploy: true,
		},
		"is patch apps v1beta1 deploy - -ve test case - valid meta task yaml with patch action & apps/v1 api version & deploy resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				// api version is not extensions/v1beta1
				"apiVersion": "apps/v1",
				"action":     "patch",
			},
			isErr: false,
			// false as wrong api version is provided
			isPatchAppsV1B1Deploy: false,
		},
		"is patch apps v1beta1 deploy - -ve test case - valid meta task yaml with patch action & apps/v1beta1 api version & pod resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "patch",
			},
			isErr: false,
			// false as the kind is not a Deployment
			isPatchAppsV1B1Deploy: false,
		},
		"is patch apps v1beta1 deploy - -ve test case - valid meta task yaml with get action & apps/v1beta1 api version & deploy resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				// wrong action
				"action": "get",
			},
			isErr: false,
			// false as action is not patch
			isPatchAppsV1B1Deploy: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is patch apps v1beta1 deploy: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isPatchAppsV1B1Deploy()
			if !mock.isErr && is != mock.isPatchAppsV1B1Deploy {
				t.Fatalf("failed to is patch apps v1beta1 deploy: expected '%t': actual '%t'", mock.isPatchAppsV1B1Deploy, is)
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
		"is put core v1 service - +ve test case - valid meta task yaml with 'put' action & 'v1' api version & 'Service' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Service
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "put",
			},
			isErr:              false,
			isPutCoreV1Service: true,
		},
		"is put core v1 service - -ve test case - valid meta task yaml with 'put' action & 'v2' api version & 'Service' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Service
action: {{ .action }}
`,
			values: map[string]interface{}{
				// api version is not right in this context
				"apiVersion": "v2",
				"action":     "put",
			},
			isErr: false,
			// false due to api version
			isPutCoreV1Service: false,
		},
		"is put v1 service - -ve test case - valid meta task yaml with 'put' action & 'v1' api version & 'Pod' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "put",
			},
			isErr: false,
			// false as the kind is not correct in this context
			isPutCoreV1Service: false,
		},
		"is put v1 service - -ve test case - valid meta task yaml with 'get' action & 'v1' api version & 'Service' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Service
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				// action is not correct in this context
				"action": "get",
			},
			isErr: false,
			// false as action is not correct in this context
			isPutCoreV1Service: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is put v1 service: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isPutCoreV1Service()
			if !mock.isErr && is != mock.isPutCoreV1Service {
				t.Fatalf("failed to is put v1 service: expected '%t': actual '%t'", mock.isPutCoreV1Service, is)
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
		"is delete extensions v1beta1 deploy - +ve test case - valid meta task yaml with 'delete' action & 'extensions/v1beta1' api version & 'Deploy' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "delete",
			},
			isErr: false,
			isDeleteExtnV1B1Deploy: true,
		},
		"is delete extensions v1beta1 deploy - -ve test case - valid meta task yaml with 'delete' action & 'v2' api version & 'Deployment' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				// api version is not right in this context
				"apiVersion": "v2",
				"action":     "delete",
			},
			isErr: false,
			// false due to api version
			isDeleteExtnV1B1Deploy: false,
		},
		"is delete extensions v1beta1 deploy - -ve test case - valid meta task yaml with 'delete' action & 'extensions/v1beta1' api version & 'Pod' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"action":     "delete",
			},
			isErr: false,
			// false as the kind is not correct in this context
			isDeleteExtnV1B1Deploy: false,
		},
		"is delete extensions v1beta1 deploy - -ve test case - valid meta task yaml with 'delete' action & 'v1beta1' api version & 'Deployment' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1beta1",
				// action is not correct in this context
				"action": "delete",
			},
			isErr: false,
			// false as action is not correct in this context
			isDeleteExtnV1B1Deploy: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is delete extensions v1beta1 deploy: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isDeleteExtnV1B1Deploy()
			if !mock.isErr && is != mock.isDeleteExtnV1B1Deploy {
				t.Fatalf("failed to is delete extensions v1beta1 deploy: expected '%t': actual '%t'", mock.isDeleteExtnV1B1Deploy, is)
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
		"is delete apps v1beta1 deploy - +ve test case - valid meta task yaml with 'delete' action & 'apps/v1beta1' api version & 'Deployment' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "delete",
			},
			isErr: false,
			isDeleteAppsV1B1Deploy: true,
		},
		"is delete apps v1beta1 deploy - -ve test case - valid meta task yaml with 'delete' action & 'v2' api version & 'Deployment' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				// api version is not right in this context
				"apiVersion": "v2",
				"action":     "delete",
			},
			isErr: false,
			// false due to api version
			isDeleteAppsV1B1Deploy: false,
		},
		"is delete apps v1beta1 deploy - -ve test case - valid meta task yaml with 'delete' action & 'apps/v1beta1' api version & 'Pod' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				"action":     "delete",
			},
			isErr: false,
			// false as the kind is not correct in this context
			isDeleteAppsV1B1Deploy: false,
		},
		"is delete apps v1beta1 deploy - -ve test case - valid meta task yaml with 'get' action & 'apps/v1beta1' api version & 'Deployment' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "apps/v1beta1",
				// action is not correct in this context
				"action": "get",
			},
			isErr: false,
			// false as action is not correct in this context
			isDeleteAppsV1B1Deploy: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is delete apps v1beta1 deploy: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isDeleteAppsV1B1Deploy()
			if !mock.isErr && is != mock.isDeleteAppsV1B1Deploy {
				t.Fatalf("failed to is delete apps v1beta1 deploy: expected '%t': actual '%t'", mock.isDeleteAppsV1B1Deploy, is)
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
		"is delete core v1 service - +ve test case - valid meta task yaml with 'delete' action & 'v1' api version & 'Service' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Service
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "delete",
			},
			isErr: false,
			isDeleteCoreV1Service: true,
		},
		"is delete v1 service - -ve test case - valid meta task yaml with 'delete' action & 'v2' api version & 'Service' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Service
action: {{ .action }}
`,
			values: map[string]interface{}{
				// api version is not right in this context
				"apiVersion": "v2",
				"action":     "delete",
			},
			isErr: false,
			// false due to api version
			isDeleteCoreV1Service: false,
		},
		"is delete v1 service - -ve test case - valid meta task yaml with 'delete' action & 'v1' api version & 'Pod' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "delete",
			},
			isErr: false,
			// false as the kind is not correct in this context
			isDeleteCoreV1Service: false,
		},
		"is delete v1 service - -ve test case - valid meta task yaml with 'get' action & 'v1' api version & 'Service' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Deployment
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				// action is not correct in this context
				"action": "get",
			},
			isErr: false,
			// false as action is not correct in this context
			isDeleteCoreV1Service: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is delete core v1 service: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isDeleteCoreV1Service()
			if !mock.isErr && is != mock.isDeleteCoreV1Service {
				t.Fatalf("failed to is delete core v1 service: expected '%t': actual '%t'", mock.isDeleteCoreV1Service, is)
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
		"is list core v1 pod - +ve test case - valid meta task yaml with 'list' action & 'v1' api version & 'Pod' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "list",
			},
			isErr:           false,
			isListCoreV1Pod: true,
		},
		"is list core v1 pod - -ve test case - valid meta task yaml with 'list' action & 'v2' api version & 'Pod' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				// api version is not right in this context
				"apiVersion": "v2",
				"action":     "list",
			},
			isErr: false,
			// false due to api version
			isListCoreV1Pod: false,
		},
		"is list core v1 pod - -ve test case - valid meta task yaml with 'list' action & 'v1' api version & 'Service' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Service
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				"action":     "list",
			},
			isErr: false,
			// false as the kind is not correct in this context
			isListCoreV1Pod: false,
		},
		"is list core v1 pod - -ve test case - valid meta task yaml with 'get' action & 'v1' api version & 'Pod' resource": {
			id: "test",
			yaml: `
apiVersion: {{ .apiVersion }}
kind: Pod
action: {{ .action }}
`,
			values: map[string]interface{}{
				"apiVersion": "v1",
				// action is not correct in this context
				"action": "get",
			},
			isErr: false,
			// false as action is not correct in this context
			isListCoreV1Pod: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to is list core v1 pod: expected 'no error': actual '%s'", err.Error())
			}

			is := mte.isListCoreV1Pod()
			if !mock.isErr && is != mock.isListCoreV1Pod {
				t.Fatalf("failed to is list core v1 pod: expected '%t': actual '%t'", mock.isListCoreV1Pod, is)
			}
		})
	}
}

// TODO
func TestIsGetOEV1alpha1SP(t *testing.T) {}

// TODO
func TestIsGetCoreV1PVC(t *testing.T) {}

func TestAsRollbackInstance(t *testing.T) {
	tests := map[string]struct {
		id             string
		yaml           string
		values         map[string]interface{}
		isErr          bool
		isRollback     bool
		rollbackAction TaskAction
	}{
		"as rollback instance - +ve test case - valid meta task with put action": {
			id: "testid",
			yaml: `
apiVersion: v1
kind: Pod
action: put
`,
			values:     map[string]interface{}{},
			isErr:      false,
			isRollback: true,
			// when original action is put its rollback is delete
			rollbackAction: DeleteTA,
		},
		"as rollback instance - -ve test case - valid meta task with delete action": {
			id: "testid",
			yaml: `
apiVersion: v1
kind: Pod
action: delete
`,
			values: map[string]interface{}{},
			isErr:  false,
			// delete action cannot be rolled back
			isRollback: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mte, err := newMetaTaskExecutor(mock.id, mock.yaml, mock.values)
			if err != nil && !mock.isErr {
				t.Fatalf("failed to rollback task instance: expected 'no error': actual '%s'", err.Error())
			}

			rmte, isRollback, err := mte.asRollbackInstance("testing")
			if !mock.isErr && err != nil {
				t.Fatalf("failed to rollback task instance: expected 'no rollback error': actual '%s'", err.Error())
			}

			if !mock.isErr && mock.isRollback != isRollback {
				t.Fatalf("failed to rollback task instance: expected rollback '%t': actual rollback '%t'", mock.isRollback, isRollback)
			}

			if !mock.isErr && isRollback && mock.rollbackAction != rmte.getMetaInfo().Action {
				t.Fatalf("failed to rollback task instance: expected rollback action '%s': actual rollback action '%s'", mock.rollbackAction, rmte.getMetaInfo().Action)
			}
		})
	}
}
