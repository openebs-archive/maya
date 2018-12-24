package server

import (
	"fmt"
	"net/http"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	snapshot "github.com/openebs/maya/pkg/snapshot/v1alpha1"
	"k8s.io/client-go/rest"
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

// Create is http handler which handles backup-create request
func (bOps *backupAPIOps) create() (interface{}, error) {
	glog.Infof("Backup create request was received")

	backup := &v1alpha1.CStorBackup{}

	err := decodeBody(bOps.req, backup)
	if err != nil {
		return nil, err
	}
	// backup name is expected
	if len(strings.TrimSpace(backup.Spec.Name)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create backup: missing backup name "))
	}

	// incrementalBackup name is expected
	if len(strings.TrimSpace(backup.Spec.IncrementalBackupName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create backup: missing incremental backup name "))
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
	snapshotName := backup.Spec.Name + backup.Spec.IncrementalBackupName
	snapOps, err := snapshot.Snapshot(&v1alpha1.SnapshotOptions{
		VolumeName: backup.Spec.VolumeName,
		Namespace:  backup.Namespace,
		CasType:    backup.Spec.CasType,
		Name:       snapshotName,
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
	backup.Spec.LastSnapshotName = snapshotName
	backup.Name = backup.Spec.Name
	glog.Infof("Snapshot created successfully: name '%s'", snap.Name)

	openebsClient, _ := loadClientFromServiceAccount()
	listOptions := v1.ListOptions{}
	bkpList, err := openebsClient.OpenebsV1alpha1().CStorBackups(backup.Namespace).List(listOptions)

	//Check if this schedule is already present
	for _, bkp := range bkpList.Items {
		if backup.Spec.Name == bkp.Spec.Name {
			bkp.Spec.LastSnapshotName = snapshotName
			bkp.Spec.IncrementalBackupName = backup.Spec.IncrementalBackupName
			openebsClient.OpenebsV1alpha1().CStorBackups(bkp.Namespace).Update(&bkp)
			glog.Infof("Creating incremental backup %s for %s volume %s poolUUID:%v",
				backup.Spec.IncrementalBackupName, backup.Spec.Name,
				backup.Spec.VolumeName, bkp.ObjectMeta.Labels["cstorpool.openebs.io/uid"])
			return "", nil
		}
	}

	//Get List of cvr's related to this pvc
	listOptions = v1.ListOptions{
		LabelSelector: "openebs.io/persistent-volume=" + backup.Spec.VolumeName,
	}
	cvrList, err := openebsClient.OpenebsV1alpha1().CStorVolumeReplicas(backup.Namespace).List(listOptions)

	//Select a healthy csr for backup
	for _, cvr := range cvrList.Items {
		if cvr.Status.Phase == v1alpha1.CVRStatusOnline {
			backup.ObjectMeta.Labels = map[string]string{
				"cstorpool.openebs.io/uid": cvr.ObjectMeta.Labels["cstorpool.openebs.io/uid"],
			}
			break
		}
	}

	glog.Infof("Creating backup %s for %s volume %q poolUUID:%v", backup.Spec.IncrementalBackupName,
		backup.Spec.Name, backup.Spec.VolumeName,
		backup.ObjectMeta.Labels["cstorpool.openebs.io/uid"])
	_, err = openebsClient.OpenebsV1alpha1().CStorBackups(backup.Namespace).Create(backup)
	if err != nil {
		glog.Errorf("Failed to create backup: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("Backup CR created successfully: name '%s'", backup.Name)
	return "", nil
}

// loadClientFromServiceAccount loads a k8s client from a ServiceAccount
// specified in the pod running
func loadClientFromServiceAccount() (*internalclientset.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	k8sClient, err := internalclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}
