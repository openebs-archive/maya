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
	"math/rand"
	"strings"
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/openebs/maya/pkg/version"

	cspi "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	cv "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
)

const (
	cvcKind = "CStorVolumeClaim"
	cvKind  = "CStorVolume"

	cstorpoolInstanceLabel = "cstorpoolinstance.openebs.io/name"
	// ReplicaCount represents replica count value
	ReplicaCount = "replicaCount"
	// pvSelector is the selector key for cstorvolumereplica belongs to a cstor
	// volume
	pvSelector = "openebs.io/persistent-volume"
)

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
// downward API where maya-apiserver has been deployed
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

	svcObj, err := svc.NewKubeClient(svc.WithNamespace(getNamespace())).
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

	svcObj, err = svc.NewKubeClient(svc.WithNamespace(getNamespace())).Create(svcObj)
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

	cvObj, err := cv.NewKubeclient(cv.WithNamespace(getNamespace())).
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
		return cv.NewKubeclient(cv.WithNamespace(getNamespace())).Create(cvObj)
	}
	return cvObj, err
}

// distributeCVRs create cstorvolume replica based on the replicaCount
// on the available cstor pools created for storagepoolclaim.
// if pools are less then desired replicaCount its return an error.
func distributeCVRs(
	pendingReplicaCount int,
	claim *apis.CStorVolumeClaim,
	service *corev1.Service,
	volume *apis.CStorVolume,
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
	for count, pool := range usablePoolList.Items {
		pool := pool
		if count < pendingReplicaCount {
			_, err = createCVR(service, volume, claim, &pool)
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
	cvrObj, err := cvr.NewKubeclient(cvr.WithNamespace(getNamespace())).
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
			WithCapacity(volume.Spec.Capacity.String()).
			WithNewVersion(version.GetVersion()).
			WithDependentsUpgraded().
			Build()
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to build cstorvolumereplica {%v}",
				cvrObj.Name,
			)
		}
		cvrObj, err = cvr.NewKubeclient(cvr.WithNamespace(getNamespace())).Create(cvrObj)
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
		poolName := cvr.GetLabels()[string(cstorpoolInstanceLabel)]
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
	cvrList, err := cvr.NewKubeclient(cvr.WithNamespace(getNamespace())).List(metav1.ListOptions{
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
	cvrList, err := cvr.NewKubeclient(cvr.WithNamespace(getNamespace())).List(metav1.ListOptions{
		LabelSelector: pvLabel,
	})
	if err != nil {
		return nil
	}
	srcPVLabel := pvSelector + "=" + srcPVName
	srcCVRList, err := cvr.NewKubeclient(cvr.WithNamespace(getNamespace())).List(metav1.ListOptions{
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

// getCVRList returns list of volume replicas related to provided volume
func getCVRList(cvObj *apis.CStorVolume) (*apis.CStorVolumeReplicaList, error) {
	pvName := cvObj.Labels[string(apis.PersistentVolumeCPK)]
	pvLabel := string(apis.PersistentVolumeCPK) + "=" + pvName
	return cvr.NewKubeclient(cvr.WithNamespace(getNamespace())).
		List(metav1.ListOptions{
			LabelSelector: pvLabel,
		})
}

// getPoolNames returns list of quorum replica pool names from cStorVolume and
// cStor volume replcia objects
func getPoolNames(cvObj *apis.CStorVolume,
	cvrList *apis.CStorVolumeReplicaList) []string {
	poolNames := []string{}
	for _, cvrObj := range cvrList.Items {
		replicaID := cvrObj.Spec.ReplicaID
		_, ok := cvObj.Status.ReplicaDetails.KnownReplicas[apis.ReplicaID(replicaID)]
		if ok {
			poolNames = append(poolNames, cvrObj.Labels[cstorpoolInstanceLabel])
		}
	}
	return poolNames
}

// getReplicaPoolNames return list of quorum replicas pool names by taking cStor
// volume as a argument and return error(if any error occured)
func getReplicaPoolNames(cvObj *apis.CStorVolume) ([]string, error) {
	pvName := cvObj.Labels[string(apis.PersistentVolumeCPK)]
	minAvailable := (cvObj.Spec.ReplicationFactor >> 1) + 1
	if len(cvObj.Status.ReplicaDetails.KnownReplicas) < minAvailable {
		return []string{}, errors.Errorf(
			"required no.of replicas are not yet connected to the target of volume %s",
			pvName,
		)
	}
	cvrList, err := getCVRList(cvObj)
	if err != nil {
		return []string{}, errors.Wrapf(err,
			"failed to list cStorVolumeReplicas related to volume %s",
			pvName)
	}
	return getPoolNames(cvObj, cvrList), nil
}

// getCStorVolumeFromCVC returns cStor Volume object by taking cstor volume
// claim as argument
func getCStorVolumeFromCVC(cvc *apis.CStorVolumeClaim) (*apis.CStorVolume, error) {
	return cv.NewKubeclient().
		WithNamespace(cvc.Namespace).
		Get(cvc.Name, metav1.GetOptions{})
}

// getPDBOwnerReference returns PDB ownerReference by building from cStor Volume
func getPDBOwnerReference(cv *apis.CStorVolume) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(cv,
			apis.SchemeGroupVersion.WithKind(cvKind)),
	}
}

// getPDBLabels return pod disruption budget required labels
func getPDBLabels(cvObj *apis.CStorVolume) map[string]string {
	pvName := cvObj.Labels[string(apis.PersistentVolumeCPK)]
	return map[string]string{
		string(apis.PersistentVolumeCPK): pvName,
	}
}

func getPDBMatchExpressions(
	cvObj *apis.CStorVolume) ([]metav1.LabelSelectorRequirement, error) {
	selectorRequirements := []metav1.LabelSelectorRequirement{}
	pools, err := getReplicaPoolNames(cvObj)
	if err != nil {
		return selectorRequirements, errors.Wrapf(err,
			"failed to get volume replica pool names")
	}
	selectorRequirements = append(
		selectorRequirements,
		metav1.LabelSelectorRequirement{
			Key:      string(apis.CStorPoolInstanceCPK),
			Operator: metav1.LabelSelectorOpIn,
			Values:   pools,
		})
	return selectorRequirements, nil
}

func getPDBSelector(cvObj *apis.CStorVolume) (*metav1.LabelSelector, error) {
	matchExpressions, err := getPDBMatchExpressions(cvObj)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to get matchExpressions for volume %s",
			cvObj.Name)
	}
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "cstor-pool",
		},
		MatchExpressions: matchExpressions,
	}, nil
}

func isReplicaPoolsModified(cvObj *apis.CStorVolume, existingPoolNames []string) bool {
	currentPoolNames, err := getReplicaPoolNames(cvObj)
	if err != nil {
		klog.Errorf("failed to get replicas pool names of volume: %s error: %v", cvObj.Name, err)
		return false
	}
	if len(currentPoolNames) != len(existingPoolNames) {
		return true
	}
	mapExistingPools := map[string]bool{}
	for _, poolName := range existingPoolNames {
		mapExistingPools[poolName] = true
	}
	for _, poolName := range currentPoolNames {
		if !mapExistingPools[poolName] {
			return true
		}
	}
	return false
}
