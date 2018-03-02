package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/openebs/maya/pkg/tracing"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

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

	tracer, closer := tracing.Init("list volumes handler (m-apiserver)")
	defer closer.Close()
	glog.Infof("Processing Volume list request")
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	span := tracer.StartSpan("list volume", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	listVolume := span.BaggageItem("operation")

	if listVolume == "" {
		listVolume = "list-volume"
	}

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
		span.LogFields(
			otlog.String("event", "pass policy enforcement"),
			otlog.Error(err),
		)
		return nil, err
	}

	vol, err = policy.Enforce(vol)
	if err != nil {
		span.LogFields(
			otlog.String("event", "enforce policies to volume"),
			otlog.Error(err),
		)
		return nil, err
	}

	// Get the persistent volume provisioner instance
	pvp, err := provisioners.GetVolumeProvisioner(nil)
	if err != nil {
		span.LogFields(
			otlog.String("event", "get provisioner instance"),
			otlog.Error(err),
		)
		return nil, err
	}
	span.LogFields(
		otlog.String("event", "get provisioner instance"),
		otlog.Object("provisioner", pvp),
	)

	// Set the volume provisioner profile to provisioner
	_, err = pvp.Profile(vol)
	if err != nil {
		span.LogFields(
			otlog.String("event", "set provisioner profile"),
			otlog.Error(err),
		)
		return nil, err
	}

	lister, ok, err := pvp.Lister()
	if err != nil {
		span.LogFields(
			otlog.String("event", "list provisioner"),
			otlog.Error(err),
		)
		return nil, err
	}

	if !ok {
		span.LogFields(
			otlog.String("event", "list provisioner"),
			otlog.Error(fmt.Errorf("Volume list is not supported by '%s:%s'", pvp.Label(), pvp.Name())),
		)
		return nil, fmt.Errorf("Volume list is not supported by '%s:%s'", pvp.Label(), pvp.Name())
	}

	l, err := lister.List()
	if err != nil {
		span.LogFields(
			otlog.String("event", "list provisioner"),
			otlog.Error(err),
		)
		return nil, err
	}
	span.LogFields(
		otlog.String("event", "list volume"),
		otlog.Bool("success", true),
		otlog.Object("value", l),
	)

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
	tracer, closer := tracing.Init("delete volume handler (mapiserver)")
	defer closer.Close()

	glog.Infof("Processing Volume delete request")
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	span := tracer.StartSpan("delete volume", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	deleteVolume := span.BaggageItem("operation")

	if deleteVolume == "" {
		deleteVolume = "delete-volume"
	}

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
		span.LogFields(
			otlog.String("event", "pass policy enforcement"),
			otlog.Error(err),
		)
		return nil, err
	}

	vol, err = policy.Enforce(vol)
	if err != nil {
		span.LogFields(
			otlog.String("event", "enforce policies to volume"),
			otlog.Error(err),
		)
		return nil, err
	}

	// Get the persistent volume provisioner instance
	pvp, err := provisioners.GetVolumeProvisioner(nil)
	if err != nil {
		span.LogFields(
			otlog.String("event", "get provisioner instance"),
			otlog.Error(err),
		)
		return nil, err
	}

	span.LogFields(
		otlog.String("event", "get provisioner instance"),
		otlog.Object("provisioner", pvp),
	)

	// Set the volume provisioner profile
	_, err = pvp.Profile(vol)
	if err != nil {
		span.LogFields(
			otlog.String("event", "set provisioner profile"),
			otlog.Error(err),
		)
		return nil, err
	}

	remover, ok, err := pvp.Remover()
	if err != nil {
		span.LogFields(
			otlog.String("event", "instantiate volume remover"),
			otlog.Error(err),
		)
		return nil, err
	}

	if !ok {
		span.LogFields(
			otlog.String("event", "instantiate volume remover"),
			otlog.Error(err),
		)
		return nil, fmt.Errorf("Volume delete is not supported by '%s:%s'", pvp.Label(), pvp.Name())
	}

	removed, err := remover.Remove()
	if err != nil {
		span.LogFields(
			otlog.String("event", "remove volume"),
			otlog.Error(err),
		)
		return nil, err
	}

	// If there was not any err & still no removal
	if !removed {
		span.LogFields(
			otlog.String("event", "get volume details"),
			otlog.Object("vol-details", removed),
			otlog.Error(fmt.Errorf("Volume '%s' not found", volName)),
		)
		return nil, CodedError(404, fmt.Sprintf("Volume '%s' not found", volName))
	}
	span.LogFields(
		otlog.String("event", "delete volume"),
		otlog.Bool("success", true),
		otlog.Object("volume-name", volName),
	)

	glog.Infof("Processed Volume delete request successfully for '" + volName + "'")

	return fmt.Sprintf("Volume '%s' deleted successfully", volName), nil
}

