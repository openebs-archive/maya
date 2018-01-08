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
	Metadata struct {
		Annotations       interface{} `json:"annotations"`
		CreationTimestamp interface{} `json:"creationTimestamp"`
		Name              string      `json:"name"`
	} `json:"metadata"`
}

// Annotations describes volume struct
type Annotations struct {
	TargetPortal     string `json:"vsm.openebs.io/targetportals"`
	ClusterIP        string `json:"vsm.openebs.io/cluster-ips"`
	Iqn              string `json:"vsm.openebs.io/iqn"`
	ReplicaCount     string `json:"vsm.openebs.io/replica-count"`
	ControllerStatus string `json:"vsm.openebs.io/controller-status"`
	ReplicaStatus    string `json:"vsm.openebs.io/replica-status"`
	VolSize          string `json:"vsm.openebs.io/volume-size"`
	ControllerIP     string `json:"vsm.openebs.io/controller-ips"`
	//	VolAddr          string `json:"vsm.openebs.io/replica-ips"`
	Replicas string `json:"vsm.openebs.io/replica-ips"`
}

const (
	timeout = 5 * time.Second
)

// getVolDetails gets response in json format of a volume from m-apiserver
func GetVolDetails(volName string, obj interface{}) error {
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
			fmt.Printf("Volume: %s not found at M_API server\n", volName)
			return errors.New("Internal Server Error")
		} else if resp.StatusCode == 503 {
			fmt.Println("M_API server not reachable")
			return errors.New("Service Unavailable")
		} else if resp.StatusCode == 404 {
			fmt.Printf("Volume: %s not found at M_API server\n", volName)
			return errors.New("Page Not Found")
		}

	} else {
		fmt.Println("M_API server not reachable")
		return err
	}

	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(obj)
}

// GetVolAnnotations gets annotations of volume
func GetVolAnnotations(volName string) (*Annotations, error) {
	var volume Volume
	var annotations Annotations
	err := GetVolDetails(volName, &volume)
	if err != nil || volume.Metadata.Annotations == nil {
		if volume.Status.Reason == "pending" {
			fmt.Println("VOLUME status Unknown to M_API server")
		}
		return nil, err
	}
	for key, value := range volume.Metadata.Annotations.(map[string]interface{}) {
		switch key {
		case "vsm.openebs.io/volume-size":
			annotations.VolSize = value.(string)
			//	case "fe.jiva.volume.openebs.io/ip":
			//		annotations.VolAddr = value.(string)
		case "vsm.openebs.io/iqn":
			annotations.Iqn = value.(string)
		case "vsm.openebs.io/replica-count":
			annotations.ReplicaCount = value.(string)
		case "vsm.openebs.io/cluster-ips":
			annotations.ClusterIP = value.(string)
		case "vsm.openebs.io/replica-ips":
			annotations.Replicas = value.(string)
		case "vsm.openebs.io/targetportals":
			annotations.TargetPortal = value.(string)
		case "vsm.openebs.io/controller-status":
			annotations.ControllerStatus = value.(string)
		case "vsm.openebs.io/replica-status":
			annotations.ReplicaStatus = value.(string)
		case "vsm.openebs.io/controller-ips":
			annotations.ControllerIP = value.(string)
		}
	}
	return &annotations, nil
}

func GetVolumeSpec(volName string) (*Annotations, error) {
	var volume Volume
	var annotations Annotations

	for key, value := range volume.Metadata.Annotations.(map[string]interface{}) {
		switch key {
		case "vsm.openebs.io/volume-size":
			annotations.VolSize = value.(string)
			//  case "fe.jiva.volume.openebs.io/ip":
			//      annotations.VolAddr = value.(string)
		case "vsm.openebs.io/iqn":
			annotations.Iqn = value.(string)
		case "vsm.openebs.io/replica-count":
			annotations.ReplicaCount = value.(string)
		case "vsm.openebs.io/cluster-ips":
			annotations.ClusterIP = value.(string)
		case "vsm.openebs.io/replica-ips":
			annotations.Replicas = value.(string)
		case "vsm.openebs.io/targetportals":
			annotations.TargetPortal = value.(string)
		case "vsm.openebs.io/controller-status":
			annotations.ControllerStatus = value.(string)
		case "vsm.openebs.io/controller-ips":
			annotations.ControllerIP = value.(string)
		}
	}
	return &annotations, nil
}
