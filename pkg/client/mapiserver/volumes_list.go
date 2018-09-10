package mapiserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// ListVolumes and return them as obj
func ListVolumes() (volumes v1alpha1.CASVolumeList, err error) {
	volumes = v1alpha1.CASVolumeList{}
	body, responseStatusCode, err := serverRequest(get, nil, GetURL()+volumePath, "")
	if err != nil {
		return
	} else if responseStatusCode != http.StatusOK {
		err = fmt.Errorf(string(body))
		return
	}

	err = json.Unmarshal(body, &volumes)
	return
}
