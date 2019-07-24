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
	"text/template"

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

type cspDeployPatchDetails struct {
	UpgradeVersion, PoolImage, PoolMgmtImage, MExporterImage string
}

func getCSPDeployPatchDetails(
	d *appsv1.Deployment,
) (*cspDeployPatchDetails, error) {
	patchDetails := &cspDeployPatchDetails{}
	cstorPoolImage, err := getBaseImage(d, "cstor-pool")
	if err != nil {
		return nil, err
	}
	cstorPoolMgmtImage, err := getBaseImage(d, "cstor-pool-mgmt")
	if err != nil {
		return nil, err
	}
	MExporterImage, err := getBaseImage(d, "maya-exporter")
	if err != nil {
		return nil, err
	}
	patchDetails.PoolImage = cstorPoolImage
	patchDetails.PoolMgmtImage = cstorPoolMgmtImage
	patchDetails.MExporterImage = MExporterImage
	return patchDetails, nil
}

func cspUpgrade(cspName, openebsNamespace string) error {
	if cspName == "" {
		return errors.Errorf("missing csp name")
	}
	if openebsNamespace == "" {
		return errors.Errorf("missing openebs namespace")
	}

	cspObj, err := cspClient.Get(cspName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	cspVersion := cspObj.Labels["openebs.io/version"]
	if (cspVersion != currentVersion) && (cspVersion != upgradeVersion) {
		return errors.Errorf(
			"cstor pool version %s is neither %s nor %s\n",
			cspVersion,
			currentVersion,
			upgradeVersion,
		)
	}
	cspLabel := "openebs.io/cstor-pool=" + cspName
	cspDeployObj, err := getDeployment(cspLabel, openebsNamespace)
	if err != nil {
		return err
	}
	if cspDeployObj.Name == "" {
		return errors.Errorf("missing deployment name for csp %s", cspName)
	}
	cspDeployVersion, err := getOpenEBSVersion(cspDeployObj)
	if err != nil {
		return err
	}
	if (cspDeployVersion != currentVersion) && (cspDeployVersion != upgradeVersion) {
		return errors.Errorf(
			"cstor pool version %s is neither %s nor %s\n",
			cspVersion,
			currentVersion,
			upgradeVersion,
		)
	}
	if cspVersion == currentVersion {
		tmpl, err := template.New("cspPatch").Parse(openebsVersionPatchTemplate)
		if err != nil {
			return err
		}
		err = tmpl.Execute(&buffer, upgradeVersion)
		if err != nil {
			return err
		}
		cspPatch := buffer.String()
		buffer.Reset()
		_, err = cspClient.Patch(
			cspName,
			types.MergePatchType,
			[]byte(cspPatch),
		)
		if err != nil {
			return err
		}
		fmt.Printf("patched csp %s\n", cspName)
	} else {
		fmt.Printf("csp %s already in %s version\n", cspName, upgradeVersion)
	}

	if cspDeployVersion == currentVersion {
		patchDetails, err := getCSPDeployPatchDetails(cspDeployObj)
		if err != nil {
			return err
		}
		patchDetails.UpgradeVersion = upgradeVersion
		tmpl, err := template.New("cspDeployPatch").Parse(cspDeployPatchTemplate)
		if err != nil {
			return err
		}
		err = tmpl.Execute(&buffer, patchDetails)
		if err != nil {
			return err
		}
		cspDeployPatch := buffer.String()
		buffer.Reset()
		err = patchDelpoyment(
			cspDeployObj.Name,
			openebsNamespace,
			types.StrategicMergePatchType,
			[]byte(cspDeployPatch),
		)
		if err != nil {
			return err
		}
		fmt.Printf("patched csp deployment %s\n", cspName)
	} else {
		fmt.Printf("csp deployment %s already in %s version\n",
			cspDeployObj.Name,
			upgradeVersion,
		)
	}
	fmt.Println("Upgrade Successful for csp", cspName)
	return nil
}
