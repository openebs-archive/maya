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

package cstorvolumeclaim

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	apispdb "github.com/openebs/maya/pkg/kubernetes/poddisruptionbudget"
	"github.com/openebs/maya/pkg/version"

	cspi "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	cv "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	cvclaim "github.com/openebs/maya/pkg/cstorvolumeclaim/v1alpha1"
	"github.com/openebs/maya/pkg/hash"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
)

const (
	cvcKind = "CStorVolumeClaim"
	cvKind  = "CStorVolume"
	// ReplicaCount represents replica count value
	ReplicaCount = "replicaCount"
	// pvSelector is the selector key for cstorvolumereplica belongs to a cstor
	// volume
	pvSelector = "openebs.io/persistent-volume"
	// minHAReplicaCount is minimum no.of replicas are required to decide
	// HighAvailable volume
	minHAReplicaCount = 3
	volumeID          = "openebs.io/volumeID"
	cspiLabel         = "cstorpoolinstance.openebs.io/name"
	cspiOnline        = "ONLINE"
)

// replicaInfo struct is used to pass replica information to
// create CVR
type replicaInfo struct {
	replicaID string
	phase     apis.CStorVolumeReplicaPhase
}

var (
	cvPorts = []corev1.ServicePort{
		corev1.ServicePort{
			Name:     "cstor-iscsi",
			Port:     3260,
			Protocol: "TCP",
			TargetPort: intstr.IntOrString{
				IntVal: 3260,
			},
		},
		corev1.ServicePort{
			Name:     "cstor-grpc",
			Port:     7777,
			Protocol: "TCP",
			TargetPort: intstr.IntOrString{
				IntVal: 7777,
			},
		},
		corev1.ServicePort{
			Name:     "mgmt",
			Port:     6060,
			Protocol: "TCP",
			TargetPort: intstr.IntOrString{
				IntVal: 6060,
			},
		},
		corev1.ServicePort{
			Name:     "exporter",
			Port:     9500,
			Protocol: "TCP",
			TargetPort: intstr.IntOrString{
				IntVal: 9500,
			},
		},
	}
	// openebsNamespace is global variable and it is initialized during starting
	// of the controller
	openebsNamespace string
)

// getTargetServiceLabels get the labels for cstor volume service
func getTargetServiceLabels(claim *apis.CStorVolumeClaim) map[string]string {
	return map[string]string{
		"openebs.io/target-service":      "cstor-target-svc",
		"openebs.io/storage-engine-type": "cstor",
		"openebs.io/cas-type":            "cstor",
		"openebs.io/persistent-volume":   claim.Name,
		"openebs.io/version":             version.GetVersion(),
	}
}

// getTargetServiceSelectors get the selectors for cstor volume service
func getTargetServiceSelectors(claim *apis.CStorVolumeClaim) map[string]string {
	return map[string]string{
		"app":                          "cstor-volume-manager",
		"openebs.io/target":            "cstor-target",
		"openebs.io/persistent-volume": claim.Name,
	}
}

// getTargetServiceOwnerReference get the ownerReference for cstorvolume service
func getTargetServiceOwnerReference(claim *apis.CStorVolumeClaim) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(claim,
			apis.SchemeGroupVersion.WithKind(cvcKind)),
	}
}

// getCVRLabels get the labels for cstorvolumereplica
func getCVRLabels(pool *apis.CStorPoolInstance, volumeName string) map[string]string {
	return map[string]string{
		"cstorpoolinstance.openebs.io/name": pool.Name,
		"cstorpoolinstance.openebs.io/uid":  string(pool.UID),
		"cstorvolume.openebs.io/name":       volumeName,
		"openebs.io/persistent-volume":      volumeName,
		"openebs.io/version":                version.GetVersion(),
	}
}

// getCVRAnnotations get the annotations for cstorvolumereplica
func getCVRAnnotations(pool *apis.CStorPoolInstance) map[string]string {
	return map[string]string{
		"cstorpoolinstance.openebs.io/hostname": pool.Labels["kubernetes.io/hostname"],
	}
}

// getCVRFinalizer get the finalizer for cstorvolumereplica
func getCVRFinalizer() []string {
	return []string{
		cvr.CStorVolumeReplicaFinalizer,
	}
}

// getCVROwnerReference get the ownerReference for cstorvolumereplica
func getCVROwnerReference(cv *apis.CStorVolume) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(cv,
			apis.SchemeGroupVersion.WithKind(cvKind)),
	}
}

// getCVLabels get the labels for cstorvolume
func getCVLabels(claim *apis.CStorVolumeClaim) map[string]string {
	return map[string]string{
		"openebs.io/persistent-volume": claim.Name,
		"openebs.io/version":           version.GetVersion(),
	}
}

