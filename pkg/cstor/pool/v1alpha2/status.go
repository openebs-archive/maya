package v1alpha2

import (
	"strings"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
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
func GetStatus(csp *api.CStorNPool) (string, error) {
	ret, err := zfs.NewPoolStatus().
		WithPool(PoolName(csp)).
		Execute()
	if err != nil {
		return "", err
	}

	switch parsePoolStatus(string(ret)) {
	case PoolStatusDegraded:
		return string(apis.CStorPoolStatusDegraded), nil
	case PoolStatusFaulted:
		return string(apis.CStorPoolStatusOffline), nil
	case PoolStatusOffline:
		return string(apis.CStorPoolStatusOffline), nil
	case PoolStatusOnline:
		return string(apis.CStorPoolStatusOnline), nil
	case PoolStatusRemoved:
		return string(apis.CStorPoolStatusDegraded), nil
	case PoolStatusUnavail:
		return string(apis.CStorPoolStatusOffline), nil
	default:
		return string(apis.CStorPoolStatusError), nil
	}
}

// parsePoolStatus parse output of `zpool status` command to extract the status of the pool.
// ToDo: Need to find some better way e.g contract for zpool command outputs.
func parsePoolStatus(output string) string {
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
	return poolStatus
}
