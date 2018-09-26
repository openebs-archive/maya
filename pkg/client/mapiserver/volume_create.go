package mapiserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	_, responseStatusCode, err := serverRequest(post, jsonValue, GetURL()+volumePath, "")
	if err != nil {
		return err
	} else if responseStatusCode != http.StatusOK {
		return fmt.Errorf("Server status error: %v", http.StatusText(responseStatusCode))
	}

	return nil
}
