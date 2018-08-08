package mapiserver

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
	yaml "gopkg.in/yaml.v2"
)

const (
	volumeCreateTimeout = 5 * time.Second
	volumePath          = "/latest/volumes/"
)

// Create a volume by invoking the API call to m-apiserver
func CreateVolume(vname string, size string) error {

	_, err := GetStatus()
	if err != nil {
		return util.MAPIADDRNotSet
	}
	// Marshal serializes the value of vs structure
	jsonValue, _ := json.Marshal(vs)

	_, err := postRequest(GetURL()+volumePath, jsonValue, "", false)
	return err
}

	c := &http.Client{
		Timeout: volume_create_timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	code := resp.StatusCode

	_, err := postRequest(GetURL()+volumePath, jsonValue, "", false)
	return err
}
