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
	"strings"
	"text/template"
	"time"

	"k8s.io/klog"

	utask "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	templates "github.com/openebs/maya/pkg/upgrade/templates/v1"
	errors "github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

type cspDeployPatchDetails struct {
	CurrentVersion, UpgradeVersion, ImageTag, PoolImage,
	BaseDir, PoolMgmtImage, MExporterImage, SPCName string
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
	patchDetails.SPCName = d.Labels[string(apis.StoragePoolClaimCPK)]
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
	err = deploy.NewForAPIObject(cspDeployObj).VerifyReplicaStatus()
	if err != nil {
		return nil, err
	}
	return cspDeployObj, nil
}

func patchCSP(cspObj *apis.CStorPool) error {
	cspVersion := cspObj.Labels["openebs.io/version"]
	if cspVersion == currentVersion {
		tmpl, err := template.New("cspPatch").
			Parse(templates.VersionDetailsPatch)
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
		klog.Infof("patched csp %s", cspObj.Name)
	} else {
		klog.Infof("csp %s already in %s version", cspObj.Name, upgradeVersion)
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
		patchDetails.CurrentVersion = currentVersion
		patchDetails.BaseDir = baseDir
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
		fmt.Println(cspDeployPatch)
		err = patchDelpoyment(
			cspDeployObj.Name,
			openebsNamespace,
			types.StrategicMergePatchType,
			[]byte(cspDeployPatch),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to patch deployment %s", cspDeployObj.Name)
		}
		klog.Infof("patched csp deployment %s", cspDeployObj.Name)
	} else {
		klog.Infof("csp deployment %s already in %s version",
			cspDeployObj.Name,
			upgradeVersion,
		)
	}
	return nil
}

type cstorCSPOptions struct {
	utaskObj     *utask.UpgradeTask
	cspObj       *apis.CStorPool
	cspDeployObj *appsv1.Deployment
}

func (c *cstorCSPOptions) preUpgrade(cspName, openebsNamespace string) error {
	var err, uerr error

	c.utaskObj, uerr = getOrCreateUpgradeTask("cstorPool", cspName, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}

	statusObj := utask.UpgradeDetailedStatuses{Step: utask.PreUpgrade}

	statusObj.Phase = utask.StepWaiting
	c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}

	statusObj.Phase = utask.StepErrored
	c.cspObj, err = getCSPObject(cspName)
	if err != nil {
		statusObj.Message = "failed to verify cstor pool"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	c.cspDeployObj, err = getCSPDeployment(cspName, openebsNamespace)
	if err != nil {
		statusObj.Message = "failed to verify cstor pool deployment"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	statusObj.Phase = utask.StepCompleted
	statusObj.Message = "Pre-upgrade steps were successful"
	statusObj.Reason = ""
	c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}
	return nil
}

func (c *cstorCSPOptions) waitForCSPCurrentVersion() error {
	// waiting for old objects to get populated with new fields
	for c.cspObj.VersionDetails.Status.Current == "" {
		klog.Infof("Waiting for csp current version to get populated.")
		// Sleep equal to the default sync time
		time.Sleep(10 * time.Second)
		obj, err := cspClient.Get(c.cspObj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		c.cspObj = obj
	}
	return nil
}

func (c *cstorCSPOptions) verifyCSPVersionReconcile(openebsNamespace string) error {
	var uerr error
	statusObj := utask.UpgradeDetailedStatuses{Step: utask.PoolInstanceUpgrade}
	statusObj.Phase = utask.StepErrored
	// waiting for the current version to be equal to desired version
	for c.cspObj.VersionDetails.Status.Current != upgradeVersion {
		klog.Infof("Verifying the reconciliation of version for %s { phase:%s }",
			c.cspObj.Name, string(c.cspObj.Status.Phase))
		// Sleep equal to the default sync time
		time.Sleep(10 * time.Second)
		obj, err := cspClient.Get(c.cspObj.Name, metav1.GetOptions{})
		if err != nil {
			statusObj.Message = "failed to get cstor pool"
			statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
			c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
			if uerr != nil && isENVPresent {
				return uerr
			}
			return err
		}
		if obj.VersionDetails.Status.Message != "" {
			statusObj.Message = obj.VersionDetails.Status.Message
			statusObj.Reason = obj.VersionDetails.Status.Reason
			c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
			if uerr != nil && isENVPresent {
				return uerr
			}
			klog.Errorf("failed to reconcile version : %s", obj.VersionDetails.Status.Reason)
		}
		c.cspObj = obj
	}
	return nil
}

func (c *cstorCSPOptions) poolInstanceUpgrade(openebsNamespace string) error {
	var err, uerr error
	statusObj := utask.UpgradeDetailedStatuses{Step: utask.PoolInstanceUpgrade}
	statusObj.Phase = utask.StepWaiting
	c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}

	statusObj.Phase = utask.StepErrored
	err = patchCSPDeploy(c.cspDeployObj, openebsNamespace)
	if err != nil {
		statusObj.Message = "failed to patch cstor pool deployment"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	err = c.waitForCSPCurrentVersion()
	if err != nil {
		statusObj.Message = "failed to verify versiondetails for cstor pool"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	err = patchCSP(c.cspObj)
	if err != nil {
		statusObj.Message = "failed to patch cstor pool"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	err = c.verifyCSPVersionReconcile(openebsNamespace)
	if err != nil {
		return err
	}

	statusObj.Phase = utask.StepCompleted
	statusObj.Message = "Pool instance upgrade was successful"
	statusObj.Reason = ""
	c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}
	return nil
}

func cspUpgrade(cspName, openebsNamespace string, utaskObj *utask.UpgradeTask) (*utask.UpgradeTask, error) {
	var err error

	options := &cstorCSPOptions{}

	options.utaskObj = utaskObj

	err = options.preUpgrade(cspName, openebsNamespace)
	if err != nil {
		return options.utaskObj, err
	}

	err = options.poolInstanceUpgrade(openebsNamespace)
	if err != nil {
		return options.utaskObj, err
	}

	klog.Infof("Upgrade Successful for csp %s", cspName)

	return options.utaskObj, nil
}
