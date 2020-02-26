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
	"encoding/json"
	"fmt"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// CVCKey returns an unique key of a CVC object,
func CVCKey(cvc *apis.CStorVolumeClaim) string {
	return fmt.Sprintf("%s/%s", cvc.Namespace, cvc.Name)
}

func getPatchData(oldObj, newObj interface{}) ([]byte, error) {
	oldData, err := json.Marshal(oldObj)
	if err != nil {
		return nil, fmt.Errorf("marshal old object failed: %v", err)
	}
	newData, err := json.Marshal(newObj)
	if err != nil {
		return nil, fmt.Errorf("mashal new object failed: %v", err)
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, oldObj)
	if err != nil {
		return nil, fmt.Errorf("CreateTwoWayMergePatch failed: %v", err)
	}
	return patchBytes, nil
}

// GetPDBPoolLabels returns the pool labels from poolNames
func GetPDBPoolLabels(poolNames []string) map[string]string {
	pdbLabels := map[string]string{}
	for _, poolName := range poolNames {
		key := fmt.Sprintf("openebs.io/%s", poolName)
		pdbLabels[key] = "true"
	}
	return pdbLabels
}

// GetPDBLabels returns the labels required for building PDB based on arguments
func GetPDBLabels(poolNames []string, cspcName string) map[string]string {
	pdbLabels := GetPDBPoolLabels(poolNames)
	pdbLabels[string(apis.CStorPoolClusterCPK)] = cspcName
	return pdbLabels
}

// GetDesiredReplicaPoolNames returns list of desired pool names
func GetDesiredReplicaPoolNames(cvc *apis.CStorVolumeClaim) []string {
	poolNames := []string{}
	for _, poolInfo := range cvc.Spec.Policy.ReplicaPoolInfo {
		poolNames = append(poolNames, poolInfo.PoolName)
	}
	return poolNames
}
