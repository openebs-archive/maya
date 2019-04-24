package utils

import (
	"fmt"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	api_core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// getNodeDetails fetches the nodeInfo for the current node
func getNodeDetails(nodeID string) (node *api_core_v1.Node, err error) {
	kc, err := m_k8s_client.NewK8sClient("")
	if err != nil {
		return nil, err
	}
	node, err = kc.GetNode(nodeID, metav1.GetOptions{})
	return
}

// FetchPVDetails gets the PV related to this VolumeID
func FetchPVDetails(volumeID string) (pv *api_core_v1.PersistentVolume, err error) {
	kc, err := m_k8s_client.NewK8sClient("")
	if err != nil {
		return nil, err
	}
	pv, err = kc.GetPV(volumeID, metav1.GetOptions{})
	return
}

// loadClientFromServiceAccount loads a k8s client from a ServiceAccount
// specified in the pod running
func loadClientFromServiceAccount() (k8sClient *internalclientset.Clientset, err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return
	}
	k8sClient, err = internalclientset.NewForConfig(config)
	if err != nil {
		return
	}
	return
}

// getVolStatus fetches the current VolumeStatus which specifies if the volume
// is ready to serve IOs
func getVolStatus(volumeID string) (string, error) {
	openebsClient, _ := loadClientFromServiceAccount()
	listOptions := v1.ListOptions{
		LabelSelector: "openebs.io/persistent-volume=" + volumeID,
	}
	volumeList, err := openebsClient.OpenebsV1alpha1().CStorVolumes(OpenEBSNamespace).List(listOptions)
	if err != nil {
		return "", err
	}
	if len(volumeList.Items) > 1 {
		return "", fmt.Errorf("Expected VolumeList count for volume: %v, Actual: %v", 1, len(volumeList.Items))
	}
	return string(volumeList.Items[0].Status.Phase), nil
}

// CreateCSIVolumeInfoCR creates a new CSIVolumeInfo CR with this nodeID
func CreateCSIVolumeInfoCR(csivol *v1alpha1.CSIVolumeInfo, nodeID, mountPath string) (err error) {

	csivol.Name = csivol.Spec.Volname + nodeID
	csivol.Labels = make(map[string]string)
	csivol.Spec.OwnerNodeID = nodeID
	csivol.Labels["Volname"] = csivol.Spec.Volname
	csivol.Labels["nodeID"] = nodeID
	nodeInfo, err := getNodeDetails(nodeID)
	if err != nil {
		return
	}
	csivol.OwnerReferences = []v1.OwnerReference{
		{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       nodeInfo.Name,
			UID:        nodeInfo.UID,
		},
	}
	csivol.Finalizers = []string{nodeID}
	openebsClient, _ := loadClientFromServiceAccount()
	_, err = openebsClient.OpenebsV1alpha1().CSIVolumeInfos(OpenEBSNamespace).Create(csivol)
	return
}

// DeleteOldCSIVolumeInfoCR deletes all CSIVolumeInfos related to this volume so
// that a new one can be created with node as current nodeID
func DeleteOldCSIVolumeInfoCR(vol *v1alpha1.CSIVolumeInfo) (err error) {
	openebsClient, _ := loadClientFromServiceAccount()
	listOptions := v1.ListOptions{
		LabelSelector: "Volname=" + vol.Name,
	}

	csivols, err := openebsClient.OpenebsV1alpha1().CSIVolumeInfos(OpenEBSNamespace).List(listOptions)
	for _, csivol := range csivols.Items {
		err = openebsClient.OpenebsV1alpha1().CSIVolumeInfos(OpenEBSNamespace).Delete(csivol.Name, &v1.DeleteOptions{})
		if err != nil {
			return
		}
	}
	return
}

// DeleteCSIVolumeInfoCR removes the CSIVolumeInfo with this nodeID as
// labelSelector from the list
func DeleteCSIVolumeInfoCR(vol *v1alpha1.CSIVolumeInfo) (err error) {
	openebsClient, _ := loadClientFromServiceAccount()
	listOptions := v1.ListOptions{
		LabelSelector: "Volname=" + vol.Name,
	}

	csivols, err := openebsClient.OpenebsV1alpha1().CSIVolumeInfos(OpenEBSNamespace).List(listOptions)
	if err != nil {
		return
	}
	for _, csivol := range csivols.Items {
		if csivol.Spec.OwnerNodeID == vol.Spec.OwnerNodeID {
			err = openebsClient.OpenebsV1alpha1().CSIVolumeInfos(OpenEBSNamespace).Delete(csivol.Name, &v1.DeleteOptions{})
			if err != nil {
				return
			}
		}
	}
	return
}

// FetchAndUpdateVolInfos gets the list of CSIVolInfos that are supposed to be
// mounted on this node and stores the info in memory
// This is required when the CSI driver has restarted to start monitoring all
// the volumes and to reject duplicate volume creation requests
func FetchAndUpdateVolInfos(nodeID string) (err error) {
	var listOptions v1.ListOptions
	openebsClient, _ := loadClientFromServiceAccount()
	if nodeID != "" {
		listOptions = v1.ListOptions{
			LabelSelector: "nodeID=" + nodeID,
		}
	}

	csivols, err := openebsClient.OpenebsV1alpha1().CSIVolumeInfos(OpenEBSNamespace).List(listOptions)
	if err != nil {
		return
	}
	for _, csivol := range csivols.Items {
		vol := csivol
		Volumes[csivol.Spec.Volname] = &vol
	}
	return
}