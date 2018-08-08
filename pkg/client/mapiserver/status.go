package mapiserver

import "github.com/openebs/maya/pkg/util"

const (
	getStatusPath = "/latest/meta-data/instance-id"
)

// GetStatus returns the status of maya-apiserver via http
func GetStatus() (string, error) {
	body, err := getRequest(GetURL()+getStatusPath, "", false)
	if err != nil {
		return "Connection failed", err
	}
	if string(body) != `"any-compute"` {
		err = util.ServerUnavailable
	}
	return string(body), err
}
