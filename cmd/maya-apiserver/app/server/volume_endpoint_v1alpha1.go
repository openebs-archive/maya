/*
Copyright 2019 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	templatefuncs "github.com/openebs/maya/pkg/templatefuncs/v1alpha1"
	"github.com/openebs/maya/pkg/usage"
	"github.com/openebs/maya/pkg/volume"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	// NamespaceKey is used in request headers to
	// get namespace
	NamespaceKey string = "namespace"
)

// isNotFound returns true if the original
// cause of error was due to castemplate's
// not found error or kubernetes not found
// error
func isNotFound(err error) bool {
	switch err := errors.Cause(err).(type) {
	case *templatefuncs.NotFoundError:
		return true
	default:
		return k8serrors.IsNotFound(err)
	}
}

type volumeAPIOpsV1alpha1 struct {
	req  *http.Request
	resp http.ResponseWriter
}

// sendEventOrIgnore sends anonymous volume (de)-provision events
func sendEventOrIgnore(cvol *v1alpha1.CASVolume, method string) {
	if menv.Truthy(menv.OpenEBSEnableAnalytics) && cvol != nil {
		usage.New().Build().ApplicationBuilder().
			SetVolumeType(cvol.Spec.CasType, method).
			SetDocumentTitle(cvol.ObjectMeta.Name).
			SetLabel(usage.EventLabelCapacity).
			SetReplicaCount(cvol.Spec.Replicas, method).
			SetCategory(method).
			SetVolumeCapacity(cvol.Spec.Capacity).Send()
	}
}

// volumeV1alpha1SpecificRequest is a http handler to handle HTTP
// requests to a OpenEBS volume.
func (s *HTTPServer) volumeV1alpha1SpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	if req == nil {
		return nil, CodedError(400, "failed to handle volume request: nil http request received")
	}

	glog.Infof("received cas volume request: http method {%s}", req.Method)

	volOp := &volumeAPIOpsV1alpha1{
		req:  req,
		resp: resp,
	}

	switch req.Method {
	case "POST":
		cvol, err := volOp.create()
		sendEventOrIgnore(cvol, usage.VolumeProvision)
		return cvol, err
	case "GET":
		return volOp.httpGet()
	case "DELETE":
		cvol, err := volOp.httpDelete()
		sendEventOrIgnore(cvol, usage.VolumeDeprovision)
		return cvol, err
	default:
		return nil, CodedError(405, http.StatusText(405))
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
func (v *volumeAPIOpsV1alpha1) httpDelete() (*v1alpha1.CASVolume, error) {
	// Extract name of volume from path after trimming
	volName := strings.TrimSpace(strings.TrimPrefix(v.req.URL.Path, "/latest/volumes/"))

	// check if req url has volume name
	if len(volName) == 0 {
		return nil, CodedError(405, "failed to delete volume: missing volume name")
	}

	return v.delete(volName)
}

func (v *volumeAPIOpsV1alpha1) create() (*v1alpha1.CASVolume, error) {
	glog.Infof("received volume create request")
	vol := &v1alpha1.CASVolume{}
	err := decodeBody(v.req, vol)
	if err != nil {
		return nil, CodedErrorWrap(400, errors.Wrap(err, "failed to create volume"))
	}

	// volume name is expected
	if len(vol.Name) == 0 {
		return nil, CodedErrorf(400, "failed to create volume: missing volume name: %s", vol)
	}

	// use run namespace from http request header if volume's namespace is still not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = v.req.Header.Get(NamespaceKey)
	}

	vOps, err := volume.NewOperation(vol)
	if err != nil {
		return nil, CodedErrorWrap(
			400,
			errors.Wrapf(err, "failed to create volume: failed to init volume operation: %s", vol),
		)
	}

	cvol, err := vOps.Create()
	if err != nil {
		return nil, CodedErrorWrap(500, errors.Wrapf(err, "failed to create volume: %s", vol))
	}

	glog.Infof("volume '%s' created successfully", cvol.Name)
	return cvol, nil
}

func (v *volumeAPIOpsV1alpha1) read(volumeName string) (*v1alpha1.CASVolume, error) {
	glog.Infof("received volume read request: %s", volumeName)

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
		return nil, CodedErrorf(400, "failed to read volume: missing volume name")
	}

	// use namespace from req headers if volume ns is still not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = hdrNS
	}

	// use StorageClass name from header if present
	scName := strings.TrimSpace(v.req.Header.Get(string(v1alpha1.StorageClassHeaderKey)))
	patchVal := v.req.Header.Get(string(v1alpha1.CASKeyIsPatchJivaReplicaNodeAffinityHeader))
	// add the StorageClass name to volume's labels
	vol.Labels = map[string]string{
		string(v1alpha1.StorageClassKey): scName,
	}

	vol.Annotations = map[string]string{
		string(v1alpha1.NodeAffinityReplicaJivaIsPatchKey): patchVal,
	}

	vOps, err := volume.NewOperation(vol)
	if err != nil {
		return nil, CodedErrorWrap(
			400,
			errors.Wrapf(err, "failed to read volume {%s}: failed to init volume operation", vol.Name),
		)
	}

	cvol, err := vOps.Read()
	if err != nil {
		if isNotFound(err) {
			return nil, CodedErrorWrap(
				404,
				errors.Errorf("failed to read volume: volume {%s} not found in namespace {%s}", vol.Name, vol.Namespace),
			)
		}
		return nil, CodedErrorWrap(500, errors.Wrap(err, "failed to handle volume read request"))
	}

	glog.Infof("volume '%s' read successfully", cvol.Name)
	return cvol, nil
}

func (v *volumeAPIOpsV1alpha1) delete(volumeName string) (*v1alpha1.CASVolume, error) {
	glog.Infof("received volume delete request")

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
		return nil, CodedErrorf(400, "failed to delete volume: missing volume name: %s", vol)
	}

	// use namespace from req headers if volume ns is still not set
	if len(vol.Namespace) == 0 {
		vol.Namespace = hdrNS
	}

	vOps, err := volume.NewOperation(vol)
	if err != nil {
		return nil, CodedErrorWrap(
			400,
			errors.Wrapf(err, "failed to delete volume: failed to init volume operation: %s", vol),
		)
	}

	cvol, err := vOps.Delete()
	if err != nil {
		if isNotFound(err) {
			return nil, CodedErrorWrap(
				404,
				errors.Errorf("failed to delete volume: volume {%s} not found in namespace {%s}", vol.Name, vol.Namespace),
			)
		}
		return nil, CodedErrorWrap(500, errors.Wrapf(err, "failed to delete volume: %s", vol))
	}

	glog.Infof("volume '%s' deleted successfully", cvol.Name)
	return cvol, nil
}

func (v *volumeAPIOpsV1alpha1) list() (*v1alpha1.CASVolumeList, error) {
	glog.Infof("received volume list request")

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
		return nil, CodedErrorWrap(
			400,
			errors.Wrapf(err, "failed to list volumes: failed to init volume operation: %s", vols),
		)
	}

	cvols, err := vOps.List()
	if err != nil {
		return nil, CodedErrorWrap(500, errors.Wrapf(err, "failed to list volumes: %s", vols))
	}

	glog.Infof("volumes listed successfully for namespace(s) {%s}", vols.Namespace)
	return cvols, nil
}

func (v *volumeAPIOpsV1alpha1) readStats(volumeName string) (interface{}, error) {
	glog.Infof("received volume stats request")
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
		return nil, CodedErrorf(400, "failed to read volume stats: missing volume name: %s", vol)
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
		return nil, CodedErrorWrap(
			400,
			errors.Wrapf(err, "failed to read volume stats: failed to init volume operation: %s", vol),
		)
	}

	stats, err := vOps.ReadStats()
	if err != nil {
		if isNotFound(err) {
			return nil, CodedErrorWrap(
				404,
				errors.Errorf("failed to read volume stats: volume {%s} not found in namespace {%s}", vol.Name, vol.Namespace),
			)
		}
		return nil, CodedErrorWrap(500, errors.Wrapf(err, "failed to read volume stats: %s", vol))
	}

	// pipelining the response
	v.resp.Write(stats)

	glog.Infof("read volume stats was successful for '%s'", volumeName)
	return nil, nil
}
