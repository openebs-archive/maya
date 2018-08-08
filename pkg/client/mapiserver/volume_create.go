package mapiserver

import (
	"encoding/json"
	"time"

	"github.com/openebs/maya/types/v1"
)

const (
	volumeCreateTimeout = 5 * time.Second
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
	jsonValue, _ := json.Marshal(vs)

	_, err := postRequest(GetURL()+"/latest/volumes/", jsonValue, "", false)
	return err
}

// CreateCloneVolume clone a volume by invoking the API call to m-apiserver
func CreateCloneVolume(vname, size, snapshotname, sourcevolume string) error {
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
		VolumeClone: v1.VolumeClone{
			Clone:        true,
			SourceVolume: sourcevolume,
			SnapshotName: snapshotname,
		},
	}

	// Marshal serializes the value of vs structure
	jsonValue, _ := json.Marshal(vs)

	_, err := postRequest(GetURL()+"/latest/volumes/", jsonValue, "", false)
	return err
}
