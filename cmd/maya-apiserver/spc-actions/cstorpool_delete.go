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

package cstorpool

import (
	"github.com/golang/glog"
	"fmt"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/storagepool"
)


func DeleteCstorpool(key string) (error) {
	// Business logic for deletion of cstor pool cr
	glog.Infof("Cstorpool delete event received for storagepoolclaim %s",key)

	// Create an empty cstor pool object
	cstorPool := &v1alpha1.CStorPool{}

	// Fill the name in cstor pool object
	// This object contains pool information for performing cstor pool deletion
	// The information used here is the storagepoolclaim name
	cstorPool.ObjectMeta.Name = key

	spcOps, err := storagepool.NewCstorPoolOperation(cstorPool)
	if err != nil {
		return fmt.Errorf("NewCstorPoolOPeration Failed error '%s'", err.Error())
	}
	_, err = spcOps.Delete()
	if err != nil {
		return fmt.Errorf("Failed to delete cas template based cstorpool: error '%s'", err.Error())
	}

	glog.Infof("Cas template based cstorpool delete successfully: name '%s'", key)
	return nil
}