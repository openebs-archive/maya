package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/types/v1"
	"github.com/openebs/maya/volume/provisioners/jiva"
)

// SnapshotSpecificRequest deals with snapshot API request w.r.t a Volume
func (s *HTTPServer) snapshotSpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Extract info from path after trimming
	path := strings.TrimPrefix(req.URL.Path, "/latest/snapshots")

	// Check request validity
	if path == req.URL.Path {
		return nil, CodedError(405, ErrInvalidMethod)
	}

	switch {
	case strings.Contains(path, "/create/"):
		return s.snapshotCreate(resp, req)
	case strings.Contains(path, "/revert/"):
		return s.snapshotRevert(resp, req)
	case strings.Contains(path, "/list"):
		volName := strings.TrimPrefix(path, "/list/")
		return s.snapshotList(resp, req, volName)
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// SnapshotCreate is http handler which handles snaphsot-create request
func (s *HTTPServer) snapshotCreate(resp http.ResponseWriter, req *http.Request) (interface{}, error) {

	if req.Method != "PUT" && req.Method != "POST" {
		return nil, CodedError(405, ErrInvalidMethod)
	}

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

	voldetails, err := s.volumeRead(resp, req, snap.Spec.VolumeName)
	if err != nil {
		return nil, err
	}

	ControllerIP := voldetails.Annotations["vsm.openebs.io/controller-ips"]

	var labelMap map[string]string
	snapinfo, err := jiva.Snapshot(snap.Metadata.Name, ControllerIP, labelMap)
	if err != nil {
		glog.Errorf("Failed to create snapshot of volume %v : %v", snap.Spec.VolumeName, err)
		return nil, err
	}

	glog.Infof("Snapshot created for volume [%s] is [%s]\n", snap.Spec.VolumeName, snap.Metadata.Name)

	return snapinfo, nil
}

// SnapshotRevert is http handler for handling snapshot-revert request.
// Volume and existing snapshot name will be passed as struct fields to
// revert to that particular snapshot
func (s *HTTPServer) snapshotRevert(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	if req.Method != "PUT" && req.Method != "POST" {
		return nil, CodedError(405, ErrInvalidMethod)
	}
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

	voldetails, err := s.volumeRead(resp, req, snap.Spec.VolumeName)
	if err != nil {
		return nil, err
	}

	ControllerIP := voldetails.Annotations["vsm.openebs.io/controller-ips"]

	err = jiva.SnapshotRevert(snap.Metadata.Name, ControllerIP)
	if err != nil {
		glog.Errorf("Failed to revert snapshot of volume %s: %v", snap.Spec.VolumeName, err)
		return nil, err
	}

	glog.Infof("Reverting to snapshot [%s] of volume [%s]", snap.Metadata.Name, snap.Spec.VolumeName)

	return fmt.Sprintf("Reverting to snapshot [%s] of volume [%s]", snap.Metadata.Name, snap.Spec.VolumeName), nil

}

// SnapshotList is http handler for listing all created snapshot specific to particular volume
func (s *HTTPServer) snapshotList(resp http.ResponseWriter, req *http.Request, volName string) (interface{}, error) {

	if req.Method != "GET" {
		return nil, CodedError(405, ErrInvalidMethod)
	}
	glog.Infof("Processing Volume snapshot-list request")

	snap := v1.VolumeSnapshot{}
	snap.Spec.VolumeName = volName

	// Name is expected to be available in snapshot specs
	if snap.Spec.VolumeName == "" {

		return nil, CodedError(400, fmt.Sprintf("Volume name missing in '%v'", snap.Spec.VolumeName))
	}

	voldetails, err := s.volumeRead(resp, req, volName)
	if err != nil {
		return nil, err
	}

	ControllerIP := voldetails.Annotations["vsm.openebs.io/controller-ips"]

	// list all created snapshot specific to particular volume
	snapChain, err := jiva.SnapshotList(snap.Spec.VolumeName, ControllerIP)
	if err != nil {
		glog.Errorf("Error getting snapshots of volume %s: %v", snap.Spec.VolumeName, err)
		return nil, err
	}

	glog.Infof("Successfully list snapshot of volume: %s", snap.Spec.VolumeName)
	return snapChain, nil

}

/*func (s *HTTPServer) getControllerIP(resp http.ResponseWriter, req *http.Request, snap.Spec.Volname string) (string, err) {
	voldetails, err := s.vsmRead(resp, req, snap.Spec.VolumeName)
	if err != nil {
		return "", err
	}

	controllerIP := voldetails.Annotations["vsm.openebs.io/controller-ips"]
	return controllerIP, nil
}*/
