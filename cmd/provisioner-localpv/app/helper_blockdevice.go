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

This code was taken from https://github.com/rancher/local-path-provisioner
and modified to work with the configuration options used by OpenEBS
*/

package app

import (
	//"fmt"
	//"path/filepath"
	//"strings"
	//"time"

	//"github.com/golang/glog"
	//"github.com/pkg/errors"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"

	//hostpath "github.com/openebs/maya/pkg/hostpath/v1alpha1"

	//container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	//pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	//volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	bdcStorageClassAnnotation = "local.openebs.io/blockdeviceclaim"
)

//TODO
//var (
//CmdTimeoutCounts specifies the duration to wait for cleanup pod to be launched.
//CmdTimeoutCounts = 120
//)

// HelperBlockDeviceOptions contains the options that
// will launch a BDC on a specific node (nodeName)
type HelperBlockDeviceOptions struct {
	nodeName   string
	name       string
	capacity   string
	deviceType string
	bdcName    string
}

// validate checks that the required fields to create BDC
// are available
func (blkDevOpts *HelperBlockDeviceOptions) validate() error {
	if blkDevOpts.name == "" || blkDevOpts.nodeName == "" {
		return errors.Errorf("invalid empty name or node")
	}
	return nil
}

// hasBDC checks if the bdcName has already been determined
func (blkDevOpts *HelperBlockDeviceOptions) hasBDC() bool {
	return blkDevOpts.bdcName != ""
}

// getBlcokDeviceClaimFromPV inspects the PV and fetches the BDC associated
//  with the Local PV.
func (blkDevOpts *HelperBlockDeviceOptions) setBlockDeviceClaimFromPV(pv *corev1.PersistentVolume) error {
	bdc, found := pv.Annotations[bdcStorageClassAnnotation]
	if found {
		blkDevOpts.bdcName = bdc
	}
	return nil
}

// createBlockDeviceClaim creates a new BlockDeviceClaim for a given
//  Local PV
func (p *Provisioner) createBlockDeviceClaim(blkDevOpts *HelperBlockDeviceOptions) error {
	if err := blkDevOpts.validate(); err != nil {
		return err
	}

	if !blkDevOpts.hasBDC() {
		//Setup the BDC Name using the provided PV name.
		//To help easily co-relate, the name format for bdc will
		//be  "bdc-<pvname>"
		blkDevOpts.bdcName = "bdc" + blkDevOpts.name
	}

	//TODO: Create BDC

	return nil
}

// getBlockDevicePath fetches the BDC associated with this Local PV
// or creates one. From the BDC, fetch the BD and get the path
func (p *Provisioner) getBlockDevicePath(blkDevOpts *HelperBlockDeviceOptions) (string, error) {

	if !blkDevOpts.hasBDC() {
		err := p.createBlockDeviceClaim(blkDevOpts)
		if err != nil {
			return "", err
		}
	}

	//TODO
	path := ""
	//Check if the BDC is created
	//Check if the BDC is associated with a BD
	//Get the BD Path.

	return path, nil
}

// deleteBlockDeviceClaim deletes the BlockDeviceClaim associated with the
//  PV being deleted.
func (p *Provisioner) deleteBlockDeviceClaim(blkDevOpts *HelperBlockDeviceOptions) error {
	if !blkDevOpts.hasBDC() {
		return nil
	}

	//TODO: Issue a delete BDC request

	return nil
}
