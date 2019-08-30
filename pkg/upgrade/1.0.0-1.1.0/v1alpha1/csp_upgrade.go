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
	"strings"
	"text/template"

	"github.com/golang/glog"

	utask "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	templates "github.com/openebs/maya/pkg/upgrade/templates/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

type cspDeployPatchDetails struct {
	UpgradeVersion, ImageTag, PoolImage, PoolMgmtImage, MExporterImage string
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
	if imageTag != "" {
		patchDetails.ImageTag = imageTag
	} else {
		patchDetails.ImageTag = upgradeVersion
	}
	return patchDetails, nil
}

func getCSPObject(cspName string) (*apis.CStorPool, error) {
	if cspName == "" {
		return nil, errors.Errorf("missing csp name")
	}
	cspObj, err := cspClient.Get(cspName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get csp %s", cspName)
	}
	cspVersion := cspObj.Labels["openebs.io/version"]
	if (cspVersion != currentVersion) && (cspVersion != upgradeVersion) {
		return nil, errors.Errorf(
			"cstor pool version %s is neither %s nor %s",
			cspVersion,
			currentVersion,
			upgradeVersion,
		)
	}
	return cspObj, nil
}

func getCSPDeployment(cspName, openebsNamespace string) (*appsv1.Deployment, error) {
	if openebsNamespace == "" {
		return nil, errors.Errorf("missing openebs namespace")
	}
	cspLabel := "openebs.io/cstor-pool=" + cspName
	cspDeployObj, err := getDeployment(cspLabel, openebsNamespace)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get deployment for csp %s", cspName)
	}
	if cspDeployObj.Name == "" {
		return nil, errors.Errorf("missing deployment name for csp %s", cspName)
	}
	cspDeployVersion, err := getOpenEBSVersion(cspDeployObj)
	if err != nil {
		return nil, err
	}
	if (cspDeployVersion != currentVersion) && (cspDeployVersion != upgradeVersion) {
		return nil, errors.Errorf(
			"cstor pool version %s is neither %s nor %s",
			cspDeployVersion,
			currentVersion,
			upgradeVersion,
		)
	}
	return cspDeployObj, nil
}

func patchCSP(cspObj *apis.CStorPool) error {
	cspVersion := cspObj.Labels["openebs.io/version"]
	if cspVersion == currentVersion {
		tmpl, err := template.New("cspPatch").
			Parse(templates.OpenebsVersionPatch)
		if err != nil {
			return errors.Wrapf(err, "failed to create template for csp patch")
		}
		err = tmpl.Execute(&buffer, upgradeVersion)
		if err != nil {
			return errors.Wrapf(err, "failed to populate template for csp patch")
		}
		cspPatch := buffer.String()
		buffer.Reset()
		_, err = cspClient.Patch(
			cspObj.Name,
			types.MergePatchType,
			[]byte(cspPatch),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to patch csp %s", cspObj.Name)
		}
		glog.Infof("patched csp %s", cspObj.Name)
	} else {
		glog.Infof("csp %s already in %s version", cspObj.Name, upgradeVersion)
	}
	return nil
}

func patchCSPDeploy(cspDeployObj *appsv1.Deployment, openebsNamespace string) error {
	cspDeployVersion := cspDeployObj.Labels["openebs.io/version"]
	if cspDeployVersion == currentVersion {
		patchDetails, err := getCSPDeployPatchDetails(cspDeployObj)
		if err != nil {
			return err
		}
		patchDetails.UpgradeVersion = upgradeVersion
		tmpl, err := template.New("cspDeployPatch").
			Parse(templates.CSPDeployPatch)
		if err != nil {
			return errors.Wrapf(err, "failed to create template for csp deployment patch")
		}
		err = tmpl.Execute(&buffer, patchDetails)
		if err != nil {
			return errors.Wrapf(err, "failed to populate template for csp deployment patch")
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
			return errors.Wrapf(err, "failed to patch deployment %s", cspDeployObj.Name)
		}
		glog.Infof("patched csp deployment %s", cspDeployObj.Name)
	} else {
		glog.Infof("csp deployment %s already in %s version",
			cspDeployObj.Name,
			upgradeVersion,
		)
	}
	return nil
}

