package mapiserver

import (
	"encoding/json"

	v1 "github.com/openebs/maya/types/v1"
)

const (
	statsVolumePath = "/latest/volumes/stats/"
)

// VolumeStats returns the VolumeMetrics fetched from apisever endpoint
func VolumeStats(volName, namespace string) (v1.VolumeMetrics, error) {
	stats := v1.VolumeMetrics{}
	body, err := getRequest(GetURL()+statsVolumePath+volName, namespace, false)
	if err != nil {
		return stats, err
	}
	err = json.Unmarshal(body, &stats)
	return stats, err
}
