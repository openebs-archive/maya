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
	"fmt"
	"time"

	"k8s.io/klog"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	bd "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstor/pool/v1alpha3"
	cspc "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	cspi "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/pkg/util/retry"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

const replicaPatch = `{
	"spec": {
		"replicas": 0
	}	
}`

// Pool ...
func Pool(spcName, openebsNamespace string) error {
	klog.Infof("Migrating spc %s to cspc", spcName)
	spcObj, err := spc.NewKubeClient().
		Get(spcName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		klog.Infof("spc %s not found.", spcName)
		_, err := cspc.NewKubeClient().
			WithNamespace(openebsNamespace).Get(spcName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to get equivalent cspc for spc %s", spcName)
		}
		klog.Infof("spc %s is already migrated to cspc", spcName)
		return nil
	}
	if err != nil {
		return err
	}
	klog.Infof("Creating equivalent cspc for spc %s", spcName)
	cspcObj, err := generateCSPC(spcObj, openebsNamespace)
	if err != nil {
		return err
	}

	cspiList, err := cspi.NewKubeClient().
		WithNamespace(openebsNamespace).
		List(metav1.ListOptions{
			LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + cspcObj.Name,
		})
	if err != nil {
		return err
	}

	for _, cspiObj := range cspiList.Items {
		if cspiObj.Status.Phase != "ONLINE" {
			err = csptocspi(&cspiObj, cspcObj, openebsNamespace)
			if err != nil {
				return err
			}
			cspcObj, err = cspc.NewKubeClient().
				WithNamespace(openebsNamespace).Get(cspcObj.Name, metav1.GetOptions{})
			for i, poolspec := range cspcObj.Spec.Pools {
				if poolspec.NodeSelector[string(apis.HostNameCPK)] ==
					cspiObj.Labels[string(apis.HostNameCPK)] {
					cspcObj.Spec.Pools[i].OldCSPUID = ""
				}
			}
			cspcObj, err = cspc.NewKubeClient().
				WithNamespace(openebsNamespace).Update(cspcObj)
			if err != nil {
				return err
			}
		}
	}
	err = spc.NewKubeClient().
		Delete(spcName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func csptocspi(cspiObj *apis.CStorPoolInstance, cspcObj *apis.CStorPoolCluster, openebsNamespace string) error {
	hostnameLabel := string(apis.HostNameCPK) + "=" + cspiObj.Labels[string(apis.HostNameCPK)]
	spcLabel := string(apis.StoragePoolClaimCPK) + "=" + cspcObj.Name
	cspLabel := hostnameLabel + "," + spcLabel

	cspObj, err := getCSP(cspLabel)
	if err != nil {
		return err
	}
	klog.Infof("Migrating csp %s to cspi %s", cspiObj.Name, cspObj.Name)
	err = scaleDownDeployment(cspObj, openebsNamespace)
	if err != nil {
		return err
	}
	for _, bdName := range cspObj.Spec.Group[0].Item {
		err = updateBDC(bdName, cspcObj, openebsNamespace)
		if err != nil {
			return err
		}
	}
	delete(cspiObj.Annotations, string(apis.OpenEBSDisableReconcileKey))
	cspiObj, err = cspi.NewKubeClient().
		WithNamespace(openebsNamespace).
		Update(cspiObj)
	if err != nil {
		return err
	}
	err = retry.
		Times(60).
		Wait(5 * time.Second).
		Try(func(attempt uint) error {
			cspiObj, err1 := cspi.NewKubeClient().
				WithNamespace(openebsNamespace).
				Get(cspiObj.Name, metav1.GetOptions{})
			if err1 != nil {
				return err1
			}
			if cspiObj.Status.Phase != "ONLINE" {
				return errors.Errorf("failed to verify cspi phase expected: Healthy got: %s",
					cspiObj.Status.Phase)
			}
			return nil
		})
	if err != nil {
		return err
	}
	err = updateCVRsLabels(cspObj.Name, openebsNamespace, cspiObj)
	if err != nil {
		return err
	}
	return nil
}

func getCSP(cspLabel string) (*apis.CStorPool, error) {
	cspClient := csp.KubeClient()
	cspList, err := cspClient.List(metav1.ListOptions{
		LabelSelector: cspLabel,
	})
	if err != nil {
		return nil, err
	}
	if len(cspList.Items) != 1 {
		return nil, fmt.Errorf("Invalid number of pools on one node: %d", len(cspList.Items))
	}
	cspObj := cspList.Items[0]
	if err != nil {
		return nil, err
	}
	return &cspObj, nil
}

func scaleDownDeployment(cspObj *apis.CStorPool, openebsNamespace string) error {
	klog.Infof("Scaling down deployemnt %s", cspObj.Name)
	cspPod, err := pod.NewKubeClient().
		WithNamespace(openebsNamespace).List(
		metav1.ListOptions{
			LabelSelector: "openebs.io/cstor-pool=" + cspObj.Name,
		})
	if err != nil {
		return err
	}
	_, err = deploy.NewKubeClient().
		WithNamespace(openebsNamespace).Patch(
		cspObj.Name,
		types.StrategicMergePatchType,
		[]byte(replicaPatch),
	)
	err = retry.
		Times(60).
		Wait(5 * time.Second).
		Try(func(attempt uint) error {
			_, err1 := pod.NewKubeClient().
				WithNamespace(openebsNamespace).
				Get(cspPod.Items[0].Name, metav1.GetOptions{})
			if !k8serrors.IsNotFound(err1) {
				return errors.Errorf("failed to get csp pod because %s", err1)
			}
			return nil
		})
	if err != nil {
		return err
	}
	return nil
}

func updateBDC(bdName apis.CspBlockDevice, cspcObj *apis.CStorPoolCluster, openebsNamespace string) error {
	bdObj, err := bd.NewKubeClient().
		WithNamespace(openebsNamespace).
		Get(bdName.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	bdcObj, err := bdc.NewKubeClient().
		WithNamespace(openebsNamespace).
		Get(bdObj.Spec.ClaimRef.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	klog.Infof("Updating bdc %s with cspc %s info.", bdcObj.Name, cspcObj.Name)
	delete(bdcObj.Labels, string(apis.StoragePoolClaimCPK))
	bdcObj.Labels[string(apis.CStorPoolClusterCPK)] = cspcObj.Name
	for i, finalizer := range bdcObj.Finalizers {
		if finalizer == "storagepoolclaim.openebs.io/finalizer" {
			bdcObj.Finalizers[i] = "cstorpoolcluster.openebs.io/finalizer"
		}
	}
	bdcObj.OwnerReferences[0].Kind = "CStorPoolCluster"
	bdcObj.OwnerReferences[0].UID = cspcObj.UID
	bdcObj, err = bdc.NewKubeClient().
		WithNamespace(openebsNamespace).
		Update(bdcObj)
	if err != nil {
		return err
	}
	return nil
}

func updateCVRsLabels(cspName, openebsNamespace string, cspiObj *apis.CStorPoolInstance) error {
	cvrList, err := cvr.NewKubeclient().
		WithNamespace(openebsNamespace).List(metav1.ListOptions{
		LabelSelector: "cstorpool.openebs.io/name=" + cspName,
	})
	if err != nil {
		return err
	}
	for _, cvrObj := range cvrList.Items {
		klog.Infof("Updating cvr %s with cspi %s info.", cvrObj.Name, cspiObj.Name)
		delete(cvrObj.Labels, "cstorpool.openebs.io/name")
		delete(cvrObj.Labels, "cstorpool.openebs.io/uid")
		cvrObj.Labels["cstorpoolinstance.openebs.io/name"] = cspiObj.Name
		cvrObj.Labels["cstorpoolinstance.openebs.io/uid"] = string(cspiObj.UID)
		delete(cvrObj.Annotations, "cstorpool.openebs.io/hostname")
		cvrObj.Annotations["cstorpoolinstance.openebs.io/hostname"] = cspiObj.Spec.HostName
		_, err = cvr.NewKubeclient().WithNamespace(openebsNamespace).
			Update(&cvrObj)
		if err != nil {
			return err
		}
	}
	return nil
}
