/*
Copyright 2018 The OpenEBS Authors

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
	"github.com/openebs/maya/pkg/client/k8s"
	cstorpool "github.com/openebs/maya/pkg/cstorpool/v1alpha1"
	disk "github.com/openebs/maya/pkg/disk/v1alpha1"
	sp "github.com/openebs/maya/pkg/sp/v1alpha1"
)

const (
	// DiskStateActive is the active state of the disks.
	DiskStateActive        = "Active"
	ProvisioningTypeManual = "manual"
	ProvisioningTypeAuto   = "auto"
)

var defaultDiskCount = map[string]int{
	string(apis.PoolTypeMirroredCPV): int(apis.MirroredDiskCountCPV),
	string(apis.PoolTypeStripedCPV):  int(apis.StripedDiskCountCPV),
	string(apis.PoolTypeRaidzCPV):    int(apis.RaidzDiskCountCPV),
	string(apis.PoolTypeRaidz2CPV):   int(apis.Raidz2DiskCountCPV),
}

type diskList struct {
	Items []string
}

type nodeDisk struct {
	NodeName string
	Disks    diskList
}
type AlgorithmConfig struct {
	Spc              *apis.StoragePoolClaim
	DiskClient       disk.DiskInterface
	SpClient         sp.StoragepoolInterface
	CspClient        cstorpool.CstorpoolInterface
	ProvisioningType string
}

func getDiskK8sClient() *disk.KubernetesClient {
	newClient, _ := k8s.NewK8sClient("")
	K8sClient := &disk.KubernetesClient{
		newClient.GetKCS(),
		newClient.GetOECS(),
	}
	return K8sClient
}
func getDiskSpcClient(spc *apis.StoragePoolClaim) *disk.SpcObjectClient {
	K8sClient := &disk.SpcObjectClient{
		getDiskK8sClient(),
		spc,
	}
	return K8sClient
}

func getSpK8sClient() *sp.KubernetesClient {
	newClient, _ := k8s.NewK8sClient("")
	K8sClient := &sp.KubernetesClient{
		newClient.GetKCS(),
		newClient.GetOECS(),
	}
	return K8sClient
}

func getCspK8sClient() *cstorpool.KubernetesClient {
	newClient, _ := k8s.NewK8sClient("")
	K8sClient := &cstorpool.KubernetesClient{
		newClient.GetKCS(),
		newClient.GetOECS(),
	}
	return K8sClient
}

func NewAlgorithmConfig(spc *apis.StoragePoolClaim) *AlgorithmConfig {
	var diskK8sClient disk.DiskInterface
	if ProvisioningType(spc) == ProvisioningTypeManual {
		diskK8sClient = getDiskK8sClient()
	} else {
		diskK8sClient = getDiskSpcClient(spc)
	}

	cspK8sClient := getCspK8sClient()
	spK8sClient := getSpK8sClient()
	pT := ProvisioningType(spc)
	ac := &AlgorithmConfig{
		Spc:              spc,
		DiskClient:       diskK8sClient,
		CspClient:        cspK8sClient,
		SpClient:         spK8sClient,
		ProvisioningType: pT,
	}
	return ac
}
