package v1alpha2

import (
	"strings"

	api "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
)

const (
	// StatusNoPoolsAvailable .. not pools available
	StatusNoPoolsAvailable = "no pools available"
	// PoolStatusDegraded .. pool is degraded
	PoolStatusDegraded = "DEGRADED"
	// PoolStatusFaulted .. pool is in faulty state
	PoolStatusFaulted = "FAULTED"
	// PoolStatusOffline .. pool is in offline state
	PoolStatusOffline = "OFFLINE"
	// PoolStatusOnline .. pool is online
	PoolStatusOnline = "ONLINE"
	// PoolStatusRemoved .. pool is in removed state
	PoolStatusRemoved = "REMOVED"
	// PoolStatusUnavail .. pool is unavailable
	PoolStatusUnavail = "UNAVAIL"
)

// GetStatus return status of the pool
func GetStatus(csp *api.CStorNPool) (api.CStorPoolStatus, error) {
	var status api.CStorPoolStatus

	ret, err := zfs.NewPoolStatus().
		WithPool(PoolName(csp)).
		Execute()
	if err != nil {
		return status, err
	}

	status.Phase = parsePoolStatus(string(ret))

	freeSize, er := getPropertyValue(csp, "free")
	if er != nil {
		err = ErrorWrapf(err, "Failed to fetch free size")
	} else {
		status.Capacity.Free = freeSize
	}

	usedSize, er := getPropertyValue(csp, "allocated")
	if er != nil {
		err = ErrorWrapf(err, "Failed to fetch used size")
	} else {
		status.Capacity.Used = usedSize
	}

	totalSize, er := getPropertyValue(csp, "size")
	if er != nil {
		err = ErrorWrapf(err, "Failed to fetch total size")
	} else {
		status.Capacity.Total = totalSize
	}

	return status, err
}

// parsePoolStatus parse output of `zpool status` command to extract the status of the pool.
// ToDo: Need to find some better way e.g contract for zpool command outputs.
func parsePoolStatus(output string) api.CStorPoolPhase {
	var outputStr []string
	var poolStatus string
	if !IsEmpty(strings.TrimSpace(output)) {
		outputStr = strings.Split(output, "\n")
		if !(len(outputStr) < 2) {
			poolStatusArr := strings.Split(outputStr[1], ":")
			if !(len(outputStr) < 2) {
				poolStatus = strings.TrimSpace(poolStatusArr[1])
			}
		}
	}
	switch string(poolStatus) {
	case PoolStatusDegraded:
		return api.CStorPoolStatusDegraded
	case PoolStatusFaulted:
		return api.CStorPoolStatusFaulted
	case PoolStatusOffline:
		return api.CStorPoolStatusOffline
	case PoolStatusOnline:
		return api.CStorPoolStatusOnline
	case PoolStatusRemoved:
		return api.CStorPoolStatusRemoved
	case PoolStatusUnavail:
		return api.CStorPoolStatusUnavail
	default:
		return api.CStorPoolStatusDegraded
	}
}

// getPropertyValue will return value of given property for given csp object's pool
func getPropertyValue(csp *api.CStorNPool, property string) (string, error) {
	ret, err := zfs.NewPoolGProperty().
		WithScriptedMode(true).
		WithField("value").
		WithProperty(property).
		WithPool(PoolName(csp)).
		Execute()
	return string(ret), err
}
