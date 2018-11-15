/*
Copyright 2017 The OpenEBS Authors.

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

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

const poolPath = "/latest/pools/"

// ListPools returns a obj StoragePoolList from api-server
func ListPools() (*v1alpha1.StoragePoolList, error) {

	body, err := getRequest(GetURL()+poolPath, "", false)
	if err != nil {
		return nil, err
	}
	pools := v1alpha1.StoragePoolList{}
	err = json.Unmarshal(body, &pools)
	return &pools, err
}

// ReadPool returns a obj of StoragePool from api-server
func ReadPool(poolName string) (*v1alpha1.StoragePool, error) {
	body, err := getRequest(GetURL()+poolPath+poolName, "", false)
	if err != nil {
		return nil, err
	}
	pool := v1alpha1.StoragePool{}
	err = json.Unmarshal(body, &pool)
	return &pool, err
}
