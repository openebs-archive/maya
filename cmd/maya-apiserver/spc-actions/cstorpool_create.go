/*
Copyright 2017 The Kubernetes Authors.

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

func CreateCstorpool(spcGot *apis.StoragePoolClaim) (error) {

	// Check wether the spc object has been processed for cstor pool creation
	if(spcGot.Status.Phase=="online"){
		return errors.New("Pool Already Exsists")
	}
	poolType := spcGot.Spec.PoolSpec.PoolType
	if(poolType==""){
		return errors.New("aborting cstor pool create operation as no poolType specified")
	}

	diskList := spcGot.Spec.Disks.DiskList
	if(len(diskList)==0){
		return errors.New("aborting cstor pool create operation as no disk specified")
	}
	err:=poolCreateWorker(spcGot)

	if err!=nil{
		return err
	}

	return nil
}

// function that creates a cstorpool CR
func poolCreateWorker(spcGot *apis.StoragePoolClaim) (error) {
	fmt.Println("Creation of cstor pool CR initiated Now")
	//fmt.Println("Creating cstorpool cr for spc %s via CASTemplate", spcGot.ObjectMeta.Name)
	glog.Infof("Creating cstorpool cr for spc %s via CASTemplate", spcGot.ObjectMeta.Name)

	// Create an empty cstor pool object
	// This object will be filled in with some details that will be required by CAS Engine to create actual
	// cstor pool cr object in kubernetes.
	// As part of building the cstor pool object some of the details that are getting filled in may not be present
	// in actual object in kubernetes. These kind of details are specific to CAS Engine usage only.
	cstorPool := &v1alpha1.CStorPool{}

	//Generate name using the prefix of StoragePoolClaim name
	cstorPool.ObjectMeta.Name = spcGot.Name

	// Add Pooltype specification
	cstorPool.Spec.PoolSpec.PoolType = spcGot.Spec.PoolSpec.PoolType

	// Fetch castemplate from spc object
	castName := spcGot.Annotations[string(v1alpha1.SPCASTemplateCK)]

	// make a map that should contain the castemplate name
	mapCastName := make(map[string]string)

	// Fill the map with castemplate name
	mapCastName[string(v1alpha1.SPCASTemplateCK)] = castName

	// Push the map to cstor pool cr object
	// This Annotation will however will not be present in actual object
	cstorPool.Annotations = mapCastName

	mapLabels := make(map[string]string)
	// Push storage pool claim name to cstor pool cr object as a label
	// This label will be present in the actual object
	mapLabels[string(v1alpha1.StoragePoolClaimCK)] = spcGot.Name

	// Add init status
	cstorPool.Status.Phase= v1alpha1.CStorPoolStatusInit


	// Push node hostname to cstor pool cr object as a label.

	// mapLabels[string(v1alpha1.CstorPoolHostNameCVK)] = spcGot.Spec.NodeSelector[nodeIndex]
	cstorPool.Labels = mapLabels

	// TODO : Select disks from nodes and push it to cstor pool cr object

	cstorOps, err := storagepool.NewCstorPoolOperation(cstorPool)
	if err != nil {
		fmt.Println("NewCstorPoolOPeration Failed with following error")
		fmt.Println(err)
	}
	cstorPoolObject, err := cstorOps.Create()
	if err != nil {
		glog.Errorf("failed to create cas template based cstorpool: error '%s'", err.Error())
		//return nil, CodedError(500, err.Error())
	} else {
		glog.Infof("cas template based cstorpool created successfully: name '%s'", cstorPoolObject.Name)
	}
	return nil
}