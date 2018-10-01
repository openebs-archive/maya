package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/snapshot"
)

type snapshotAPIOpsV1alpha1 struct {
	req  *http.Request
	resp http.ResponseWriter
}

// SnapshotSpecificRequest deals with snapshot API request w.r.t a Volume
func (s *HTTPServer) snapshotV1alpha1SpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	snapOp := &snapshotAPIOpsV1alpha1{
		req:  req,
		resp: resp,
	}
	// volName := req.URL.Query().Get("volume")
	// snapName := strings.Split(strings.TrimPrefix(req.URL.Path, "/latest/snapshots/"), "?")[0]
	switch req.Method {
	case "POST":
		return snapOp.create(resp, req)
	// case "GET":
	// 	// The volume name is expected to be present as request parameter
	// 	// eg http://1.1.1.1:5656/latest/snapshots/?volume=myvol
	// 	if snapName == "" {
	// 		return snapOp.list(resp, req, volName)
	// 	}
	// 	return snapOp.get(resp, req, snapName, volName)
	// case "DELETE":
	// 	// The volume name is expected to be present as request parameter
	// 	// eg http://1.1.1.1:5656/latest/snapshots/?volume=myvol
	// 	//
	// 	// TODO: Use some http framework to extract snapshot name. strings method is not a good way
	// 	return snapOp.delete(resp, req, snapName, volName)
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// Create is http handler which handles snaphsot-create request
func (v *snapshotAPIOpsV1alpha1) create(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	glog.Infof("cas template based snapshot create request was received")

	snap := &v1alpha1.CASSnapshot{}

	// do initial validation with passed params
	// create snapshotOperation object and then create the snapshot using the object
	err := decodeBody(v.req, snap)
	if err != nil {
		return nil, err
	}
	glog.V(2).Infof("CASSnapshot object received: %+v", req)
	// snapshot name is expected
	if len(strings.TrimSpace(snap.Name)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create snapshot: missing snapshot name "))
	}

	// volume name is expected
	if len(strings.TrimSpace(snap.Spec.VolumeName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create snapshot '%v': missing volume name", snap.Name))
	}

	snapOps, err := snapshot.Snapshot(snap)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	snap, err = snapOps.Create()
	if err != nil {
		glog.Errorf("failed to create snapshot: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("snapshot created successfully: name '%s'", snap.Name)
	return snap, nil
}

/*
// list is http handler for listing all created snapshot specific to particular volume
func (v *snapshotAPIOpsV1alpha1) list(resp http.ResponseWriter, req *http.Request, volName string) (interface{}, error) {
	glog.Infof("cas template based snapshot list request received")
	snapList := &v1alpha1.CASSnapshotList{}

	// hdrNS := ""
	// // get namespace from http request
	// if v.req != nil {
	// 	decodeBody(v.req, snapList)
	// 	hdrNS = v.req.Header.Get(NamespaceKey)
	// }

	// snapList.Spec.Namespace = hdrNS
	// if snapList.Spec.Namespace == "" {
	// 	return nil, CodedError(400, fmt.Sprintf("failed to list snapshot for volume %q: missing volume namespace", snapList.Spec.VolumeName))
	// }

	glog.Infof("Processing snapshot list request for volume: %q", snapList.Spec.VolumeName)
	// volDetails, err := v.read(snapList.Spec.VolumeName)
	// if err != nil {
	// 	return "", err
	// }

	snapOps, err := snapshot.NewSnapshotListOperation(snapList)
	if err != nil {
		glog.Errorf("Error creating SnapshotOps: %s", err)
		return nil, err
	}

	glog.Infof("Listing snapshots for volume %q", snap.Spec.VolumeName)
	output, err := snapOps.SnapshotList()
	if err != nil {
		glog.Errorf("Failed to list snapshots for volume %q: %s", snap.Spec.VolumeName, err)
		return nil, err
	}
	return output, nil
}

// read is http handler for listing all created snapshot specific to particular volume
func (v *snapshotAPIOpsV1alpha1) get(resp http.ResponseWriter, req *http.Request, snapName, volName string) (interface{}, error) {
	glog.Infof("Received request for snapshot list")
	snap := v1.VolumeSnapshot{}
	snap.Spec.VolumeName = volName

	// Name is expected to be available in snapshot specs
	if snap.Spec.VolumeName == "" {
		return nil, CodedError(400, fmt.Sprintf("Volume name missing in '%v'", snap.Spec.VolumeName))
	}

	glog.Infof("Processing snapshot list request for volume: %q", snap.Spec.VolumeName)
	volDetails, err := v.read(snap.Spec.VolumeName)
	if err != nil {
		return "", err
	}

	snapOps, err := snapshot.GetSnapshotOps(volDetails.Spec.CasType)
	if err != nil {
		glog.Errorf("Error creating SnapshotOps: %s", err)
		return nil, err
	}

	glog.Infof("Listing snapshots for volume %q", snap.Spec.VolumeName)
	output, err := snapOps.SnapshotList(snap, volDetails.Spec.TargetIP)
	if err != nil {
		glog.Errorf("Failed to list snapshots for volume %q: %s", snap.Spec.VolumeName, err)
		return nil, err
	}
	return output, nil
}

func (v *snapshotAPIOpsV1alpha1) delete(resp http.ResponseWriter, req *http.Request, snapName, volName string) (interface{}, error) {
	glog.Infof("Received request for snapshot delete")
	snap := v1.VolumeSnapshot{}
	snap.Metadata.Name = snapName
	snap.Spec.VolumeName = volName
	// Name is expected to be available in snapshot specs
	if snap.Metadata.Name == "" {
		return nil, CodedError(400, fmt.Sprintf("Snapshot name missing"))
	}
	if snap.Spec.VolumeName == "" {
		return nil, CodedError(400, fmt.Sprintf("Volume name missing"))
	}

	glog.Infof("Processing snapshot delete request for volume: %s", snap.Spec.VolumeName)
	volDetails, err := v.read(snap.Spec.VolumeName)
	if err != nil {
		return "", err
	}

	snapOps, err := snapshot.GetSnapshotOps(volDetails.Spec.CasType)
	if err != nil {
		glog.Errorf("Error creating SnapshotOps: %s", err)
		return nil, err
	}

	glog.Infof("Deleting snapshot %q of volume %q", snap.Metadata.Name, snap.Spec.VolumeName)
	output, err := snapOps.SnapshotDelete(snap, volDetails.Spec.TargetIP)
	if err != nil {
		glog.Errorf("Failed to delete snapshot %q for volume %q: %s", snap.Metadata.Name, snap.Spec.VolumeName, err)
		return nil, err
	}
	return output, nil
}
*/
