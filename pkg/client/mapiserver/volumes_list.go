package mapiserver

import (
	"encoding/json"

	"time"

	"github.com/openebs/maya/types/v1"
)

const (
	timeoutVolumesList = 5 * time.Second
)

// ListVolumes and return them as obj
func ListVolumes(obj interface{}) error {

	body, err := getRequest(GetURL()+"/latest/volumes/", v1.DefaultNamespaceForListOps, false)

	if err != nil {
		return err
	}

	return json.Unmarshal(body, &obj)
}
