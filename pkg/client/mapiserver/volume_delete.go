package mapiserver

import (
	"time"
)

const (
	timeoutVolumeDelete = 5 * time.Second

//	deleteVolumePath    = "/latest/volumes/delete/"
)

// DeleteVolume will request maya-apiserver to delete volume (vname)
func DeleteVolume(vname string, namespace string) error {
	err := deleteRequest(GetURL()+volumePath+vname, namespace)
	return err
}
