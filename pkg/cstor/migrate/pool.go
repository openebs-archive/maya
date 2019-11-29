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

const (
	replicaPatch = `{
	"spec": {
		"replicas": 0
	}	
}`
	cspNameLabel           = "cstorpool.openebs.io/name"
	cspUIDLabel            = "cstorpool.openebs.io/uid"
	cspHostnameAnnotation  = "cstorpool.openebs.io/hostname"
	cspiNameLabel          = "cstorpoolinstance.openebs.io/name"
	cspiUIDLabel           = "cstorpoolinstance.openebs.io/uid"
	cspiHostnameAnnotation = "cstorpoolinstance.openebs.io/hostname"
	spcFinalizer           = "storagepoolclaim.openebs.io/finalizer"
	cspcFinalizer          = "cstorpoolcluster.openebs.io/finalizer"
)

// Pool migrates the pool from SPC schema to CSPC schema
func Pool(spcName, openebsNamespace string) error {

	spcObj, err := spc.NewKubeClient().
		Get(spcName, metav1.GetOptions{})
	// verify if the spc is already migrated.
	if k8serrors.IsNotFound(err) {
		klog.Infof("spc %s not found.", spcName)
		_, err = cspc.NewKubeClient().
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
	// List all cspi created with reconcile off
	cspiList, err := cspi.NewKubeClient().
		WithNamespace(openebsNamespace).
		List(metav1.ListOptions{
			LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + cspcObj.Name,
		})
	if err != nil {
		return err
	}

	// For each cspi perform the migration from csp that present was on
	// node on which cspi came up.
	for _, cspiObj := range cspiList.Items {
		// Skip the migration if cspi is already in ONLINE state,
		// which implies the migration is done and makes it idempotent
		if cspiObj.Status.Phase != "ONLINE" {
			err = csptocspi(&cspiObj, cspcObj, openebsNamespace)
			if err != nil {
				return err
			}
			cspcObj, err = cspc.NewKubeClient().
				WithNamespace(openebsNamespace).Get(cspcObj.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			// remove the OldCSPUID name to avoid renaming in case
			// cspi is deleted and comes up after reconciliation.
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
	// Clean up old SPC resources after the migration is complete
	err = spc.NewKubeClient().
		Delete(spcName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// csptocspi migrates a CSP to CSPI based on hostname
func csptocspi(cspiObj *apis.CStorPoolInstance, cspcObj *apis.CStorPoolCluster, openebsNamespace string) error {
	hostnameLabel := string(apis.HostNameCPK) + "=" + cspiObj.Labels[string(apis.HostNameCPK)]
	spcLabel := string(apis.StoragePoolClaimCPK) + "=" + cspcObj.Name
	cspLabel := hostnameLabel + "," + spcLabel
	var err1 error
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
	// once the old pool pod is scaled down and bdcs are patched
	// bring up the cspi pod so that the old pool can be renamed and imported.
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
			cspiObj, err1 = cspi.NewKubeClient().
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

// get csp for cspi on the basis of cspLabel, which is the combination of
// hostname label on which cspi came up and the spc label.
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
	return &cspObj, nil
}

// The old pool pod should be scaled down before the new cspi pod comes up
// to avoid importing the pool at two places at the same time.
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
	if len(cspPod.Items) > 0 {
		_, err = deploy.NewKubeClient().WithNamespace(openebsNamespace).
			Patch(
				cspObj.Name,
				types.StrategicMergePatchType,
				[]byte(replicaPatch),
			)
		if err != nil {
			return err
		}
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
	}
	return nil
}

// Update the bdc with the cspc labels instead of spc labels to allow
// filtering of bds claimed by the migrated cspc.
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
		if finalizer == spcFinalizer {
			bdcObj.Finalizers[i] = cspcFinalizer
		}
	}
	bdcObj.OwnerReferences[0].Kind = "CStorPoolCluster"
	bdcObj.OwnerReferences[0].UID = cspcObj.UID
	_, err = bdc.NewKubeClient().
		WithNamespace(openebsNamespace).
		Update(bdcObj)
	if err != nil {
		return err
	}
	return nil
}

// Update the cvrs on the old csp with the migrated cspi labels and annotations
// to allow backward compatibility with old external provisioned volumes.
func updateCVRsLabels(cspName, openebsNamespace string, cspiObj *apis.CStorPoolInstance) error {
	cvrList, err := cvr.NewKubeclient().
		WithNamespace(openebsNamespace).List(metav1.ListOptions{
		LabelSelector: cspNameLabel + "=" + cspName,
	})
	if err != nil {
		return err
	}
	for _, cvrObj := range cvrList.Items {
		klog.Infof("Updating cvr %s with cspi %s info.", cvrObj.Name, cspiObj.Name)
		delete(cvrObj.Labels, cspNameLabel)
		delete(cvrObj.Labels, cspUIDLabel)
		cvrObj.Labels[cspiNameLabel] = cspiObj.Name
		cvrObj.Labels[cspiUIDLabel] = string(cspiObj.UID)
		delete(cvrObj.Annotations, cspHostnameAnnotation)
		cvrObj.Annotations[cspiHostnameAnnotation] = cspiObj.Spec.HostName
		_, err = cvr.NewKubeclient().WithNamespace(openebsNamespace).
			Update(&cvrObj)
		if err != nil {
			return err
		}
	}
	return nil
}
