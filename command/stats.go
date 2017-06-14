package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Volume is a command implementation struct
type Volume struct {
	Spec struct {
		AccessModes interface{} `json:"AccessModes"`
		Capacity    interface{} `json:"Capacity"`
		ClaimRef    interface{} `json:"ClaimRef"`
		OpenEBS     struct {
			VolumeID string `json:"volumeID"`
		} `json:"OpenEBS"`
		PersistentVolumeReclaimPolicy string `json:"PersistentVolumeReclaimPolicy"`
		StorageClassName              string `json:"StorageClassName"`
	} `json:"Spec"`

	Status struct {
		Message string `json:"Message"`
		Phase   string `json:"Phase"`
		Reason  string `json:"Reason"`
	} `json:"Status"`
	Annotations       interface{} `json:"annotations"`
	CreationTimestamp interface{} `json:"creationTimestamp"`
	Name              string      `json:"name"`
}

// Annotations describes volume struct
type Annotations struct {
	VolSize      string   `json:"be.jiva.volume.openebs.io/vol-size"`
	VolAddr      string   `json:"fe.jiva.volume.openebs.io/ip"`
	Iqn          string   `json:"iqn"`
	Targetportal string   `json:"targetportal"`
	Replicas     []string `json:"JIVA_REP_IP_*"`
	ReplicaCount string   `json:"be.jiva.volume.openebs.io/count"`
}

const (
	timeout = 5 * time.Second
)

// getVolDetails gets response in json format of a volume from m-apiserver
func getVolDetails(volName string, obj interface{}) error {
	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		fmt.Println(err)
		return err
	}
	url := addr + "/latest/volumes/info/" + volName
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if resp != nil {
		if resp.StatusCode == 500 {
			fmt.Printf("VSM %s not found at M_API server\n", volName)
			return err
		} else if resp.StatusCode == 503 {
			fmt.Println("M_API server not reachable")
			return err
		}
	} else {
		fmt.Println("M_API server not reachable")
		return err
	}

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(obj)
}

// GetVolAnnotations gets annotations of volume
func GetVolAnnotations(volName string) (*Annotations, error) {
	var volume Volume
	var annotations Annotations
	err := getVolDetails(volName, &volume)
	if err != nil || volume.Annotations == nil {
		if volume.Status.Reason == "pending" {
			fmt.Println("VSM status Unknown to M_API server")
		}
		return nil, err
	}
	for key, value := range volume.Annotations.(map[string]interface{}) {
		switch key {
		case "be.jiva.volume.openebs.io/vol-size":
			annotations.VolSize = value.(string)
		case "fe.jiva.volume.openebs.io/ip":
			annotations.VolAddr = value.(string)
		case "iqn":
			annotations.Iqn = value.(string)
		case "be.jiva.volume.openebs.io/count":
			annotations.ReplicaCount = value.(string)
		case "JIVA_REP_IP_0":
			annotations.Replicas = append(annotations.Replicas, value.(string))
		case "JIVA_REP_IP_1":
			annotations.Replicas = append(annotations.Replicas, value.(string))

		}
	}
	return &annotations, nil
}
