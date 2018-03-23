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
	"fmt"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/template"
)

// TaskResultQuery helps in extracting specific data from the task's result
//
// NOTE:
//  A TaskResult is the result obtained after this task's execution.
type TaskResultQuery struct {
	// Alias is the name/key used to hold the extracted data from the task's
	// result
	Alias string `json:"alias"`
	// Path contains the path to the property of the taskresult, whose value(s)
	// need to be extracted.
	//
	// NOTE:
	//  Path will be a string type. It will vary depending on the query language
	// used. e.g. It can represent a jsonpath or a go template function.
	//
	// NOTE:
	//  Path can be optional i.e. commonly used Paths can be set as constants &
	// be retrieved from the query's Alias property. Refer keyToJsonPathMap.
	Path string `json:"path"`
	// TaskResultVerify will verify the data collected after querying
	TaskResultVerify `json:"verify"`
}

// KeyToJsonPathMap holds often used jsonpath(s) against some predefined keys
var keyToJsonPathMap = map[string]string{
	// All the K8s objects have name at this path
	string(v1alpha1.ObjectNameTRTP): "{.metadata.name}",
	// StoragePool (i.e. a Custom Resource) path
	"poolPath": "{.spec.path}",
	// K8s Service's cluster IP path
	"clusterIP": "{.spec.clusterIP}",
}

func jsonPathFromKey(key string) (jsonpath string) {
	return keyToJsonPathMap[key]
}

// taskResultQueryExecutor queries data from the task result.
type taskResultQueryExecutor struct {
	// taskID is the identity of the task
	taskID string
	// result is the task's result after executing this task
	result []byte
	// queries holds the info about the data that needs to be
	// extracted from the task's result.
	queries []TaskResultQuery
}

func newTaskResultQueryExecutor(taskID string, queries []TaskResultQuery, result []byte) *taskResultQueryExecutor {
	return &taskResultQueryExecutor{
		taskID:  taskID,
		queries: queries,
		result:  result,
	}
}

// queryAndVerify will run jsonpath query against the task result & verify this
// result. Each of the query will be run in an iteractive manner. All the query
// outputs will be aggregated & returned.
//
// NOTE:
//  This is currently coupled to JsonPath Query!!!
func (t *taskResultQueryExecutor) queryAndVerify() (map[string]string, error) {
	var outputs = map[string]string{}

	for _, q := range t.queries {
		// get the jsonpath
		path := q.Path
		if len(path) == 0 {
			path = jsonPathFromKey(q.Alias)
		}

		// check again
		if len(path) == 0 {
			err := fmt.Errorf("jsonpath not found for key '%s': can not query against task result", q.Alias)
			return nil, err
		}

		// t.taskID is provided as a context that can act as an identifier
		// result is the json doc against which the jsonpath is run
		jq := template.NewJsonQuery(t.taskID, t.result, path)
		op, err := jq.Query()
		if err != nil {
			return nil, err
		}

		v := newTaskResultVerifyExecutor(t.taskID, q.Alias, op, q.TaskResultVerify)
		_, err = v.verify()
		if err != nil {
			return nil, err
		}

		outputs[q.Alias] = op
	}

	return outputs, nil
}

// execute will query & validate the data extracted from specific
// properties of the task result. This query data will be returned as a map with
// taskID as the key
func (t *taskResultQueryExecutor) execute() (storage map[string]interface{}, err error) {
	outputs, err := t.queryAndVerify()
	if err != nil {
		return
	}

	// attach extracted data with the task ID
	storage = map[string]interface{}{
		t.taskID: outputs,
	}

	return
}
