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

package cstorpoolit

import (
	"github.com/openebs/CITF"
	apis "github.com/openebs/CITF/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ToDo: The 'citf' variable which is an instance of citf should have a behavioural binding and not passed as arg
// TODo: Add comment for each function
func getCstorPoolCount(spcName string, citf citf.CITF) (int, error) {
	cspObject, err := citf.K8S.ListCStorPool(v1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spcName})
	if err != nil {
		return 0, err
	}
	return len(cspObject.Items), nil
}

func getPoolDeployCount(spcName string, citf citf.CITF) (int, error) {
	deployObject, err := citf.K8S.ListDeployments("openebs", v1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spcName})
	if err != nil {
		return 0, err
	}
	return len(deployObject.Items), nil
}

func getStoragePoolCount(spcName string, citf citf.CITF) (int, error) {
	spObject, err := citf.K8S.ListStoragePool(v1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spcName})
	if err != nil {
		return 0, err
	}
	return len(spObject.Items), nil
}

func getCstorPoolStatus(spcName string, citf citf.CITF) (int, error) {
	var onlineCspCount int
	onlineCspCount = 0
	cspObject, err := citf.K8S.ListCStorPool(v1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spcName})
	if err != nil {
		return 0, err
	}
	for _, obj := range cspObject.Items {
		if obj.Status.Phase == "Online" {
			onlineCspCount++
		}
	}
	return onlineCspCount, nil
}
