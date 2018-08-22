package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/mapiserver"

	"github.com/openebs/maya/pkg/util"
)

// Client interface defines the GetVolAnnotation method which can be used to get the details about the volumes from maya-apiserver.
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

// GetVolDetails gets response in json format of a volume from m-apiserver
func GetVolDetails(volName string, namespace string, obj interface{}) error {
	url := mapiserver.GetURL() + "/latest/volumes/" + volName

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("namespace", namespace)

	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("Can't get a response, error found: %v", err)
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
func (annotations *Annotations) GetVolAnnotations(volName string, namespace string) (v1alpha1.CASVolume, error) {
	var volume v1alpha1.CASVolume
	// var volumeold v1.Volume
	err := GetVolDetails(volName, namespace, &volume)
	if err != nil || volume.ObjectMeta.Annotations == nil {
		if volume.Status.Reason == "pending" {
			fmt.Println("VOLUME status Unknown to M_API server")
		}
		return volume, err
	}
	for key, value := range volume.ObjectMeta.Annotations {
		switch key {
		case "openebs.io/capacity":
			annotations.VolSize = value
		case "openebs.io/jiva-iqn":
			annotations.Iqn = value
		case "openebs.io/jiva-replica-count":
			annotations.ReplicaCount = value
		case "openebs.io/jiva-controller-cluster-ip":
			annotations.ClusterIP = value
		case "openebs.io/jiva-replica-ips":
			annotations.Replicas = value
		case "openebs.io/jiva-target-portal":
			annotations.TargetPortal = value
		case "openebs.io/jiva-controller-status":
			annotations.ControllerStatus = value
		case "openebs.io/jiva-replica-status":
			annotations.ReplicaStatus = value
		case "openebs.io/jiva-controller-ips":
			annotations.ControllerIP = value
		}
	}
	return volume, nil
}
