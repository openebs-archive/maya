package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/openebs/maya/pkg/template"
	"github.com/openebs/maya/pkg/usage"
	"github.com/openebs/maya/pkg/volume"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	// NamespaceKey is used in request headers to get the
	// namespace
	NamespaceKey string = "namespace"
)

func isNotFound(err error) bool {
	if _, ok := err.(*template.NotFoundError); ok {
		return ok
	}

	return errors.IsNotFound(err)
}

type volumeAPIOpsV1alpha1 struct {
	req  *http.Request
	resp http.ResponseWriter
}

func volumeEvents(cvol *v1alpha1.CASVolume, method string, err error) {
	if menv.Truthy(menv.OpenEBSEnableAnalytics) {
		sendObj := usage.New().Build().ApplicationBuilder().
			SetApplicationName(cvol.Spec.CasType).
			SetDocumentTitle(cvol.ObjectMeta.Name).
			SetLabel("Capacity")
		if method == "create" {
			sendObj.SetCategory("volume-provision-replica-count:" + cvol.Spec.Replicas)
		} else if method == "delete" {
			sendObj.SetCategory("volume-deprovision-replica-count:" + cvol.Spec.Replicas)
		}
		if err != nil {
			sendObj.SetAction(err.Error())
		} else {
			sendObj.SetAction("success")
		}
		sendObj.SetVolumeCapacity(cvol.Spec.Capacity).Send()
	}

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
	case "POST":
		return volOp.create()
	case "GET":
		return volOp.httpGet()
	case "DELETE":
		return volOp.httpDelete()
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// httpGet deals with http GET request
func (v *volumeAPIOpsV1alpha1) httpGet() (interface{}, error) {
	// Extract name of volume from path after trimming
	path := strings.TrimSpace(strings.TrimPrefix(v.req.URL.Path, "/latest/volumes"))
	// list cas volumes
	if path == "/" {
		return v.list()
	} else if strings.Contains(path, "/stats/") {
		return v.readStats(strings.TrimPrefix(path, "/stats/"))
	}

	// read a cas volume
	volName := strings.TrimPrefix(path, "/")
	return v.read(volName)
}

// httpDelete deals with http DELETE request
func (v *volumeAPIOpsV1alpha1) httpDelete() (interface{}, error) {
	// Extract name of volume from path after trimming
	volName := strings.TrimSpace(strings.TrimPrefix(v.req.URL.Path, "/latest/volumes/"))

	// check if req url has volume name
	if len(volName) == 0 {
		return nil, CodedError(405, ErrInvalidMethod)
	}

	return v.delete(volName)
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

	// use run namespace from http request header if volume's namespace is still not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = v.req.Header.Get(NamespaceKey)
	}

	vOps, err := volume.NewOperation(vol)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	cvol, err := vOps.Create()
	volumeEvents(cvol, "create", err)
	if err != nil {
		glog.Errorf("failed to create cas template based volume: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}
	glog.Infof("cas template based volume created successfully: name '%s'", cvol.Name)

	return cvol, nil
}

func (v *volumeAPIOpsV1alpha1) read(volumeName string) (*v1alpha1.CASVolume, error) {
	glog.Infof("cas template based volume read request was received")

	vol := &v1alpha1.CASVolume{}
	// hdrNS will store namespace from http header
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

	// use namespace from req headers if volume ns is still not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = hdrNS
	}

	// use StorageClass name from header if present
	scName := strings.TrimSpace(v.req.Header.Get(string(v1alpha1.StorageClassHeaderKey)))
	// add the StorageClass name to volume's labels
	vol.Labels = map[string]string{
		string(v1alpha1.StorageClassKey): scName,
	}

	vOps, err := volume.NewOperation(vol)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	cvol, err := vOps.Read()
	if err != nil {
		glog.Errorf("failed to read cas template based volume: error '%s'", err.Error())
		if isNotFound(err) {
			return nil, CodedError(404, fmt.Sprintf("volume '%s' not found at namespace '%s'", vol.Name, vol.Namespace))
		}
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("cas template based volume was read successfully: name '%s'", cvol.Name)
	return cvol, nil
}

func (v *volumeAPIOpsV1alpha1) delete(volumeName string) (*v1alpha1.CASVolume, error) {
	glog.Infof("cas template based volume delete request was received")

	vol := &v1alpha1.CASVolume{}
	// hdrNS will store namespace from http header
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

	// use namespace from req headers if volume ns is still not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = hdrNS
	}

	vOps, err := volume.NewOperation(vol)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	cvol, err := vOps.Delete()
	volumeEvents(cvol, "delete", err)
	if err != nil {
		glog.Errorf("failed to delete cas template based volume: error '%s'", err.Error())
		if isNotFound(err) {
			return nil, CodedError(404, fmt.Sprintf("volume '%s' not found at namespace '%s'", vol.Name, vol.Namespace))
		}
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("cas template based volume was deleted successfully: name '%s'", cvol.Name)
	return cvol, nil
}

func (v *volumeAPIOpsV1alpha1) list() (*v1alpha1.CASVolumeList, error) {
	glog.Infof("cas template based volume list request was received")

	vols := &v1alpha1.CASVolumeList{}
	// hdrNS will store namespace from http header
	hdrNS := ""

	// extract volume list details from http request
	if v.req != nil {
		decodeBody(v.req, vols)
		hdrNS = v.req.Header.Get(NamespaceKey)
	}

	// use namespace from req headers if volume ns is still not set
	if len(vols.Namespace) == 0 {
		vols.Namespace = hdrNS
	}

	vOps, err := volume.NewListOperation(vols)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	cvols, err := vOps.List()
	if err != nil {
		glog.Errorf("failed to list cas template based volumes at namespaces '%s': error '%s'", vols.Namespace, err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("cas template based volumes were listed successfully: namespaces '%s'", vols.Namespace)
	return cvols, nil
}

func (v *volumeAPIOpsV1alpha1) readStats(volumeName string) (interface{}, error) {
	glog.Infof("CASTemplate based volume stats request received")
	vol := &v1alpha1.CASVolume{}
	// hdrNS will store namespace from http header
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

	// use namespace from req headers if volume ns is still not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = hdrNS
	}

	// use StorageClass name from header if present
	scName := strings.TrimSpace(v.req.Header.Get(string(v1alpha1.StorageClassHeaderKey)))
	// add the StorageClass name to volume's labels
	vol.Labels = map[string]string{
		string(v1alpha1.StorageClassKey): scName,
	}

	vOps, err := volume.NewOperation(vol)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}
	stats, err := vOps.ReadStats()
	if err != nil {
		glog.Errorf("failed to read cas template based volume: error '%s'", err.Error())
		if isNotFound(err) {
			return nil, CodedError(404, fmt.Sprintf("volume '%s' not found at namespace '%s'", vol.Name, vol.Namespace))
		}
		return nil, CodedError(500, err.Error())
	}

	// pipelining the response
	v.resp.Write(stats)
	glog.Infof("cas template based volume stats read successful '%s'", volumeName)
	return nil, err
}
