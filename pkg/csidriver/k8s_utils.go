package driver

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

func getNodeDetails(nodeID string) (node *api_core_v1.Node, err error) {
	kc, err := m_k8s_client.NewK8sClient("")
	if err != nil {
		return nil, err
	}
	node, err = kc.GetNode(nodeID, metav1.GetOptions{})
	return
}

func fetchPVDetails(volumeID string) (pv *api_core_v1.PersistentVolume, err error) {
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

func getVolStatus(volumeID string) (string, error) {
	openebsClient, _ := loadClientFromServiceAccount()
	listOptions := v1.ListOptions{
		LabelSelector: "openebs.io/persistent-volume=" + volumeID,
	}
	volumeList, err := openebsClient.OpenebsV1alpha1().CStorVolumes("openebs").List(listOptions)
	if err != nil {
		return "", err
	}
	if len(volumeList.Items) > 1 {
		return "", fmt.Errorf("Expected VolumeList count for volume: %v, Actual: %v", 1, len(volumeList.Items))
	}
	return string(volumeList.Items[0].Status.Phase), nil
}

func createCSIVolumeInfoCR(csivol *v1alpha1.CSIVolumeInfo, nodeID string, mountPath string) (err error) {

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
	_, err = openebsClient.OpenebsV1alpha1().CSIVolumeInfos("openebs").Create(csivol)
	return
}

func deleteOldCSIVolumeInfoCR(vol *v1alpha1.CSIVolumeInfo) (err error) {
	openebsClient, _ := loadClientFromServiceAccount()
	listOptions := v1.ListOptions{
		LabelSelector: "Volname=" + vol.Name,
	}

	csivols, err := openebsClient.OpenebsV1alpha1().CSIVolumeInfos("openebs").List(listOptions)
	for _, csivol := range csivols.Items {
		err = openebsClient.OpenebsV1alpha1().CSIVolumeInfos("openebs").Delete(csivol.Name, &v1.DeleteOptions{})
		if err != nil {
			return
		}
	}
	return
}

func deleteCSIVolumeInfoCR(vol *v1alpha1.CSIVolumeInfo) (err error) {
	openebsClient, _ := loadClientFromServiceAccount()
	listOptions := v1.ListOptions{
		LabelSelector: "Volname=" + vol.Name,
	}

	csivols, err := openebsClient.OpenebsV1alpha1().CSIVolumeInfos("openebs").List(listOptions)
	if err != nil {
		return
	}
	for _, csivol := range csivols.Items {
		if csivol.Spec.OwnerNodeID == vol.Spec.OwnerNodeID {
			err = openebsClient.OpenebsV1alpha1().CSIVolumeInfos("openebs").Delete(csivol.Name, &v1.DeleteOptions{})
			if err != nil {
				return
			}
		}
	}
	return
}

func fetchAndUpdateVolInfos(nodeID string) (err error) {
	openebsClient, _ := loadClientFromServiceAccount()
	listOptions := v1.ListOptions{
		LabelSelector: "nodeID=" + nodeID,
	}

	csivols, err := openebsClient.OpenebsV1alpha1().CSIVolumeInfos("openebs").List(listOptions)
	if err != nil {
		return
	}
	for _, csivol := range csivols.Items {
		Volumes[csivol.Spec.Volname] = &csivol
	}
	return
}
