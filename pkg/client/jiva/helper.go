package client

import (
	"encoding/json"
	"errors"
)

// GetStatus is the helper function for mayactl.It is used to get the response of
// the replica created in json format and then the response is then decoded to
// the desired structure.
func GetStatus(address string, obj interface{}) (int, error) {
	replica, err := NewReplicaClient(address)
	if err != nil {
		return -1, err
	}
	url := replica.Address + "/stats"
	resp, err := replica.httpClient.Get(url)
	if resp != nil {
		if resp.StatusCode == 500 {
			return 500, errors.New("Internal Server Error")
		} else if resp.StatusCode == 503 {
			return 503, errors.New("Service Unavailable")
		}
	} else {
		return -1, errors.New("Server Not Reachable")
	}
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	return 0, json.NewDecoder(resp.Body).Decode(obj)
}

// GetVolumeStats is used to get the status of volume controller.It is used to
// get the response in json format and then the response is then decoded to the
// desired structure.
func GetVolumeStats(address string, obj interface{}) (int, error) {
	controller, err := NewControllerClient(address)
	if err != nil {
		return -1, err
	}
	url := controller.Address + "/stats"
	resp, err := controller.httpClient.Get(url)
	if resp != nil {
		if resp.StatusCode == 500 {
			return 500, errors.New("Internal Server Error")
		} else if resp.StatusCode == 503 {
			return 503, errors.New("Service Unavailable")
		}
	} else {
		return -1, errors.New("Server Not Reachable")
	}
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	rc := json.NewDecoder(resp.Body).Decode(obj)
	return 0, rc
}
