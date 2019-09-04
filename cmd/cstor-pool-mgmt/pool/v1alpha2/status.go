/*
Copyright 2019 The OpenEBS Authors.

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

package v1alpha2

import (
	"strings"

	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
)

// GetPropertyValue will return value of given property for given pool
func GetPropertyValue(poolName, property string) (string, error) {
	ret, err := zfs.NewPoolGetProperty().
		WithScriptedMode(true).
		WithField("value").
		WithProperty(property).
		WithPool(poolName).
		Execute()
	if err != nil {
		return "", err
	}
	outStr := strings.Split(string(ret), "\n")
	return outStr[0], nil
}
