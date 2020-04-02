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
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	utask "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	jivaClient "github.com/openebs/maya/pkg/client/jiva"
	templates "github.com/openebs/maya/pkg/upgrade/templates/v1"
	errors "github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"

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
		klog.Infof("%s patched", replicaObj.name)
	} else {
		klog.Infof("replica deployment already in %s version", upgradeVersion)
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
		klog.Infof("%s patched", controllerObj.name)
	} else {
		klog.Infof("controller deployment already in %s version\n", upgradeVersion)
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

func getReplicationFactor(ctrlLabel, namespace string) (int, error) {
	ctrlList, err := podClient.WithNamespace(namespace).List(
		metav1.ListOptions{
			LabelSelector: ctrlLabel,
		})
	if err != nil {
		return 0, err
	}
	if len(ctrlList.Items) == 0 {
		return 0, errors.Errorf("no deployments found for %s in %s", ctrlLabel, namespace)
	}
	ctrlPod := ctrlList.Items[0]
	// the only env in jiva target pod is "REPLICATION_FACTOR"
	return strconv.Atoi(ctrlPod.Spec.Containers[0].Env[0].Value)
}

func getAPIURL(svcLabel, namespace string) (string, error) {
	svcList, err := serviceClient.WithNamespace(namespace).List(
		metav1.ListOptions{
			LabelSelector: svcLabel,
		})
	if err != nil {
		return "", err
	}
	if len(svcList.Items) == 0 {
		return "", errors.Errorf("no service found for %s in %s", svcLabel, namespace)
	}
	targetIP := svcList.Items[0].Spec.ClusterIP
	apiURL := "http://" + targetIP + ":9501/v1/replicas"
	return apiURL, nil
}

func validateSync(pvLabel, namespace string) error {
	klog.Infof("Verifying replica sync")
	ctrlLabel := "openebs.io/controller=jiva-controller," + pvLabel
	svcLabel := "openebs.io/controller-service=jiva-controller-svc," + pvLabel
	quorum := false
	syncedReplicas := 0
	replicationFactor, err := getReplicationFactor(ctrlLabel, namespace)
	if err != nil {
		return err
	}
	apiURL, err := getAPIURL(svcLabel, namespace)
	if err != nil {
		return err
	}
	for syncedReplicas != replicationFactor {
		syncedReplicas = 0
		httpClient := &http.Client{Timeout: 30 * time.Second}
		resp, err := httpClient.Get(apiURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		replicas := jivaClient.ReplicaCollection{}
		err = json.NewDecoder(resp.Body).Decode(&replicas)
		if err != nil {
			return err
		}
		for _, replica := range replicas.Data {
			if replica.Mode == "RW" {
				syncedReplicas = syncedReplicas + 1
			}
		}
		if !quorum && syncedReplicas > (replicationFactor/2) {
			klog.Infof("Synced replica quorum is reached")
			quorum = true
		}
	}
	klog.Infof("Replica syncing complete")
	return nil
}

type jivaVolumeOptions struct {
	utaskObj      *utask.UpgradeTask
	replicaObj    *replicaDetails
	controllerObj *controllerDetails
	ns            string
}

func (j *jivaVolumeOptions) preupgrade(pvName, openebsNamespace string) error {
	var (
		pvLabel         = "openebs.io/persistent-volume=" + pvName
		replicaLabel    = "openebs.io/replica=jiva-replica," + pvLabel
		controllerLabel = "openebs.io/controller=jiva-controller," + pvLabel
		uerr, err       error
	)

	statusObj := utask.UpgradeDetailedStatuses{Step: utask.PreUpgrade}

	statusObj.Phase = utask.StepErrored
	j.ns, err = getPVCDeploymentsNamespace(pvName, pvLabel, openebsNamespace)
	if err != nil {
		statusObj.Message = "failed to get namespace for pvc deployments"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return errors.Wrapf(err, "failed to get namespace for pvc deployments")
	}

	// fetching replica deployment details
	j.replicaObj, err = getReplica(replicaLabel, j.ns)
	if err != nil {
		statusObj.Message = "failed to get replica details"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}
	j.replicaObj.patchDetails.PVName = pvName

	// fetching controller deployment details
	j.controllerObj, err = getController(controllerLabel, j.ns)
	if err != nil {
		statusObj.Message = "failed to get target details"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	if currentVersion < "1.9.0" {
		klog.Infof("Scaling down old deployments")
		err = j.preMigration(openebsNamespace)
		if err != nil {
			statusObj.Message = "failed to get scale down old deployments"
			statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
			j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
			if uerr != nil && isENVPresent {
				return uerr
			}
			return err
		}
	}

	statusObj.Phase = utask.StepCompleted
	statusObj.Message = "Pre-upgrade steps were successful"
	statusObj.Reason = ""
	j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}
	return nil
}

func (j *jivaVolumeOptions) replicaUpgrade(openebsNamespace string) error {
	var err, uerr error
	statusObj := utask.UpgradeDetailedStatuses{Step: utask.ReplicaUpgrade}
	statusObj.Phase = utask.StepWaiting
	j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}

	statusObj.Phase = utask.StepErrored

	if currentVersion < "1.9.0" {
		err = j.migrateReplica(openebsNamespace)
		if err != nil {
			statusObj.Message = "failed to migrate replica to openebs namespace"
			statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
			j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
			if uerr != nil && isENVPresent {
				return uerr
			}
			return errors.Wrap(err, "failed to migrate replica to openebs namespace")
		}
	} else {

		// Scaling down controller deployment before patching replica deployment
		// if the replica is not upgraded already.
		if j.replicaObj.version == currentVersion {
			err = scaleDeploy(j.controllerObj.name, j.ns, ctrlDeployLabel, 0)
			if err != nil {
				statusObj.Message = "failed to scale down target depoyment"
				statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
				j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
				if uerr != nil && isENVPresent {
					return uerr
				}
				return errors.Wrap(err, "failed to scale down target depoyment")
			}
		}

		// replica patch
		err = patchReplica(j.replicaObj, j.ns)
		if err != nil {
			// If patching of replica fails the controller needs to be reverted
			// as the upgrade will not proceed with controller patch.
			scaleErr := scaleDeploy(j.controllerObj.name, j.ns, ctrlDeployLabel, 1)
			if scaleErr != nil {
				klog.Infof("failed to scale up controller delpoyment. Please scale up deployment "+
					"%s in %s namespace to 1 manually.", j.controllerObj.name, j.ns)
			}
			statusObj.Message = "failed to patch replica depoyment"
			statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
			j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
			if uerr != nil && isENVPresent {
				return uerr
			}
			return err
		}
	}

	statusObj.Phase = utask.StepCompleted
	statusObj.Message = "Replica upgrade was successful"
	statusObj.Reason = ""
	j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}
	return nil
}

