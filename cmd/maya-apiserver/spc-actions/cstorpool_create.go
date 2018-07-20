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
	"errors"
	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/storagepool"
	"fmt"
)

const(
	onlineStatus = "Online"
)
// Cas template is a custom resource which has a list of runTasks.

// runTasks are configmaps which has defined yaml templates for resources that needs
// to be created or deleted for a cstorpool creation or deletion respectively.

// CreateCstorpool is a function that does following:
// 1. It receives storagepoolclaim object from the spc watcher event handler.

// 2. There are three sources of information for creation of a cstorpool i.e.
//    storagepoolclaim object,default values in cas template and hardcoded
//    values in run tasks.
//    The function will validate the contained information in storagepoolcalim object.

// 3. After successful validation, it will call a worker function for actual cstorpool creation.
func CreateCstorpool(spcGot *apis.StoragePoolClaim) (error) {

	glog.Infof("Cstorpool create event received for storagepoolclaim %s",spcGot.ObjectMeta.Name)

	// Check wether the spc object has been processed for cstor pool creation
	if(spcGot.Status.Phase==onlineStatus){
		return errors.New("Cstorpool already exists since the status on storagepoolclaim object is Online")
	}

	// Check for poolType
	poolType := spcGot.Spec.PoolSpec.PoolType
	if(poolType==""){
		return errors.New("Aborting cstorpool create operation as no poolType is specified")
	}

	// Check for disks
	diskList := spcGot.Spec.Disks.DiskList
	if(len(diskList)==0){
		return errors.New("Aborting cstorpool create operation as no disk is specified")
	}

	// The name of cas template should be provided as annotation in storagepoolclaim yaml
	// so that it can be used.

	// Check for cas template
	castTemplateName := spcGot.Annotations[string(v1alpha1.SPCASTemplateCK)]
	if(castTemplateName==""){
		return errors.New("Aborting cstorpool create operation as no cas template is specified")
	}

	// Calling worker function to create cstorpool
	err:=poolCreateWorker(spcGot,castTemplateName)

	if err!=nil{
		return err
	}

	return nil
}

// poolCreateWorker is a worker function which will create a cstorpool
// successful creation of a cstorpool should involve following successful resource creation:
// 1. cstorpool ( A custom resource).
// 2. cstorpool deployment
// 3. storagepool (A custom resource)

func poolCreateWorker(spcGot *apis.StoragePoolClaim, castTemplateName string) (error) {

	glog.Infof("Creating cstorpool for storagepoolclaim %s via CASTemplate", spcGot.ObjectMeta.Name)

	// Create an empty cstorpool object.

	// This object will be filled in with some of the details present in storagepoolclaim object
	// that will be required by CAS Engine to create actual cstorpool cr object in kubernetes.

	// As part of building the cstor pool object, some of the details that are getting filled in,
	// may not be present in actual cstorpool object that lives in kubernetes. These kind of
	// details are specific to CAS Engine usage only.
	cstorPool := &v1alpha1.CStorPool{}

	//Generate name using the prefix of StoragePoolClaim name
	cstorPool.ObjectMeta.Name = spcGot.Name

	// Add Pooltype specification
	cstorPool.Spec.PoolSpec.PoolType = spcGot.Spec.PoolSpec.PoolType

	// Add overProvisioning which is a bool (e.g. true or false)
	cstorPool.Spec.PoolSpec.OverProvisioning = spcGot.Spec.PoolSpec.OverProvisioning


	// make a map that should contain the castemplate name
	mapcastTemplateName := make(map[string]string)

	// Fill the map with castemplate name
	// e.g.
	// openebs.io/create-template : cast-standard-cstorpool-0.6.0
	mapcastTemplateName[string(v1alpha1.SPCASTemplateCK)] = castTemplateName

	// Push the map to cstorpool cr object
	// This Annotation will however not be present in actual object
	// This information is reuired by CAS engine
	cstorPool.Annotations = mapcastTemplateName

	cstorOps, err := storagepool.NewCstorPoolOperation(cstorPool)
	if err != nil {
		return fmt.Errorf("NewCstorPoolOPeration Failed error '%s'", err.Error())

	}
	_, err = cstorOps.Create()
	if err != nil {
		return fmt.Errorf("Failed to create cas template based cstorpool: error '%s'", err.Error())

	}

	glog.Infof("Cas template based cstorpool created successfully: name '%s'",spcGot.Name )
	return nil
}