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
	"github.com/openebs/maya/pkg/util"
	errors "github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type replicaPatchDetails struct {
	UpgradeVersion, ImageTag, PVName, ReplicaContainerName, ReplicaImage string
}

type controllerPatchDetails struct {
	UpgradeVersion, ImageTag, ControllerContainerName, ControllerImage, MExporterImage string
	IsMonitorEnabled                                                                   bool
}

type replicaDetails struct {
	patchDetails  *replicaPatchDetails
	version, name string
	replicas      map[string]string
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
	patchDetails.IsMonitorEnabled = true
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
		if err.Error() == "image not found for container maya-volume-exporter" {
			patchDetails.IsMonitorEnabled = false
		} else {
			return nil, err
		}
	}
	patchDetails.MExporterImage = image
	if imageTag != "" {
		patchDetails.ImageTag = imageTag
	} else {
		patchDetails.ImageTag = upgradeVersion
	}
	return patchDetails, nil
}

func getPatchDetailsForReplicaDeploy(pvName string, deployObj *appsv1.Deployment) (*replicaPatchDetails, error) {
	patchDetails, err := getReplicaPatchDetails(deployObj)
	if err != nil {
		return nil, err
	}
	patchDetails.UpgradeVersion = upgradeVersion
	patchDetails.PVName = pvName
	return patchDetails, nil
}

func validateReplicaDeployVersion(d *appsv1.Deployment) (string, error) {
	version, err := getOpenEBSVersion(d)
	if err != nil {
		return "", err
	}
	if (version != currentVersion) && (version != upgradeVersion) {
		return "", errors.Errorf(
			"replica %s version %s is neither %s nor %s\n",
			d.Name,
			version,
			currentVersion,
			upgradeVersion,
		)
	}
	return version, nil
}

