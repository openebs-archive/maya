package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/openebs/maya/pkg/util"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NewControllerClient create the new controller client
func NewControllerClient(address string) (*ControllerClient, error) {
	address = strings.TrimPrefix(address, "tcp://")

	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}

	if !strings.HasSuffix(address, "/v1") {
		address += "/v1"
	}

	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(u.Host, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("Invalid address %s, must have a port in it", address)
	}

	return &ControllerClient{
		Host:       parts[0],
		Address:    address,
		httpClient: &http.Client{Timeout: 2 * time.Second},
	}, nil
}

// Post sends a POST request to the specified path and stores body in the value
// pointed to by resp.
func (c *ControllerClient) Post(path string, req, resp interface{}) error {
	return c.Do("POST", path, req, resp)
}

// Do sends a request to the specified path and it stores JSON-decoded body
// from the response into the value pointed to by resp.
func (c *ControllerClient) Do(method, path string, req, resp interface{}) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	bodyType := "application/json"
	url := path
	if !strings.HasPrefix(url, "http") {
		url = c.Address + path

	}

	httpReq, err := http.NewRequest(method, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", bodyType)

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 300 {
		content, _ := ioutil.ReadAll(httpResp.Body)
		return fmt.Errorf("Bad response: %d %s: %s", httpResp.StatusCode, httpResp.Status, content)
	}

	if resp == nil {
		return nil
	}
	return json.NewDecoder(httpResp.Body).Decode(resp)
}

// GetVolume returns Volumes from the specified path.
func GetVolume(path string) (*Volumes, error) {
	var volume VolumeCollection
	var c ControllerClient

	err := c.Get(path+"/volumes", &volume)
	if err != nil {
		return nil, err
	}

	if len(volume.Data) == 0 {
		return nil, errors.New("No volume found")
	}

	return &volume.Data[0], nil
}

// Get sends a request to the specified path and stores body in the value
// pointed to by obj.
func (c *ControllerClient) Get(path string, obj interface{}) error {
	resp, err := http.Get(path)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(obj)
}

// ListReplicas to get the details of all the existing replicas
// which contains address and mode of those replicas (RW/R/W) as well as
// resource information.
func (c *ControllerClient) ListReplicas(path string) ([]Replica, error) {
	var resp ReplicaCollection

	err := c.Get(path+"/replicas", &resp)

	return resp.Data, err
}

// GetVolumeStats is used to get the status of volume controller.It is used to
// get the response in json format and then the response is then decoded to the
// desired structure.
func (c *ControllerClient) GetVolumeStats(address string, api string, obj interface{}) (int, error) {
	controller, err := NewControllerClient(address)
	if err != nil {
		return -1, err
	}
	url := controller.Address + api
	resp, err := controller.httpClient.Get(url)
	if resp != nil {
		if resp.StatusCode == 500 {
			return 500, util.ErrInternalServerError
		} else if resp.StatusCode == 503 {
			return 503, util.ErrServerUnavailable
		}
	} else {
		return -1, util.ErrServerNotReachable
	}
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	rc := json.NewDecoder(resp.Body).Decode(obj)
	return 0, rc
}
