package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/types/v1"
	policies_v1 "github.com/openebs/maya/volume/policies/v1"
	"github.com/openebs/maya/volume/provisioners"
)

const (
	// NamespaceKey is used in request headers to get the
	// namespace
	NamespaceKey string = "namespace"
)

// VolumeSpecificRequest is a http handler implementation. It deals with HTTP
// requests w.r.t a single Volume.
//
// TODO
//    Should it return specific types than interface{} ?
func (s *HTTPServer) volumeSpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {

	fmt.Println("[DEBUG] Processing", req.Method, "request")

	switch req.Method {
	case "PUT", "POST":
		return s.volumeAdd(resp, req)
	case "GET":
		return s.volumeSpecificGetRequest(resp, req)
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// VolumeSpecificGetRequest deals with HTTP GET request w.r.t a single Volume
func (s *HTTPServer) volumeSpecificGetRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Extract info from path after trimming
	path := strings.TrimPrefix(req.URL.Path, "/latest/volumes")

	// Is req valid ?
	if path == req.URL.Path {
		return nil, CodedError(405, ErrInvalidMethod)
	}

	switch {

	case strings.Contains(path, "/info/"):
		volName := strings.TrimPrefix(path, "/info/")
		return s.volumeRead(resp, req, volName)
	case strings.Contains(path, "/delete/"):
		volName := strings.TrimPrefix(path, "/delete/")
		return s.volumeDelete(resp, req, volName)
	case path == "/":
		return s.volumeList(resp, req)
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// VolumeList is the http handler that lists Volumes
func (s *HTTPServer) volumeList(resp http.ResponseWriter, req *http.Request) (interface{}, error) {

	glog.Infof("Processing Volume list request")

	// Get the namespace if provided
	ns := ""
	if req != nil {
		ns = req.Header.Get(NamespaceKey)
	}

	if ns == "" {
		// We shall override if empty. This seems to be simple enough
		// that works for most of the usecases.
		// Otherwise we need to introduce logic to decide for default
		// namespace depending on operation type.
		ns = v1.DefaultNamespaceForListOps
	}

	// Create a Volume
	vol := &v1.Volume{}
	vol.Namespace = ns

	// Pass through the policy enforcement logic
	policy, err := policies_v1.VolumeGenericPolicy()
	if err != nil {
		return nil, err
	}

	vol, err = policy.Enforce(vol)
	if err != nil {
		return nil, err
	}

	// Get the persistent volume provisioner instance
	pvp, err := provisioners.GetVolumeProvisioner(nil)
	if err != nil {
		return nil, err
	}

	// Set the volume provisioner profile to provisioner
	_, err = pvp.Profile(vol)
	if err != nil {
		return nil, err
	}

	lister, ok, err := pvp.Lister()
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("Volume list is not supported by '%s:%s'", pvp.Label(), pvp.Name())
	}

	l, err := lister.List()
	if err != nil {
		return nil, err
	}

	glog.Infof("Processed Volume list request successfully")

	return l, nil
}

// VolumeRead is the http handler that fetches the details of a Volume
func (s *HTTPServer) volumeRead(resp http.ResponseWriter, req *http.Request, volName string) (*v1.Volume, error) {

	glog.Infof("Processing Volume read request")

	if volName == "" {
		return nil, CodedError(400, fmt.Sprintf("Volume name is missing"))
	}

	// Get the namespace if provided
	ns := ""
	if req != nil {
		ns = req.Header.Get(NamespaceKey)
	}

	// Create a Volume
	vol := &v1.Volume{}
	vol.Name = volName
	vol.Namespace = ns

	// Pass through the policy enforcement logic
	policy, err := policies_v1.VolumeGenericPolicy()
	if err != nil {
		return nil, err
	}

	vol, err = policy.Enforce(vol)
	if err != nil {
		return nil, err
	}

	// Get persistent volume provisioner instance
	pvp, err := provisioners.GetVolumeProvisioner(nil)
	if err != nil {
		return nil, err
	}

	// Set the volume provisioner profile to provisioner
	_, err = pvp.Profile(vol)
	if err != nil {
		return nil, err
	}

	reader, ok := pvp.Reader()
	if !ok {
		return nil, fmt.Errorf("Volume read is not supported by '%s:%s'", pvp.Label(), pvp.Name())
	}

	// TODO
	// vol should not be passed again !!
	details, err := reader.Read(vol)
	if err != nil {
		return nil, err
	}

	if details == nil {
		return nil, CodedError(404, fmt.Sprintf("Volume '%s' not found", volName))
	}

	glog.Infof("Processed Volume read request successfully for '" + volName + "'")

	return details, nil
}

// VolumeDelete is the http handler that fetches the details of a Volume
func (s *HTTPServer) volumeDelete(resp http.ResponseWriter, req *http.Request, volName string) (interface{}, error) {

	glog.Infof("Processing Volume delete request")

	if volName == "" {
		return nil, CodedError(400, fmt.Sprintf("Volume name is missing"))
	}

	// Get the namespace if provided
	ns := ""
	if req != nil {
		ns = req.Header.Get(NamespaceKey)
	}

	// Create a Volume
	vol := &v1.Volume{}
	vol.Name = volName
	vol.Namespace = ns

	// Pass through the policy enforcement logic
	policy, err := policies_v1.VolumeGenericPolicy()
	if err != nil {
		return nil, err
	}

	vol, err = policy.Enforce(vol)
	if err != nil {
		return nil, err
	}

	// Get the persistent volume provisioner instance
	pvp, err := provisioners.GetVolumeProvisioner(nil)
	if err != nil {
		return nil, err
	}

	// Set the volume provisioner profile
	_, err = pvp.Profile(vol)
	if err != nil {
		return nil, err
	}

	remover, ok, err := pvp.Remover()
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("Volume delete is not supported by '%s:%s'", pvp.Label(), pvp.Name())
	}

	removed, err := remover.Remove()
	if err != nil {
		return nil, err
	}

	// If there was not any err & still no removal
	if !removed {
		return nil, CodedError(404, fmt.Sprintf("Volume '%s' not found", volName))
	}

	glog.Infof("Processed Volume delete request successfully for '" + volName + "'")

	return fmt.Sprintf("Volume '%s' deleted successfully", volName), nil
}

// VolumeAdd is the http handler that fetches the details of a Volume
func (s *HTTPServer) volumeAdd(resp http.ResponseWriter, req *http.Request) (interface{}, error) {

	glog.Infof("Processing Volume add request")

	vol := &v1.Volume{}

	// The yaml/json spec is decoded to vol struct
	if err := decodeBody(req, vol); err != nil {
		return nil, CodedError(400, err.Error())
	}

	// Name is expected to be available even in the minimalist specs
	if vol.Name == "" {
		return nil, CodedError(400, fmt.Sprintf("Volume name missing in '%v'", vol))
	}

	// Pass through the policy enforcement logic
	policy, err := policies_v1.VolumeAddPolicy()
	if err != nil {
		return nil, err
	}

	vol, err = policy.Enforce(vol)
	if err != nil {
		return nil, err
	}

	// Get persistent volume provisioner instance
	pvp, err := provisioners.GetVolumeProvisioner(nil)
	if err != nil {
		return nil, err
	}

	// Set the volume provisioner profile to provisioner
	_, err = pvp.Profile(vol)
	if err != nil {
		return nil, err
	}

	adder, ok := pvp.Adder()
	if !ok {
		return nil, fmt.Errorf("Volume add operation is not supported by '%s:%s'", pvp.Label(), pvp.Name())
	}

	// TODO
	// vol should not be passed again !!
	details, err := adder.Add(vol)
	if err != nil {
		return nil, err
	}

	glog.Infof("Processed Volume add request successfully for '" + vol.Name + "'")

	return details, nil
}