// VolumeAdd is the http handler that fetches the details of a Volume
func (s *HTTPServer) volumeAdd(resp http.ResponseWriter, req *http.Request) (interface{}, error) {

	tracer, closer := tracing.Init("create volume handler (m-apiserver)")
	defer closer.Close()
	glog.Infof("Processing Volume add request")

	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	span := tracer.StartSpan("create", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	createVolume := span.BaggageItem("operation")

	if createVolume == "" {
		createVolume = "create-volume"
	}

	vol := &v1.Volume{}

	// The yaml/json spec is decoded to vol struct
	if err := decodeBody(req, vol); err != nil {

		span.LogFields(
			otlog.String("event", "decode-request"),
			otlog.Error(err),
		)
		return nil, CodedError(400, err.Error())
	}

	// Name is expected to be available even in the minimalist specs
	if vol.Name == "" {
		span.LogFields(
			otlog.String("event", "get volume name"),
			otlog.Error(CodedError(400, fmt.Sprintf("Volume name missing in '%v'", vol))),
		)
		return nil, CodedError(400, fmt.Sprintf("Volume name missing in '%v'", vol))
	}

	// Pass through the policy enforcement logic
	policy, err := policies_v1.VolumeAddPolicy()
	if err != nil {

		span.LogFields(
			otlog.String("event", "pass policy enforcement"),
			otlog.Error(err),
		)
		return nil, err
	}

	vol, err = policy.Enforce(vol)
	if err != nil {

		span.LogFields(
			otlog.String("event", "enforce policies to volume"),
			otlog.Error(err),
		)
		return nil, err
	}

	// Get persistent volume provisioner instance
	pvp, err := provisioners.GetVolumeProvisioner(nil)
	if err != nil {

		span.LogFields(
			otlog.String("event", "get provisioner instance"),
			otlog.Error(err),
		)
		return nil, err
	}
	span.LogFields(
		otlog.String("event", "get provisioner instance"),
		otlog.Object("provisioner", pvp),
	)

	// Set the volume provisioner profile to provisioner
	_, err = pvp.Profile(vol)
	if err != nil {

		span.LogFields(
			otlog.String("event", "set provisioner profile"),
			otlog.Error(err),
		)
		return nil, err
	}

	adder, ok := pvp.Adder()
	if !ok {

		span.LogFields(
			otlog.String("event", "add volume to provisioner"),
			otlog.Error(fmt.Errorf("Volume add operation is not supported by '%s:%s'", pvp.Label(), pvp.Name())),
		)
		return nil, fmt.Errorf("Volume add operation is not supported by '%s:%s'", pvp.Label(), pvp.Name())
	}

	// TODO
	// vol should not be passed again !!
	details, err := adder.Add(vol)
	if err != nil {

		span.LogFields(
			otlog.String("event", "get volume details"),
			otlog.Error(err),
		)
		return nil, err
	}
	span.LogFields(
		otlog.String("event", "get volume details"),
		otlog.Bool("success", true),
		otlog.Object("vol-details", details),
	)

	glog.Infof("Processed Volume add request successfully for '" + vol.Name + "'")

	return details, nil
}
