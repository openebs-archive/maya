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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// Update will update the deployed pool according to given cspi object
func Update(cspi *apis.CStorPoolInstance) (*apis.CStorPoolInstance, error) {
	var err error
	var isObjChanged bool
	var isRaidGroupChanged bool

	// first we will check if there any bdev is replaced or removed
	for raidIndex := 0; raidIndex < len(cspi.Spec.RaidGroups); raidIndex++ {
		isRaidGroupChanged = false
		raidGroup := cspi.Spec.RaidGroups[raidIndex]

		for bdevIndex := 0; bdevIndex < len(raidGroup.BlockDevices); bdevIndex++ {
			bdev := raidGroup.BlockDevices[bdevIndex]

			// Let's check if bdev name is empty
			// if yes then remove relevant disk from pool
			/*
				// TODO revisit for day-2 ops
					if IsEmpty(bdev.BlockDeviceName) {
						// block device name is empty
						// Let's remove it
						// TODO should we offline it only?
						if er := removePoolVdev(cspi, bdev); er != nil {
							err = ErrorWrapf(err, "Failed to remove bdev {%s}.. %s", bdev.DevLink, er.Error())
							continue
						}
						// remove this entry since it's been already removed from pool
						raidGroup.BlockDevices = append(raidGroup.BlockDevices[:bdevIndex], raidGroup.BlockDevices[bdevIndex+1:]...)
						// We just remove the bdevIndex entry from BlockDevices
						// let's decrement the index to handle above removal
						bdevIndex--
						isRaidGroupChanged = true
						continue
					}
			*/

			// Let's check if any replacement is needed for this BDev
			newpath, er := getPathForBDev(bdev.BlockDeviceName)
			if er != nil {
				err = ErrorWrapf(err, "Failed to check bdev change {%s}.. %s", bdev.BlockDeviceName, er.Error())
			} else {
				if diskPath, er := replacePoolVdev(cspi, bdev, newpath); er != nil {
					err = ErrorWrapf(err, "Failed to replace bdev for {%s}.. %s", bdev.BlockDeviceName, er.Error())
				} else {
					if !IsEmpty(diskPath) && diskPath != bdev.DevLink {
						// Let's update devLink with new path for this bdev
						raidGroup.BlockDevices[bdevIndex].DevLink = diskPath
						isRaidGroupChanged = true
					}
				}
			}
		}
		// If raidGroup is changed then update the cspi.spec.raidgroup entry
		// If raidGroup doesn't have any blockdevice then remove that raidGroup
		// and set isObjChanged
		if isRaidGroupChanged {
			if len(raidGroup.BlockDevices) == 0 {
				cspi.Spec.RaidGroups = append(cspi.Spec.RaidGroups[:raidIndex], cspi.Spec.RaidGroups[raidIndex+1:]...)
				// We removed the raidIndex entry cspi.Spec.raidGroup
				raidIndex--
			}
			isObjChanged = true
		}
	}

	//TODO revisit for day 2 ops
	if er := addNewVdevFromCSP(cspi); er != nil {
		err = ErrorWrapf(err, "Failed to execute add operation.. %s", er.Error())
	}

	if isObjChanged {
		if ncspi, er := OpenEBSClient.
			OpenebsV1alpha1().
			CStorPoolInstances(cspi.Namespace).
			Update(cspi); er != nil {
			err = ErrorWrapf(err, "Failed to update object.. err {%s}", er.Error())
		} else {
			cspi = ncspi
		}
	}
	return cspi, err
}
