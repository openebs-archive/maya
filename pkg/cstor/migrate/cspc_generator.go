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

package migrate

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstor/pool/v1alpha3"
	cspc "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	typeMap = map[string]string{
		string(apis.PoolTypeStripedCPV):  string(apis.PoolStriped),
		string(apis.PoolTypeMirroredCPV): string(apis.PoolMirrored),
		string(apis.PoolTypeRaidzCPV):    string(apis.PoolRaidz),
		string(apis.PoolTypeRaidz2CPV):   string(apis.PoolRaidz2),
	}
)

func getBDCList(cspObj apis.CStorPool) []apis.CStorPoolClusterBlockDevice {
	list := []apis.CStorPoolClusterBlockDevice{}
	for _, bdcObj := range cspObj.Spec.Group[0].Item {
		list = append(list,
			apis.CStorPoolClusterBlockDevice{
				BlockDeviceName: bdcObj.Name,
			},
		)
	}
	return list
}

func getCSPCSpec(spc *apis.StoragePoolClaim) (*apis.CStorPoolCluster, error) {
	cspClient := csp.KubeClient()
	cspList, err := cspClient.List(metav1.ListOptions{
		LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spc.Name,
	})
	if err != nil {
		return nil, err
	}
	cspcObj := &apis.CStorPoolCluster{}
	cspcObj.Name = spc.Name
	cspcObj.Annotations = map[string]string{
		// This label will be used to disable reconciliation on the dependants
		// In this case that will be CSPI
		"reconcile.openebs.io/dependants": "false",
	}
	for _, cspObj := range cspList.Items {
		cspcObj.Spec.Pools = append(cspcObj.Spec.Pools,
			apis.PoolSpec{
				NodeSelector: map[string]string{
					string(apis.HostNameCPK): cspObj.Labels[string(apis.HostNameCPK)],
				},
				RaidGroups: []apis.RaidGroup{
					apis.RaidGroup{
						Type:         typeMap[cspObj.Spec.PoolSpec.PoolType],
						BlockDevices: getBDCList(cspObj),
					},
				},
				PoolConfig: apis.PoolConfig{
					CacheFile:            cspObj.Spec.PoolSpec.CacheFile,
					DefaultRaidGroupType: typeMap[cspObj.Spec.PoolSpec.PoolType],
					OverProvisioning:     cspObj.Spec.PoolSpec.OverProvisioning,
				},
				OldCSPUID: string(cspObj.UID),
			},
		)

	}
	return cspcObj, nil
}

func generateCSPC(spcName, openebsNamespace string) error {
	spcObj, err := spc.NewKubeClient().Get(spcName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	cspcObj, err := getCSPCSpec(spcObj)
	if err != nil {
		return err
	}
	_, err = cspc.NewKubeClient().WithNamespace(openebsNamespace).Create(cspcObj)
	if err != nil {
		return err
	}
	return nil
}
