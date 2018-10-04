package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/snapshot"
)

type snapshotAPIOps struct {
	req  *http.Request
	resp http.ResponseWriter
}

// SnapshotSpecificRequest deals with snapshot API request w.r.t a Volume
func (s *HTTPServer) snapshotRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	snapOp := &snapshotAPIOps{
		req:  req,
		resp: resp,
	}
	// volName is the volume name in the query params
	volName := req.URL.Query().Get("volume")
	// snapName is expected to be used only in case of delete and get of a particular snapshot
	snapName := strings.Split(strings.TrimPrefix(req.URL.Path, "/latest/snapshots/"), "?")[0]
	switch req.Method {
	case "POST":
		return snapOp.create(resp, req)
	case "GET":
		// The volume name is expected to be present as request parameter
		// eg http://1.1.1.1:5656/latest/snapshots/?volume=myvol
		if snapName == "" {
			return snapOp.list(resp, req, volName)
		}

		return snapOp.get(resp, req, snapName, volName)
	case "DELETE":
		// The volume name is expected to be present as request parameter
		// eg http://1.1.1.1:5656/latest/snapshots/?volume=myvol
		//
		// TODO: Use some http framework to extract snapshot name. strings method is not a good way
		// TODO: Uncomment the below line when we start supporting deletion of snapshot
		return snapOp.delete(resp, req, snapName, volName)
		//return nil, errors.Errorf("snapshot deletion not supported")
	}
	return nil, CodedError(405, ErrInvalidMethod)
}

// list is http handler for listing all created snapshot specific to particular volume
func (sOps *snapshotAPIOps) list(resp http.ResponseWriter, req *http.Request, volName string) (interface{}, error) {
	glog.Infof("Snapshot list request was received")

	snaps := &v1alpha1.CASSnapshotList{}

	err := decodeBody(req, snaps)
	if err != nil {
		return nil, err
	}
	// Volume name is expected
	if len(strings.TrimSpace(snaps.Options.VolumeName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to list snapshot: missing snapshot name "))
	}

	glog.Infof("Listing snapshots for volume %q ", snaps.Options.VolumeName)

	snapOps, err := snapshot.SnapshotList(snaps)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	snaps, err = snapOps.List()
	if err != nil {
		glog.Errorf("Failed to list snapshots: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("Snapshots listed successfully for volume '%s'", snaps.Options.VolumeName)
	return snaps, nil
}

// Create is http handler which handles snaphsot-create request
func (sOps *snapshotAPIOps) create(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	glog.Infof("Snapshot create request was received")

	snap := &v1alpha1.CASSnapshot{}

	err := decodeBody(req, snap)
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

	glog.Infof("Creating snapshot %q for %s volume %q ", snap.Name, snap.Spec.CasType, snap.Spec.VolumeName)

	snapOps, err := snapshot.Snapshot(snap)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	glog.Infof("Creating %s volume %q snapshot", snap.Spec.CasType, snap.Spec.VolumeName)

	snap, err = snapOps.Create()
	if err != nil {
		glog.Errorf("Failed to create snapshot: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("Snapshot created successfully: name '%s'", snap.Name)
	return snap, nil
}

// read is http handler for reading a snapshot specific to particular volume
func (sOps *snapshotAPIOps) get(resp http.ResponseWriter, req *http.Request, snapName, volName string) (interface{}, error) {
	glog.Infof("Received request for snapshot get")
	snap := &v1alpha1.CASSnapshot{}
	snap.Name = snapName
	snap.Spec.VolumeName = volName

	// snapshot name is expected
	if len(strings.TrimSpace(snap.Name)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create snapshot: missing snapshot name "))
	}

	// volume name is expected
	if len(strings.TrimSpace(snap.Spec.VolumeName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create snapshot '%v': missing volume name", snap.Name))
	}

	glog.Infof("Processing snapshot get request for volume: %q", snap.Spec.VolumeName)

	snapOps, err := snapshot.Snapshot(snap)
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	snap, err = snapOps.Read()
	if err != nil {
		glog.Errorf("Failed to get snapshot: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("Getting %s volume %q snapshot", snap.Spec.CasType, snap.Spec.VolumeName)

	return nil, nil
}

func (sOps *snapshotAPIOps) delete(resp http.ResponseWriter, req *http.Request, snapName, volName string) (interface{}, error) {
	glog.Infof("Received request for snapshot delete")
	snap := &v1alpha1.CASSnapshot{}
	snap.Name = snapName
	snap.Spec.VolumeName = volName
	// snapshot name is expected
	if len(strings.TrimSpace(snap.Name)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create snapshot: missing snapshot name "))
	}

	// volume name is expected
	if len(strings.TrimSpace(snap.Spec.VolumeName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create snapshot '%v': missing volume name", snap.Name))
	}
	snapOps, err := snapshot.Snapshot(snap)

	glog.Infof("Deleting snapshot %q of volume %q", snap.Name, snap.Spec.VolumeName)
	output, err := snapOps.Delete()
	if err != nil {
		glog.Errorf("Failed to delete snapshot %q for volume %q: %s", snap.Name, snap.Spec.VolumeName, err)
		return nil, err
	}
	return output, nil
}
