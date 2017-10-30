package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/openebs/maya/types/v1"
	"github.com/openebs/maya/volume/provisioners/jiva"
)

// SnapshotSpecificGetRequest deals with HTTP GET request w.r.t a Volume Snapshot
func (s *HTTPServer) SnapshotSpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Extract info from path after trimming
	path := strings.TrimPrefix(req.URL.Path, "/latest/snapshot")

	// Is req valid ?
	if path == req.URL.Path {
		fmt.Println("Request coming", path)
		return nil, CodedError(405, ErrInvalidMethod)
	}

	switch {
	case strings.Contains(path, "/create/"):
		return s.SnapshotCreate(resp, req)
	case strings.Contains(path, "/revert/"):
		return s.SnapshotRevert(resp, req)
	case strings.Contains(path, "/list"):
		volName := strings.TrimPrefix(path, "/list/")
		return s.SnapshotList(resp, req, volName)
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// SnapshotCreate is http handler which handles snaphsot-create request
func (s *HTTPServer) SnapshotCreate(resp http.ResponseWriter, req *http.Request) (interface{}, error) {

	if req.Method != "PUT" && req.Method != "POST" {
		return nil, CodedError(405, ErrInvalidMethod)
	}

	fmt.Println("[DEBUG] Processing Volume Snapshot request")

	snap := v1.VolumeSnapshot{}

	// The yaml/json spec is decoded to pvc struct
	if err := decodeBody(req, &snap); err != nil {

		return nil, CodedError(400, err.Error())
	}

	// SnapshotName is expected to be available even in the minimalist specs
	if snap.Metadata.Name == "" {
		return nil, CodedError(400, fmt.Sprintf("Snapshot name missing in '%v'", snap.SnapshotName))
	}

	// Name is expected to be available even in the minimalist specs
	if snap.Spec.VolumeName == "" {

		return nil, CodedError(400, fmt.Sprintf("PVC Volume name missing in '%v'", snap.Spec.VolumeName))
	}

	fmt.Println("Volume Name :", snap.Spec.VolumeName)
	fmt.Println("[DEBUG] Processing snapshot-create request of volume:", snap.Spec.VolumeName)

	voldetails, err := s.vsmRead(resp, req, snap.Spec.VolumeName)
	if err != nil {
		return nil, err
	}

	ControllerIP := voldetails.Annotations["vsm.openebs.io/controller-ips"]

	var labelMap map[string]string
	snapinfo, err := jiva.Snapshot(snap.Spec.VolumeName, snap.Metadata.Name, ControllerIP, labelMap)
	if err != nil {
		log.Printf("Error running create snapshot command: %v", err)
		return nil, err
	}

	fmt.Println("[DEBUG] Snapshot created:", snap.Metadata.Name)

	return snapinfo, nil
}

// SnapshotRevert is http handler for handling snapshot-revert request.
// Volume and existing snapshot name will be passed as struct fields to
// revert to that particular snapshot
func (s *HTTPServer) SnapshotRevert(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	if req.Method != "PUT" && req.Method != "POST" {
		return nil, CodedError(405, ErrInvalidMethod)
	}
	fmt.Println("[DEBUG] Processing Volume snapshot-revert request")

	snap := v1.VolumeSnapshot{}

	// The yaml/json spec is decoded to pvc struct
	if err := decodeBody(req, &snap); err != nil {

		return nil, CodedError(400, err.Error())
	}

	// SnapshotName is expected to be available even in the minimalist specs
	if snap.Metadata.Name == "" {
		return nil, CodedError(400, fmt.Sprintf("[ERROR] Snapshot name missing in '%v'", snap.Metadata.Name))
	}

	// Name is expected to be available even in the minimalist specs
	if snap.Spec.VolumeName == "" {

		return nil, CodedError(400, fmt.Sprintf("[ERROR] Volume name missing in '%v'", snap))
	}

	fmt.Println("Volume Name :", snap.Spec.VolumeName)
	fmt.Println("[DEBUG] Processing snapshot-revert request of volume:", snap.Spec.VolumeName)

	voldetails, err := s.vsmRead(resp, req, snap.Spec.VolumeName)
	if err != nil {
		return nil, err
	}

	ControllerIP := voldetails.Annotations["vsm.openebs.io/controller-ips"]

	err = jiva.SnapshotRevert(snap.Spec.VolumeName, snap.Metadata.Name, ControllerIP)
	if err != nil {
		log.Fatalf("[ERROR] Error running revert snapshot command: %v", err)
		return nil, err
	}

	fmt.Println("[DEBUG] Reverting to snapshot:", snap.Metadata.Name)
	return nil, nil

}

// SnapshotList is http handler for listing all created snapshot specific to particular volume
func (s *HTTPServer) SnapshotList(resp http.ResponseWriter, req *http.Request, volName string) (interface{}, error) {

	if req.Method != "GET" {
		return nil, CodedError(405, ErrInvalidMethod)
	}
	fmt.Println("[DEBUG] Processing Volume snapshot-list request")

	snap := v1.VolumeSnapshot{}
	snap.Spec.VolumeName = volName

	// Name is expected to be available in specs
	if snap.Spec.VolumeName == "" {

		return nil, CodedError(400, fmt.Sprintf("[ERROR] Volume name missing in '%v'", snap.Spec.VolumeName))
	}

	voldetails, err := s.vsmRead(resp, req, volName)
	if err != nil {
		return nil, err
	}

	ControllerIP := voldetails.Annotations["vsm.openebs.io/controller-ips"]

	// list all created snapshot specific to particular volume
	snapChain, err := jiva.SnapshotList(snap.Spec.VolumeName, ControllerIP)
	if err != nil {
		log.Fatalf("[ERROR] Error running list snapshot command: %v", err)
		return nil, err
	}

	fmt.Println("[DEBUG] Successfully created snapshot of volume", snap.Spec.VolumeName)
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
