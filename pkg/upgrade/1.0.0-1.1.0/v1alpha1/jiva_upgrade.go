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
	templates "github.com/openebs/maya/pkg/upgrade/templates/v1"

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type replicaPatchDetails struct {
	UpgradeVersion, ImageTag, PVName, ReplicaContainerName, ReplicaImage string
}

type controllerPatchDetails struct {
	UpgradeVersion, ImageTag, ControllerContainerName, ControllerImage, MExporterImage string
}

type replicaDetails struct {
	patchDetails  *replicaPatchDetails
	version, name string
}

type controllerDetails struct {
	patchDetails  *controllerPatchDetails
	version, name string
}

func getReplicaPatchDetails(d *appsv1.Deployment) (
	*replicaPatchDetails,
	error,
) {
	patchDetails := &replicaPatchDetails{}
	// verify delpoyment name
	if d.Name == "" {
		return nil, errors.New("missing deployment name")
	}
	name, err := getContainerName(d)
	if err != nil {
		return nil, err
	}
	patchDetails.ReplicaContainerName = name
	image, err := getBaseImage(d, patchDetails.ReplicaContainerName)
	if err != nil {
		return nil, err
	}
	patchDetails.ReplicaImage = image
	if imageTag != "" {
		patchDetails.ImageTag = imageTag
	} else {
		patchDetails.ImageTag = upgradeVersion
	}
	return patchDetails, nil
}

func getControllerPatchDetails(d *appsv1.Deployment) (
	*controllerPatchDetails,
	error,
) {
	patchDetails := &controllerPatchDetails{}
	// verify delpoyment name
	if d.Name == "" {
		return nil, errors.New("missing deployment name")
	}
	name, err := getContainerName(d)
	if err != nil {
		return nil, err
	}
	patchDetails.ControllerContainerName = name
	image, err := getBaseImage(d, patchDetails.ControllerContainerName)
	if err != nil {
		return nil, err
	}
	patchDetails.ControllerImage = image
	image, err = getBaseImage(d, "maya-volume-exporter")
	if err != nil {
		return nil, err
	}
	patchDetails.MExporterImage = image
	if imageTag != "" {
		patchDetails.ImageTag = imageTag
	} else {
		patchDetails.ImageTag = upgradeVersion
	}
	return patchDetails, nil
}

