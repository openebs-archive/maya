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
	volume_create_timeout = 5 * time.Second
)

// Create a volume by invoking the API call to m-apiserver
func CreateVolume(vname string, size string) error {

	_, err := GetStatus()
	if err != nil {
		return util.MAPIADDRNotSet
	}

	var vs v1.VolumeAPISpec

	vs.Kind = "Volume"
	vs.APIVersion = "v1"
	vs.Metadata.Name = vname
	vs.Metadata.Labels.Storage = size

	//Marshal serializes the value provided into a YAML document
	yamlValue, _ := yaml.Marshal(vs)

	//fmt.Printf("Volume Spec Created:\n%v\n", string(yamlValue))

	url := GetURL() + "/latest/volumes/"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(yamlValue))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/yaml")

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

	if code != http.StatusOK {
		return fmt.Errorf("Status error: %v", http.StatusText(code))
	}
	return nil
}
