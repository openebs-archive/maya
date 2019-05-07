// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"fmt"

	internalclientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
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
	if len(volumeList.Items) != 1 {
		return "", fmt.Errorf("Expected VolumeList count for volume: %v, Actual: %v", 1, len(volumeList.Items))
	}
	return string(volumeList.Items[0].Status.Phase), nil
}

/*
// CreateCSIVolumeCR creates a new CSIVolume CR with this nodeID
func CreateCSIVolumeCR(csivol *v1alpha1.CSIVolume, nodeID, mountPath string) (err error) {

	csivol.Name = csivol.Spec.Volume.Volname + "-" + nodeID
	csivol.Labels = make(map[string]string)
	csivol.Spec.Volume.OwnerNodeID = nodeID
	csivol.Labels["Volname"] = csivol.Spec.Volume.Volname
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
	_, err = openebsClient.OpenebsV1alpha1().CSIVolumes(OpenEBSNamespace).Create(csivol)
	return
}

// DeleteOldCSIVolumeCR deletes all CSIVolumes related to this volume so
// that a new one can be created with node as current nodeID
func DeleteOldCSIVolumeCR(vol *v1alpha1.CSIVolume) (err error) {
	openebsClient, _ := loadClientFromServiceAccount()
	listOptions := v1.ListOptions{
		LabelSelector: "Volname=" + vol.Name,
	}

	csivols, err := openebsClient.OpenebsV1alpha1().CSIVolumes(OpenEBSNamespace).List(listOptions)
	for _, csivol := range csivols.Items {
		err = openebsClient.OpenebsV1alpha1().CSIVolumes(OpenEBSNamespace).Delete(csivol.Name, &v1.DeleteOptions{})
		if err != nil {
			return
		}
	}
	return
}

// DeleteCSIVolumeCR removes the CSIVolume with this nodeID as
// labelSelector from the list
func DeleteCSIVolumeCR(vol *v1alpha1.CSIVolume) (err error) {
	openebsClient, _ := loadClientFromServiceAccount()
	var csivols *v1alpha1.CSIVolumeList
	listOptions := v1.ListOptions{
		LabelSelector: "Volname=" + vol.Spec.Volume.Volname,
	}

	csivols, err = openebsClient.OpenebsV1alpha1().CSIVolumes(OpenEBSNamespace).List(listOptions)
	if err != nil {
		return
	}
	for _, csivol := range csivols.Items {
		if csivol.Spec.Volume.OwnerNodeID == vol.Spec.Volume.OwnerNodeID {
			csivol.Finalizers = nil
			_, err = openebsClient.OpenebsV1alpha1().CSIVolumes(OpenEBSNamespace).Update(&csivol)
			if err != nil {
				return
			}

			err = openebsClient.OpenebsV1alpha1().CSIVolumes(OpenEBSNamespace).Delete(csivol.Name, &v1.DeleteOptions{})
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

	csivols, err := openebsClient.OpenebsV1alpha1().CSIVolumes(OpenEBSNamespace).List(listOptions)
	if err != nil {
		return
	}
	for _, csivol := range csivols.Items {
		vol := csivol
		Volumes[csivol.Spec.Volume.Volname] = &vol
	}
	return
}
*/