// getCVOwnerReference get the ownerReference for cstorvolume
func getCVOwnerReference(cvc *apis.CStorVolumeClaim) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(cvc,
			apis.SchemeGroupVersion.WithKind(cvcKind)),
	}
}

// getNamespace gets the namespace OPENEBS_NAMESPACE env value which is set by the
// downward API where CVC-Operator has been deployed
func getNamespace() string {
	return menv.Get(menv.OpenEBSNamespace)
}

// getCSPC gets CStorPoolCluster name from cstorvolumeclaim resource
func getCSPC(
	claim *apis.CStorVolumeClaim,
) string {

	cspcName := claim.Labels[string(apis.CStorPoolClusterCPK)]
	return cspcName
}

// getPDBName returns the PDB name from cStor Volume Claim label
func getPDBName(claim *apis.CStorVolumeClaim) string {
	return claim.GetLabels()[string(apis.PodDisruptionBudgetKey)]
}

// listCStorPools get the list of available pool using the storagePoolClaim
// as labelSelector.
func listCStorPools(
	cspcName string,
	replicaCount int,
) (*apis.CStorPoolInstanceList, error) {

	if cspcName == "" {
		return nil, errors.New("failed to list cstorpool: cspc name missing")
	}

	cstorPoolList, err := cspi.NewKubeClient().List(metav1.ListOptions{
		LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + cspcName,
	})

	//	cspList, err := ncsp.NewKubeClient().WithNamespace(getNamespace()).
	//		List(metav1.ListOptions{
	//			LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + cspcName,
	//		})

	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to list cstorpool for cspc {%s}",
			cspcName,
		)
	}
	if len(cstorPoolList.Items) < replicaCount {
		return nil, errors.New("not enough pools available to create replicas")
	}
	return cstorPoolList, nil
}

// getOrCreateTargetService creates cstor volume target service
func getOrCreateTargetService(
	claim *apis.CStorVolumeClaim,
) (*corev1.Service, error) {

	svcObj, err := svc.NewKubeClient(svc.WithNamespace(openebsNamespace)).
		Get(claim.Name, metav1.GetOptions{})

	if err == nil {
		return svcObj, nil
	}

	// error other than 'not found', return err
	if !k8serror.IsNotFound(err) {
		return nil, errors.Wrapf(
			err,
			"failed to get cstorvolume service {%v}",
			svcObj.Name,
		)
	}

	// Not found case, so need to create
	svcObj, err = svc.NewBuilder().
		WithName(claim.Name).
		WithLabelsNew(getTargetServiceLabels(claim)).
		WithOwnerReferenceNew(getTargetServiceOwnerReference(claim)).
		WithSelectorsNew(getTargetServiceSelectors(claim)).
		WithPorts(cvPorts).
		Build()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to build target service {%v}",
			svcObj,
		)
	}

	svcObj, err = svc.NewKubeClient(svc.WithNamespace(openebsNamespace)).Create(svcObj)
	return svcObj, err
}

// getOrCreateCStorVolumeResource creates CStorVolume resource for a cstor volume
func getOrCreateCStorVolumeResource(
	service *corev1.Service,
	claim *apis.CStorVolumeClaim,
) (*apis.CStorVolume, error) {
	var (
		srcVolume string
		err       error
	)

	qCap := claim.Spec.Capacity[corev1.ResourceStorage]

	// get the replicaCount from cstorvolume claim
	rfactor := claim.Spec.ReplicaCount
	desiredRF := claim.Spec.ReplicaCount
	cfactor := rfactor/2 + 1

	volLabels := getCVLabels(claim)
	if len(claim.Spec.CstorVolumeSource) != 0 {
		srcVolume, _, err = getSrcDetails(claim.Spec.CstorVolumeSource)
		if err != nil {
			return nil, err
		}
		volLabels[string(apis.SourceVolumeKey)] = srcVolume
	}

	cvObj, err := cv.NewKubeclient(cv.WithNamespace(openebsNamespace)).
		Get(claim.Name, metav1.GetOptions{})
	if err != nil && !k8serror.IsNotFound(err) {
		return nil, errors.Wrapf(
			err,
			"failed to get cstorvolume {%v}",
			cvObj.Name,
		)
	}

	if k8serror.IsNotFound(err) {
		cvObj, err = cv.NewBuilder().
			WithName(claim.Name).
			WithLabelsNew(volLabels).
			WithOwnerRefernceNew(getCVOwnerReference(claim)).
			WithTargetIP(service.Spec.ClusterIP).
			WithCapacity(qCap.String()).
			WithCStorIQN(claim.Name).
			WithNodeBase(cv.CStorNodeBase).
			WithTargetPortal(service.Spec.ClusterIP + ":" + cv.TargetPort).
			WithTargetPort(cv.TargetPort).
			WithReplicationFactor(rfactor).
			WithDesiredReplicationFactor(desiredRF).
			WithConsistencyFactor(cfactor).
			WithNewVersion(version.GetVersion()).
			WithDependentsUpgraded().
			Build()
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to build cstorvolume {%v}",
				claim.Name,
			)
		}
		return cv.NewKubeclient(cv.WithNamespace(openebsNamespace)).Create(cvObj)
	}
	return cvObj, err
}

