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
		return err
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
		return fmt.Errorf("Status Error: %v", http.StatusText(code))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &obj)
}
