package mapiserver

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/openebs/maya/pkg/util"
)

// GetStatus returns the status of maya-apiserver via http
func GetStatus() (string, error) {

	var url bytes.Buffer
	addr := GetURL()

	if addr == "" {
		return "", util.ServerUnavailable
	}
	url.WriteString(addr + "/latest/meta-data/instance-id")
	resp, err := http.Get(url.String())

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}
	if string(body) != `"any-compute"` {
		err = util.ServerUnavailable
	}
	return string(body), err
}
