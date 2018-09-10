package mapiserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// ReadVolume reads a volume info from m-api server and return them as CASVolume obj
func ReadVolume(volumeName, namespace string) (volume v1alpha1.CASVolume, err error) {
	volume = v1alpha1.CASVolume{}
	body, responseStatusCode, err := serverRequest(get, nil, GetURL()+volumePath+volumeName, namespace)
	if err != nil {
		return
	} else if responseStatusCode != http.StatusOK {
		if responseStatusCode == http.StatusInternalServerError {
			err = fmt.Errorf("Sorry something went wrong with service. Please raise an issue on: https://github.com/openebs/openebs/issues")
			return
		} else if responseStatusCode == http.StatusServiceUnavailable {
			err = fmt.Errorf("Maya apiservice not reachable at %q", GetURL())
			return
		} else if responseStatusCode == http.StatusNotFound {
			err = fmt.Errorf("Volume: %s not found at namespace: %q error: %s", volumeName, namespace, http.StatusText(responseStatusCode))
			return
		}
		err = fmt.Errorf("Received an error from maya apiservice: statuscode: %d", responseStatusCode)
		return
	}

	err = json.Unmarshal(body, &volume)
	return
}