func (j *jivaVolumeOptions) targetUpgrade(pvName, openebsNamespace string) error {
	var err, uerr error
	statusObj := utask.UpgradeDetailedStatuses{Step: utask.TargetUpgrade}
	statusObj.Phase = utask.StepWaiting
	j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}

	statusObj.Phase = utask.StepErrored

	if currentVersion < "1.9.0" && j.ns != openebsNamespace {
		err = j.migrateTarget(openebsNamespace)
		if err != nil {
			statusObj.Message = "failed to migrate target to openebs namespace"
			statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
			j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
			if uerr != nil && isENVPresent {
				return uerr
			}
			return err
		}
	} else {
		// controller patch
		err = patchController(j.controllerObj, j.ns)
		if err != nil {
			statusObj.Message = "failed to patch target depoyment"
			statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
			j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
			if uerr != nil && isENVPresent {
				return uerr
			}
			return err
		}
		pvLabel := "openebs.io/persistent-volume=" + pvName
		serviceLabel := "openebs.io/controller-service=jiva-controller-svc," + pvLabel

		err = patchService(serviceLabel, j.ns)
		if err != nil {
			statusObj.Message = "failed to patch target service"
			statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
			j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
			if uerr != nil && isENVPresent {
				return uerr
			}
			return err
		}
	}

	statusObj.Phase = utask.StepCompleted
	statusObj.Message = "Target upgrade was successful"
	statusObj.Reason = ""
	j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}
	return nil
}

