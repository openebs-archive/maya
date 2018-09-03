package mapiserver

import (
	"encoding/json"
	"time"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

const (
	timeoutVolumesList = 5 * time.Second
	listVolumePath     = "/latest/volumes/"
)

// ListVolumes and return them as obj
func ListVolumes() (v1alpha1.CASVolumeList, error) {

	cvols := v1alpha1.CASVolumeList{}

	body, err := getRequest(GetURL()+listVolumePath, "", true)
	if err != nil {
		return cvols, err
	}

	err = json.Unmarshal(body, &cvols)

	return cvols, err
}
