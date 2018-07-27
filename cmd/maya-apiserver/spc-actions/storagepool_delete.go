/*
Copyright 2017 The OpenEBS Authors

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

package storagepoolactions

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/storagepool"
)

func DeleteStoragePool(key string) error {
	// Business logic for deletion of storagepool cr
	glog.Infof("Storagepool delete event received for storagepoolclaim %s", key)

	// Create an empty  CasPool object
	pool := &v1alpha1.CasPool{}

	// Fill the name in CasPool object
	// This object contains pool information for performing storagepool deletion
	// The information used here is the storagepoolclaim name
	pool.StoragePoolClaim = key

	storagepoolOps, err := storagepool.NewCasPoolOperation(pool)
	if err != nil {
		return fmt.Errorf("NewCasPoolOperation failed error '%s'", err.Error())
	}
	_, err = storagepoolOps.Delete()
	if err != nil {
		return fmt.Errorf("Failed to delete cas template based storagepool: error '%s'", err.Error())
	}

	glog.Infof("Cas template based storagepool delete successfully: name '%s'", key)
	return nil
}
