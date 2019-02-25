// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mapiserver

import (
	"encoding/json"
	"time"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	volumeCreateTimeout = 60 * time.Second
	volumePath          = "/latest/volumes/"
)

// CreateVolume creates a volume by invoking the API call to m-apiserver
func CreateVolume(vname, size, namespace string) error {
	// Filling structure with values
	cVol := v1alpha1.CASVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vname,
			Namespace: namespace,
		},
		Spec: v1alpha1.CASVolumeSpec{
			Capacity: size,
		},
	}
	// Marshal serializes the value of vs structure
	jsonValue, err := json.Marshal(cVol)
	if err != nil {
		return err
	}

	_, err = sendRequest(requestType, GetURL()+volumePath, jsonValue, "", false)
	return err
}
