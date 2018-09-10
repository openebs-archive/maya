package mapiserver

import (
	"net/http"

	"github.com/openebs/maya/pkg/util"
)

const (
	getStatusPath = "/latest/meta-data/instance-id"
)

// GetStatus returns the status of maya-apiserver via http
func GetStatus() (string, error) {
	body, responseStatusCode, err := serverRequest(get, nil, GetURL()+getStatusPath, "")
	if err != nil {
		return "Connection failed", err
	} else if responseStatusCode != http.StatusOK && string(body) != `"any-compute"` {
		err = util.ServerUnavailable
	}
	return string(body), err
}
