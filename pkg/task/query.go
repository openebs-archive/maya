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

// @Deprecated
//  All the structures, functions and methods present here are deprecated in
// favour of 'RunTask.post' property
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
	// Alias is the name/key used to **hold** the extracted value after jsonpath
	// parsing. This jsonpath is run against the task's result.
	Alias string `json:"alias"`
	// Path contains the path to the property of the task result, whose value(s)
	// need to be extracted.
	//
	// NOTE:
	//  Path represents a jsonpath
	//
	// NOTE:
	//  Path can be optional if they are commonly used paths. Some of the common
	// paths can be retrieved from the query's Alias property. This implies one
	// can just set the alias without bothering to set the corresponding path.
	//
	// Refer keyToJsonPathMap
	Path string `json:"path"`
	// TaskResultVerify will verify the resulting value after the task result is
	// run through the jsonpath parser
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

// queryExecutor queries specific properties (via jsonpaths) from the task result
type queryExecutor struct {
	// taskResult is the task's result after execution of a task
	taskResult []byte
	// queries holds the info about the data that needs to be
	// extracted from the task's result.
	queries []TaskResultQuery
}

func newQueryExecutor(queries []TaskResultQuery, taskResult []byte) *queryExecutor {
	return &queryExecutor{
		queries:    queries,
		taskResult: taskResult,
	}
}

// queryAndVerify will run jsonpath query against the task result & verify this
// result. The queries will be run iteratively against the same task result. Each
// of the query's output will get set against the query's alias, then aggregated
// & returned.
//
// NOTE:
//  A query path refers to a JsonPath
func (t *queryExecutor) queryAndVerify() (map[string]string, error) {
	pathResults := map[string]string{}

	// loop through all the queries & set the resulting query output against
	// the corresponding alias
	for _, q := range t.queries {
		// get the jsonpath
		path := q.Path
		if len(path) == 0 {
			path = jsonPathFromKey(q.Alias)
		}

		// check again
		if len(path) == 0 {
			return nil, fmt.Errorf("jsonpath not found for key '%s': can not query against task result", q.Alias)
		}

		// t.Alias is provided as a context that can act as an identifier;
		// result is a json doc against which jsonpath is run
		jq := template.NewJsonQuery(q.Alias, t.taskResult, path)
		jqOp, err := jq.Query()
		if err != nil {
			return nil, err
		}

		v := newTaskResultVerifyExecutor(q.Alias, jqOp, q.TaskResultVerify)
		_, err = v.verify()
		if err != nil {
			return nil, err
		}

		// stick the json query output with the alias
		pathResults[q.Alias] = jqOp
	}

	return pathResults, nil
}

// execute will query & validate the data extracted from specific
// properties of the task result. This query result will be returned.
func (t *queryExecutor) result() (pathResults map[string]string, err error) {
	return t.queryAndVerify()
}

// queryExecFormatter queries specific properties (via jsonpaths) from the task
// result & prepares a formatted output based on these query results
type queryExecFormatter struct {
	// index is the index that forms the query result
	index string
	// queryExecutor forms the rest of the properties required to execute one
	// or more queries
	queryExecutor
}

func newQueryExecFormatter(index string, queries []TaskResultQuery, taskResult []byte) *queryExecFormatter {
	return &queryExecFormatter{
		index: index,
		queryExecutor: queryExecutor{
			queries:    queries,
			taskResult: taskResult,
		},
	}
}

// formattedResult will query & validate the data extracted from specific
// properties of the task result. This query result will be returned as a map
// with index as the key
func (t *queryExecFormatter) formattedResult() (pathResults map[string]interface{}, err error) {
	r, err := t.queryAndVerify()
	if err != nil {
		return
	}

	// query path results are set against the index
	pathResults = map[string]interface{}{
		t.index: r,
	}
	return
}