// distributeCVRs create cstorvolume replica based on the replicaCount
// on the available cstor pools created for storagepoolclaim.
// if pools are less then desired replicaCount its return an error.
func (c *CVCController) distributeCVRs(
	pendingReplicaCount int,
	claim *apis.CStorVolumeClaim,
	service *corev1.Service,
	volume *apis.CStorVolume,
	policy *apis.CStorVolumePolicy,
) error {
	var (
		usablePoolList *apis.CStorPoolInstanceList
		srcVolName     string
		err            error
	)

	cspcName := getCSPC(claim)
	if len(cspcName) == 0 {
		return errors.New("failed to get cspc name from cstorvolumeclaim")
	}

	poolList, err := listCStorPools(cspcName, claim.Spec.ReplicaCount)
	if err != nil {
		return err
	}

	if claim.Spec.CstorVolumeSource != "" {
		srcVolName, _, err = getSrcDetails(claim.Spec.CstorVolumeSource)
		if err != nil {
			return err
		}
		usablePoolList = getUsablePoolListForClone(volume.Name, srcVolName, poolList)
	} else {
		usablePoolList = getUsablePoolList(volume.Name, poolList)
	}

	// randomizePoolList to get the pool list in random order
	usablePoolList = randomizePoolList(usablePoolList)

	// prioritized pool instances matched to the given
	// nodeName in case of replica affinity is enabled via cstor volume policy
	if c.isReplicaAffinityEnabled(policy) {
		usablePoolList = prioritizedPoolList(claim.Publish.NodeID, usablePoolList)
	}
	for count, pool := range usablePoolList.Items {
		pool := pool
		if count < pendingReplicaCount {
			rInfo := replicaInfo{
				phase: apis.CVRStatusEmpty,
			}
			_, err = createCVR(service, volume, claim, &pool, rInfo)
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
	return nil
}

func getSrcDetails(cstorVolumeSrc string) (string, string, error) {
	volSrc := strings.Split(cstorVolumeSrc, "@")
	if len(volSrc) == 0 {
		return "", "", errors.New(
			"failed to get volumeSource",
		)
	}
	return volSrc[0], volSrc[1], nil
}

// createCVR is actual method to create cstorvolumereplica resource on a given
// cstor pool
func createCVR(
	service *corev1.Service,
	volume *apis.CStorVolume,
	claim *apis.CStorVolumeClaim,
	pool *apis.CStorPoolInstance,
	rInfo replicaInfo,
) (*apis.CStorVolumeReplica, error) {
	var (
		isClone             string
		srcVolume, snapName string
		err                 error
	)
	annotations := getCVRAnnotations(pool)
	labels := getCVRLabels(pool, volume.Name)

	if claim.Spec.CstorVolumeSource != "" {
		isClone = "true"
		srcVolume, snapName, err = getSrcDetails(claim.Spec.CstorVolumeSource)
		if err != nil {
			return nil, err
		}
		annotations[string(apis.SourceVolumeKey)] = srcVolume
		annotations[string(apis.SnapshotNameKey)] = snapName
		labels[string(apis.CloneEnableKEY)] = isClone
	}
	cvrObj, err := cvr.NewKubeclient(cvr.WithNamespace(openebsNamespace)).
		Get(volume.Name+"-"+string(pool.Name), metav1.GetOptions{})

	if err != nil && !k8serror.IsNotFound(err) {
		return nil, errors.Wrapf(
			err,
			"failed to get cstorvolumereplica {%v}",
			cvrObj.Name,
		)
	}
	if k8serror.IsNotFound(err) {
		cvrObj, err = cvr.NewBuilder().
			WithName(volume.Name + "-" + pool.Name).
			WithLabelsNew(labels).
			WithAnnotationsNew(annotations).
			WithOwnerRefernceNew(getCVROwnerReference(volume)).
			WithFinalizers(getCVRFinalizer()).
			WithTargetIP(service.Spec.ClusterIP).
			WithReplicaID(rInfo.replicaID).
			WithCapacity(volume.Spec.Capacity.String()).
			WithNewVersion(version.GetVersion()).
			WithDependentsUpgraded().
			WithStatusPhase(rInfo.phase).
			Build()
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to build cstorvolumereplica {%v}",
				cvrObj.Name,
			)
		}
		cvrObj, err = cvr.NewKubeclient(cvr.WithNamespace(openebsNamespace)).Create(cvrObj)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to create cstorvolumereplica {%v}",
				cvrObj.Name,
			)
		}
		return cvrObj, nil
	}
	return cvrObj, nil
}

