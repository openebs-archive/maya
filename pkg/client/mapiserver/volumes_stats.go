package mapiserver

import (
	"encoding/json"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

const (
	statsVolumePath = "/latest/volumes/stats/"
)

// VolumeStats returns the VolumeMetrics fetched from apisever endpoint
func VolumeStats(volName, namespace string) (v1alpha1.VolumeMetricsList, error) {
	stats := v1alpha1.VolumeMetricsList{}
	body, err := getRequest(GetURL()+statsVolumePath+volName, namespace, false)
	if err != nil {
		return stats, err
	}
	err = json.Unmarshal(body, &stats)
	return stats, err
}
