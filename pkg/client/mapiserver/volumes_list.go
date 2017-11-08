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
)

// ListVolumes and return them as obj
func ListVolumes(obj interface{}) error {

	_, err := GetStatus()
	if err != nil {
		return fmt.Errorf("Unable to contact maya-apiserver: %s", GetURL())
	}

	url := GetURL() + "/latest/volumes/"
	client := &http.Client{
		Timeout: timeoutVolumesList,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	code := resp.StatusCode
	if code != http.StatusOK {
		return fmt.Errorf("Status error: %v", http.StatusText(code))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &obj)
	if err != nil {
		return err
	}

	return nil
}
