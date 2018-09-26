package mapiserver

import (
	"fmt"
	"net/http"
)

// DeleteVolume will request maya-apiserver to delete volume (vname)
func DeleteVolume(vname string, namespace string) error {
	_, responseStatusCode, err := serverRequest(get, nil, GetURL()+volumePath+vname, namespace)
	if err != nil {
		return err
	} else if responseStatusCode != http.StatusOK {
		return fmt.Errorf("Server status error: %v", http.StatusText(responseStatusCode))
	}
	return nil
}
