package mapiserver

import (
	"time"
)

const (
	timeoutVolumeDelete = 5 * time.Second
	deleteVolumePath    = "/latest/volumes/delete/"
)

// DeleteVolume will request maya-apiserver to delete volume (vname)
func DeleteVolume(vname string, namespace string) error {
	_, err := getRequest(GetURL()+deleteVolumePath+vname, namespace, false)
	return err
}
