package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// NewReplicaClient create the new replica client
func NewReplicaClient(address string) (*ReplicaClient, error) {
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
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	syncAgent := strings.Replace(address, fmt.Sprintf(":%d", port), fmt.Sprintf(":%d", port+2), -1)

	timeout := time.Duration(2 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	return &ReplicaClient{
		Host:       parts[0],
		Address:    address,
		SyncAgent:  syncAgent,
		httpClient: client,
	}, nil
}

func (c *ReplicaClient) GetReplica() (InfoReplica, error) {
	var replica InfoReplica

	err := c.Get(c.Address+"/replicas/1", &replica)

	return replica, err
}

func (c *ReplicaClient) Get(url string, obj interface{}) error {
	if !strings.HasPrefix(url, "http") {
		url = c.Address + url
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(obj)
}

func (c *ReplicaClient) Post(path string, req, resp interface{}) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	bodyType := "application/json"
	url := path

	if !strings.HasPrefix(url, "http") {
		url = c.Address + path
	}

	httpResp, err := c.httpClient.Post(url, bodyType, bytes.NewBuffer(b))
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

func (c *ReplicaClient) MarkDiskAsRemoved(disk string) error {

	_, err := c.GetReplica()
	if err != nil {
		return err
	}
	//url := "/replicas/1?action=markdiskasremoved"
	url := "/replicas/1?action=removedisk"

	return c.Post(url, &MarkDiskAsRemovedInput{
		Name: disk,
	}, nil)
}
