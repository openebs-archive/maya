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
	"fmt"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func spcUpgrade(spcName, openebsNamespace string) error {

	spcLabel := "openebs.io/storage-pool-claim=" + spcName
	cspList, err := cspClient.List(metav1.ListOptions{
		LabelSelector: spcLabel,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to list csp for spc %s", spcName)
	}
	if len(cspList.Items) == 0 {
		return errors.Errorf("no csp found for spc %s: no csp found", spcName)
	}
	err = verifyCSPNodeName(cspList)
	if err != nil {
		return err
	}
	for _, cspObj := range cspList.Items {
		if cspObj.Name == "" {
			return errors.Errorf("missing csp name")
		}
		err = cspUpgrade(cspObj.Name, openebsNamespace)
		if err != nil {
			return err
		}
	}
	fmt.Println("Upgrade Successful for spc", spcName)
	return nil
}