// getPoolmapFromCVRList returns a list of cstor pool
// name corresponding to cstor volume replica
// instances
func getPoolMapFromCVRList(cvrList *apis.CStorVolumeReplicaList) map[string]bool {
	var poolMap = make(map[string]bool)
	for _, cvr := range cvrList.Items {
		poolName := cvr.GetLabels()[string(apis.CStorpoolInstanceLabel)]
		if poolName != "" {
			poolMap[poolName] = true
		}
	}
	return poolMap
}

// GetUsablePoolList returns a list of usable cstorpools
// which hasn't been used to create cstor volume replica
// instances
func getUsablePoolList(pvName string, poolList *apis.CStorPoolInstanceList) *apis.CStorPoolInstanceList {
	usablePoolList := &apis.CStorPoolInstanceList{}

	pvLabel := pvSelector + "=" + pvName
	cvrList, err := cvr.NewKubeclient(cvr.WithNamespace(openebsNamespace)).List(metav1.ListOptions{
		LabelSelector: pvLabel,
	})
	if err != nil {
		return nil
	}

	usedPoolMap := getPoolMapFromCVRList(cvrList)
	for _, pool := range poolList.Items {
		if !usedPoolMap[pool.Name] {
			usablePoolList.Items = append(usablePoolList.Items, pool)
		}
	}
	return usablePoolList
}

// GetUsablePoolListForClones returns a list of usable cstorpools
// which hasn't been used to create cstor volume replica
// instances
func getUsablePoolListForClone(pvName, srcPVName string, poolList *apis.CStorPoolInstanceList) *apis.CStorPoolInstanceList {
	usablePoolList := &apis.CStorPoolInstanceList{}

	pvLabel := pvSelector + "=" + pvName
	cvrList, err := cvr.NewKubeclient(cvr.WithNamespace(openebsNamespace)).List(metav1.ListOptions{
		LabelSelector: pvLabel,
	})
	if err != nil {
		return nil
	}
	srcPVLabel := pvSelector + "=" + srcPVName
	srcCVRList, err := cvr.NewKubeclient(cvr.WithNamespace(openebsNamespace)).List(metav1.ListOptions{
		LabelSelector: srcPVLabel,
	})
	if err != nil {
		return nil
	}

	srcVolPoolMap := getPoolMapFromCVRList(srcCVRList)
	usedPoolMap := getPoolMapFromCVRList(cvrList)
	for _, pool := range poolList.Items {
		if !usedPoolMap[pool.Name] && srcVolPoolMap[pool.Name] {
			usablePoolList.Items = append(usablePoolList.Items, pool)
		}
	}
	return usablePoolList
}

// prioritizedPoolList prioritized pool instance name matched to the given
// nodeName in case of replica affinity is enabled via volume policy
func prioritizedPoolList(nodeName string, list *apis.CStorPoolInstanceList) *apis.CStorPoolInstanceList {
	for i, pool := range list.Items {
		if pool.Spec.HostName != nodeName {
			continue
		}
		list.Items[0], list.Items[i] = list.Items[i], list.Items[0]
		break
	}
	return list
}

// randomizePoolList returns randomized pool list
func randomizePoolList(list *apis.CStorPoolInstanceList) *apis.CStorPoolInstanceList {
	res := &apis.CStorPoolInstanceList{}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	perm := r.Perm(len(list.Items))
	for _, randomIdx := range perm {
		res.Items = append(res.Items, list.Items[randomIdx])
	}
	return res
}

