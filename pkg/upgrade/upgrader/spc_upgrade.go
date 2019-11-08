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
	"time"

	utask "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// to verify that no two csp are on same node
func verifyCSPNodeName(cspList *apis.CStorPoolList) error {
	nodeMap := map[string]bool{}
	for _, cspObj := range cspList.Items {
		nodeName := cspObj.Labels[string(apis.HostNameCPK)]
		if nodeMap[nodeName] {
			return errors.Errorf("more than one csp on %s node."+
				" please make sure all csp are on different nodes", nodeName)
		}
		nodeMap[nodeName] = true
	}
	return nil
}

func spcUpgrade(spcName, openebsNamespace string) (*utask.UpgradeTask, error) {

	spcLabel := "openebs.io/storage-pool-claim=" + spcName
	cspList, err := cspClient.List(metav1.ListOptions{
		LabelSelector: spcLabel,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list csp for spc %s", spcName)
	}
	if len(cspList.Items) == 0 {
		return nil, errors.Errorf("no csp found for spc %s: no csp found", spcName)
	}
	err = waitForSPCCurrentVersion(spcName)
	if err != nil {
		return nil, err
	}
	err = verifyCSPNodeName(cspList)
	if err != nil {
		return nil, err
	}
	for _, cspObj := range cspList.Items {
		if cspObj.Name == "" {
			return nil, errors.Errorf("missing csp name")
		}
		utaskObj, uerr := getOrCreateUpgradeTask("cstorPool", cspObj.Name, openebsNamespace)
		if uerr != nil && isENVPresent {
			return nil, uerr
		}

		statusObj := utask.UpgradeDetailedStatuses{Step: utask.PreUpgrade}

		statusObj.Phase = utask.StepWaiting
		utaskObj, uerr = updateUpgradeDetailedStatus(utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return nil, uerr
		}
		utaskObj, err = cspUpgrade(cspObj.Name, openebsNamespace, utaskObj)
		if err != nil {
			return utaskObj, err
		}
		if utaskObj != nil {
			utaskObj.Status.Phase = utask.UpgradeSuccess
			utaskObj.Status.CompletedTime = metav1.Now()
			_, uerr := utaskClient.WithNamespace(openebsNamespace).
				Update(utaskObj)
			if uerr != nil && isENVPresent {
				return nil, uerr
			}
		}
	}
	err = updateSPCVersion(spcName)
	if err != nil {
		return nil, err
	}
	err = verifySPCVersionReconcile(spcName)
	if err != nil {
		return nil, err
	}
	klog.Infof("Upgrade Successful for spc %s", spcName)
	return nil, nil
}

func updateSPCVersion(name string) error {
	client := spc.NewKubeClient()
	spcObj, err := client.Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	spcObj.VersionDetails.Desired = upgradeVersion
	spcObj.VersionDetails.Status.State = apis.ReconcilePending
	_, err = client.Update(spcObj)
	if err != nil {
		return err
	}
	return nil
}

func waitForSPCCurrentVersion(name string) error {
	client := spc.NewKubeClient()
	spcObj, err := client.Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// waiting for old objects to get populated with new fields
	for spcObj.VersionDetails.Status.Current == "" {
		klog.Infof("Waiting for spc current version to get populated.")
		// Sleep equal to the default sync time
		time.Sleep(10 * time.Second)
		spcObj, err = client.Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func verifySPCVersionReconcile(name string) error {
	client := spc.NewKubeClient()
	spcObj, err := client.Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// waiting for the current version to be equal to desired version
	for spcObj.VersionDetails.Status.Current != upgradeVersion {
		klog.Infof("Verifying the reconciliation of version for %s", spcObj.Name)
		// Sleep equal to the default sync time
		time.Sleep(10 * time.Second)
		spcObj, err = client.Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if spcObj.VersionDetails.Status.Message != "" {
			klog.Errorf("failed to reconcile: %s", spcObj.VersionDetails.Status.Reason)
		}
	}
	return nil
}
