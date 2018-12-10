package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	snapshot "github.com/openebs/maya/pkg/snapshot/v1alpha1"
)

type backupAPIOps struct {
	req  *http.Request
	resp http.ResponseWriter
}

// backupV1alpha1SpecificRequest deals with backup API requests
func (s *HTTPServer) backupV1alpha1SpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	backupOp := &backupAPIOps{
		req:  req,
		resp: resp,
	}

	switch req.Method {
	case "POST":
		return backupOp.create()
	}
	return nil, CodedError(405, ErrInvalidMethod)
}

// Create is http handler which handles snaphsot-create request
func (bOps *backupAPIOps) create() (interface{}, error) {
	glog.Infof("Backup create request was received")

	backup := &v1alpha1.CStorBackup{}

	err := decodeBody(bOps.req, backup)
	if err != nil {
		return nil, err
	}
	// snapshot name is expected
	if len(strings.TrimSpace(backup.Name)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create backup: missing snapshot name "))
	}

	// volume name is expected
	if len(strings.TrimSpace(backup.Spec.VolumeName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create backup '%v': missing volume name", backup.Name))
	}

	// backupIP is expected
	if len(strings.TrimSpace(backup.Spec.BackupIP)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create backup '%v': missing backupIP", backup.Name))
	}

	// namespace is expected
	if len(strings.TrimSpace(backup.Namespace)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create backup '%v': missing namespace", backup.Name))
	}

	//TODO Create snapname randomly
	snapOps, err := snapshot.Snapshot(&v1alpha1.SnapshotOptions{
		VolumeName: backup.Spec.VolumeName,
		Namespace:  backup.Namespace,
		CasType:    backup.Spec.CasType,
		Name:       "Snap1",
	})
	if err != nil {
		return nil, CodedError(400, err.Error())
	}

	glog.Infof("Creating %s volume %q snapshot", backup.Spec.CasType, backup.Spec.VolumeName)

	snap, err := snapOps.Create()
	if err != nil {
		glog.Errorf("Failed to create snapshot: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("Snapshot created successfully: name '%s'", snap.Name)

	glog.Infof("Creating backup %q for %s volume %q ", backup.Name, backup.Spec.VolumeName)
	clientset := internalclientset.Clientset{}
	_, err = clientset.OpenebsV1alpha1().CStorBackups(backup.Namespace).Create(backup)
	if err != nil {
		glog.Errorf("Failed to create backup: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("Backup CR created successfully: name '%s'", "")
	return "", nil
}