// getOrCreatePodDisruptionBudget will does following things
// 1. It tries to get the PDB that was created among volume replica pools.
// 2. If PDB exist it returns the PDB.
// 3. If PDB doesn't exist it creates new PDB(With CSPC hash)
func getOrCreatePodDisruptionBudget(
	cvObj *apis.CStorVolume, cspcName string, poolNames []string) (*policy.PodDisruptionBudget, error) {
	// poolNames, err := cvr.GetVolumeReplicaPoolNames(pvName, openebsNamespace)
	// if err != nil {
	// 	return nil, errors.Wrapf(err,
	// 		"failed to get volume replica pool names of volume %s",
	// 		cvObj.Name)
	// }
	pdbLabels := cvclaim.GetPDBPoolLabels(poolNames)
	labelSelector := apispdb.GetPDBLabelSelector(pdbLabels)
	pdbList, err := apispdb.KubeClient().
		WithNamespace(openebsNamespace).
		List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to list PDB belongs to pools %v", pdbLabels)
	}
	if len(pdbList.Items) > 1 {
		return nil, errors.Wrapf(err,
			"current PDB count %d of pools %v",
			len(pdbList.Items),
			pdbLabels)
	}
	if len(pdbList.Items) == 1 {
		return &pdbList.Items[0], nil
	}
	return createPDB(poolNames, cspcName)
}

// createPDB creates PDB for cStorVolumes based on arguments
func createPDB(poolNames []string, cspcName string) (*policy.PodDisruptionBudget, error) {
	// Calculate minAvailable value from cStorVolume replica count
	//minAvailable := (cvObj.Spec.ReplicationFactor >> 1) + 1
	maxUnavailableIntStr := intstr.FromInt(1)

	//build podDisruptionBudget for volume
	pdbObj := policy.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cspcName,
			Labels:       cvclaim.GetPDBLabels(poolNames, cspcName),
		},
		Spec: policy.PodDisruptionBudgetSpec{
			MaxUnavailable: &maxUnavailableIntStr,
			Selector:       getPDBSelector(poolNames),
		},
	}
	// Create podDisruptionBudget
	return apispdb.KubeClient().
		WithNamespace(openebsNamespace).
		Create(&pdbObj)
}

// getPDBSelector returns PDB label selector from list of pools
func getPDBSelector(pools []string) *metav1.LabelSelector {
	selectorRequirements := []metav1.LabelSelectorRequirement{}
	selectorRequirements = append(
		selectorRequirements,
		metav1.LabelSelectorRequirement{
			Key:      string(apis.CStorPoolInstanceCPK),
			Operator: metav1.LabelSelectorOpIn,
			Values:   pools,
		})
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "cstor-pool",
		},
		MatchExpressions: selectorRequirements,
	}
}

// addReplicaPoolInfo updates in-memory replicas pool information on spec and
// status of CVC
func addReplicaPoolInfo(cvcObj *apis.CStorVolumeClaim, poolNames []string) {
	for i, poolName := range poolNames {
		cvcObj.Spec.Policy.ReplicaPool.PoolInfo[i].PoolName = poolName
	}
	cvcObj.Status.PoolInfo = append(cvcObj.Status.PoolInfo, poolNames...)
}

// addPDBLabelOnCVC will add PodDisruptionBudget label on CVC
func addPDBLabelOnCVC(
	cvcObj *apis.CStorVolumeClaim, pdbObj *policy.PodDisruptionBudget) {
	cvcLabels := cvcObj.GetLabels()
	if cvcLabels == nil {
		cvcLabels = map[string]string{}
	}
	cvcLabels[apis.PodDisruptionBudgetKey] = pdbObj.Name
	cvcObj.SetLabels(cvcLabels)
}

// isHAVolume returns true if no.of replicas are greater than or equal to 3.
// If CVC doesn't hold any volume replica pool information then verify with
// ReplicaCount. If CVC holds any volume replica pool information then verify
// with Status.PoolInfo
func isHAVolume(cvcObj *apis.CStorVolumeClaim) bool {
	if len(cvcObj.Status.PoolInfo) == 0 {
		return cvcObj.Spec.ReplicaCount >= minHAReplicaCount
	}
	return len(cvcObj.Status.PoolInfo) >= minHAReplicaCount
}

