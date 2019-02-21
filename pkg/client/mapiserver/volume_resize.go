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

package mapiserver

import (
	"encoding/json"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	resizeVolumePath = "/latest/volumes/resize"
)

// ResizeVolume will request maya-apiserver to resize volume
func ResizeVolume(volName, size, namespace string) error {
	resizeVol := apis.CASVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      volName,
			Namespace: namespace,
		},
		Spec: apis.CASVolumeSpec{
			Capacity: size,
		},
	}

	// Marshal serializes the values
	jsonValue, err := json.Marshal(resizeVol)
	if err != nil {
		return err
	}

	requestType := "UPDATE"
	_, err = sendRequest(requestType, GetURL()+resizeVolumePath, jsonValue, namespace, true)
	return err
}
