package mapiserver

import (
	"encoding/json"
	"time"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	volumeCreateTimeout = 60 * time.Second
	volumePath          = "/latest/volumes/"
)

// CreateVolume creates a volume by invoking the API call to m-apiserver
func CreateVolume(vname, size, namespace string) error {
	// Filling structure with values
	cVol := v1alpha1.CASVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vname,
			Namespace: namespace,
		},
		Spec: v1alpha1.CASVolumeSpec{
			Capacity: size,
		},
	}
	// Marshal serializes the value of vs structure
	jsonValue, err := json.Marshal(cVol)
	if err != nil {
		return err
	}
	_, err = postRequest(GetURL()+volumePath, jsonValue, "", false)
	return err
}