// 1. If Volume was already pointing to a PDB then check is that same PDB will be
//    applicable after scalingup/scalingdown(case might be from 4 to 3
//    replicas) if applicable then return same pdb name. If not applicable do
//    following changes:
//    1.1 Delete PDB if no other CVC is pointing to PDB.
// 2. If current volume was not pointing to any PDB then do nothing.
// 3. If current volume is HAVolume then check is there any PDB already
//    existing among the current replica pools. If PDB exists then return
//    that PDB name. If PDB doesn't exist then create new PDB and return newely
//    created PDB name.
// 4. If current volume is not HAVolume then return nothing.
func updatePDBForVolume(cvcObj *apis.CStorVolumeClaim,
	cvObj *apis.CStorVolume) (string, error) {
	pdbName, hasPDB := cvcObj.GetLabels()[string(apis.PodDisruptionBudgetKey)]
	pdbLabels := cvclaim.GetPDBPoolLabels(cvcObj.Status.PoolInfo)
	labelSelector := apispdb.GetPDBLabelSelector(pdbLabels)
	if hasPDB {
		pdbList, err := apispdb.KubeClient().
			WithNamespace(openebsNamespace).
			List(metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			return "", errors.Wrapf(err,
				"failed to get PDB present among pools %v",
				cvcObj.Status.PoolInfo,
			)
		}
		if pdbList.Items[0].Name == pdbName {
			return pdbName, nil
		}
		err = deletePDBIfNotInUse(cvcObj)
		if err != nil {
			return "", err
		}
	}
	if !isHAVolume(cvcObj) {
		return "", nil
	}
	pdbObj, err := getOrCreatePodDisruptionBudget(cvObj,
		getCSPC(cvcObj), cvcObj.Status.PoolInfo)
	if err != nil {
		return "", err
	}
	return pdbObj.Name, nil
}

// isCVCScalePending returns true if there is change in desired replica pool
// names and current replica pool names
// 1. Below function will check whether there is any change in desired replica
//    pool names and current replica pool names.
func (c *CVCController) isCVCScalePending(cvc *apis.CStorVolumeClaim) bool {
	desiredPoolNames := cvclaim.GetDesiredReplicaPoolNames(cvc)
	return util.IsChangeInLists(desiredPoolNames, cvc.Status.PoolInfo)
}

// handlePostScalingProcess will does the following changes:
// 1. Handle PDB updation based on no.of volume replicas. It should handle in
//    following ways:
//    1.1 If Volume was already pointing to a PDB then check is that same PDB will be
//        applicable after scalingup/scalingdown(case might be from 4 to 3
//        replicas) if applicable then return same pdb name. If not applicable do
//        following changes:
//        1.1.1 Delete PDB if no other CVC is pointing to PDB.
//    1.2 If current volume was not pointing to any PDB then do nothing.
//    1.3 If current volume is HAVolume then check is there any PDB already
//        existing among the current replica pools. If PDB exists then return
//        that PDB name. If PDB doesn't exist then create new PDB and return newely
//        created PDB name.
// 2. Update CVC label to point it to newely PDB got from above step and also
//    replicas pool information on status of CVC.
func handlePostScalingProcess(cvc *apis.CStorVolumeClaim,
	cvObj *apis.CStorVolume, currentRPNames []string) error {
	var err error
	cvcCopy := cvc.DeepCopy()
	cvc.Status.PoolInfo = []string{}
	cvc.Status.PoolInfo = append(cvc.Status.PoolInfo, currentRPNames...)
	pdbName, err := updatePDBForVolume(cvc, cvObj)
	if err != nil {
		return errors.Wrapf(err,
			"failed to handle PDB for scaled volume %s",
			cvc.Name,
		)
	}
	delete(cvc.Labels, string(apis.PodDisruptionBudgetKey))
	if pdbName != "" {
		cvc.Labels[string(apis.PodDisruptionBudgetKey)] = pdbName
	}
	cvc, err = cvclaim.NewKubeclient().WithNamespace(cvc.Namespace).Update(cvc)
	if err != nil {
		// If error occured point it to old cvc object it self
		cvc = cvcCopy
		return errors.Wrapf(err,
			"failed to update %s CVC status with scaledup replica pool names",
			cvc.Name,
		)
	}
	return nil
}

