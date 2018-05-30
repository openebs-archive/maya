package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/volume"
)

type volumeAPIOpsV1alpha1 struct {
	req  *http.Request
	resp http.ResponseWriter
}

// volumeV1alpha1SpecificRequest is a http handler to handle HTTP
// requests to a OpenEBS volume.
func (s *HTTPServer) volumeV1alpha1SpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	glog.Infof("cas template based volume request was received: method '%s'", req.Method)

	if req == nil {
		return nil, CodedError(400, "nil http request was received")
	}

	volOp := &volumeAPIOpsV1alpha1{
		req:  req,
		resp: resp,
	}

	switch req.Method {
	case "PUT", "POST":
		return volOp.create()
	case "GET":
		return volOp.httpGet()
	case "DELETE":
		return volOp.httpDelete()
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// get deals with HTTP GET request
func (v *volumeAPIOpsV1alpha1) httpGet() (interface{}, error) {
	// Extract info from path after trimming
	path := strings.TrimPrefix(v.req.URL.Path, "/latest/volumes")

	// check if req has valid get path
	if path == v.req.URL.Path {
		return nil, CodedError(405, ErrInvalidMethod)
	}

	switch {

	case strings.Contains(path, "/info/"):
		volName := strings.TrimPrefix(path, "/info/")
		// read a cas volume
		return v.read(volName)
	case path == "/":
		// list cas volumes
		return v.list()
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// httpDelete deals with http delete request method
func (v *volumeAPIOpsV1alpha1) httpDelete() (interface{}, error) {
	// Extract info from path after trimming
	path := strings.TrimPrefix(v.req.URL.Path, "/latest/volumes")

	// check if req has valid delete path
	if path == v.req.URL.Path {
		return nil, CodedError(405, ErrInvalidMethod)
	}

	switch {
	case strings.Contains(path, "/delete/"):
		volName := strings.TrimPrefix(path, "/delete/")
		// delete a cas volume
		return v.delete(volName)
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

func (v *volumeAPIOpsV1alpha1) create() (*v1alpha1.CASVolume, error) {
	glog.Infof("cas template based volume create request was received")

	vol := &v1alpha1.CASVolume{}
	err := decodeBody(v.req, vol)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	// volume name is expected
	if len(vol.Name) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create volume: missing volume name '%v'", vol))
	}

	// use run namespace from labels if volume's namespace is not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = vol.Labels[string(v1alpha1.NamespaceCVK)]
	}

	// use run namespace from http request header if volume's namespace is still not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = v.req.Header.Get(NamespaceKey)
	}

	vOps, err := volume.NewVolumeOperation(vol)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	cvol, err := vOps.Create()
	if err != nil {
		glog.Errorf("failed to create cas template based volume: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("cas template based volume created successfully: name '%s'", cvol.Name)
	return cvol, nil
}

func (v *volumeAPIOpsV1alpha1) read(volumeName string) (interface{}, error) {
	glog.Infof("cas template based volume read request was received")

	vol := &v1alpha1.CASVolume{}
	hdrNS := ""

	// get volume related details from http request
	if v.req != nil {
		decodeBody(v.req, vol)
		hdrNS = v.req.Header.Get(NamespaceKey)
	}

	vol.Name = volumeName

	// volume name is expected
	if len(vol.Name) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to read volume: missing volume name '%v'", vol))
	}

	// use namespace from labels if volume ns is not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = vol.Labels[string(v1alpha1.NamespaceCVK)]
	}

	// use namespace from req headers if volume ns is still not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = hdrNS
	}

	vOps, err := volume.NewVolumeOperation(vol)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	cvol, err := vOps.Read()
	if err != nil {
		glog.Errorf("failed to read cas template based volume: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("cas template based volume was read successfully: name '%s'", cvol.Name)
	return cvol, nil
}

func (v *volumeAPIOpsV1alpha1) delete(volumeName string) (interface{}, error) {
	glog.Infof("cas template based volume delete request was received")

	vol := &v1alpha1.CASVolume{}
	hdrNS := ""

	// get volume related details from http request
	if v.req != nil {
		decodeBody(v.req, vol)
		hdrNS = v.req.Header.Get(NamespaceKey)
	}

	vol.Name = volumeName

	// volume name is expected
	if len(vol.Name) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to delete volume: missing volume name '%v'", vol))
	}

	// use run namespace from labels if volume ns is not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = vol.Labels[string(v1alpha1.NamespaceCVK)]
	}

	// use run namespace from req headers if volume ns is still not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = hdrNS
	}

	vOps, err := volume.NewVolumeOperation(vol)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	cvol, err := vOps.Delete()
	if err != nil {
		glog.Errorf("failed to delete cas template based volume: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("cas template based volume was deleted successfully: name '%s'", cvol.Name)
	return cvol, nil
}

func (v *volumeAPIOpsV1alpha1) list() (interface{}, error) {
	glog.Infof("cas template based volume list request was received")

	vols := &v1alpha1.CASVolumeList{}
	hdrNS := ""

	// extract volume list details from http request
	if v.req != nil {
		decodeBody(v.req, vols)
		hdrNS = v.req.Header.Get(NamespaceKey)
	}

	// use run namespace from labels if volume ns is not set
	if len(vols.Namespace) == 0 {
		vols.Namespace = vols.Labels[string(v1alpha1.NamespaceCVK)]
	}

	// use run namespace from req headers if volume ns is still not set
	if len(vols.Namespace) == 0 {
		vols.Namespace = hdrNS
	}

	vOps, err := volume.NewVolumeListOperation(vols)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	cvols, err := vOps.List()
	if err != nil {
		glog.Errorf("failed to list cas template based volumes: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("cas template based volumes were listed successfully")
	return cvols, nil
}
