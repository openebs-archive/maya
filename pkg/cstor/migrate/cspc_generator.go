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
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstor/pool/v1alpha3"
	cspc "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	cspi "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	"github.com/openebs/maya/pkg/util/retry"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

func getBDList(cspObj apis.CStorPool) []apis.CStorPoolClusterBlockDevice {
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

func getCSPCSpecForSPC(spc *apis.StoragePoolClaim, openebsNamespace string) (*apis.CStorPoolCluster, error) {
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
		string(apis.OpenEBSDisableDependantsReconcileKey): "true",
	}
	for _, cspObj := range cspList.Items {
		cspDeployList, err := deploy.NewKubeClient().WithNamespace(openebsNamespace).
			List(&metav1.ListOptions{
				LabelSelector: "openebs.io/cstor-pool=" + cspObj.Name,
			})
		if err != nil {
			return nil, err
		}
		if len(cspDeployList.Items) != 1 {
			return nil, errors.Errorf("invalid number of csp deployment found: %d", len(cspDeployList.Items))
		}
		poolSpec := apis.PoolSpec{
			NodeSelector: map[string]string{
				string(apis.HostNameCPK): cspObj.Labels[string(apis.HostNameCPK)],
			},
			RaidGroups: []apis.RaidGroup{
				apis.RaidGroup{
					Type:         typeMap[cspObj.Spec.PoolSpec.PoolType],
					BlockDevices: getBDList(cspObj),
				},
			},
			PoolConfig: apis.PoolConfig{
				CacheFile:            cspObj.Spec.PoolSpec.CacheFile,
				DefaultRaidGroupType: typeMap[cspObj.Spec.PoolSpec.PoolType],
				OverProvisioning:     cspObj.Spec.PoolSpec.OverProvisioning,
				Resources:            getCSPResources(cspDeployList.Items[0]),
				Tolerations:          cspDeployList.Items[0].Spec.Template.Spec.Tolerations,
				// AuxResources:         getCSPAuxResources(cspDeployList.Items[0]),
				// PriorityClassName:    cspDeployList.Items[0].Spec.Template.Spec.PriorityClassName,
			},
		}
		// if the csp does not have a cachefile then add cachefile
		if poolSpec.PoolConfig.CacheFile == "" {
			poolSpec.PoolConfig.CacheFile = "/tmp/pool1.cache"
		}
		cspcObj.Spec.Pools = append(cspcObj.Spec.Pools, poolSpec)
	}
	return cspcObj, nil
}

func getCSPResources(cspDeploy appsv1.Deployment) *corev1.ResourceRequirements {
	for _, con := range cspDeploy.Spec.Template.Spec.Containers {
		if con.Name == "cstor-pool" {
			return &con.Resources
		}
	}
	return nil
}

// func getCSPAuxResources(cspDeploy appsv1.Deployment) *corev1.ResourceRequirements {
// 	for _, con := range cspDeploy.Spec.Template.Spec.Containers {
// 		if con.Name == "cstor-pool-mgmt" {
// 			return &con.Resources
// 		}
// 	}
// 	return nil
// }

// generateCSPC creates an equivalent cspc for the given spc object
func generateCSPC(spcObj *apis.StoragePoolClaim, openebsNamespace string) (
	*apis.CStorPoolCluster, error) {
	cspcObj, err := cspc.NewKubeClient().
		WithNamespace(openebsNamespace).Get(spcObj.Name, metav1.GetOptions{})
	if err == nil {
		return cspcObj, nil
	}
	if !k8serrors.IsNotFound(err) {
		return nil, err
	}
	cspcObj, err = getCSPCSpecForSPC(spcObj, openebsNamespace)
	if err != nil {
		return nil, err
	}
	cspcObj, err = cspc.NewKubeClient().
		WithNamespace(openebsNamespace).Create(cspcObj)
	if err != nil {
		return nil, err
	}
	// verify the number of cspi created is correct
	err = retry.
		Times(60).
		Wait(5 * time.Second).
		Try(func(attempt uint) error {
			cspiList, err1 := cspi.NewKubeClient().
				WithNamespace(openebsNamespace).List(
				metav1.ListOptions{
					LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + cspcObj.Name,
				})
			if err1 != nil {
				return err1
			}
			if len(cspiList.Items) != len(cspcObj.Spec.Pools) {
				return errors.Errorf("failed to verify cspi count expected: %d got: %d",
					len(cspcObj.Spec.Pools),
					len(cspiList.Items),
				)
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	cspcObj, err = cspc.NewKubeClient().
		WithNamespace(openebsNamespace).Get(spcObj.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// after all the cspi come up which reconcilation disabled delete the
	// OpenEBSDisableDependantsReconcileKey annotation so that in future when
	// a cspi is delete and it comes back on reconciliation it should not have
	// reconciliation disabled
	delete(cspcObj.Annotations, string(apis.OpenEBSDisableDependantsReconcileKey))
	cspcObj, err = cspc.NewKubeClient().
		WithNamespace(openebsNamespace).
		Update(cspcObj)
	if err != nil {
		return nil, err
	}
	return cspcObj, nil
}
