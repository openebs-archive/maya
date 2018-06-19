package mapiserver

import (
	"github.com/openebs/maya/pkg/util"
)

// GetStatus returns the status of maya-apiserver via http
func GetStatus() (string, error) {
	body, err := getRequest(GetURL()+"/latest/meta-data/instance-id", "", false)
	if err != nil {
		return "Connection failed", err
	}
	if string(body) != `"any-compute"` {
		err = util.ServerUnavailable
	}
	return string(body), err
}
