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
	templates "github.com/openebs/maya/pkg/upgrade/templates/v1"

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

type cstorTargetPatchDetails struct {
	UpgradeVersion, ImageTag, IstgtImage, MExporterImage, VolumeMgmtImage string
}

func verifyCSPVersion(pvLabel, namespace string) error {
	cvrList, err := cvrClient.WithNamespace(namespace).List(
		metav1.ListOptions{
			LabelSelector: pvLabel,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "failed to list cvr for %s", pvLabel)
	}
	if len(cvrList.Items) == 0 {
		return errors.Errorf("no cvr found for %s in %s", pvLabel, namespace)
	}
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
	if imageTag != "" {
		patchDetails.ImageTag = imageTag
	} else {
		patchDetails.ImageTag = upgradeVersion
	}
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
		tmpl, err := template.New("targetPatch").
			Parse(templates.CstorTargetPatch)
		if err != nil {
			return errors.Wrapf(err, "failed to create template for cstor target deployment patch")
		}
		patchDetails, err := getTargetDeployPatchDetails(d)
		if err != nil {
			return err
		}
		patchDetails.UpgradeVersion = upgradeVersion
		err = tmpl.Execute(&buffer, patchDetails)
		if err != nil {
			return errors.Wrapf(err, "failed to populate template for cstor target deployment patch")
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
			return errors.Wrapf(err, "failed to patch target deployment %s", d.Name)
		}
		glog.Infof("target deployment %s patched", d.Name)
	} else {
		glog.Infof("target deployment already in %s version", upgradeVersion)
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
			"cstorvolume version %s is neither %s nor %s",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	if version == currentVersion {
		tmpl, err := template.New("cvPatch").
			Parse(templates.OpenebsVersionPatch)
		if err != nil {
			return errors.Wrapf(err, "failed to create template for cstorvolume patch")
		}
		err = tmpl.Execute(&buffer, upgradeVersion)
		if err != nil {
			return errors.Wrapf(err, "failed to populate template for cstorvolume patch")
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
			return errors.Wrapf(err, "failed to patch cstorvolume %s", cvObject.Items[0].Name)
		}
		glog.Infof("cstorvolume %s patched", cvObject.Items[0].Name)
	} else {
		glog.Infof("cstorvolume already in %s version", upgradeVersion)
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
			Parse(templates.OpenebsVersionPatch)
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
		glog.Infof("cstorvolumereplica %s patched", cvrObject.Name)
	} else {
		glog.Infof("cstorvolume replica already in %s version", upgradeVersion)
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
	return cvrList, nil
}

func cstorVolumeUpgrade(pvName, openebsNamespace string) (*utask.UpgradeTask, error) {
	var (
		pvLabel            = "openebs.io/persistent-volume=" + pvName
		targetLabel        = pvLabel + ",openebs.io/target=cstor-target"
		targetServiceLabel = pvLabel + ",openebs.io/target-service=cstor-target-svc"
		err, uerr          error
		utaskObj           *utask.UpgradeTask
	)

	utaskObj, uerr = getOrCreateUpgradeTask("cstorVolume", pvName, openebsNamespace)
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

	err = verifyCSPVersion(pvLabel, openebsNamespace)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.PreUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to verify csp for cstor volume",
					Reason:  strings.ReplaceAll(err.Error(), ":", ""),
				},
			},
			openebsNamespace,
		)
		if uerr != nil && isENVPresent {
			return nil, uerr
		}
		return utaskObj, err
	}

	ns, err := getPVCDeploymentsNamespace(pvName, pvLabel, openebsNamespace)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.PreUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to get namespace for pvc deployments",
					Reason:  strings.ReplaceAll(err.Error(), ":", ""),
				},
			},
			openebsNamespace,
		)
		if uerr != nil && isENVPresent {
			return nil, uerr
		}
		return utaskObj, err
	}

	targetDeployObj, err := getDeployment(targetLabel, ns)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.PreUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to get target details",
					Reason:  strings.ReplaceAll(err.Error(), ":", ""),
				},
			},
			openebsNamespace,
		)
		if uerr != nil && isENVPresent {
			return nil, uerr
		}
		return utaskObj, err
	}

	cvrList, err := getCVRList(pvLabel, openebsNamespace)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.PreUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to get replica details",
					Reason:  strings.ReplaceAll(err.Error(), ":", ""),
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

	err = patchTargetDeploy(targetDeployObj, ns)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.TargetUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to patch target deployment",
					Reason:  strings.ReplaceAll(err.Error(), ":", ""),
				},
			},
			openebsNamespace,
		)
		if uerr != nil && isENVPresent {
			return nil, uerr
		}
		return utaskObj, err
	}

	err = patchService(targetServiceLabel, ns)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.TargetUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to patch target service",
					Reason:  strings.ReplaceAll(err.Error(), ":", ""),
				},
			},
			openebsNamespace,
		)
		if uerr != nil && isENVPresent {
			return nil, uerr
		}
		return utaskObj, err
	}

	err = patchCV(pvLabel, ns)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.TargetUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to patch cstor volume",
					Reason:  strings.ReplaceAll(err.Error(), ":", ""),
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

	utaskObj, uerr = updateUpgradeDetailedStatus(
		utaskObj,
		utask.UpgradeDetailedStatuses{
			Step: utask.ReplicaUpgrade,
			Status: utask.Status{
				Phase: utask.StepWaiting,
			},
		},
		openebsNamespace,
	)
	if uerr != nil && isENVPresent {
		return nil, uerr
	}

	for _, cvrObj := range cvrList.Items {
		err = patchCVR(cvrObj.Name, openebsNamespace)
		if err != nil {
			utaskObj, uerr = updateUpgradeDetailedStatus(
				utaskObj,
				utask.UpgradeDetailedStatuses{
					Step: utask.ReplicaUpgrade,
					Status: utask.Status{
						Phase:   utask.StepErrored,
						Message: "failed to patch cstor volume replica",
						Reason:  strings.ReplaceAll(err.Error(), ":", ""),
					},
				},
				openebsNamespace,
			)
			if uerr != nil && isENVPresent {
				return nil, uerr
			}
			return utaskObj, err
		}
	}

	utaskObj, uerr = updateUpgradeDetailedStatus(
		utaskObj,
		utask.UpgradeDetailedStatuses{
			Step: utask.ReplicaUpgrade,
			Status: utask.Status{
				Phase:   utask.StepCompleted,
				Message: "Replica upgrade was successful",
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
	glog.Infof("Upgrade Successful for cstor volume %s", pvName)
	return utaskObj, nil
}
