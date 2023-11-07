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

	utask "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	templates "github.com/openebs/maya/pkg/upgrade/templates/v1"
	"github.com/openebs/maya/pkg/util"
	"k8s.io/klog/v2"

	errors "github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

type cstorTargetPatchDetails struct {
	CurrentVersion, UpgradeVersion, ImageTag, IstgtImage,
	BaseDir, MExporterImage, VolumeMgmtImage, PVName string
	IsMonitorEnabled bool
}

const (
	pvLabelKey = "openebs.io/persistent-volume"
)

func verifyCSPVersion(cvrList *apis.CStorVolumeReplicaList, namespace string) error {
	for _, cvrObj := range cvrList.Items {
		cspName := cvrObj.Labels["cstorpool.openebs.io/name"]
		if cspName == "" {
			return errors.Errorf("missing csp name for %s", cvrObj.Name)
		}
		cspDeployObj, err := deployClient.WithNamespace(namespace).
			Get(cspName)
		if err != nil {
			return errors.Wrapf(err, "failed to get deployment for csp %s", cspName)
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
	patchDetails.IsMonitorEnabled = true
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
		if err.Error() == "image not found for container maya-volume-exporter" {
			patchDetails.IsMonitorEnabled = false
		} else {
			return nil, err
		}
	}
	patchDetails.MExporterImage = mexporterImage
	volumeMgmtImage, err := getBaseImage(d, "cstor-volume-mgmt")
	if err != nil {
		return nil, err
	}
	patchDetails.VolumeMgmtImage = volumeMgmtImage
	if imageTag != "" {
		patchDetails.ImageTag = imageTag
	} else {
		patchDetails.ImageTag = upgradeVersion
	}
	patchDetails.PVName = d.Labels[pvLabelKey]
	return patchDetails, nil
}

func patchTargetDeploy(d *appsv1.Deployment, ns string) error {
	version, err := getOpenEBSVersion(d)
	if err != nil {
		return err
	}
	if (version != currentVersion) && (version != upgradeVersion) {
		return errors.Errorf(
			"target deployment version %s is neither %s nor %s",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	if version == currentVersion {
		tmpl, err := template.New("targetPatch").Funcs(template.FuncMap{
			"isCurrentLessThanNewVersion": util.IsCurrentLessThanNewVersion,
		}).Parse(templates.CstorTargetPatch)
		if err != nil {
			return errors.Wrapf(err, "failed to create template for cstor target deployment patch")
		}
		patchDetails, err := getTargetDeployPatchDetails(d)
		if err != nil {
			return err
		}
		patchDetails.UpgradeVersion = upgradeVersion
		patchDetails.CurrentVersion = currentVersion
		patchDetails.BaseDir = baseDir
		err = tmpl.Execute(&buffer, patchDetails)
		if err != nil {
			return errors.Wrapf(err, "failed to populate template for cstor target deployment patch")
		}
		targetPatch := buffer.String()
		buffer.Reset()
		err = patchDelpoyment(
			d.Name,
			ns,
			types.StrategicMergePatchType,
			[]byte(targetPatch),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to patch target deployment %s", d.Name)
		}
		klog.Infof("target deployment %s patched", d.Name)
	} else {
		klog.Infof("target deployment already in %s version", upgradeVersion)
	}
	return nil
}

func (c *cstorVolumeOptions) patchCV() error {
	version := c.cv.Labels["openebs.io/version"]
	if version == currentVersion {
		tmpl, err := template.New("cvPatch").
			Parse(templates.VersionDetailsPatch)
		if err != nil {
			return errors.Wrapf(err, "failed to create template for cstorvolume patch")
		}
		err = tmpl.Execute(&buffer, upgradeVersion)
		if err != nil {
			return errors.Wrapf(err, "failed to populate template for cstorvolume patch")
		}
		cvPatch := buffer.String()
		buffer.Reset()
		_, err = cvClient.WithNamespace(c.ns).Patch(
			c.cv.Name,
			c.ns,
			types.MergePatchType,
			[]byte(cvPatch),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to patch cstorvolume %s", c.cv.Name)
		}
		klog.Infof("cstorvolume %s patched", c.cv.Name)
	} else {
		klog.Infof("cstorvolume already in %s version", upgradeVersion)
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
			"cstorvolume version %s is neither %s nor %s",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	if version == currentVersion {
		tmpl, err := template.New("cvPatch").
			Parse(templates.VersionDetailsPatch)
		if err != nil {
			return errors.Wrapf(err, "failed to create template for cstorvolumereplica patch")
		}
		err = tmpl.Execute(&buffer, upgradeVersion)
		if err != nil {
			return errors.Wrapf(err, "failed to populate template for cstorvolumereplica patch")
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
			return errors.Wrapf(err, "failed to patch cstorvolumereplica %s", cvrObject.Name)
		}
		klog.Infof("cstorvolumereplica %s patched", cvrObject.Name)
	} else {
		klog.Infof("cstorvolume replica already in %s version", upgradeVersion)
	}
	return nil
}

func getCVRList(pvLabel, openebsNamespace string) (*apis.CStorVolumeReplicaList, error) {
	cvrList, err := cvrClient.WithNamespace(openebsNamespace).List(
		metav1.ListOptions{
			LabelSelector: pvLabel,
		},
	)
	if err != nil {
		return nil, err
	}
	if len(cvrList.Items) == 0 {
		return nil, errors.Errorf("no cvr found for label %s, in %s", pvLabel, openebsNamespace)
	}
	for _, cvrObj := range cvrList.Items {
		if cvrObj.Name == "" {
			return nil, errors.Errorf("missing cvr name for %v", cvrObj)
		}
	}
	err = verifyCSPVersion(cvrList, openebsNamespace)
	if err != nil {
		return nil, err
	}
	return cvrList, nil
}

type cstorVolumeOptions struct {
	utaskObj        *utask.UpgradeTask
	ns              string
	targetDeployObj *appsv1.Deployment
	cvrList         *apis.CStorVolumeReplicaList
	cv              *apis.CStorVolume
}

func (c *cstorVolumeOptions) preUpgrade(pvName, openebsNamespace string) error {
	var (
		err, uerr   error
		pvLabel     = pvLabelKey + "=" + pvName
		targetLabel = pvLabel + ",openebs.io/target=cstor-target"
	)

	statusObj := utask.UpgradeDetailedStatuses{Step: utask.PreUpgrade}

	statusObj.Phase = utask.StepErrored
	c.ns, err = getPVCDeploymentsNamespace(pvName, pvLabel, openebsNamespace)
	if err != nil {
		statusObj.Message = "failed to get namespace for pvc deployments"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	c.targetDeployObj, err = getDeployment(targetLabel, c.ns)
	if err != nil {
		statusObj.Message = "failed to get target details"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	c.cvrList, err = getCVRList(pvLabel, openebsNamespace)
	if err != nil {
		statusObj.Message = "failed to get replica details"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}
	c.cv, err = c.getCV(pvLabel)
	if err != nil {
		statusObj.Message = "failed to get cstorvolume"
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

func (c *cstorVolumeOptions) getCV(pvLabel string) (*apis.CStorVolume, error) {
	cvList, err := cvClient.WithNamespace(c.ns).List(
		metav1.ListOptions{
			LabelSelector: pvLabel,
		},
	)
	if err != nil {
		return nil, err
	}
	if len(cvList.Items) != 1 {
		return nil, errors.Errorf("invalid number of cstorvolume found : %d", len(cvList.Items))
	}
	version := cvList.Items[0].Labels["openebs.io/version"]
	if (version != currentVersion) && (version != upgradeVersion) {
		return nil, errors.Errorf(
			"cstorvolume version %s is neither %s nor %s",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	return &cvList.Items[0], nil
}

func (c *cstorVolumeOptions) verifyCVVersionReconcile(openebsNamespace string) error {
	var uerr error
	statusObj := utask.UpgradeDetailedStatuses{Step: utask.TargetUpgrade}
	statusObj.Phase = utask.StepErrored
	// waiting for the current version to be equal to desired version
	for c.cv.VersionDetails.Status.Current != upgradeVersion {
		klog.Infof("Verifying the reconciliation of version for %s", c.cv.Name)
		// Sleep equal to the default sync time
		time.Sleep(10 * time.Second)
		obj, err := cvClient.Get(c.cv.Name, metav1.GetOptions{})
		if err != nil {
			statusObj.Message = "failed to get cstor volume"
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
		c.cv = obj
		printCVStatus(c.cv.Status)
	}
	return nil
}

func (c *cstorVolumeOptions) waitForCVCurrentVersion(pvLabel, namespace string) error {
	var err error
	c.cv, err = c.getCV(pvLabel)
	if err != nil {
		return err
	}
	// waiting for old objects to get populated with new fields
	for c.cv.VersionDetails.Status.Current == "" {
		// Sleep equal to the default sync time
		klog.Infof("Waiting for cv current version to get populated.")
		time.Sleep(10 * time.Second)
		obj, err := cvClient.Get(c.cv.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		c.cv = obj
	}
	return nil
}

func (c *cstorVolumeOptions) targetUpgrade(pvName, openebsNamespace string) error {
	var (
		err, uerr          error
		pvLabel            = pvLabelKey + "=" + pvName
		targetServiceLabel = pvLabel + ",openebs.io/target-service=cstor-target-svc"
	)
	statusObj := utask.UpgradeDetailedStatuses{Step: utask.TargetUpgrade}
	statusObj.Phase = utask.StepWaiting
	c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}
	printCVStatus(c.cv.Status)
	statusObj.Phase = utask.StepErrored
	err = patchTargetDeploy(c.targetDeployObj, c.ns)
	if err != nil {
		statusObj.Message = "failed to patch target deployment"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	err = patchService(targetServiceLabel, c.ns)
	if err != nil {
		statusObj.Message = "failed to patch target service"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	err = c.waitForCVCurrentVersion(pvLabel, c.ns)
	if err != nil {
		statusObj.Message = "failed to verify version details for cstor volume"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	err = c.patchCV()
	if err != nil {
		statusObj.Message = "failed to patch cstor volume"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}
	err = c.verifyCVVersionReconcile(openebsNamespace)
	if err != nil {
		return err
	}

	statusObj.Phase = utask.StepCompleted
	statusObj.Message = "Target upgrade was successful"
	statusObj.Reason = ""
	c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}
	return nil
}

func waitForCVRCurrentVersion(name, openebsNamespace string) error {
	cvrObj, err := cvrClient.WithNamespace(openebsNamespace).
		Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// waiting for old objects to get populated with new fields
	for cvrObj.VersionDetails.Status.Current == "" {
		klog.Infof("Waiting for cvr current version to get populated.")
		// Sleep equal to the default sync time
		time.Sleep(10 * time.Second)
		cvrObj, err = cvrClient.WithNamespace(openebsNamespace).
			Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *cstorVolumeOptions) verifyCVRVersionReconcile(name, openebsNamespace string) error {
	var uerr error
	statusObj := utask.UpgradeDetailedStatuses{Step: utask.ReplicaUpgrade}
	statusObj.Phase = utask.StepErrored
	cvrObj, err := cvrClient.WithNamespace(openebsNamespace).
		Get(name, metav1.GetOptions{})
	if err != nil {
		statusObj.Message = "failed to get cstor volume replica"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}
	// waiting for the current version to be equal to desired version
	for cvrObj.VersionDetails.Status.Current != upgradeVersion {
		klog.Infof("Verifying the reconciliation of %s { id:%s phase:%s lastTransition:%s }",
			cvrObj.Name, cvrObj.Spec.ReplicaID, string(cvrObj.Status.Phase), cvrObj.Status.LastTransitionTime.String())
		// Sleep equal to the default sync time
		time.Sleep(10 * time.Second)
		cvrObj, err = cvrClient.WithNamespace(openebsNamespace).
			Get(name, metav1.GetOptions{})
		if err != nil {
			statusObj.Message = "failed to get cstor volume replica"
			statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
			c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
			if uerr != nil && isENVPresent {
				return uerr
			}
			return err
		}
		if cvrObj.VersionDetails.Status.Message != "" {
			statusObj.Message = cvrObj.VersionDetails.Status.Message
			statusObj.Reason = cvrObj.VersionDetails.Status.Reason
			c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
			if uerr != nil && isENVPresent {
				return uerr
			}
			klog.Errorf("failed to reconcile version : %s", cvrObj.VersionDetails.Status.Reason)
		}
	}
	return nil
}

func (c *cstorVolumeOptions) replicaUpgrade(openebsNamespace string) error {
	var uerr, err error
	statusObj := utask.UpgradeDetailedStatuses{Step: utask.ReplicaUpgrade}
	statusObj.Phase = utask.StepWaiting
	c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}

	statusObj.Phase = utask.StepErrored
	for _, cvrObj := range c.cvrList.Items {
		err = waitForCVRCurrentVersion(cvrObj.Name, cvrObj.Namespace)
		if err != nil {
			return err
		}
		err = patchCVR(cvrObj.Name, cvrObj.Namespace)
		if err != nil {
			statusObj.Message = "failed to patch cstor volume replica"
			statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
			c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
			if uerr != nil && isENVPresent {
				return uerr
			}
			return err
		}
		err = c.verifyCVRVersionReconcile(cvrObj.Name, cvrObj.Namespace)
		if err != nil {
			return err
		}
	}

	statusObj.Phase = utask.StepCompleted
	statusObj.Message = "Replica upgrade was successful"
	statusObj.Reason = ""
	c.utaskObj, uerr = updateUpgradeDetailedStatus(c.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}
	return nil
}

func printCVStatus(s apis.CStorVolumeStatus) {
	replicaStatusFormat := "{ id:%s mode:%s chckIOSeq:%s r/w/s:%s/%s/%s uptime:%d quorum:%s }"
	status := "CStor volume { phase:" + string(s.Phase) + " lastTransition:" + s.LastTransitionTime.String() + " } replicaStatuses: "
	for _, r := range s.ReplicaStatuses {
		status = status + "\n\t\t" + fmt.Sprintf(
			replicaStatusFormat,
			r.ID, r.Mode, r.CheckpointedIOSeq, r.InflightRead, r.InflightWrite,
			r.InflightSync, r.UpTime, r.Quorum,
		)
	}
	klog.Info(status)
}

func cstorVolumeUpgrade(pvName, openebsNamespace string, utaskObj *utask.UpgradeTask) (*utask.UpgradeTask, error) {
	var err error

	options := &cstorVolumeOptions{}

	options.utaskObj = utaskObj

	// PreUpgrade
	err = options.preUpgrade(pvName, openebsNamespace)
	if err != nil {
		return options.utaskObj, err
	}

	// ReplicaUpgrade
	err = options.replicaUpgrade(openebsNamespace)
	if err != nil {
		return options.utaskObj, err
	}

	// TargetUpgrade
	err = options.targetUpgrade(pvName, openebsNamespace)
	if err != nil {
		return options.utaskObj, err
	}

	klog.Infof("Upgrade Successful for cstor volume %s", pvName)
	return options.utaskObj, nil
}
