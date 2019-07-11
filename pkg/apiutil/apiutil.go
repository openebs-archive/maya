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
//TODO: Get better home from reviews
package apiutil

import (
	"encoding/json"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// RemoveFinalizer will removes the finalizer from object meta
func RemoveFinalizer(
	objMeta *metav1.ObjectMeta,
	removeFinalizers ...string) {
	objFinalizers := objMeta.GetFinalizers()
	if len(objFinalizers) == 0 || len(removeFinalizers) == 0 {
		return
	}
	updatedFinalizerList := RemoveSlices(objFinalizers, removeFinalizers)
	objMeta.SetFinalizers(updatedFinalizerList)
}

// RemoveSlices will remove removeList from originalList
func RemoveSlices(originalList, removeList []string) []string {
	removeListMap := map[string]bool{}
	resultList := []string{}
	for _, value := range removeList {
		removeListMap[value] = true
	}
	for _, value := range originalList {
		if !removeListMap[value] {
			resultList = append(resultList, value)
		}
	}
	return resultList
}

// GetPatchData returns byte difference in object
func GetPatchData(oldObj, newObj interface{}) ([]byte, error) {
	oldData, err := json.Marshal(oldObj)
	if err != nil {
		return nil, errors.Wrapf(err, "marshal old object failed")
	}
	newData, err := json.Marshal(newObj)
	if err != nil {
		return nil, errors.Wrapf(err, "mashal new object failed")
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, oldObj)
	if err != nil {
		return nil, errors.Wrapf(err, "CreateTwoWayMergePatch failed")
	}
	return patchBytes, nil
}
