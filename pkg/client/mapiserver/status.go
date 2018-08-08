package mapiserver

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

const (
	getStatusPath = "/latest/meta-data/instance-id"
)

const (
	getStatusPath = "/latest/meta-data/instance-id"
)

// GetStatus returns the status of maya-apiserver via http
func GetStatus() (string, error) {
	body, err := getRequest(GetURL()+getStatusPath, "", false)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	return string(body[:]), err
}
