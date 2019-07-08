package v1alpha2

import (
	"github.com/golang/glog"
	api "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
	"github.com/pkg/errors"
)

const (
	// DeviceTypeSpare .. spare device type
	DeviceTypeSpare = "spare"
	// DeviceTypeReadCache .. read cache device type
	DeviceTypeReadCache = "cache"
	// DeviceTypeWriteCache .. write cache device type
	DeviceTypeWriteCache = "log"
	// DeviceTypeEmpty .. empty device type.. data disk
	DeviceTypeEmpty = ""
)

// addRaidGroup add given raidGroup to pool
func addRaidGroup(csp *api.CStorNPool, r *api.RaidGroup) error {
	var err error

	if r.IsReadCache {
		if e := addVdevToPool(csp, r, DeviceTypeReadCache); e != nil {
			err = ErrorWrapf(err, "Failed add readcache {%s}", e.Error())
		}
	} else if r.IsSpare {
		if e := addVdevToPool(csp, r, DeviceTypeSpare); e != nil {
			err = ErrorWrapf(err, "Failed to add spare disk {%s}", e.Error())
		}
	} else if r.IsWriteCache {
		if e := addVdevToPool(csp, r, DeviceTypeWriteCache); e != nil {
			err = ErrorWrapf(err, "Failed to add write cache {%s}", e.Error())
		}
	} else {
		if e := addVdevToPool(csp, r, DeviceTypeEmpty); e != nil {
			err = ErrorWrapf(err, "Failed to add additional disk {%s}", e.Error())
		}
	}
	return err
}

// addVdev will add devices to pool
func addVdevToPool(csp *api.CStorNPool, r *api.RaidGroup, deviceType string) error {
	ptype := r.Type
	if len(ptype) == 0 {
		return errors.Errorf("No type mentioned in raidGroup for pool {%s}", PoolName(csp))
	}

	vlist, err := getPathForCSPBdevList(r.BlockDevices)
	if err != nil {
		glog.Errorf("Failed to get list of disk-path : %s", err.Error())
		return err
	}

	_, err = zfs.NewPoolExpansion().
		WithDeviceType(deviceType).
		WithType(ptype).
		WithPool(PoolName(csp)).
		WithVdevList(vlist).
		Execute()
	return err
}

// addNewVdevFromCSP will add new disk, which is not being used in pool, from csp to given pool
func addNewVdevFromCSP(csp *api.CStorNPool) error {
	var err error

	poolTopology, err := zfs.NewPoolDump().WithPool(PoolName(csp)).Execute()
	if err != nil {
		return errors.Errorf("Failed to fetch pool topology.. %s", err.Error())
	}

	for _, raidGroup := range csp.Spec.RaidGroups {
		wholeGroup := true
		var devlist []string

		for _, bdev := range raidGroup.BlockDevices {
			if isUsed := checkIfDeviceUsed(bdev.DevLink, poolTopology); !isUsed {
				devlist = append(devlist, bdev.DevLink)
			} else {
				wholeGroup = false
			}
		}
		if wholeGroup {
			if er := addRaidGroup(csp, &raidGroup); er != nil {
				err = ErrorWrapf(err, "Failed to add raidGroup{%s}.. %s", raidGroup.Name, er.Error())
			}
		} else if len(devlist) != 0 {
			if _, er := zfs.NewPoolExpansion().
				WithVdevList(devlist).
				WithPool(PoolName(csp)).
				Execute(); er != nil {
				err = ErrorWrapf(err, "Failed to add devlist %v.. err {%s}", devlist, er.Error())
			}
		}
	}

	return err
}

func removePoolVdev(csp *api.CStorNPool, bdev api.CStorPoolClusterBlockDevice) error {
	if _, err := zfs.NewPoolRemove().
		WithDevice(bdev.DevLink).
		WithPool(PoolName(csp)).
		Execute(); err != nil {
		return err
	}

	// Let's clear the label for removed disk
	if _, err := zfs.NewPoolLabelClear().
		WithForceFully(true).
		WithVdev(bdev.DevLink).
		Execute(); err != nil {
		// Let's just log the error
		glog.Errorf("Failed to perform label clear for disk {%s}", bdev.DevLink)
	}

	return nil
}

func replacePoolVdev(csp *api.CStorNPool, bdev api.CStorPoolClusterBlockDevice, npath string) error {
	if IsEmpty(npath) || IsEmpty(bdev.DevLink) {
		return errors.Errorf("Empty path for bdev")
	}

	_, err := zfs.NewPoolDiskReplace().
		WithOldVdev(bdev.DevLink).
		WithNewVdev(npath).
		WithPool(PoolName(csp)).
		WithForcefully(true).
		Execute()
	if err != nil {
		if _, er := zfs.NewPoolLabelClear().
			WithForceFully(true).
			WithVdev(bdev.DevLink).
			Execute(); er != nil {
			// Let's log the error
			glog.Errorf("Failed to perform label clear for disk {%s}", bdev.DevLink)
		}
	}
	return err
}
