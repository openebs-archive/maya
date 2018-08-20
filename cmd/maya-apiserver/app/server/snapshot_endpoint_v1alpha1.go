package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/types/v1"
	"github.com/openebs/maya/volume/provisioners/cstor"
	"github.com/openebs/maya/volume/provisioners/jiva"
)

// SnapshotSpecificRequest deals with snapshot API request w.r.t a Volume
func (s *HTTPServer) snapshotV1alpha1SpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	glog.Infof("Feature gated request received for volume snapshot")
	volOp := &volumeAPIOpsV1alpha1{
		req:  req,
		resp: resp,
	}
	switch req.Method {
	case "POST":
		return volOp.snapshotCreateV1Alpha1(resp, req)
	case "PUT":
		return volOp.snapshotRevertV1Alpha1(resp, req)
	case "GET":
		volName := req.Header.Get("volume-name")
		return volOp.snapshotListV1Alpha1(resp, req, volName)
	case "DELETE":
		return volOp.snapshotDeleteV1Alpha1(resp, req, strings.TrimPrefix(req.URL.Path, "/latest/snapshots/"))
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// Create is http handler which handles snaphsot-create request
func (v *volumeAPIOpsV1alpha1) snapshotCreateV1Alpha1(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	glog.Infof("Processing Volume create snapshot request")

	snap := v1.VolumeSnapshot{}

	// The yaml/json spec is decoded to VolumeSnapshot struct
	if err := decodeBody(req, &snap); err != nil {
		return nil, CodedError(400, err.Error())
	}

	// SnapshotName is expected to be available even in the minimalist specs
	if snap.Metadata.Name == "" {
		return nil, CodedError(400, fmt.Sprintf("Error: snapshot name missing in '%v'", snap.SnapshotName))
	}

	// Name is expected to be available even in the minimalist specs
	if snap.Spec.VolumeName == "" {
		return nil, CodedError(400, fmt.Sprintf("Error: volume name missing in '%v'", snap.Spec.VolumeName))
	}

	glog.Infof("Processing snapshot-create request of volume: %s", snap.Spec.VolumeName)
	volDetails, err := v.read(snap.Spec.VolumeName)
	if err != nil {
		return "", err
	}

	volType := volDetails.Spec.EngineType
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	switch volType {
	case string(v1.CStorVolumeType):
		return cStorSnapshot(resp, req, snap)
	case string(v1.JivaVolumeType):
		return jivaSnapshot(resp, req, snap)
	}
	return nil, fmt.Errorf("invalid volType for volume '%s'", snap.Spec.VolumeName)
}

// List is http handler for listing all created snapshot specific to particular volume
func (v *volumeAPIOpsV1alpha1) snapshotListV1Alpha1(resp http.ResponseWriter, req *http.Request, volName string) (interface{}, error) {
	snap := v1.VolumeSnapshot{}
	snap.Spec.VolumeName = volName

	// Name is expected to be available in snapshot specs
	if snap.Spec.VolumeName == "" {
		return nil, CodedError(400, fmt.Sprintf("Volume name missing in '%v'", snap.Spec.VolumeName))
	}

	glog.Infof("Processing snapshot-list request of volume: %s", snap.Spec.VolumeName)
	volDetails, err := v.read(snap.Spec.VolumeName)
	if err != nil {
		return "", err
	}

	volType := volDetails.Spec.EngineType

	switch volType {
	case string(v1.CStorVolumeType):
		return cstorList(resp, req, snap)
	case string(v1.JivaVolumeType):
		return jivaList(resp, req, snap)
	}
	return nil, fmt.Errorf("invalid volType for volume '%s'", snap.Spec.VolumeName)
}

func (v *volumeAPIOpsV1alpha1) snapshotDeleteV1Alpha1(resp http.ResponseWriter, req *http.Request, snapName string) (interface{}, error) {
	snap := v1.VolumeSnapshot{}
	snap.Metadata.Name = snapName
	snap.Spec.VolumeName = strings.TrimSpace(req.Header.Get("volume-name"))
	// Name is expected to be available in snapshot specs
	if snap.Metadata.Name == "" {
		return nil, CodedError(400, fmt.Sprintf("Snapshot name missing"))
	}
	if snap.Spec.VolumeName == "" {
		return nil, CodedError(400, fmt.Sprintf("Volume name missing"))
	}

	glog.Infof("Processing snapshot-delete request of volume: %s", snap.Spec.VolumeName)
	volDetails, err := v.read(snap.Spec.VolumeName)
	if err != nil {
		return "", err
	}

	volType := volDetails.Spec.EngineType

	switch volType {
	case string(v1.CStorVolumeType):
		return cStorDelete(resp, req, snap)
	case string(v1.JivaVolumeType):
		return jivaDelete(resp, req, snap)
	}
	return nil, fmt.Errorf("invalid volType for volume '%s'", snap.Spec.VolumeName)
}

func (v *volumeAPIOpsV1alpha1) snapshotRevertV1Alpha1(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	glog.Infof("Processing Volume snapshot-revert request")

	snap := v1.VolumeSnapshot{}

	// The yaml/json spec is decoded to VolumeSnapshot struct
	if err := decodeBody(req, &snap); err != nil {
		return nil, CodedError(400, err.Error())
	}

	// SnapshotName is expected to be available even in the minimalist specs
	if snap.Metadata.Name == "" {
		return nil, CodedError(400, fmt.Sprintf("ERROR: Snapshot name missing in '%v'", snap.Metadata.Name))
	}

	// Name is expected to be available even in the minimalist specs
	if snap.Spec.VolumeName == "" {
		return nil, CodedError(400, fmt.Sprintf("ERROR: Volume name missing in '%v'", snap))
	}

	glog.Infof("Processing snapshot-revert request of volume: %s", snap.Spec.VolumeName)

	volDetails, err := v.read(snap.Spec.VolumeName)
	if err != nil {
		return "", err
	}

	volType := volDetails.Spec.EngineType

	switch volType {
	case string(v1.CStorVolumeType):
		return jivaRevert(resp, req, snap)
	case string(v1.JivaVolumeType):
		return cStorRevert(resp, req, snap)
	}
	return nil, fmt.Errorf("invalid volType for volume '%s'", snap.Spec.VolumeName)
}

func jivaSnapshot(resp http.ResponseWriter, req *http.Request, snap v1.VolumeSnapshot) (interface{}, error) {
	targetIP, err := getTargetIP(resp, req, snap.Spec.VolumeName)
	if err != nil {
		return "", err
	}

	var labelMap map[string]string
	snapinfo, err := jiva.Snapshot(snap.Metadata.Name, targetIP, labelMap)
	if err != nil {
		glog.Errorf("Failed to create snapshot of volume %v : %v", snap.Spec.VolumeName, err)
		return nil, err
	}

	glog.Infof("Snapshot created for volume [%s] is [%s]\n", snap.Spec.VolumeName, snap.Metadata.Name)

	return snapinfo, nil
}

func cStorSnapshot(resp http.ResponseWriter, req *http.Request, snap v1.VolumeSnapshot) (interface{}, error) {
	glog.Infof("[DEBUG] cStorSnapshot called")
	targetIP, err := getTargetIP(resp, req, snap.Spec.VolumeName)
	if err != nil {
		return nil, err
	}

	return cstor.CreateSnapshot(snap.Spec.VolumeName, snap.Metadata.Name, targetIP)
}

func getTargetIP(resp http.ResponseWriter, req *http.Request, volName string) (string, error) {
	vOps := volumeAPIOpsV1alpha1{req: req, resp: resp}
	volDetails, err := vOps.read(volName)
	if err != nil {
		return "", err
	}
	glog.Infof("[DEBUG] getTargetIP volDetails:%#v", volDetails)
	return volDetails.Spec.TargetIP, nil
}

func jivaList(resp http.ResponseWriter, req *http.Request, snap v1.VolumeSnapshot) (interface{}, error) {
	targetIP, err := getTargetIP(resp, req, snap.Spec.VolumeName)

	if err != nil {
		return nil, err
	}
	// list all created snapshot specific to particular volume
	snapChain, err := jiva.SnapshotList(snap.Spec.VolumeName, targetIP)
	if err != nil {
		glog.Errorf("Error getting snapshots of volume %s: %v", snap.Spec.VolumeName, err)
		return nil, err
	}

	glog.Infof("Successfully list snapshot of volume: %s", snap.Spec.VolumeName)
	return snapChain, nil
}

func cstorList(resp http.ResponseWriter, req *http.Request, snap v1.VolumeSnapshot) (interface{}, error) {
	return nil, fmt.Errorf("cstor snapshot list not supported")
}

func jivaRevert(resp http.ResponseWriter, req *http.Request, snap v1.VolumeSnapshot) (interface{}, error) {
	targetIP, err := getTargetIP(resp, req, snap.Spec.VolumeName)
	if err != nil {
		return nil, err
	}
	err = jiva.SnapshotRevert(snap.Metadata.Name, targetIP)
	if err != nil {
		glog.Errorf("Failed to revert snapshot of volume %s: %v", snap.Spec.VolumeName, err)
		return nil, err
	}

	glog.Infof("Reverting to snapshot [%s] of volume [%s]", snap.Metadata.Name, snap.Spec.VolumeName)

	return fmt.Sprintf("Reverting to snapshot [%s] of volume [%s]", snap.Metadata.Name, snap.Spec.VolumeName), nil
}

func cStorRevert(resp http.ResponseWriter, req *http.Request, snap v1.VolumeSnapshot) (interface{}, error) {
	return nil, fmt.Errorf("cstor snapshot revert not supported")
}

func jivaDelete(resp http.ResponseWriter, req *http.Request, snap v1.VolumeSnapshot) (interface{}, error) {
	return "Not implemented", nil
}

func cStorDelete(resp http.ResponseWriter, req *http.Request, snap v1.VolumeSnapshot) (interface{}, error) {
	glog.Infof("[DEBUG] cStorSnapshotDelete called. Header values : %v", req.Header)
	targetIP, err := getTargetIP(resp, req, snap.Spec.VolumeName)
	if err != nil {
		return nil, err
	}

	return cstor.DeleteSnapshot(snap.Spec.VolumeName, snap.Metadata.Name, targetIP)
}
