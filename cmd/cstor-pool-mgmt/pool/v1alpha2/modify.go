package v1alpha2

import (
	api "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
)

// Update will update the deployed pool according to given csp object
func Update(csp *api.CStorNPool) error {
	var err error
	var isObjChanged bool
	var isRaidGroupChanged bool

	// first we will check if there any bdev is replaced or removed
	for raidIndex := 0; raidIndex < len(csp.Spec.RaidGroups); raidIndex++ {
		isRaidGroupChanged = false
		raidGroup := csp.Spec.RaidGroups[raidIndex]

		for bdevIndex := 0; bdevIndex < len(raidGroup.BlockDevices); bdevIndex++ {
			bdev := raidGroup.BlockDevices[bdevIndex]

			// Let's check if bdev name is empty
			// if yes then remove relevant disk from pool
			if IsEmpty(bdev.BlockDeviceName) {
				// block device name is empty
				// Let's remove it
				// TODO should we offline it only?
				if er := removePoolVdev(csp, bdev); er != nil {
					err = ErrorWrapf(err, "Failed to remove bdev {%s}.. %s", bdev.DevLink, er.Error())
				}
				// remove this entry since it's been already removed from pool
				raidGroup.BlockDevices = append(raidGroup.BlockDevices[:bdevIndex], raidGroup.BlockDevices[bdevIndex+1:]...)
				// We just remove the bdevIndex entry from BlockDevices
				// let's decrement the index to handle above removal
				bdevIndex--
				isRaidGroupChanged = true
				continue
			}

			// Let's check if bdev path is changed or not
			newpath, isChanged, er := isBdevPathChanged(bdev)
			if er != nil {
				err = ErrorWrapf(err, "Failed to check bdev change {%s}.. %s", bdev.BlockDeviceName, er.Error())
			} else if isChanged {
				if er := replacePoolVdev(csp, bdev, newpath); err != nil {
					err = ErrorWrapf(err, "Failed to replace bdev for {%s}.. %s", bdev.BlockDeviceName, er.Error())
				} else {
					// Let's update devLink with new path for this bdev
					raidGroup.BlockDevices[bdevIndex].DevLink = newpath
					isRaidGroupChanged = true
				}
			}
		}
		// If raidGroup is changed then update the csp.spec.raidgroup entry
		// If raidGroup doesn't have any blockdevice then remove that raidGroup
		// and set isObjChanged
		if isRaidGroupChanged {
			if len(raidGroup.BlockDevices) == 0
				csp.Spec.RaidGroups = append(csp.Spec.RaidGroups[:raidIndex], csp.Spec.RaidGroups[raidIndex+1:]...)
				// We removed the raidIndex entry csp.Spec.raidGroup
				raidIndex--
			}
			isObjChanged = true
		}
	}

	if er := addNewVdevFromCSP(csp); er != nil {
		err = ErrorWrapf(err, "Failed to execute add operation.. %s", er.Error())
	}

	if isObjChanged {
		if _, er := OpenEBSClient2.
			OpenebsV1alpha2().
			CStorNPools(csp.Namespace).
			Update(csp); er != nil {
			err = ErrorWrapf(err, "Failed to update object.. err {%s}", er.Error())
		}
	}
	return err
}