func cspUpgrade(cspName, openebsNamespace string) (*utask.UpgradeTask, error) {
	var err, uerr error
	var utaskObj *utask.UpgradeTask
	utaskObj, uerr = getOrCreateUpgradeTask("cstorPool", cspName, openebsNamespace)
	if uerr != nil && isENVPresent {
		return nil, uerr
	}

	utaskObj, uerr = updateUpgradeDetailedStatus(
		utaskObj,
		utask.UpgradeDetailedStatuses{
			Step: utask.PreUpgrade,
			Status: utask.Status{
				Phase: utask.StepWaiting,
			},
		},
		openebsNamespace,
	)
	if uerr != nil && isENVPresent {
		return nil, uerr
	}

	cspObj, err := getCSPObject(cspName)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.PreUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to verify cstor pool",
					Reason:  strings.Replace(err.Error(), ":", "", -1),
				},
			},
			openebsNamespace,
		)
		if uerr != nil && isENVPresent {
			return nil, uerr
		}
		return utaskObj, err
	}

	cspDeployObj, err := getCSPDeployment(cspName, openebsNamespace)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.PreUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to verify cstor pool deployment",
					Reason:  strings.Replace(err.Error(), ":", "", -1),
				},
			},
			openebsNamespace,
		)
		if uerr != nil && isENVPresent {
			return nil, uerr
		}
		return utaskObj, err
	}

	utaskObj, uerr = updateUpgradeDetailedStatus(
		utaskObj,
		utask.UpgradeDetailedStatuses{
			Step: utask.PreUpgrade,
			Status: utask.Status{
				Phase:   utask.StepCompleted,
				Message: "Pre-upgrade steps were successful",
			},
		},
		openebsNamespace,
	)
	if uerr != nil && isENVPresent {
		return nil, uerr
	}

	utaskObj, uerr = updateUpgradeDetailedStatus(
		utaskObj,
		utask.UpgradeDetailedStatuses{
			Step: utask.TargetUpgrade,
			Status: utask.Status{
				Phase: utask.StepWaiting,
			},
		},
		openebsNamespace,
	)
	if uerr != nil && isENVPresent {
		return nil, uerr
	}

	err = patchCSPDeploy(cspDeployObj, openebsNamespace)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.TargetUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to patch cstor pool deployment",
					Reason:  strings.Replace(err.Error(), ":", "", -1),
				},
			},
			openebsNamespace,
		)
		if uerr != nil && isENVPresent {
			return nil, uerr
		}
		return utaskObj, err
	}

	err = patchCSP(cspObj)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.TargetUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to patch cstor pool",
					Reason:  strings.Replace(err.Error(), ":", "", -1),
				},
			},
			openebsNamespace,
		)
		if uerr != nil && isENVPresent {
			return nil, uerr
		}
		return utaskObj, err
	}

	utaskObj, uerr = updateUpgradeDetailedStatus(
		utaskObj,
		utask.UpgradeDetailedStatuses{
			Step: utask.TargetUpgrade,
			Status: utask.Status{
				Phase:   utask.StepCompleted,
				Message: "Target upgrade was successful",
			},
		},
		openebsNamespace,
	)
	if uerr != nil && isENVPresent {
		return nil, uerr
	}

	utaskObj.Status.Phase = utask.UpgradeSuccess
	utaskObj.Status.CompletedTime = metav1.Now()
	utaskObj, uerr = utaskClient.WithNamespace(openebsNamespace).
		Update(utaskObj)
	if uerr != nil && isENVPresent {
		return nil, uerr
	}
	glog.Infof("Upgrade Successful for csp %s", cspName)
	return utaskObj, nil
}
