package mapiserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"time"
)

const (
	timeoutVolumesList = 5 * time.Second
	listVolumePath     = "/latest/volumes/"
)

// ListVolumes and return them as obj
func ListVolumes(obj interface{}) error {

	body, err := getRequest(GetURL()+listVolumePath, v1.DefaultNamespaceForListOps, false)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &obj)
}
