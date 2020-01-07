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
)

func cleanupCSPIResources(cspcObj *apis.CStorPoolCluster) error {
	cspiList, err := cspi.NewKubeClient().WithNamespace(cspcObj.Namespace).List(
		metav1.ListOptions{
			LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + cspcObj.Name,
		},
	)
	if err != nil {
		return err
	}
	opts := []cspiCleanupOptions{cleanupBDC}
	for _, cspiObj := range cspiList.Items {
		cspiObj := cspiObj // pin it
		if cspiObj.DeletionTimestamp != nil && len(cspiObj.Finalizers) == 1 {
			for _, o := range opts {
				err = o(cspiObj)
				if err != nil {
					return errors.Wrap(err, "failed to cleanup cspi resources")
				}
			}
			cspiItem := &cspiObj
			cspiItem.Finalizers = []string{}
			_, err = cspi.NewKubeClient().WithNamespace(cspiObj.Namespace).Update(cspiItem)
			if err != nil {
				return errors.Wrap(err, "failed to remove finalizer from cspi")
			}
		}
	}
	return nil
}

type cspiCleanupOptions func(apis.CStorPoolInstance) error

func cleanupBDC(cspiObj apis.CStorPoolInstance) error {
	bdcList, err := bdc.NewKubeClient().WithNamespace(cspiObj.Namespace).List(
		metav1.ListOptions{},
	)
	if err != nil {
		return err
	}
	for _, bdcItem := range bdcList.Items {
		bdcItem := bdcItem // pin it
		if isBDCForCSPI(bdcItem.Spec.BlockDeviceName, cspiObj) {
			bdcObj := &bdcItem
			bdcObj.Finalizers = util.RemoveString(bdcObj.Finalizers, apiscspc.CSPCFinalizer)
			bdcObj, err = bdc.NewKubeClient().WithNamespace(cspiObj.Namespace).Update(bdcObj)
			if err != nil {
				return errors.Wrap(err, "failed to remove finalizers from bdc")
			}
			err = bdc.NewKubeClient().WithNamespace(cspiObj.Namespace).Delete(bdcObj.Name, &metav1.DeleteOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to delete bdc")
			}
		}
	}
	return err
}

func isBDCForCSPI(bdName string, cspiObj apis.CStorPoolInstance) bool {
	for _, raidGroup := range cspiObj.Spec.RaidGroups {
		for _, bdcObj := range raidGroup.BlockDevices {
			if bdcObj.BlockDeviceName == bdName {
				return true
			}
		}
	}
	return false
}
