package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/snapshot/v1alpha1"
)

type snapshotAPIOps struct {
	req  *http.Request
	resp http.ResponseWriter
}

// snapshotV1alpha1SpecificRequest deals with snapshot API requests
func (s *HTTPServer) snapshotV1alpha1SpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	snapOp := &snapshotAPIOps{
		req:  req,
		resp: resp,
	}

	// The params extracted below are going to be used for RUD operations
	// volName is the volume name in the query params
	volName := req.URL.Query().Get("volume")
	// namespace is the namespace of volume in the query params
	namespace := req.URL.Query().Get("namespace")
	// casType is the cas type of volume in the query params
	casType := req.URL.Query().Get("casType")
	// snapName is expected to be used only in case of delete and get of a particular snapshot
	// TODO: Use some http framework to extract snapshot name. strings method is not a good way
	snapName := strings.Split(strings.TrimPrefix(req.URL.Path, "/latest/snapshots/"), "?")[0]

	switch req.Method {
	case "POST":
		return snapOp.create()
	case "GET":
		// If snapshot name is missing, assume it to be list request
		if snapName == "" {
			return snapOp.list(volName, namespace, casType)
		}
		return snapOp.get(snapName, volName, namespace, casType)
	case "DELETE":
		return snapOp.delete(snapName, volName, namespace, casType)
	}
	return nil, CodedError(405, ErrInvalidMethod)
}

// list is http handler for listing all created snapshot specific to particular volume
func (sOps *snapshotAPIOps) list(volName, namespace, casType string) (interface{}, error) {
	glog.Infof("Snapshot list request was received")

	// Volume name is expected
	if len(strings.TrimSpace(volName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to list snapshot: missing snapshot name "))
	}

	// namespace is expected
	if len(strings.TrimSpace(namespace)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to list snapshot: missing namespace "))
	}

	glog.Infof("Listing snapshots for volume %q ", volName)

	snapOps, err := snapshot.Snapshot(&v1alpha1.SnapshotOptions{
		CasType:    casType,
		Namespace:  namespace,
		VolumeName: volName,
	})
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	snaps, err := snapOps.List()
	if err != nil {
		glog.Errorf("Failed to list snapshots: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("Snapshots listed successfully for volume '%s'", volName)
	return snaps, nil
}

// Create is http handler which handles snaphsot-create request
func (sOps *snapshotAPIOps) create() (interface{}, error) {
	glog.Infof("Snapshot create request was received")

	snap := &v1alpha1.CASSnapshot{}

	err := decodeBody(sOps.req, snap)
	if err != nil {
		return nil, err
	}
	glog.V(2).Infof("CASSnapshot object received: %+v", sOps.req)
	// snapshot name is expected
	if len(strings.TrimSpace(snap.Name)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create snapshot: missing snapshot name "))
	}

	// volume name is expected
	if len(strings.TrimSpace(snap.Spec.VolumeName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create snapshot '%v': missing volume name", snap.Name))
	}

	// namespace is expected
	if len(strings.TrimSpace(snap.Namespace)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create snapshot '%v': missing volume name", snap.Name))
	}

	glog.Infof("Creating snapshot %q for %s volume %q ", snap.Name, snap.Spec.CasType, snap.Spec.VolumeName)

	snapOps, err := snapshot.Snapshot(&v1alpha1.SnapshotOptions{
		VolumeName: snap.Spec.VolumeName,
		Namespace:  snap.Namespace,
		CasType:    snap.Spec.CasType,
		Name:       snap.Name,
	})
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
func (sOps *snapshotAPIOps) get(snapName, volName, namespace, casType string) (interface{}, error) {
	glog.Infof("Received request for snapshot get")

	// snapshot name is expected
	if len(strings.TrimSpace(snapName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to get snapshot: missing snapshot name "))
	}

	// volume name is expected
	if len(strings.TrimSpace(volName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to get snapshot '%v': missing volume name", snapName))
	}

	// namespace is expected
	if len(strings.TrimSpace(namespace)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to get snapshot '%v': missing namespace", snapName))
	}

	glog.Infof("Processing snapshot %q get request for volume: %q", snapName, volName)

	snapOps, err := snapshot.Snapshot(&v1alpha1.SnapshotOptions{
		CasType:    casType,
		Namespace:  namespace,
		VolumeName: volName,
		Name:       snapName,
	})
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	glog.Infof("Getting %s volume %q snapshot %q", casType, volName, snapName)
	snap, err := snapOps.Read()
	if err != nil {
		glog.Errorf("Failed to get snapshot: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("Snapshot created successfully: name '%s'", snap.Name)
	return snap, nil
}

func (sOps *snapshotAPIOps) delete(snapName, volName, namespace, casType string) (interface{}, error) {
	glog.Infof("Received request for snapshot delete")
	// snapshot name is expected
	if len(strings.TrimSpace(snapName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to delete snapshot: missing snapshot name"))
	}

	// volume name is expected
	if len(strings.TrimSpace(volName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to delete snapshot '%v': missing volume name", snapName))
	}

	// namespace is expected
	if len(strings.TrimSpace(namespace)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to delete snapshot '%v': missing namespace", snapName))
	}

	snapOps, err := snapshot.Snapshot(&v1alpha1.SnapshotOptions{
		CasType:    casType,
		Namespace:  namespace,
		VolumeName: volName,
		Name:       snapName,
	})
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	glog.Infof("Deleting snapshot %q of %s volume %q", snapName, casType, volName)
	output, err := snapOps.Delete()
	if err != nil {
		glog.Errorf("Failed to delete snapshot %q for volume %q: %s", snapName, volName, err)
		return nil, err
	}
	glog.Infof("Snapshot deleted successfully: name '%s'", snapName)
	return output, nil
}
