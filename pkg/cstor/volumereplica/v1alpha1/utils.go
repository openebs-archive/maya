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
	errors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getCVRList returns list of volume replicas related to provided volume
func getCVRList(cvObj *apis.CStorVolume) (*apis.CStorVolumeReplicaList, error) {
	pvName := cvObj.Labels[string(apis.PersistentVolumeCPK)]
	pvLabel := string(apis.PersistentVolumeCPK) + "=" + pvName
	return NewKubeclient(WithNamespace(cvObj.Namespace)).
		List(metav1.ListOptions{
			LabelSelector: pvLabel,
		})
}

// getPoolNames returns list of pool names from cStor volume replcia list
func getPoolNames(cvrList *apis.CStorVolumeReplicaList) []string {
	poolNames := []string{}
	for _, cvrObj := range cvrList.Items {
		poolNames = append(poolNames, cvrObj.Labels[string(apis.CstorpoolInstanceLabel)])
	}
	return poolNames
}

// GetReplicaPoolNames return list of replicas pool names by taking cStor
// volume claim as a argument and return error(if any error occured)
func GetReplicaPoolNames(cvObj *apis.CStorVolume) ([]string, error) {
	pvName := cvObj.Labels[string(apis.PersistentVolumeCPK)]
	cvrList, err := getCVRList(cvObj)
	if err != nil {
		return []string{}, errors.Wrapf(err,
			"failed to list cStorVolumeReplicas related to volume %s",
			pvName)
	}
	return getPoolNames(cvrList), nil
}
