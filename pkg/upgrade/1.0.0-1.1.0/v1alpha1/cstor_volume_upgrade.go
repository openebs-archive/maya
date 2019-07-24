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

type cstorTargetPatchDetails struct {
	UpgradeVersion, IstgtImage, MExporterImage, VolumeMgmtImage string
}

func verifyCSPVersion(pvLabel, namespace string) error {
	cvrList, err := cvrClient.List(
		metav1.ListOptions{
			LabelSelector: pvLabel,
		},
	)
	if err != nil {
		return err
	}
	for _, cvrObj := range cvrList.Items {
		cspName := cvrObj.Labels["cstorpool.openebs.io/name"]
		if cspName == "" {
			return errors.Errorf("missing csp name for %s", cvrObj.Name)
		}
		cspDeployObj, err := deployClient.WithNamespace(namespace).
			Get(cspName)
		if err != nil {
			return err
		}
		if cspDeployObj.Labels["openebs.io/version"] != upgradeVersion {
			return errors.Errorf(
				"csp deployment %s not in %s version",
				cspDeployObj.Name,
				upgradeVersion,
			)
		}
	}
	return nil
}

func getTargetDeployPatchDetails(
	d *appsv1.Deployment,
) (*cstorTargetPatchDetails, error) {
	patchDetails := &cstorTargetPatchDetails{}
	if d.Name == "" {
		return nil, errors.Errorf("missing deployment name")
	}
	istgtImage, err := getBaseImage(d, "cstor-istgt")
	if err != nil {
		return nil, err
	}
	patchDetails.IstgtImage = istgtImage
	mexporterImage, err := getBaseImage(d, "maya-volume-exporter")
	if err != nil {
		return nil, err
	}
	patchDetails.MExporterImage = mexporterImage
	volumeMgmtImage, err := getBaseImage(d, "cstor-volume-mgmt")
	if err != nil {
		return nil, err
	}
	patchDetails.VolumeMgmtImage = volumeMgmtImage
	return patchDetails, nil
}

