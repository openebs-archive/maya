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
	
	"github.com/openebs/maya/pkg/template"
)

// resultExecuteMock is the mock structure to test task result
// templating
type resultExecuteMock struct {
	// taskID is the task ID
	taskID string
	// bytes hold the data that will be fed while executing the go template
	bytes []byte
	// alias is the key against which the value is set,
	// value is the one that is derived after running jsonpath
	alias string
	// path represents the go template function
	path string
	// expected is the resulting value i.e. after
	// executing path against the bytes
	expected string
}

func TestResultExecute(t *testing.T) {
  // this yml should not interfere with the json query to be done later
  var jsonPathFeederYml = `
data:
  meta: |
    runNamespace: {{ .Volume.runNamespace }}
    apiVersion: openebs.io/v1alpha1
    kind: PersistentVolumeClaim
    objectName: {{ .Volume.pvc }}
    action: get
    queries:
    - alias: affinity
      path: "{.metadata.annotations.controller\.openebs\.io/affinity}"
    - alias: affinityTopology
      path: "{.metadata.annotations.controller\.openebs\.io/affinity-topology}"
`

	var myPodJson = []byte(`{
  "kind": "Pod",
  "apiVersion": "v1",
  "metadata": {
    "name": "kubectl-tester",
    "annotations": {
      "simple": "value",
      "controller.openebs.io/affinity": "mypin",
      "controller.openebs.io/affinity-topology": "kubernetes.io/hostname"
    }
  },
  "spec": {
    "containers": [
      {
        "name": "bb",
        "image": "k8s.gcr.io/busybox",
        "command": [
          "sh", "-c", "sleep 5; wget -O - ${KUBERNETES_RO_SERVICE_HOST}:${KUBERNETES_RO_SERVICE_PORT}/api/v1/pods/; sleep 10000"
        ],
        "ports": [
          {
            "containerPort": 8080
          }
        ],
        "env": [
          {
            "name": "KUBERNETES_RO_SERVICE_HOST",
            "value": "127.0.0.1"
          },
          {
            "name": "KUBERNETES_RO_SERVICE_PORT",
            "value": "8001"
          }
        ],
        "volumeMounts": [
          {
            "name": "test-volume",
            "mountPath": "/mount/test-volume"
          }
        ]
      },
      {
        "name": "kubectl",
        "image": "k8s.gcr.io/kubectl:v0.18.0-120-gaeb4ac55ad12b1-dirty",
        "imagePullPolicy": "Always",
        "args": [
          "proxy", "-p", "8001"
        ]
      }
    ],
    "volumes": [
      {
        "name": "test-volume",
        "emptyDir": {}
      }
    ]
  }
}`)

	tests := map[string]resultExecuteMock{
		"Test 'name' in yaml": {
			taskID:   "mypod",
			alias:    "name",
			bytes:    myPodJson,
			path:     "{.metadata.name}",
			expected: "kubectl-tester",
		},
		"Test 'objectName without jsonpath' in yaml": {
			taskID:   "mypod",
			alias:    "objectName",
			bytes:    myPodJson,
			path:     "",
			expected: "kubectl-tester",
		},
		"Test 'image with condition' in yaml": {
			taskID:   "mypod",
			alias:    "containerImage",
			bytes:    myPodJson,
			path:     "{.spec.containers[?(@.name=='bb')].image}",
			expected: "k8s.gcr.io/busybox",
		},
		"Test 'mountpath with condition' in yaml": {
			taskID:   "mypod",
			alias:    "mountPath",
			bytes:    myPodJson,
			path:     "{.spec.containers[?(@.name=='bb')].volumeMounts[?(@.name=='test-volume')].mountPath}",
			expected: "/mount/test-volume",
		},
		"Test 'annotation' in yaml": {
			taskID:   "mypod",
			alias:    "simple",
			bytes:    myPodJson,
			path:     "{.metadata.annotations.simple}",
			expected: "value",
		},
		"Test 'complex annotation' in yaml": {
			taskID:   "mypod",
			alias:    "affinity",
			bytes:    myPodJson,
			path:     `{.metadata.annotations.controller\.openebs\.io/affinity}`,
			expected: "mypin",
		},
		"Test 'complex annotation 2' in yaml": {
			taskID:   "mypod",
			alias:    "affinityTopology",
			bytes:    myPodJson,
			path:     `{.metadata.annotations.controller\.openebs\.io/affinity-topology}`,
			expected: "kubernetes.io/hostname",
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
		
		  // go template is run to check if it interferes with jsonpath querying 
		  // later. go template should not try to execute the jsonquery strings and 
		  // pass them as-is
		  _, err := template.AsMapOfObjects(jsonPathFeederYml, map[string]interface{}{
		    "test": "check",
		  })
			if err != nil {
				t.Fatalf("Expected: 'no interference error' Actual: '%s'", err)
			}
      
      // Now test the jsonpath querying which is done internally in 
      // TaskResultStorage.store() method
			q := TaskResultQuery{
				Alias: mock.alias,
				Path:  mock.path,
			}

			s := NewTaskResultStorage(mock.taskID, []TaskResultQuery{q}, mock.bytes)
			mActual, err := s.store()
			if err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%#v'", err)
			}

			// Get back to expected yaml & unmarshall the yaml into
			// this object
			mExpected := map[string]interface{}{
				mock.taskID: map[string]string{
					mock.alias: mock.expected,
				},
			}

			// Now Compare
			ok := reflect.DeepEqual(mExpected, mActual)
			if !ok {
				t.Fatalf("\nExpected: '%#v' \nActual: '%#v'", mExpected, mActual)
			}
		}) // end of run
	}
}
