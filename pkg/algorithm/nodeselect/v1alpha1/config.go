/*
Copyright 2019 The OpenEBS Authors

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

package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha1"
	"github.com/openebs/maya/pkg/client/k8s"
	cstorpool "github.com/openebs/maya/pkg/cstor/pool/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	sp "github.com/openebs/maya/pkg/sp/v1alpha1"
)

const (
	// DiskStateActive is the active state of the disks.
	DiskStateActive = "Active"
	// ProvisioningTypeManual is the manual provisioned SPC.
	ProvisioningTypeManual = "manual"
	// ProvisioningTypeAuto is the auto provisioned SPC.
	ProvisioningTypeAuto = "auto"
)

type blockDeviceList struct {
	Items []string
}

type nodeBlockDevice struct {
	NodeName     string
	BlockDevices blockDeviceList
}

// Config embeds clients for disk,csp and sp and contains, SPC object and ProvisioningType field which should tell
// provisioning type manual or auto.
type Config struct {
	// Spc is the StoragePoolClaim object.
	Spc *apis.StoragePoolClaim
	// BlockDeviceClient is the client for Disk to perform CRUD operations on Disk object.
	BlockDeviceClient blockdevice.BlockDeviceInterface
	// SpClient is the client for SP to perform CRUD operations on SP object.
	SpClient sp.StoragepoolInterface
	// CspClient is the client for CSP to perform CRUD operations on CSP object.
	CspClient cstorpool.CstorpoolInterface
	// ProvisioningType tells the type of provisioning i.e. manual or auto.
	ProvisioningType string
}

// getDiskK8sClient returns an instance of kubernetes client for Disk.
func getBlockDeviceK8sClient() *blockdevice.KubernetesClient {
	namespace := env.Get(env.OpenEBSNamespace)
	newClient, _ := k8s.NewK8sClient("")
	K8sClient := &blockdevice.KubernetesClient{
		Kubeclientset: newClient.GetKCS(),
		Clientset:     newClient.GetNDMCS(),
		Namespace:     namespace,
	}
	return K8sClient
}

// getDiskSpcClient returns an instance of SPC client for Block Device.
// NOTE : SPC is a typed client which embeds regular kubernetes block device client and SPC object.
// This client is used in manual provisioning of SPC.
func getBlockDeviceSpcClient(spc *apis.StoragePoolClaim) *blockdevice.SpcObjectClient {
	K8sClient := &blockdevice.SpcObjectClient{
		KubernetesClient: getBlockDeviceK8sClient(),
		Spc:              spc,
	}
	return K8sClient
}

// getSpK8sClient returns an instance of kubernetes client for SP.
// TODO: Deprecate SP
func getSpK8sClient() *sp.KubernetesClient {
	newClient, _ := k8s.NewK8sClient("")
	K8sClient := &sp.KubernetesClient{
		Kubeclientset: newClient.GetKCS(),
		Clientset:     newClient.GetOECS(),
	}
	return K8sClient
}

// getCspK8sClient returns an instance of kubernetes client for CSP.
func getCspK8sClient() *cstorpool.KubernetesClient {
	newClient, _ := k8s.NewK8sClient("")
	K8sClient := &cstorpool.KubernetesClient{
		Kubeclientset: newClient.GetKCS(),
		Clientset:     newClient.GetOECS(),
	}
	return K8sClient
}

// NewConfig returns an instance of Config based on SPC object.
func NewConfig(spc *apis.StoragePoolClaim) *Config {
	var bdClient blockdevice.BlockDeviceInterface

	// If provisioning type is manual blockdeviceClient is assigned SPC
	// blockdevice client
	// else it is assigned kubernetes blockdevice client.
	if ProvisioningType(spc) == ProvisioningTypeManual {
		bdClient = getBlockDeviceSpcClient(spc)
	} else {
		bdClient = getBlockDeviceK8sClient()
	}

	cspK8sClient := getCspK8sClient()
	spK8sClient := getSpK8sClient()
	pT := ProvisioningType(spc)
	ac := &Config{
		Spc:               spc,
		BlockDeviceClient: bdClient,
		CspClient:         cspK8sClient,
		SpClient:          spK8sClient,
		ProvisioningType:  pT,
	}
	return ac
}