func (j *jivaVolumeOptions) verify(pvLabel, openebsNamespace string) error {
	// after the new ctrl and svc it takes few seconds for the
	// tcp connection to start
	time.Sleep(10 * time.Second)
	var err, uerr error
	statusObj := utask.UpgradeDetailedStatuses{Step: utask.Verify}
	statusObj.Phase = utask.StepWaiting
	j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}

	statusObj.Phase = utask.StepErrored
	// Verify synced replicas
	err = validateSync(pvLabel, openebsNamespace)
	if err != nil {
		statusObj.Message = "failed to verify synced replicas. Please check it manually using the steps mentioned in https://docs.openebs.io/docs/next/mayactl.html"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		if k8serror.IsForbidden(err) {
			klog.Warningf("failed to verify replica sync : %v\n Please check it manually using the steps mentioned in https://docs.openebs.io/docs/next/mayactl.html", err)
			return nil
		}
		return err
	}

	statusObj.Phase = utask.StepCompleted
	statusObj.Message = "Replica sync was successful"
	statusObj.Reason = ""
	j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}
	return nil
}

var (
	ctrlDeployLabel, replicaDeployLabel, ctrlSVCLabel string
)

func jivaUpgrade(pvName, openebsNamespace string, utaskObj *utask.UpgradeTask) (*utask.UpgradeTask, error) {

	var (
		pvLabel      = "openebs.io/persistent-volume=" + pvName
		ctrlLabel    = "openebs.io/controller=jiva-controller,"
		replicaLabel = "openebs.io/replica=jiva-replica,"
		svcLabel     = "openebs.io/controller-service=jiva-controller-svc,"
		err          error
	)

	ctrlDeployLabel = ctrlLabel + pvLabel
	replicaDeployLabel = replicaLabel + pvLabel
	ctrlSVCLabel = svcLabel + pvLabel

	options := &jivaVolumeOptions{}

	options.utaskObj = utaskObj

	// PreUpgrade
	err = options.preupgrade(pvName, openebsNamespace)
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

	// Verify
	err = options.verify(pvLabel, openebsNamespace)
	if err != nil {
		return options.utaskObj, err
	}

	err = options.cleanup(openebsNamespace)
	if err != nil {
		return options.utaskObj, err
	}

	klog.Info("Upgrade Successful for", pvName)
	return options.utaskObj, nil
}

