package mapiserver

import (
	"encoding/json"
	"time"

	"github.com/openebs/maya/types/v1"
)

const (
	volumeCreateTimeout = 5 * time.Second
	volumePath          = "/latest/volumes/"
)

// CreateVolume creates a volume by invoking the API call to m-apiserver
func CreateVolume(vname, size string) error {
	// Filling structure with values
	vs := v1.Volume{
		TypeMeta: v1.TypeMeta{
			Kind:       "Volume",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: vname,
		},
		Capacity: size,
	}

	// Marshal serializes the value of vs structure
	jsonValue, err := json.Marshal(vs)
	if err != nil {
		return err
	}
	_, err = postRequest(GetURL()+volumePath, jsonValue, "", false)
	return err
}