// verifyAndUpdateScaleUpInfo does the following changes:
// 1. Get list of new replica pool names by using CVC(spec and status)
// 2. Verify status of ScalingUp Replica(by using CV object) based on the status
//    does following changes:
//    2.1: If scalingUp was completed then update PDB accordingly(only if it was
//         HAVolume) and update the replica pool info on CVC(API calls).
//    2.2: If scalingUp was going then return error saying scalingUp was in
//      progress.
func verifyAndUpdateScaleUpInfo(cvc *apis.CStorVolumeClaim, cvObj *apis.CStorVolume) error {
	// scaledRPNames contains the new replica pool names where entier data was
	// reconstructed from other replicas
	scaledRPNames := []string{}
	pvName := cvc.GetAnnotations()[volumeID]
	desiredPoolNames := cvclaim.GetDesiredReplicaPoolNames(cvc)
	newPoolNames := util.ListDiff(desiredPoolNames, cvc.Status.PoolInfo)
	for _, poolName := range newPoolNames {
		cvrName := pvName + "-" + poolName
		cvrObj, err := cvr.NewKubeclient().
			WithNamespace(getNamespace()).
			Get(cvrName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("failed to get CVR %s error: %v", cvrName, err)
			continue
		}
		_, isIDExists := cvObj.Status.ReplicaDetails.KnownReplicas[apis.ReplicaID(cvrObj.Spec.ReplicaID)]
		// ScalingUp was completed only if CVR replicaID exists on CV status
		// and also CVR should be Healthy(there might be cases of replica
		// migration in that case replicaID will be same zvol guid will be
		// different)
		if isIDExists && cvrObj.Status.Phase == apis.CVRStatusOnline {
			scaledRPNames = append(scaledRPNames, poolName)
		}
	}
	if len(scaledRPNames) > 0 {
		var currentRPNames []string
		currentRPNames = append(currentRPNames, cvc.Status.PoolInfo...)
		currentRPNames = append(currentRPNames, scaledRPNames...)
		// handlePostScalingProcess will handle PDB and CVC status
		err := handlePostScalingProcess(cvc, cvObj, currentRPNames)
		if err != nil {
			return errors.Wrapf(
				err,
				"failed to handle post volume replicas scale up process",
			)
		}
		return nil
	}
	return errors.Errorf(
		"scaling replicas from %d to %d in progress",
		len(cvc.Status.PoolInfo),
		len(cvc.Spec.Policy.ReplicaPool.PoolInfo),
	)
}

func getScaleDownCVR(cvc *apis.CStorVolumeClaim) (*apis.CStorVolumeReplica, error) {
	pvName := cvc.GetAnnotations()[volumeID]
	desiredPoolNames := cvclaim.GetDesiredReplicaPoolNames(cvc)
	removedPoolNames := util.ListDiff(cvc.Status.PoolInfo, desiredPoolNames)
	cvrName := pvName + removedPoolNames[0]
	return cvr.NewKubeclient().
		WithNamespace(getNamespace()).
		Get(cvrName, metav1.GetOptions{})
}

// handleVolumeReplicaCreation does the following changes:
// 1. Get the list of new pool names(i.e poolNames which are in spec but not in
//    status of CVC).
// 2. Creates new CVR on new pools only if CVR on that pool doesn't exists. If
//    CVR already created then do nothing.
func handleVolumeReplicaCreation(cvc *apis.CStorVolumeClaim, cvObj *apis.CStorVolume) error {
	pvName := cvc.GetAnnotations()[volumeID]
	desiredPoolNames := cvclaim.GetDesiredReplicaPoolNames(cvc)
	newPoolNames := util.ListDiff(desiredPoolNames, cvc.Status.PoolInfo)
	errs := []error{}
	var errorMsg string

	svcObj, err := svc.NewKubeClient(svc.WithNamespace(openebsNamespace)).
		Get(cvc.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get service object %s", cvc.Name)
	}

	cvrApiList, err := cvr.NewKubeclient().
		WithNamespace(getNamespace()).
		List(metav1.ListOptions{LabelSelector: pvSelector + "=" + pvName})
	if err != nil {
		return errors.Wrapf(err, "failed to list cstorvolumereplicas of volume %s", pvName)
	}
	cvrListbuilder := cvr.NewListBuilder().
		WithAPIList(cvrApiList)

	for _, poolName := range newPoolNames {
		if cvrListbuilder.
			WithFilter(cvr.HasLabel(cspiLabel, poolName)).
			List().Len() == 0 {
			cspiObj, err := cspi.NewKubeClient().
				WithNamespace(getNamespace()).
				Get(poolName, metav1.GetOptions{})
			if err != nil {
				errorMsg = fmt.Sprintf("failed to get cstorpoolinstance %s error: %v", poolName, err)
				errs = append(errs, errors.Errorf("%v", errorMsg))
				klog.Errorf("%s", errorMsg)
				continue
			}
			if cspiObj.Status.Phase != cspiOnline {
				errorMsg = fmt.Sprintf(
					"failed to create cstorvolumerplica on pool %s error: pool is not in %s",
					cspiObj.Name,
					cspiOnline,
				)
				errs = append(errs, errors.Errorf("%v", errorMsg))
				klog.Errorf("%s", errorMsg)
				continue
			}
			hash, err := hash.Hash(pvName + "-" + poolName)
			if err != nil {
				errorMsg = fmt.Sprintf(
					"failed to calculate of hase for new volume replica error: %v",
					err)
				errs = append(errs, errors.Errorf("%v", errorMsg))
				klog.Errorf("%s", errorMsg)
				continue
			}
			// TODO: Add a check for ClonedVolumeReplica scaleup case
			// Create replica with Recreate state
			rInfo := replicaInfo{
				replicaID: hash,
				phase:     apis.CVRStatusRecreate,
			}
			cvr, err := createCVR(svcObj, cvObj, cvc, cspiObj, rInfo)
			if err != nil {
				errorMsg = fmt.Sprintf(
					"failed to create new replica on pool %s error: %v",
					poolName,
					err,
				)
				errs = append(errs, errors.Errorf("%v", errorMsg))
				klog.Errorf("%s", errorMsg)
				continue
			}
			// Update cvrListbuilder with new replicas
			cvrListbuilder = cvrListbuilder.AppendListBuilder(cvr)
		}
	}
	if len(errs) > 0 {
		return errors.Errorf("%+v", errs)
	}
	return nil
}

