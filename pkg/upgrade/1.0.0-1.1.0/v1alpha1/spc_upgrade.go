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

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func spcUpgrade(spcName, openebsNamespace string) error {

	spcLabel := "openebs.io/storage-pool-claim=" + spcName
	cspList, err := cspClient.List(metav1.ListOptions{
		LabelSelector: spcLabel,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to list csp for spc %s", spcName)
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
