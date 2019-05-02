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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	snapshot "github.com/openebs/maya/pkg/snapshot/v1alpha1"
	"k8s.io/client-go/kubernetes"
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
	case "GET":
		return backupOp.get()
	}
	return nil, CodedError(405, ErrInvalidMethod)
}

// Create is http handler which handles backup create request
func (bOps *backupAPIOps) create() (interface{}, error) {
	bkp := &v1alpha1.BackupCStor{}

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

	// backupIP is expected
	if len(strings.TrimSpace(bkp.Spec.BackupDest)) == 0 {
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

	glog.Infof("Creating backup %s for volume %q poolUUID:%v", bkp.Spec.SnapName,
		bkp.Spec.VolumeName,
		bkp.ObjectMeta.Labels["cstorpool.openebs.io/uid"])

	_, err = openebsClient.OpenebsV1alpha1().BackupCStors(bkp.Namespace).Create(bkp)
	if err != nil {
		glog.Errorf("Failed to create backup: error '%s'", err.Error())
		return nil, CodedError(500, err.Error())
	}

	glog.Infof("Backup resource:'%s' created successfully", bkp.Name)
	return "", nil
}

// createSnapshotForBackup will create a snapshot for given backup
func createSnapshotForBackup(bkp *v1alpha1.BackupCStor) error {
	snapOps, err := snapshot.Snapshot(&v1alpha1.SnapshotOptions{
		VolumeName: bkp.Spec.VolumeName,
		Namespace:  bkp.Namespace,
		CasType:    string(v1alpha1.CstorVolume),
		Name:       bkp.Spec.SnapName,
	})
	if err != nil {
		return CodedError(400, err.Error())
	}

	glog.Infof("Creating backup snapshot %s for volume %q", bkp.Spec.SnapName, bkp.Spec.VolumeName)

	snap, err := snapOps.Create()
	if err != nil {
		glog.Errorf("Failed to create snapshot:%s error '%s'", bkp.Spec.SnapName, err.Error())
		return CodedError(500, err.Error())
	}
	glog.Infof("Snapshot:'%s' created successfully for backup:%s", snap.Name, bkp.Name)
	return nil
}

// loadClientFromServiceAccount loads a k8s and openebs client from a ServiceAccount
// specified in the pod running
func loadClientFromServiceAccount() (*internalclientset.Clientset, *kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("Failed to fetch k8s cluster config. %+v", err)
		return nil, nil, err
	}

	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Errorf("Failed to create k8s client: %v", err)
		return nil, nil, err
	}

	openebsClient, err := internalclientset.NewForConfig(cfg)
	if err != nil {
		glog.Errorf("Failed to create openeEBS client. %+v", err)
		return nil, nil, err
	}

	return openebsClient, k8sClient, nil
}

