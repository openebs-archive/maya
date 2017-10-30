package mapiserver

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Get the status of maya-apiserver via http
func GetStatus() (string, error) {

	//getting the m-apiserver env variable
	addr := os.Getenv("MAPI_ADDR")

	var url bytes.Buffer
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

	return string(body[:]), err
}
