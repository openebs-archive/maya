/*
Copyright 2019 The OpenEBS Authors.

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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	snapshot "github.com/openebs/maya/pkg/snapshot/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
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
	case "GET":
		return backupOp.get()
	case "DELETE":
		return backupOp.delete()
	}
	return nil, CodedError(405, ErrInvalidMethod)
}

// Create is http handler which handles backup create request
func (bOps *backupAPIOps) create() (interface{}, error) {
	bkp := &v1alpha1.CStorBackup{}

	err := decodeBody(bOps.req, bkp)
	if err != nil {
		return nil, err
	}

	// backup name is expected
	if len(strings.TrimSpace(bkp.Spec.BackupName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("Failed to create backup: missing backup name "))
	}

	// namespace is expected
	if len(strings.TrimSpace(bkp.Namespace)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("Failed to create backup '%v': missing namespace", bkp.Spec.BackupName))
	}

	// volume name is expected
	if len(strings.TrimSpace(bkp.Spec.VolumeName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("Failed to create backup '%v': missing volume name", bkp.Spec.BackupName))
	}

	// backupIP is expected for remote snapshot
	if !bkp.Spec.LocalSnap && len(strings.TrimSpace(bkp.Spec.BackupDest)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("Failed to create backup '%v': missing backupIP", bkp.Spec.BackupName))
	}

	// snapshot name is expected
	if len(strings.TrimSpace(bkp.Spec.SnapName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("Failed to create backup '%v': missing snapName", bkp.Spec.BackupName))
	}

	openebsClient, _, err := loadClientFromServiceAccount()
	if err != nil {
		return nil, CodedError(400, fmt.Sprintf("Failed to create openEBSClient '%v'", err))
	}

	if err = createSnapshotForBackup(bkp); err != nil {
		return nil, CodedError(400, fmt.Sprintf("Failed to create snapshot '%v'", err))
	}

	bkp.Name = bkp.Spec.SnapName + "-" + bkp.Spec.VolumeName

	// find healthy CVR
	cvr, err := findHealthyCVR(openebsClient, bkp.Spec.VolumeName)
	if err != nil {
		return nil, CodedError(400, fmt.Sprintf("Failed to find healthy replica"))
	}

	if bkp.Spec.LocalSnap {
		return "", nil
	}

	bkp.ObjectMeta.Labels = map[string]string{
		"cstorpool.openebs.io/uid":     cvr.ObjectMeta.Labels["cstorpool.openebs.io/uid"],
		"openebs.io/persistent-volume": cvr.ObjectMeta.Labels["openebs.io/persistent-volume"],
		"openebs.io/backup":            bkp.Spec.BackupName,
	}

	// Find last backup snapshot name
	lastsnap, err := getLastBackupSnap(openebsClient, bkp)
	if err != nil {
		return nil, CodedError(400, fmt.Sprintf("Failed create lastbackup"))
	}

	// Initialize backup status as pending
	bkp.Status = v1alpha1.BKPCStorStatusPending
	bkp.Spec.PrevSnapName = lastsnap

	klog.Infof("Creating backup %s for volume %q poolUUID:%v", bkp.Spec.SnapName,
		bkp.Spec.VolumeName,
		bkp.ObjectMeta.Labels["cstorpool.openebs.io/uid"])

	_, err = openebsClient.OpenebsV1alpha1().CStorBackups(bkp.Namespace).
		Create(context.TODO(), bkp, v1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create backup: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	klog.Infof("Backup resource:'%s' created successfully", bkp.Name)
	return "", nil
}

// createSnapshotForBackup will create a snapshot for given backup
func createSnapshotForBackup(bkp *v1alpha1.CStorBackup) error {
	snapOps, err := snapshot.Snapshot(&v1alpha1.SnapshotOptions{
		VolumeName: bkp.Spec.VolumeName,
		Namespace:  bkp.Namespace,
		CasType:    string(v1alpha1.CstorVolume),
		Name:       bkp.Spec.SnapName,
	})
	if err != nil {
		return CodedError(400, err.Error())
	}

	klog.Infof("Creating backup snapshot %s for volume %q", bkp.Spec.SnapName, bkp.Spec.VolumeName)

	snap, err := snapOps.Create()
	if err != nil {
		klog.Errorf("Failed to create snapshot:%s error '%s'", bkp.Spec.SnapName, err.Error())
		return CodedError(500, err.Error())
	}
	klog.Infof("Backup snapshot:'%s' created successfully for volume:%s", snap.Name, bkp.Spec.VolumeName)
	return nil
}

// loadClientFromServiceAccount loads a k8s and openebs client from a ServiceAccount
// specified in the pod running
func loadClientFromServiceAccount() (*versioned.Clientset, *kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorf("Failed to fetch k8s cluster config. %+v", err)
		return nil, nil, err
	}

	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("Failed to create k8s client: %v", err)
		return nil, nil, err
	}

	openebsClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("Failed to create openeEBS client. %+v", err)
		return nil, nil, err
	}

	return openebsClient, k8sClient, nil
}

// findHealthyCVR will find a healthy CVR for a given volume
func findHealthyCVR(openebsClient *versioned.Clientset, volume string) (v1alpha1.CStorVolumeReplica, error) {
	listOptions := v1.ListOptions{
		LabelSelector: "openebs.io/persistent-volume=" + volume,
	}

	cvrList, err := openebsClient.OpenebsV1alpha1().CStorVolumeReplicas("").
		List(context.TODO(), listOptions)
	if err != nil {
		return v1alpha1.CStorVolumeReplica{}, err
	}

	// Select a healthy cvr for backup
	for _, cvr := range cvrList.Items {
		if cvr.Status.Phase == v1alpha1.CVRStatusOnline {
			return cvr, nil
		}
	}

	return v1alpha1.CStorVolumeReplica{}, errors.New("unable to find healthy CVR")
}

// getLastBackupSnap will fetch the last successful backup's snapshot name
func getLastBackupSnap(openebsClient *versioned.Clientset, bkp *v1alpha1.CStorBackup) (string, error) {
	lastbkpname := bkp.Spec.BackupName + "-" + bkp.Spec.VolumeName
	b, err := openebsClient.OpenebsV1alpha1().CStorCompletedBackups(bkp.Namespace).
		Get(context.TODO(), lastbkpname, v1.GetOptions{})
	if err != nil {
		bk := &v1alpha1.CStorCompletedBackup{
			ObjectMeta: v1.ObjectMeta{
				Name:      lastbkpname,
				Namespace: bkp.Namespace,
				Labels:    bkp.Labels,
			},
			Spec: v1alpha1.CStorBackupSpec{
				BackupName: bkp.Spec.BackupName,
				VolumeName: bkp.Spec.VolumeName,
			},
		}

		_, err := openebsClient.OpenebsV1alpha1().CStorCompletedBackups(bk.Namespace).
			Create(context.TODO(), bk, v1.CreateOptions{})
		if err != nil {
			klog.Errorf("Error creating last completed-backup resource for backup:%v err:%v", bk.Spec.BackupName, err)
			return "", err
		}
		klog.Infof("LastBackup resource created for backup:%s volume:%s", bk.Spec.BackupName, bk.Spec.VolumeName)
		return "", nil
	}

	// PrevSnapName stores the last completed backup snapshot
	return b.Spec.PrevSnapName, nil
}

// get is http handler which handles backup get request
// It will delete the snapshot created by the given backup if backup is done/failed
func (bOps *backupAPIOps) get() (interface{}, error) {
	bkp := &v1alpha1.CStorBackup{}

	err := decodeBody(bOps.req, bkp)
	if err != nil {
		return nil, err
	}

	// backup name is expected
	if len(strings.TrimSpace(bkp.Spec.BackupName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("Failed to get backup: missing backup name "))
	}

	// namespace is expected
	if len(strings.TrimSpace(bkp.Namespace)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("Failed to get backup '%v': missing namespace", bkp.Spec.BackupName))
	}

	// volume name is expected
	if len(strings.TrimSpace(bkp.Spec.VolumeName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("Failed to get backup '%v': missing volume name", bkp.Spec.BackupName))
	}

	openebsClient, k8sClient, err := loadClientFromServiceAccount()
	if err != nil {
		return nil, CodedError(400, fmt.Sprintf("Failed to create openEBSClient '%v'", err))
	}

	bkp.Name = bkp.Spec.SnapName + "-" + bkp.Spec.VolumeName
	b, err := openebsClient.OpenebsV1alpha1().CStorBackups(bkp.Namespace).
		Get(context.TODO(), bkp.Name, v1.GetOptions{})
	if err != nil {
		return nil, CodedError(400, fmt.Sprintf("Failed to fetch backup error:%v", err))
	}

	if !isBackupCompleted(b) {
		// check if node is running or not
		bkpNodeDown := checkIfCSPPoolNodeDown(k8sClient, b.Labels["cstorpool.openebs.io/uid"])
		// check if cstor-pool-mgmt container is running or not
		bkpPodDown := checkIfCSPPoolPodDown(k8sClient, b.Labels["cstorpool.openebs.io/uid"])

		if bkpNodeDown || bkpPodDown {
			// Backup is stalled, let's find last completed-backup status
			laststat := findLastBackupStat(openebsClient, b)
			// Update Backup status according to last completed-backup
			updateBackupStatus(openebsClient, b, laststat)

			// Get updated backup object
			b, err = openebsClient.OpenebsV1alpha1().CStorBackups(bkp.Namespace).
				Get(context.TODO(), bkp.Name, v1.GetOptions{})
			if err != nil {
				return nil, CodedError(400, fmt.Sprintf("Failed to fetch backup error:%v", err))
			}
		}
	}

	out, err := json.Marshal(b)
	if err == nil {
		_, err = bOps.resp.Write(out)
		if err != nil {
			return nil, CodedError(400, fmt.Sprintf("Failed to send response data"))
		}
		return nil, nil
	}

	return nil, CodedError(400, fmt.Sprintf("Failed to encode response data"))
}

// checkIfCSPPoolNodeDown will check if CSP pool node is running or not
func checkIfCSPPoolNodeDown(k8sclient *kubernetes.Clientset, cstorID string) bool {
	var nodeDown = true

	pod, err := findPodFromCStorID(k8sclient, cstorID)
	if err != nil {
		klog.Errorf("Failed to find pod for cstorID:%v err:%s", cstorID, err.Error())
		return nodeDown
	}

	if pod.Spec.NodeName == "" {
		return nodeDown
	}

	node, err := k8sclient.CoreV1().Nodes().
		Get(context.TODO(), pod.Spec.NodeName, v1.GetOptions{})
	if err != nil {
		klog.Infof("Failed to fetch node info for cstorID:%v: %v", cstorID, err)
		return nodeDown
	}
	for _, nodestat := range node.Status.Conditions {
		if nodestat.Type == corev1.NodeReady && nodestat.Status != corev1.ConditionTrue {
			klog.Infof("Node:%v is not in ready state", node.Name)
			return nodeDown
		}
	}
	return !nodeDown
}

// checkIfCSPPoolPodDown will check if pool pod is running or not
func checkIfCSPPoolPodDown(k8sclient *kubernetes.Clientset, cstorID string) bool {
	var podDown = true

	pod, err := findPodFromCStorID(k8sclient, cstorID)
	if err != nil {
		klog.Errorf("Failed to find pod for cstorID:%v err:%s", cstorID, err.Error())
		return podDown
	}

	for _, containerstatus := range pod.Status.ContainerStatuses {
		if containerstatus.Name == "cstor-pool-mgmt" {
			return !containerstatus.Ready
		}
	}

	return podDown
}

// findPodFromCStorID will find the Pod having given cstorID
func findPodFromCStorID(k8sclient *kubernetes.Clientset, cstorID string) (corev1.Pod, error) {
	cstorPodLabel := "app=cstor-pool"
	podlistops := v1.ListOptions{
		LabelSelector: cstorPodLabel,
	}

	openebsNs := os.Getenv("OPENEBS_NAMESPACE")
	if openebsNs == "" {
		return corev1.Pod{}, errors.New("Failed to fetch operator namespace")
	}

	podlist, err := k8sclient.CoreV1().Pods(openebsNs).
		List(context.TODO(), podlistops)
	if err != nil {
		klog.Errorf("Failed to fetch pod list :%v", err)
		return corev1.Pod{}, errors.New("Failed to fetch pod list")
	}

	for _, pod := range podlist.Items {
		for _, env := range pod.Spec.Containers[0].Env {
			if env.Name == "OPENEBS_IO_CSTOR_ID" && env.Value == cstorID {
				return pod, nil
			}
		}
	}
	return corev1.Pod{}, errors.New("No Pod exists")
}

// findLastBackupStat will find the status of given backup from last completed-backup
func findLastBackupStat(clientset versioned.Interface, bkp *v1alpha1.CStorBackup) v1alpha1.CStorBackupStatus {
	lastbkpname := bkp.Spec.BackupName + "-" + bkp.Spec.VolumeName
	lastbkp, err := clientset.OpenebsV1alpha1().CStorCompletedBackups(bkp.Namespace).
		Get(context.TODO(), lastbkpname, v1.GetOptions{})
	if err != nil {
		// Unable to fetch the last backup, so we will return fail state
		klog.Errorf("Failed to fetch last completed-backup:%s error:%s", lastbkpname, err.Error())
		return v1alpha1.BKPCStorStatusFailed
	}

	// lastbkp stores the last(PrevSnapName) and 2nd last(SnapName) completed snapshot
	// let's check if last backup's snapname/PrevSnapName  matches with current snapshot name
	if bkp.Spec.SnapName == lastbkp.Spec.SnapName || bkp.Spec.SnapName == lastbkp.Spec.PrevSnapName {
		return v1alpha1.BKPCStorStatusDone
	}

	// lastbackup snap/prevsnap doesn't match with bkp snapname
	return v1alpha1.BKPCStorStatusFailed
}

// updateBackupStatus will update the backup status to given status
func updateBackupStatus(clientset versioned.Interface, bkp *v1alpha1.CStorBackup, status v1alpha1.CStorBackupStatus) {
	bkp.Status = status

	_, err := clientset.OpenebsV1alpha1().CStorBackups(bkp.Namespace).
		Update(context.TODO(), bkp, v1.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update backup:%s with status:%v", bkp.Name, status)
		return
	}
}

// delete is http handler which handles backup delete request
func (bOps *backupAPIOps) delete() (interface{}, error) {
	// Extract name of backup from path after trimming
	backup := strings.TrimSpace(strings.TrimPrefix(bOps.req.URL.Path, "/latest/backups/"))

	// volname is the volume name in the query params
	volname := bOps.req.URL.Query().Get("volume")

	// namespace is the namespace(pvc namespace) name in the query params
	namespace := bOps.req.URL.Query().Get("namespace")

	// schedule name is the schedule name for the given backup, for non-scheduled backup it will be backup name
	scheduleName := bOps.req.URL.Query().Get("schedule")

	if len(backup) == 0 || len(volname) == 0 || len(namespace) == 0 || len(scheduleName) == 0 {
		return nil, CodedError(405, "failed to delete backup: Insufficient info -- required values volume_name, backup_name, namespace, schedule_name")
	}

	klog.Infof("Deleting backup=%s for volume=%s with namesace=%s and schedule=%s", backup, volname, namespace, scheduleName)

	openebsClient, _, err := loadClientFromServiceAccount()
	if err != nil {
		return nil, CodedError(500, fmt.Sprintf("Failed to create openEBSClient '%v'", err))
	}

	err = deleteBackup(openebsClient, backup, volname, namespace, scheduleName)
	if err != nil {
		klog.Errorf("Error deleting backup=%s for volume=%s with namesace=%s and schedule=%s..  error=%s", backup, volname, namespace, scheduleName, err)
		return nil, CodedError(500, fmt.Sprintf("Error deleting backup=%s for volume=%s with namesace=%s and schedule=%s..  error=%s", backup, volname, namespace, scheduleName, err))
	}
	return "", nil
}

// deleteBackup delete the relevant cstorBackup/cstorCompletedBackup resource and cstor snapshot for the given backup
func deleteBackup(client *versioned.Clientset, backup, volname, ns, schedule string) error {
	lastCompletedBackup := schedule + "-" + volname

	// Let's get the cstorCompletedBackup resource for the given backup
	// CStorCompletedBackups resource stores the information about last two completed snapshots
	lastbkp, err := client.OpenebsV1alpha1().CStorCompletedBackups(ns).
		Get(context.TODO(), lastCompletedBackup, v1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to fetch last-completed-backup=%s resource", lastCompletedBackup)
	}

	// lastbkp stores the last(PrevSnapName) and 2nd last(SnapName) completed snapshot
	// If given backup is the last backup of scheduled backup (lastbkp.Spec.PrevSnapName == backup) or
	// completedBackup doesn't have successful backup(len(lastbkp.Spec.PrevSnapName) == 0) then we will delete the lastbkp CR
	// Deleting this CR make sure that next backup of the schedule will be full backup
	if lastbkp != nil && (lastbkp.Spec.PrevSnapName == backup || len(lastbkp.Spec.PrevSnapName) == 0) {
		err := client.OpenebsV1alpha1().CStorCompletedBackups(ns).
			Delete(context.TODO(), lastCompletedBackup, v1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return errors.Wrapf(err, "failed to delete last-completed-backup=%s resource", lastCompletedBackup)
		}
	}

	// Snapshot Name and backup name are same
	err = deleteSnapshot(backup, volname, ns)
	if err != nil {
		return errors.Wrapf(err, "failed to delete snapshot=%s for volume=%s", backup, volname)
	}

	cstorBackup := backup + "-" + volname
	err = client.OpenebsV1alpha1().CStorBackups(ns).Delete(context.TODO(), cstorBackup, v1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete cstorbackup=%s resource", cstorBackup)
	}
	return nil
}

// deleteSnapshot will delete the given snapshot for the volume
func deleteSnapshot(snapname, volname, ns string) error {
	snapOps, err := snapshot.Snapshot(&v1alpha1.SnapshotOptions{
		VolumeName: volname,
		Namespace:  ns,
		CasType:    string(v1alpha1.CstorVolume),
		Name:       snapname,
	})
	if err != nil {
		return CodedError(400, err.Error())
	}

	klog.Infof("Deleting backup snapshot %s for volume %q", snapname, volname)

	_, err = snapOps.Delete()
	if err != nil {
		return errors.Wrapf(err, "Failed to delete snapshot")
	}
	klog.Infof("Snapshot:'%s' deleted successfully for volume:%s", snapname, volname)
	return nil
}

// isBackupCompleted returns true if backup execution is completed
func isBackupCompleted(bkp *v1alpha1.CStorBackup) bool {
	if isBackupFailed(bkp) || isBackupSucceeded(bkp) {
		return true
	}
	return false
}

// isBackupFailed returns true if backup failed
func isBackupFailed(bkp *v1alpha1.CStorBackup) bool {
	if bkp.Status == v1alpha1.BKPCStorStatusFailed {
		return true
	}
	return false
}

// isBackupSucceeded returns true if backup completed successfully
func isBackupSucceeded(bkp *v1alpha1.CStorBackup) bool {
	if bkp.Status == v1alpha1.BKPCStorStatusDone {
		return true
	}
	return false
}