// findHealthyCVR will find a healthy CVR for a given volume
func findHealthyCVR(openebsClient *internalclientset.Clientset, volume string) (v1alpha1.CStorVolumeReplica, error) {
	listOptions := v1.ListOptions{
		LabelSelector: "openebs.io/persistent-volume=" + volume,
	}

	cvrList, err := openebsClient.OpenebsV1alpha1().CStorVolumeReplicas("").List(listOptions)
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
func getLastBackupSnap(openebsClient *internalclientset.Clientset, bkp *v1alpha1.BackupCStor) (string, error) {
	lastbkpname := bkp.Spec.BackupName + "-" + bkp.Spec.VolumeName
	b, err := openebsClient.OpenebsV1alpha1().BackupCStorLasts(bkp.Namespace).Get(lastbkpname, v1.GetOptions{})
	if err != nil {
		bk := &v1alpha1.BackupCStorLast{
			ObjectMeta: v1.ObjectMeta{
				Name:      lastbkpname,
				Namespace: bkp.Namespace,
				Labels:    bkp.Labels,
			},
			Spec: v1alpha1.BackupCStorSpec{
				BackupName:   bkp.Spec.BackupName,
				VolumeName:   bkp.Spec.VolumeName,
				PrevSnapName: bkp.Spec.SnapName,
			},
		}

		_, err := openebsClient.OpenebsV1alpha1().BackupCStorLasts(bk.Namespace).Create(bk)
		if err != nil {
			glog.Errorf("Error creating last-backup resource for backup:%v err:%v", bk.Spec.BackupName, err)
			return "", err
		}
		glog.Infof("LastBackup resource created for backup:%s volume:%s", bk.Spec.BackupName, bk.Spec.VolumeName)
		return "", nil
	}
	return b.Spec.PrevSnapName, nil
}

// get is http handler which handles backup get request
func (bOps *backupAPIOps) get() (interface{}, error) {
	bkp := &v1alpha1.BackupCStor{}

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
	b, err := openebsClient.OpenebsV1alpha1().BackupCStors(bkp.Namespace).Get(bkp.Name, v1.GetOptions{})
	if err != nil {
		return nil, CodedError(400, fmt.Sprintf("Failed to fetch backup error:%v", err))
	}

	if b.Status != v1alpha1.BKPCStorStatusDone && b.Status != v1alpha1.BKPCStorStatusFailed {
		// check if node is running or not
		bkpNodeDown := checkIfCSPPoolNodeDown(k8sClient, b.Labels["cstorpool.openebs.io/uid"])
		// check if cstor-pool-mgmt container is running or not
		bkpPodDown := checkIfCSPPoolPodDown(k8sClient, b.Labels["cstorpool.openebs.io/uid"])

		if bkpNodeDown || bkpPodDown {
			// Backup is stalled, let's find last-backup status
			laststat := findLastBackupStat(openebsClient, b)
			// Update Backup status according to last-backup
			updateBackupStatus(openebsClient, b, laststat)

			// Get updated backup object
			b, err = openebsClient.OpenebsV1alpha1().BackupCStors(bkp.Namespace).Get(bkp.Name, v1.GetOptions{})
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
		glog.Errorf("Failed to find pod for cstorID:%v err:%s", cstorID, err.Error())
		return nodeDown
	}

	if pod.Spec.NodeName == "" {
		return nodeDown
	}

	node, err := k8sclient.CoreV1().Nodes().Get(pod.Spec.NodeName, v1.GetOptions{})
	if err != nil {
		glog.Infof("Failed to fetch node info for cstorID:%v: %v", cstorID, err)
		return nodeDown
	}
	for _, nodestat := range node.Status.Conditions {
		if nodestat.Type == corev1.NodeReady && nodestat.Status != corev1.ConditionTrue {
			glog.Infof("Node:%v is not in ready state", node.Name)
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
		glog.Errorf("Failed to find pod for cstorID:%v err:%s", cstorID, err.Error())
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

	podlist, err := k8sclient.CoreV1().Pods(openebsNs).List(podlistops)
	if err != nil {
		glog.Errorf("Failed to fetch pod list :%v", err)
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

// findLastBackupStat will find the status of given backup from last-backup
func findLastBackupStat(clientset internalclientset.Interface, bkp *v1alpha1.BackupCStor) v1alpha1.BackupCStorStatus {
	lastbkpname := bkp.Spec.BackupName + "-" + bkp.Spec.VolumeName
	lastbkp, err := clientset.OpenebsV1alpha1().BackupCStorLasts(bkp.Namespace).Get(lastbkpname, v1.GetOptions{})
	if err != nil {
		// Unable to fetch the last backup, so we will return fail state
		glog.Errorf("Failed to fetch last-backup:%s error:%s", lastbkpname, err.Error())
		return v1alpha1.BKPCStorStatusFailed
	}

	// let's check if snapname matches with current snapshot name
	if bkp.Spec.SnapName == lastbkp.Spec.SnapName || bkp.Spec.SnapName == lastbkp.Spec.PrevSnapName {
		return v1alpha1.BKPCStorStatusDone
	}

	// lastbackup snap/prevsnap doesn't match with bkp snapname
	return v1alpha1.BKPCStorStatusFailed
}

// updateBackupStatus will update the backup status to given status
func updateBackupStatus(clientset internalclientset.Interface, bkp *v1alpha1.BackupCStor, status v1alpha1.BackupCStorStatus) {
	bkp.Status = status

	_, err := clientset.OpenebsV1alpha1().BackupCStors(bkp.Namespace).Update(bkp)
	if err != nil {
		glog.Errorf("Failed to update backup:%s with status:%v", bkp.Name, status)
		return
	}
}
