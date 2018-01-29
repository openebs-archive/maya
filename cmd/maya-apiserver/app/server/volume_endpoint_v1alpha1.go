package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/volume"
	"github.com/openebs/maya/types/v1"
)

type volumeAPIOpsV1alpha1 struct {
	req  *http.Request
	resp http.ResponseWriter
}

// volumeV1alpha1SpecificRequest is a http handler to handle HTTP
// requests to a OpenEBS volume.
func (s *HTTPServer) volumeV1alpha1SpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {

	//fmt.Println("[DEBUG] processing v1alpha1", req.Method, "request")
	glog.Infof("Received volume request: Method: '%s' Version: 'v1alpha1'", req.Method)

	if req == nil {
		return nil, CodedError(400, "Nil http request")
	}

	volOp := &volumeAPIOpsV1alpha1{
		req:  req,
		resp: resp,
	}

	switch req.Method {
	case "PUT", "POST":
		return volOp.create()
	case "GET":
		return volOp.get()
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// TODO Move the delete out to HTTP DELETE
//
// get deals with HTTP GET request
func (v *volumeAPIOpsV1alpha1) get() (interface{}, error) {
	// Extract info from path after trimming
	path := strings.TrimPrefix(v.req.URL.Path, "/0.6.0/volumes")

	// Is req valid ?
	if path == v.req.URL.Path {
		return nil, CodedError(405, ErrInvalidMethod)
	}

	switch {

	case strings.Contains(path, "/info/"):
		volName := strings.TrimPrefix(path, "/info/")
		return v.read(volName)
	case strings.Contains(path, "/delete/"):
		volName := strings.TrimPrefix(path, "/delete/")
		return v.delete(volName)
	case path == "/":
		return v.list()
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

func (v *volumeAPIOpsV1alpha1) create() (interface{}, error) {

	glog.Infof("Received volume add request: Version: 'v1alpha1'")

	vol := &v1.Volume{}

	// unmarshall the request to vol
	if err := decodeBody(v.req, vol); err != nil {
		return nil, CodedError(400, err.Error())
	}

	// Name is expected
	if vol.Name == "" {
		return nil, CodedError(400, fmt.Sprintf("Missing volume name '%v'", vol))
	}

	vOps, err := volume.NewVolumeOperation(vol)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	vol, err = vOps.Create()
	if err != nil {
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("Volume added successfully: Name: '%s' Version: 'v1alpha1'", vol.Name)

	return vol, nil
}

func (v *volumeAPIOpsV1alpha1) read(volumeName string) (interface{}, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (v *volumeAPIOpsV1alpha1) delete(volumeName string) (interface{}, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (v *volumeAPIOpsV1alpha1) list() (interface{}, error) {
	return nil, fmt.Errorf("Not implemented")
}