func getReplica(replicaLabel, namespace string) (*replicaDetails, error) {
	replicaObj := &replicaDetails{}
	deployObj, err := getDeployment(replicaLabel, namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get replica deployment")
	}
	if deployObj.Name == "" {
		return nil, errors.Errorf("missing deployment name for replica")
	}
	replicaObj.name = deployObj.Name
	version, err := getOpenEBSVersion(deployObj)
	if err != nil {
		return nil, err
	}
	if (version != currentVersion) && (version != upgradeVersion) {
		return nil, errors.Errorf(
			"replica version %s is neither %s nor %s\n",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	replicaObj.version = version
	patchDetails, err := getReplicaPatchDetails(deployObj)
	if err != nil {
		return nil, err
	}
	replicaObj.patchDetails = patchDetails
	replicaObj.patchDetails.UpgradeVersion = upgradeVersion
	return replicaObj, nil
}

func getController(controllerLabel, namespace string) (*controllerDetails, error) {
	controllerObj := &controllerDetails{}
	deployObj, err := getDeployment(controllerLabel, namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get controller deployment")
	}
	if deployObj.Name == "" {
		return nil, errors.Errorf("missing deployment name for controller")
	}
	controllerObj.name = deployObj.Name
	version, err := getOpenEBSVersion(deployObj)
	if err != nil {
		return nil, err
	}
	if (version != currentVersion) && (version != upgradeVersion) {
		return nil, errors.Errorf(
			"controller version %s is neither %s nor %s\n",
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	controllerObj.version = version
	patchDetails, err := getControllerPatchDetails(deployObj)
	if err != nil {
		return nil, err
	}
	controllerObj.patchDetails = patchDetails
	controllerObj.patchDetails.UpgradeVersion = upgradeVersion
	return controllerObj, nil
}

func patchReplica(replicaObj *replicaDetails, namespace string) error {
	if replicaObj.version == currentVersion {
		tmpl, err := template.New("replicaPatch").
			Parse(templates.JivaReplicaPatch)
		if err != nil {
			return errors.Wrapf(err, "failed to create template for replica patch")
		}
		err = tmpl.Execute(&buffer, replicaObj.patchDetails)
		if err != nil {
			return errors.Wrapf(err, "failed to populate template for replica patch")
		}
		replicaPatch := buffer.String()
		buffer.Reset()
		err = patchDelpoyment(
			replicaObj.name,
			namespace,
			types.StrategicMergePatchType,
			[]byte(replicaPatch),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to patch replica deployment")
		}
		glog.Infof("%s patched", replicaObj.name)
	} else {
		glog.Infof("replica deployment already in %s version", upgradeVersion)
	}
	return nil
}

func patchController(controllerObj *controllerDetails, namespace string) error {
	if controllerObj.version == currentVersion {
		tmpl, err := template.New("controllerPatch").
			Parse(templates.JivaTargetPatch)
		if err != nil {
			return errors.Wrapf(err, "failed to create template for controller patch")
		}
		err = tmpl.Execute(&buffer, controllerObj.patchDetails)
		if err != nil {
			return errors.Wrapf(err, "failed to populate template for controller patch")
		}
		controllerPatch := buffer.String()
		buffer.Reset()
		err = patchDelpoyment(
			controllerObj.name,
			namespace,
			types.StrategicMergePatchType,
			[]byte(controllerPatch),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to patch replica deployment")
		}
		glog.Infof("%s patched", controllerObj.name)
	} else {
		glog.Infof("controller deployment already in %s version\n", upgradeVersion)
	}
	return nil
}

func getPVCDeploymentsNamespace(
	pvName,
	pvLabel,
	openebsNamespace string) (ns string, err error) {
	pvObj, err := pvClient.Get(pvName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	// verifying whether the pvc is deployed with DeployInOpenebsNamespace cas config
	deployList, err := deployClient.WithNamespace(openebsNamespace).List(
		&metav1.ListOptions{
			LabelSelector: pvLabel,
		})
	if err != nil {
		return "", err
	}
	// check whether pvc pods are openebs namespace or not
	if len(deployList.Items) > 0 {
		ns = openebsNamespace
		return ns, nil
	}
	// if pvc pods are not in openebs namespace take the namespace of pvc
	if pvObj.Spec.ClaimRef.Namespace == "" {
		return "", errors.Errorf("namespace missing for pv %s", pvName)
	}
	ns = pvObj.Spec.ClaimRef.Namespace
	// check for pv deployments in pv refclaim namespace
	deployList, err = deployClient.WithNamespace(ns).List(
		&metav1.ListOptions{
			LabelSelector: pvLabel,
		})
	if err != nil {
		return "", err
	}
	if len(deployList.Items) == 0 {
		return "", errors.Errorf(
			"failed to get deployments for pv %s in %s or %s namespace",
			pvName,
			openebsNamespace,
			ns,
		)
	}
	return ns, nil
}

func jivaUpgrade(pvName, openebsNamespace string) (*utask.UpgradeTask, error) {

	var (
		pvLabel         = "openebs.io/persistent-volume=" + pvName
		replicaLabel    = "openebs.io/replica=jiva-replica," + pvLabel
		controllerLabel = "openebs.io/controller=jiva-controller," + pvLabel
		serviceLabel    = "openebs.io/controller-service=jiva-controller-svc," + pvLabel
		ns              string
		err, uerr       error
	)

	var utaskObj *utask.UpgradeTask
	utaskObj, uerr = getOrCreateUpgradeTask("jivaVolume", pvName, openebsNamespace)
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

	ns, err = getPVCDeploymentsNamespace(pvName, pvLabel, openebsNamespace)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.PreUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to get namespace for pvc deployments",
					Reason:  strings.Replace(err.Error(), ":", "", -1),
				},
			},
			openebsNamespace,
		)
		if uerr != nil && isENVPresent {
			return nil, uerr
		}
		return utaskObj, errors.Wrapf(err, "failed to get namespace for pvc deployments")
	}

	// fetching replica deployment details
	replicaObj, err := getReplica(replicaLabel, ns)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.PreUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to get replica details",
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
	replicaObj.patchDetails.PVName = pvName

	// fetching controller deployment details
	controllerObj, err := getController(controllerLabel, ns)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.PreUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to get target details",
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

	// replica patch
	err = patchReplica(replicaObj, ns)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.ReplicaUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to patch replica depoyment",
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

	// controller patch
	err = patchController(controllerObj, ns)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.TargetUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to patch target depoyment",
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

	err = patchService(serviceLabel, ns)
	if err != nil {
		utaskObj, uerr = updateUpgradeDetailedStatus(
			utaskObj,
			utask.UpgradeDetailedStatuses{
				Step: utask.TargetUpgrade,
				Status: utask.Status{
					Phase:   utask.StepErrored,
					Message: "failed to patch target service",
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

	glog.Info("Upgrade Successful for", pvName)
	return utaskObj, nil
}