// scaleUpVolumeReplicas does the following work
// 1. Fetch corresponding CStorVolume object of CVC.
// 2. Verify is there need to update desiredReplicationFactor of CVC(In etcd).
// 3. Create CVRs if doesn't created on scaled cStor
//    pool(handleVolumeReplicaCreation will handle new CVR creations).
// 4. If scalingUp volume replicas was completed then do following
//    things(verifyAndUpdateScaleUpInfo will does following things). If
//    scalingUp of volume replicas was not completed then return error
//    4.1.1 Update PDB according to the new pools(only if volume is HAVolume).
//    4.1.2 Update PDB label on CVC and replica pool information on status.
func scaleUpVolumeReplicas(cvc *apis.CStorVolumeClaim) error {
	drCount := len(cvc.Spec.Policy.ReplicaPool.PoolInfo)
	cvObj, err := cv.NewKubeclient().
		WithNamespace(getNamespace()).
		Get(cvc.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get cstorvolumes object %s", cvc.Name)
	}
	if cvObj.Spec.DesiredReplicationFactor < drCount {
		cvObj.Spec.DesiredReplicationFactor = drCount
		cvObj, err = updateCStorVolumeInfo(cvObj)
		if err != nil {
			return err
		}
	}
	err = handleVolumeReplicaCreation(cvc, cvObj)
	if err != nil {
		return err
	}
	err = verifyAndUpdateScaleUpInfo(cvc, cvObj)
	return err
}

// scaleDownVolumeReplicas will process the following steps
// 1. Verify whether operation made by user is valid for scale down
//    process(Only one replica scaledown at a time is allowed).
// 2. Update the CV object by decreasing the DRF and removing the
//    replicaID entry.
// 3. Check the status of scale down if scale down was completed then
//    perform post scaling process(updating PDB if applicable and CVC
//    replica pool status).
func scaleDownVolumeReplicas(cvc *apis.CStorVolumeClaim) error {
	drCount := len(cvc.Spec.Policy.ReplicaPool.PoolInfo)
	cvObj, err := cv.NewKubeclient().
		WithNamespace(getNamespace()).
		Get(cvc.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get cstorvolumes object %s", cvc.Name)
	}
	// If more than one replica was scale down at a time keep on return the error
	if (cvObj.Spec.ReplicationFactor - drCount) > 1 {
		return errors.Wrapf(err,
			"cann't perform %d replicas scaledown at a time",
			(cvObj.Spec.DesiredReplicationFactor - drCount),
		)
	}
	if cvObj.Spec.DesiredReplicationFactor > drCount {
		cvrObj, err := getScaleDownCVR(cvc)
		if err != nil {
			return errors.Wrapf(err, "failed to get scale down CVR object")
		}
		cvObj.Spec.DesiredReplicationFactor = drCount
		delete(cvObj.Spec.ReplicaDetails.KnownReplicas, apis.ReplicaID(cvrObj.Spec.ReplicaID))
		cvObj, err = updateCStorVolumeInfo(cvObj)
		if err != nil {
			return err
		}
	}
	if !cv.IsScaleDownInProgress(cvObj) {
		desiredPoolNames := cvclaim.GetDesiredReplicaPoolNames(cvc)
		err = handlePostScalingProcess(cvc, cvObj, desiredPoolNames)
		if err != nil {
			return errors.Wrapf(err,
				"failed to handle post volume replicas scale down process")
		}
		return nil
	}
	return errors.Errorf(
		"Scaling down volume replicas from %d to %d is in progress",
		len(cvc.Status.PoolInfo),
		drCount,
	)
}

// UpdateCStorVolumeInfo modifies the CV Object in etcd by making update API call
// Note: Caller code should handle the error
func updateCStorVolumeInfo(cvObj *apis.CStorVolume) (*apis.CStorVolume, error) {
	return cv.NewKubeclient().
		WithNamespace(getNamespace()).
		Update(cvObj)
}
