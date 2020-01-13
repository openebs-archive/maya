/*
Copyright 2020 The OpenEBS Authors

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

package app

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	apiscspc "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	cspi "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	util "github.com/openebs/maya/pkg/util"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func cleanupCSPIResources(cspcObj *apis.CStorPoolCluster) error {
	cspiList, err := cspi.NewKubeClient().WithNamespace(cspcObj.Namespace).List(
		metav1.ListOptions{
			LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + cspcObj.Name,
		},
	)
	if err != nil {
		klog.Errorf("failed to list cspi for cspc %s to perform cleanup: %s", cspcObj.Name, err.Error())
		return nil
	}
	opts := []cspiCleanupOptions{cleanupBDC}
	for _, cspiItem := range cspiList.Items {
		cspiItem := cspiItem // pin it
		cspiObj := &cspiItem
		if cspiObj.DeletionTimestamp != nil && canPerformCSPICleanup(cspiObj) {
			for _, o := range opts {
				err = o(cspiObj)
				if err != nil {
					return errors.Wrapf(err, "failed to cleanup cspi %s for cspc %s", cspiObj.Name, cspcObj.Name)
				}
			}
			cspiObj.Finalizers = util.RemoveString(cspiObj.Finalizers, apiscspc.CSPCFinalizer)
			_, err = cspi.NewKubeClient().WithNamespace(cspiObj.Namespace).Update(cspiObj)
			if err != nil {
				return errors.Wrapf(err, "failed to remove finalizer from cspi %s", cspiObj.Name)
			}
		}
	}
	return nil
}

func canPerformCSPICleanup(cspiObj *apis.CStorPoolInstance) bool {
	predicates := []cspiCleanupPredicates{
		hasCSPCFinalizer,
		hasNoPoolProtectionFinalizer,
	}
	for _, p := range predicates {
		if !p(cspiObj) {
			return false
		}
	}
	return true
}

type cspiCleanupPredicates func(*apis.CStorPoolInstance) bool

func hasCSPCFinalizer(cspiObj *apis.CStorPoolInstance) bool {
	return util.ContainsString(cspiObj.Finalizers, apiscspc.CSPCFinalizer)
}

func hasNoPoolProtectionFinalizer(cspiObj *apis.CStorPoolInstance) bool {
	return !util.ContainsString(cspiObj.Finalizers, apiscspc.PoolProtectionFinalizer)
}

type cspiCleanupOptions func(*apis.CStorPoolInstance) error

func cleanupBDC(cspiObj *apis.CStorPoolInstance) error {
	bdcList, err := bdc.NewKubeClient().WithNamespace(cspiObj.Namespace).List(
		metav1.ListOptions{
			LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + cspiObj.Labels[string(apis.CStorPoolClusterCPK)],
		},
	)
	if err != nil {
		return err
	}
	cspiBDMap := map[string]bool{}
	for _, raidGroup := range cspiObj.Spec.RaidGroups {
		for _, bdcObj := range raidGroup.BlockDevices {
			cspiBDMap[bdcObj.BlockDeviceName] = true
		}
	}
	for _, bdcItem := range bdcList.Items {
		bdcItem := bdcItem // pin it
		if cspiBDMap[bdcItem.Spec.BlockDeviceName] {
			bdcObj := &bdcItem
			bdcObj.Finalizers = util.RemoveString(bdcObj.Finalizers, apiscspc.CSPCFinalizer)
			bdcObj, err = bdc.NewKubeClient().WithNamespace(cspiObj.Namespace).Update(bdcObj)
			if err != nil {
				return errors.Wrapf(err, "failed to remove finalizers from bdc %s", bdcObj.Name)
			}
			err = bdc.NewKubeClient().WithNamespace(cspiObj.Namespace).Delete(bdcObj.Name, &metav1.DeleteOptions{})
			if err != nil {
				return errors.Wrapf(err, "failed to delete bdc %s", bdcObj.Name)
			}
		}
	}
	return err
}
