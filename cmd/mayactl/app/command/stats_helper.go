package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
)

type Client interface {
	GetVolAnnotations(string) (*Annotations, error)
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
	Replicas         string `json:"vsm.openebs.io/replica-ips"`
}

const (
	timeout = 5 * time.Second
)

// getVolDetails gets response in json format of a volume from m-apiserver
func GetVolDetails(volName string, obj interface{}) error {
	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := util.MAPIADDRNotSet
		fmt.Printf("error getting env variable: %v", err)
		return err
	}

	url := addr + "/latest/volumes/info/" + volName
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)

	if err != nil {
		fmt.Printf("Could not get response, found error: %v", err)
		return err
	}

	if resp != nil {
		if resp.StatusCode == 500 {
			fmt.Printf("Volume: %s not found at M_API server\n", volName)
			return util.InternalServerError
		} else if resp.StatusCode == 503 {
			fmt.Println("M_API server not reachable")
			return util.ServerUnavailable
		} else if resp.StatusCode == 404 {
			fmt.Printf("Volume: %s not found at M_API server\n", volName)
			return util.PageNotFound
		}

	} else {
		fmt.Println("M_API server not reachable")
		return err
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(obj)
}

// GetVolAnnotations maps annotations of volume to Annotations structure.
func (annotations *Annotations) GetVolAnnotations(volName string) error {
	var volume v1.Volume
	err := GetVolDetails(volName, &volume)
	if err != nil || volume.ObjectMeta.Annotations == nil {
		if volume.Status.Reason == "pending" {
			fmt.Println("VOLUME status Unknown to M_API server")
		}
		return err
	}
	for key, value := range volume.ObjectMeta.Annotations {
		switch key {
		case "vsm.openebs.io/volume-size":
			annotations.VolSize = value
		case "vsm.openebs.io/iqn":
			annotations.Iqn = value
		case "vsm.openebs.io/replica-count":
			annotations.ReplicaCount = value
		case "vsm.openebs.io/cluster-ips":
			annotations.ClusterIP = value
		case "vsm.openebs.io/replica-ips":
			annotations.Replicas = value
		case "vsm.openebs.io/targetportals":
			annotations.TargetPortal = value
		case "vsm.openebs.io/controller-status":
			annotations.ControllerStatus = value
		case "vsm.openebs.io/replica-status":
			annotations.ReplicaStatus = value
		case "vsm.openebs.io/controller-ips":
			annotations.ControllerIP = value
		}
	}
	return nil
}