func (j *jivaVolumeOptions) cleanup(openebsNamespace string) error {
	var err error
	if j.replicaObj.version == currentVersion {
		err = deployClient.WithNamespace(j.ns).Delete(j.replicaObj.name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	if j.controllerObj.version == currentVersion && j.ns != openebsNamespace {
		err = deployClient.WithNamespace(j.ns).Delete(j.controllerObj.name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (j *jivaVolumeOptions) preMigration(openebsNamespace string) error {
	var err error
	if j.replicaObj.version == currentVersion {
		err = scaleDeploy(j.controllerObj.name, j.ns, ctrlDeployLabel, 0)
		if err != nil {
			return err
		}
	}
	if j.controllerObj.version == currentVersion {
		err = scaleDeploy(j.replicaObj.name, j.ns, replicaDeployLabel, 0)
		if err != nil {
			return err
		}
	}

	return nil
}

func (j *jivaVolumeOptions) migrateReplica(openebsNamespace string) error {
	if j.replicaObj.version == currentVersion {
		klog.Infof("splitting replica deployment")
		deployObj, err := deployClient.WithNamespace(j.ns).Get(j.replicaObj.name)
		if err != nil {
			return err
		}
		nodeNames := deployObj.Spec.Template.Spec.Affinity.NodeAffinity.
			RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values
		var one int32 = 1
		for i, node := range nodeNames {
			replicaDeploy := deployObj.DeepCopy()
			replicaDeploy.Name = replicaDeploy.Name + "-" + strconv.Itoa(i+1)
			replicaDeploy.Namespace = openebsNamespace
			replicaDeploy.ResourceVersion = ""
			replicaDeploy.Spec.Replicas = &one
			// updating the replica deployments image and version before
			// creating them in openebs namespace.
			lastIndex := strings.LastIndex(replicaDeploy.Spec.Template.Spec.Containers[0].Image, ":")
			replicaDeploy.Spec.Template.Spec.Containers[0].Image = replicaDeploy.Spec.
				Template.Spec.Containers[0].Image[:lastIndex+1] + imageTag
			replicaDeploy.Labels["openebs.io/version"] = upgradeVersion
			replicaDeploy.Spec.Template.Spec.Affinity.NodeAffinity.
				RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values = []string{node}
			klog.Infof("creating replica deployment %s in %s namespace", replicaDeploy.Name, openebsNamespace)
			replicaDeploy, err := deployClient.WithNamespace(openebsNamespace).Create(replicaDeploy)
			if err != nil {
				return err
			}
			err = waitUntilPodsAreRunning(openebsNamespace, replicaDeployLabel, i+1)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (j *jivaVolumeOptions) migrateTarget(openebsNamespace string) error {
	if j.controllerObj.version == currentVersion {
		deployObj, err := deployClient.WithNamespace(j.ns).Get(j.controllerObj.name)
		if err != nil {
			return err
		}
		var one int32 = 1
		deployObj.Spec.Replicas = &one
		deployObj.Namespace = openebsNamespace
		deployObj.ResourceVersion = ""
		// updating the controller deploymentsimage and version before
		// creating them in openebs namespace.
		lastIndex := strings.LastIndex(deployObj.Spec.Template.Spec.Containers[0].Image, ":")
		deployObj.Spec.Template.Spec.Containers[0].Image = deployObj.Spec.
			Template.Spec.Containers[0].Image[:lastIndex+1] + imageTag
		lastIndex = strings.LastIndex(deployObj.Spec.Template.Spec.Containers[1].Image, ":")
		deployObj.Spec.Template.Spec.Containers[1].Image = deployObj.Spec.
			Template.Spec.Containers[1].Image[:lastIndex+1] + imageTag
		// if target-affinity is set for the pvc them openebs namespace
		// needs to be added as a bug fix.
		if deployObj.Spec.Template.Spec.Affinity != nil {
			if deployObj.Spec.Template.Spec.Affinity.PodAffinity.
				RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.
				MatchExpressions[0].Key == "openebs.io/target-affinity" {
				deployObj.Spec.Template.Spec.Affinity.PodAffinity.
					RequiredDuringSchedulingIgnoredDuringExecution[0].
					Namespaces = []string{j.ns}
			}
		}

		deployObj.Labels["openebs.io/version"] = upgradeVersion
		klog.Infof("creating controller deployment %s in %s namespace", deployObj.Name, openebsNamespace)
		deployObj, err = deployClient.WithNamespace(openebsNamespace).Create(deployObj)
		if err != nil {
			return err
		}
		err = waitUntilPodsAreRunning(openebsNamespace, ctrlDeployLabel, 1)
		if err != nil {
			return err
		}
		svcList, err := serviceClient.WithNamespace(j.ns).List(metav1.ListOptions{
			LabelSelector: ctrlSVCLabel,
		})
		if err != nil {
			return err
		}
		svcObj := &svcList.Items[0]
		svcObj.Namespace = openebsNamespace
		svcObj.ResourceVersion = ""
		svcObj.Labels["openebs.io/version"] = upgradeVersion
		klog.Infof("removing controller service %s in %s namespace", svcObj.Name, j.ns)
		err = serviceClient.WithNamespace(j.ns).Delete(svcObj.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		klog.Infof("creating controller service %s in %s namespace", svcObj.Name, openebsNamespace)
		svcObj, err = serviceClient.WithNamespace(openebsNamespace).Create(svcObj)
		if err != nil {
			return err
		}
	}
	return nil
}
