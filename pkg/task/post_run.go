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

// PostRunTemplate composes a set of template functions that are executed
// against the result of a task execution.
//
// These is an example of PostRunTemplates:
// - {{ jsonpath .JsonDoc "{.apiVersion}" | saveAs "TaskResult.tskId.apiVersion" .Values }}
// - {{ .Values.TaskResult.tskId.kinds | splitList " " | len }}
//
// NOTE:
//  Refer to pkg/template/template_test.go for more details
type PostRunTemplate string

// PostRunTemplates are a set of actions that will be performed against the result
// that is obtained after running this task
type PostRunTemplates struct {
	Runs []PostRunTemplate `json:"runs"`
}

type PostRunExecutor struct {
}
