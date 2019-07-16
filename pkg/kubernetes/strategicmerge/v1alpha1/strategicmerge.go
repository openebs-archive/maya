/*
Copyright 2019 The OpenEBS Authors

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

package v1alpha1

import (
	"encoding/json"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// GetPatchData returns byte difference in object
func GetPatchData(oldObj, newObj interface{}) ([]byte, error) {
	oldData, err := json.Marshal(oldObj)
	if err != nil {
		return nil, errors.Wrapf(err, "marshalling old object failed")
	}
	newData, err := json.Marshal(newObj)
	if err != nil {
		return nil, errors.Wrapf(err, "marshalling new object failed")
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, oldObj)
	if err != nil {
		return nil, errors.Wrapf(err, "CreateTwoWayMergePatch failed")
	}
	return patchBytes, nil
}