func patchTargetDeploy(d *appsv1.Deployment, ns string) error {
	version, err := getOpenEBSVersion(d)
	if err != nil {
		return err
	}
	if (version != currentVersion) && (version != upgradeVersion) {
		return errors.Errorf(
			"target deployment version %s is neither %s nor %s\n",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	if version == currentVersion {
		tmpl, err := template.New("targetPatch").Parse(cstorTargetPatchTemplate)
		if err != nil {
			return err
		}
		patchDetails, err := getTargetDeployPatchDetails(d)
		if err != nil {
			return err
		}
		patchDetails.UpgradeVersion = upgradeVersion
		err = tmpl.Execute(&buffer, patchDetails)
		if err != nil {
			return err
		}
		replicaPatch := buffer.String()
		buffer.Reset()
		err = patchDelpoyment(
			d.Name,
			ns,
			types.StrategicMergePatchType,
			[]byte(replicaPatch),
		)
		if err != nil {
			return err
		}
		fmt.Printf("target deployment %s patched\n", d.Name)
	} else {
		fmt.Printf("target deployment already in %s version\n", upgradeVersion)
	}
	return nil
}

func patchCV(pvLabel, namespace string) error {
	cvObject, err := cvClient.WithNamespace(namespace).List(
		metav1.ListOptions{
			LabelSelector: pvLabel,
		},
	)
	if err != nil {
		return err
	}
	if len(cvObject.Items) == 0 {
		return errors.Errorf("cstorvolume not found")
	}
	version := cvObject.Items[0].Labels["openebs.io/version"]
	if (version != currentVersion) && (version != upgradeVersion) {
		return errors.Errorf(
			"cstorvolume version %s is neither %s nor %s\n",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	if version == currentVersion {
		tmpl, err := template.New("cvPatch").Parse(openebsVersionPatchTemplate)
		if err != nil {
			return err
		}
		err = tmpl.Execute(&buffer, upgradeVersion)
		if err != nil {
			return err
		}
		cvPatch := buffer.String()
		buffer.Reset()
		_, err = cvClient.WithNamespace(namespace).Patch(
			cvObject.Items[0].Name,
			namespace,
			types.MergePatchType,
			[]byte(cvPatch),
		)
		if err != nil {
			return err
		}
		fmt.Printf("cstorvolume %s patched\n", cvObject.Items[0].Name)
	} else {
		fmt.Printf("cstorvolume already in %s version\n", upgradeVersion)
	}
	return nil
}

func patchCVR(cvrName, namespace string) error {
	cvrObject, err := cvrClient.WithNamespace(namespace).Get(cvrName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	version := cvrObject.Labels["openebs.io/version"]
	if (version != currentVersion) && (version != upgradeVersion) {
		return errors.Errorf(
			"cstorvolume version %s is neither %s nor %s\n",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	if version == currentVersion {
		tmpl, err := template.New("cvPatch").Parse(openebsVersionPatchTemplate)
		if err != nil {
			return err
		}
		err = tmpl.Execute(&buffer, upgradeVersion)
		if err != nil {
			return err
		}
		cvPatch := buffer.String()
		buffer.Reset()
		_, err = cvrClient.WithNamespace(namespace).Patch(
			cvrObject.Name,
			namespace,
			types.MergePatchType,
			[]byte(cvPatch),
		)
		if err != nil {
			return err
		}
		fmt.Printf("cstorvolumereplica %s patched\n", cvrObject.Name)
	} else {
		fmt.Printf("cstorvolume replica already in %s version\n", upgradeVersion)
	}
	return nil
}

func patchService(targetServiceLabel, namespace string) error {
	targetServiceObj, err := serviceClient.WithNamespace(namespace).List(
		metav1.ListOptions{
			LabelSelector: targetServiceLabel,
		},
	)
	if err != nil {
		return err
	}
	targetServiceName := targetServiceObj.Items[0].Name
	if targetServiceName == "" {
		return errors.Errorf("missing service name")
	}
	version := targetServiceObj.Items[0].
		Labels["openebs.io/version"]
	if version != currentVersion && version != upgradeVersion {
		return errors.Errorf(
			"service version %s is neither %s nor %s\n",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	if version == currentVersion {
		tmpl, err := template.New("servicePatch").Parse(openebsVersionPatchTemplate)
		if err != nil {
			return err
		}
		err = tmpl.Execute(&buffer, upgradeVersion)
		if err != nil {
			return err
		}
		servicePatch := buffer.String()
		buffer.Reset()
		_, err = serviceClient.WithNamespace(namespace).Patch(
			targetServiceName,
			types.StrategicMergePatchType,
			[]byte(servicePatch),
		)
		if err != nil {
			return err
		}
		fmt.Printf("targetservice %s patched\n", targetServiceName)
	} else {
		fmt.Printf("service already in %s version\n", upgradeVersion)
	}
	return nil
}

func cstorVolumeUpgrade(pvName, openebsNamespace string) error {
	pvLabel := "openebs.io/persistent-volume=" + pvName
	targetLabel := pvLabel + ",openebs.io/target=cstor-target"
	targetServiceLabel := pvLabel + ",openebs.io/target-service=cstor-target-svc"

	err := verifyCSPVersion(pvLabel, openebsNamespace)
	if err != nil {
		return err
	}

	ns, err := getPVCDeploymentsNamespace(pvName, pvLabel, openebsNamespace)
	if err != nil {
		return err
	}

	targetDeployObj, err := getDeployment(targetLabel, ns)
	if err != nil {
		return err
	}

	err = patchTargetDeploy(targetDeployObj, ns)
	if err != nil {
		return err
	}

	err = patchService(targetServiceLabel, ns)
	if err != nil {
		return err
	}

	err = patchCV(pvLabel, ns)
	if err != nil {
		return err
	}

	cvrList, err := cvrClient.WithNamespace(openebsNamespace).List(
		metav1.ListOptions{
			LabelSelector: pvLabel,
		},
	)
	if err != nil {
		return err
	}
	for _, cvrObj := range cvrList.Items {
		if cvrObj.Name == "" {
			return errors.Errorf("missing cvr name")
		}
		err = patchCVR(cvrObj.Name, openebsNamespace)
		if err != nil {
			return err
		}

	}

	return nil
}
