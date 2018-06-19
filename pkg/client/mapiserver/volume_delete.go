package mapiserver

import (
	"time"
)

const (
	timeoutVolumeDelete = 5 * time.Second
)

// DeleteVolume will request maya-apiserver to delete volume (vname)
func DeleteVolume(vname string, namespace string) error {
	_, err := getRequest(GetURL()+"/latest/volumes/delete/"+vname, namespace, false)
	return err
}
