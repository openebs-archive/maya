package mapiserver

import (
	"fmt"
	"net/http"
	"time"
)

const (
	timeoutVolumeDelete = 5 * time.Second
)

// DeleteVolume will request maya-apiserver to delete volume (vname)
func DeleteVolume(vname string) error {

	_, err := GetStatus()
	if err != nil {
		return fmt.Errorf("Unable to contact maya-apiserver: %s", GetURL())
	}

	url := GetURL() + "/latest/volumes/delete/" + vname
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	c := &http.Client{
		Timeout: timeoutVolumeDelete,
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	code := resp.StatusCode
	if code != http.StatusOK {
		return fmt.Errorf("Status error: %v", http.StatusText(code))
	}
	return nil
}
