package apiutil

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
