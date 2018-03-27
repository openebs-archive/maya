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
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/types"
)

type TaskPatchType string

const (
	// JsonTPT refers to a generic json patch type that is understood
	// by Kubernetes API as well
	JsonTPT TaskPatchType = "json"
	// MergeTPT refers to a generic json merge patch type that is
	// understood by Kubernetes API as well
	MergeTPT TaskPatchType = "merge"
	// StrategicTPT refers to a patch type that is understood
	// by Kubernetes API only
	StrategicTPT TaskPatchType = "strategic"
)

var taskPatchTypes = map[TaskPatchType]types.PatchType{
	JsonTPT:      types.JSONPatchType,
	MergeTPT:     types.MergePatchType,
	StrategicTPT: types.StrategicMergePatchType,
}

// TaskPatches will consist of patches that gets applied
// against the task object
type TaskPatch struct {
	// Type determines the type of patch to be applied
	Type TaskPatchType `json:"type"`
	// Specs is a yaml document that provides the patch specifications
	//
	//  Below is a sample patch as yaml document
	//  ```yaml
	//      spec:
	//        template:
	//          spec:
	//            affinity:
	//              nodeAffinity:
	//                requiredDuringSchedulingIgnoredDuringExecution:
	//                  nodeSelectorTerms:
	//                  - matchExpressions:
	//                    - key: kubernetes.io/hostname
	//                      operator: In
	//                      values:
	//                      - amit-thinkpad-l470
	//              podAntiAffinity: null
	//  ```
	Specs string `json:"pspec"`
}

type taskPatchExecutor struct {
	patch TaskPatch
}

func isValidPatchType(patch TaskPatch) bool {
	return patch.Type == JsonTPT || patch.Type == StrategicTPT || patch.Type == MergeTPT
}

func newTaskPatchExecutor(patch TaskPatch) (*taskPatchExecutor, error) {
	if !isValidPatchType(patch) {
		return nil, fmt.Errorf("Failed to create patch executor: Invalid patch type '%s'", patch.Type)
	}

	return &taskPatchExecutor{
		patch: patch,
	}, nil
}

// build converts the patch in yaml document format to corresponding
// json document
func (p *taskPatchExecutor) build() ([]byte, error) {
	m := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(p.patch.Specs), &m)
	if err != nil {
		return nil, err
	}

	return json.Marshal(m)
}

func (p *taskPatchExecutor) patchType() types.PatchType {
	return taskPatchTypes[p.patch.Type]
}