func getReplica(pvName, replicaLabel, volumeNamespace, openebsNamespace string) (*replicaDetails, error) {
	replicaObj := &replicaDetails{
		replicas: map[string]string{},
	}
	var err error
	// check if old replica is present for currentVersion < 1.9.0
	// if present then migration is not complete and store the old
	// replica details
	// replicaObj.name and replicaObj.version would be empty if old replica got
	// deleted as part of upgrade.
	// So, later on code uses replicaObj.name to perform replica related migration.
	if util.IsCurrentLessThanNewVersion(currentVersion, "1.9.0") {
		deployObj, err := deployClient.WithNamespace(volumeNamespace).Get(pvName + "-rep")

		if err != nil && !k8serror.IsNotFound(err) {
			return nil, errors.Wrapf(err, "failed to get replica deployment")
		}
		if err == nil {
			version, err := validateReplicaDeployVersion(deployObj)
			if err != nil {
				return nil, err
			}
			replicaObj.patchDetails, err = getPatchDetailsForReplicaDeploy(pvName, deployObj)
			if err != nil {
				return nil, err
			}
			replicaObj.name = deployObj.Name
			replicaObj.version = version
		}
	}
	replicaList, err := deployClient.WithNamespace(openebsNamespace).List(&metav1.ListOptions{
		LabelSelector: replicaLabel,
	})
	if err != nil {
		return nil, err
	}
	for _, replica := range replicaList.Items {
		// skip the old deployment as that will
		// be removed and not patched
		if replica.Name != pvName+"-rep" {
			deployObj := &replica
			version, err := validateReplicaDeployVersion(deployObj)
			if err != nil {
				return nil, err
			}
			//
			replicaObj.replicas[deployObj.Name] = version
			if replicaObj.patchDetails == nil {
				replicaObj.patchDetails, err = getPatchDetailsForReplicaDeploy(pvName, deployObj)
				if err != nil {
					return nil, err
				}
			}
		}
	}
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

func patchReplica(name, namespace string, replicaObj *replicaDetails) error {
	if replicaObj.replicas[name] == upgradeVersion {
		klog.Infof("replica deployment %s already in %s version", name, upgradeVersion)
		return nil
	}
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
		name,
		namespace,
		types.StrategicMergePatchType,
		[]byte(replicaPatch),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to patch replica deployment %s", name)
	}
	klog.Infof("%s patched", name)
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
	openebsNamespace string) (string, error) {
	pvObj, err := pvClient.Get(pvName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if pvObj.Spec.ClaimRef.Namespace == "" {
		return "", errors.Errorf("namespace missing for pv %s", pvName)
	}
	ns := pvObj.Spec.ClaimRef.Namespace
	// check for pv deployments in pv refclaim namespace
	deployList, err := deployClient.WithNamespace(ns).List(
		&metav1.ListOptions{
			LabelSelector: pvLabel,
		})
	if err != nil {
		return "", err
	}
	// check whether pvc pods are pvc namespace or not
	if len(deployList.Items) > 0 {
		return ns, nil
	}
	// if pvc pods are not in pvc namespace
	// verifying whether the pvc is deployed with DeployInOpenebsNamespace cas config
	deployList, err = deployClient.WithNamespace(openebsNamespace).List(
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
	return openebsNamespace, nil
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
	j.replicaObj, err = getReplica(pvName, replicaLabel, j.ns, openebsNamespace)
	if err != nil {
		statusObj.Message = "failed to get replica details"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

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

	if util.IsCurrentLessThanNewVersion(currentVersion, "1.9.0") {
		err = j.migrate(pvName, openebsNamespace)
		if err != nil {
			statusObj.Message = "failed to migrate deployments in openebes namespace"
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

// preReplicaUpgradeLessThan190 scales down old replica deployment
// and migrates the target service to openebs namespace
// before bringing up the new separate deployments
func (j *jivaVolumeOptions) preReplicaUpgradeLessThan190(pvName, openebsNamespace string) (string, error) {
	// if the upgrade is successful till replica cleanup and restarts
	// after that old replica will be missing and if replica cleanup
	// was done then service was also migrated successfully
	if util.IsCurrentLessThanNewVersion(currentVersion, "1.9.0") && j.replicaObj.name != "" {
		err := scaleDeploy(j.replicaObj.name, j.ns, replicaDeployLabel, 0)
		if err != nil {
			return "failed to get scale down replica deployment", err

		}
		err = j.migrateTargetSVC(pvName, openebsNamespace)
		if err != nil {
			return "failed to get migrate target service", err
		}
	}
	return "", nil
}

func (j *jivaVolumeOptions) replicaUpgrade(pvName, openebsNamespace string) error {
	var err, uerr error
	statusObj := utask.UpgradeDetailedStatuses{Step: utask.ReplicaUpgrade}
	statusObj.Phase = utask.StepWaiting
	j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
	if uerr != nil && isENVPresent {
		return uerr
	}
	statusObj.Phase = utask.StepErrored

	// Continue with replica upgrade only if the controller is not upgraded
	// otherwise return from here itself,
	// as controller is always upgarded after replicas are successfully upgraded.
	if j.controllerObj.version == upgradeVersion {
		klog.Infof("replicas already in %s version", upgradeVersion)
		statusObj.Phase = utask.StepCompleted
		statusObj.Message = "Replica upgrade was successful"
		statusObj.Reason = ""
		j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return nil
	}

	// Scaling down controller ensures no I/O occurs
	// which make volume to come in RW mode early
	err = scaleDeploy(j.controllerObj.name, j.ns, ctrlDeployLabel, 0)
	if err != nil {
		statusObj.Message = "failed to get scale down target deployment"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	msg, err := j.preReplicaUpgradeLessThan190(pvName, openebsNamespace)
	if err != nil {
		statusObj.Message = msg
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
	}

	for name := range j.replicaObj.replicas {
		// replica patch
		klog.Info("patching replica deployments")
		err = patchReplica(name, openebsNamespace, j.replicaObj)
		if err != nil {
			statusObj.Message = "failed to patch replica depoyment " + name
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

	// controller patch
	err = patchController(j.controllerObj, openebsNamespace)
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

	err = patchService(serviceLabel, openebsNamespace)
	if err != nil {
		statusObj.Message = "failed to patch target service"
		statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
		j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
		if uerr != nil && isENVPresent {
			return uerr
		}
		return err
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
	if util.IsCurrentLessThanNewVersion(currentVersion, "1.9.0") {
		err = j.cleanup(openebsNamespace)
		if err != nil {
			statusObj.Message = "failed to clean up old replica deployemts"
			statusObj.Reason = strings.Replace(err.Error(), ":", "", -1)
			j.utaskObj, uerr = updateUpgradeDetailedStatus(j.utaskObj, statusObj, openebsNamespace)
			if uerr != nil && isENVPresent {
				return uerr
			}
			return err
		}
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
	err = options.replicaUpgrade(pvName, openebsNamespace)
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

	klog.Info("Upgrade Successful for", pvName)
	return options.utaskObj, nil
}

func (j *jivaVolumeOptions) cleanup(openebsNamespace string) error {
	var err error
	if j.replicaObj.version == currentVersion && j.replicaObj.name != "" {
		klog.Info("cleaning old replica deployment")
		err = deployClient.WithNamespace(j.ns).Delete(j.replicaObj.name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	if j.controllerObj.version == currentVersion && j.ns != openebsNamespace {
		klog.Info("cleaning old controller deployment")
		err = deployClient.WithNamespace(j.ns).Delete(j.controllerObj.name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (j *jivaVolumeOptions) migrate(pvName, openebsNamespace string) error {
	var err error
	// if old replica is missing then migration was
	// successful till replica cleanup in previous iteration
	if j.replicaObj.name != "" {
		err = j.migrateReplica(openebsNamespace)
		if err != nil {
			return err
		}
	}
	// if pvc deployed in openebs namespace no need
	// to migrate the controller deployment
	if j.ns != openebsNamespace {
		err = j.migrateTarget(pvName, openebsNamespace)
	}
	return err
}

func getNodeNames(deployObj *appsv1.Deployment) (int, []string) {
	matchExp := deployObj.Spec.Template.Spec.Affinity.NodeAffinity.
		RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions
	for i, exp := range matchExp {
		if exp.Key == "kubernetes.io/hostname" {
			return i, exp.Values
		}
	}
	return -1, nil
}

func (j *jivaVolumeOptions) migrateReplica(openebsNamespace string) error {
	// get the old replica deployment by name
	oldDeployObj, err := deployClient.WithNamespace(j.ns).Get(j.replicaObj.name)
	if err != nil {
		return err
	}
	index, nodeNames := getNodeNames(oldDeployObj)
	if index == -1 {
		return errors.New("unable to find kubernetes.io/hostname key in nodeAffinity")
	}
	replicaCount := len(nodeNames)
	// get the separate replica deployments in openebs namespace
	deployList, err := deployClient.WithNamespace(openebsNamespace).List(&metav1.ListOptions{
		LabelSelector: replicaDeployLabel,
	})
	if err != nil {
		return err
	}
	replicasCreated := len(deployList.Items)
	// if the volume was deployed in openebs namespace while provisioning
	// as the old deployment also has the same label
	if j.ns == openebsNamespace {
		replicasCreated = replicasCreated - 1
	}

	// replica deployment pv-name-rep will be split into multiple replicas like
	// pv-name-rep-1, pv-name-rep-2,... pv-name-rep-n,
	// where n is the replica count for this volume.
	klog.Infof("splitting replica deployment")
	var zero int32
	for i := replicasCreated; i < replicaCount; i++ {
		replicaDeploy := oldDeployObj.DeepCopy()
		replicaDeploy.Name = replicaDeploy.Name + "-" + strconv.Itoa(i+1)
		replicaDeploy.Namespace = openebsNamespace
		replicaDeploy.ResourceVersion = ""
		replicaDeploy.Spec.Replicas = &zero
		replicaDeploy.Spec.Template.Spec.Affinity.NodeAffinity.
			RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[index].Values = []string{nodeNames[i]}
		klog.Infof("creating replica deployment %s in %s namespace", replicaDeploy.Name, openebsNamespace)
		replicaDeploy, err := deployClient.WithNamespace(openebsNamespace).Create(replicaDeploy)
		if err != nil {
			return err
		}
		j.replicaObj.replicas[replicaDeploy.Name] = replicaDeploy.Labels["openebs.io/version"]
	}

	return nil
}

func (j *jivaVolumeOptions) migrateTarget(pvName, openebsNamespace string) error {
	// get the controller deployment in openebs namespace
	// controllerObj.name cannot be nil as after successful upgrade
	// controller is removed and no deploy or svc is present in pvc namespace
	// so controllerObj will be in openebs namespace
	deployObj, err := deployClient.WithNamespace(openebsNamespace).Get(j.controllerObj.name)
	if err == nil {
		klog.Info("controller deployment already migrated to openebs namespace")
		return nil
	}
	if err != nil && !k8serror.IsNotFound(err) {
		return err
	}
	// if the deployment is not found in openebs namespace migrate it

	var zero int32
	deployObj, err = deployClient.WithNamespace(j.ns).Get(j.controllerObj.name)
	if err != nil {
		return err
	}
	deployObj.Namespace = openebsNamespace
	deployObj.ResourceVersion = ""
	deployObj.Spec.Replicas = &zero
	// if target-affinity is set for the pvc them openebs namespace
	// needs to be added as a bug fix.
	if deployObj.Spec.Template.Spec.Affinity != nil {
		deployObj.Spec.Template.Spec.Affinity.PodAffinity.
			RequiredDuringSchedulingIgnoredDuringExecution[0].
			Namespaces = []string{j.ns}
	}

	klog.Infof("creating controller deployment %s in %s namespace", deployObj.Name, openebsNamespace)
	_, err = deployClient.WithNamespace(openebsNamespace).Create(deployObj)
	return err
}

func (j *jivaVolumeOptions) migrateTargetSVC(pvName, openebsNamespace string) error {
	// migrate service only if service not in openebs namespace
	if j.ns == openebsNamespace {
		return nil
	}
	// get the original service and if present remove it
	svcObj, err := serviceClient.WithNamespace(j.ns).
		Get(j.controllerObj.name+"-svc", metav1.GetOptions{})
	if err != nil && !k8serror.IsNotFound(err) {
		return err
	}
	if err == nil {
		klog.Infof("removing controller service %s in %s namespace", svcObj.Name, j.ns)
		err = serviceClient.WithNamespace(j.ns).Delete(svcObj.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	// get the controller service in openebs namespace
	_, err = serviceClient.WithNamespace(openebsNamespace).
		Get(j.controllerObj.name+"-svc", metav1.GetOptions{})
	if err != nil && !k8serror.IsNotFound(err) {
		return err
	}
	// if the service is not found in openebs namespace create it
	if k8serror.IsNotFound(err) {
		svcObj, err := getTargetSVC(pvName, openebsNamespace)
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

func getTargetSVC(pvName, openebsNamespace string) (*corev1.Service, error) {
	pvObj, err := pvClient.Get(pvName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	storageClass := pvObj.Spec.StorageClassName
	svcObj := &corev1.Service{}
	svcObj.ObjectMeta = metav1.ObjectMeta{
		Name: pvName + "-ctrl-svc",
		Annotations: map[string]string{
			"openebs.io/storage-class-ref": `|
          	   name: ` + storageClass,
		},
		Labels: map[string]string{
			"openebs.io/storage-engine-type":     "jiva",
			"openebs.io/cas-type":                "jiva",
			"openebs.io/controller-service":      "jiva-controller-svc",
			"openebs.io/persistent-volume":       pvName,
			"openebs.io/persistent-volume-claim": pvObj.Spec.ClaimRef.Name,
			"pvc":                                pvObj.Spec.ClaimRef.Name,
			"openebs.io/version":                 currentVersion,
		},
	}
	svcObj.Spec = corev1.ServiceSpec{
		ClusterIP: strings.Split(pvObj.Spec.ISCSI.TargetPortal, ":")[0],
		Ports: []corev1.ServicePort{
			corev1.ServicePort{
				Name:       "iscsi",
				Port:       3260,
				Protocol:   "TCP",
				TargetPort: intstr.FromInt(3260),
			},
			corev1.ServicePort{
				Name:       "api",
				Port:       9501,
				Protocol:   "TCP",
				TargetPort: intstr.FromInt(9501),
			},
			corev1.ServicePort{
				Name:       "exporter",
				Port:       9500,
				Protocol:   "TCP",
				TargetPort: intstr.FromInt(9500),
			},
		},
		Selector: map[string]string{
			"openebs.io/controller":        "jiva-controller",
			"openebs.io/persistent-volume": pvName,
		},
	}
	return svcObj, nil
}
