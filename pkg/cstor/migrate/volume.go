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

package migrate

import (
	"strconv"
	"strings"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cv "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	cvc "github.com/openebs/maya/pkg/cstorvolumeclaim/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	pv "github.com/openebs/maya/pkg/kubernetes/persistentvolume/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	"github.com/openebs/maya/pkg/migrate"
	"github.com/openebs/maya/pkg/util"
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

var (
	trueBool         = true
	openebsNamespace = "openebs"
	cvNamespace      = "openebs"
	cstorCSIDriver   = "cstor.csi.openebs.io"
	cvcKind          = "CStorVolumeClaim"
	cvKind           = "CStorVolume"
	storageClass     *storagev1.StorageClass
)

// Volume migrates the volume from non-CSI schema to CSI schema
func Volume(pvName, openebsNS string) error {
	var pvObj *corev1.PersistentVolume
	openebsNamespace = openebsNS
	pvcObj, pvPresent, err := validatePVName(pvName)
	if err != nil {
		return errors.Wrapf(err, "failed to validate pvname")
	}
	err = populateCVNamespace(pvName)
	if err != nil {
		return errors.Wrapf(err, "failed to cv namespace")
	}
	if pvPresent {
		klog.Infof("Checking volume is not mounted on any application")
		pvObj, err = migrate.IsVolumeMounted(pvName)
		if err != nil {
			return errors.Wrapf(err, "failed to verify mount status for pv {%s}", pvName)
		}
		if pvObj.Spec.PersistentVolumeSource.CSI == nil {
			klog.Infof("Retaining PV to migrate into csi volume")
			err = migrate.RetainPV(pvObj)
			if err != nil {
				return errors.Wrapf(err, "failed to retain pv {%s}", pvName)
			}
		}
		err = updateStorageClass(pvObj.Name, pvObj.Spec.StorageClassName)
		if err != nil {
			return errors.Wrapf(err, "failed to update storageclass {%s}", pvObj.Spec.StorageClassName)
		}
		pvcObj, err = migratePVC(pvObj)
		if err != nil {
			return err
		}
	} else {
		klog.Infof("PVC and storageclass already migrated to csi format")
	}
	storageClass, err = sc.NewKubeClient().Get(*pvcObj.Spec.StorageClassName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, err = migratePV(pvcObj)
	if err != nil {
		return err
	}
	klog.Infof("Creating CVC to bound the volume and trigger CSI driver")
	err = createCVC(pvName)
	if err != nil {
		return err
	}
	return nil
}

func migratePVC(pvObj *corev1.PersistentVolume) (*corev1.PersistentVolumeClaim, error) {
	klog.Infof("Generating equivalent CSI PVC")
	pvcObj, recreateRequired, err := generateCSIPVC(pvObj.Name)
	if err != nil {
		return nil, err
	}
	if recreateRequired {
		klog.Infof("Recreating equivalent CSI PVC")
		pvcObj, err = migrate.RecreatePVC(pvcObj)
		if err != nil {
			return nil, err
		}
	}
	return pvcObj, nil
}

func migratePV(pvcObj *corev1.PersistentVolumeClaim) (*corev1.PersistentVolume, error) {
	klog.Infof("Generating equivalent CSI PV")
	pvObj, recreateRequired, err := generateCSIPV(pvcObj.Spec.VolumeName, pvcObj)
	if err != nil {
		return nil, err
	}
	if recreateRequired {
		klog.Infof("Recreating equivalent CSI PV")
		_, err = migrate.RecreatePV(pvObj)
		if err != nil {
			return nil, err
		}
	}
	return pvObj, nil
}

func generateCSIPVC(pvName string) (*corev1.PersistentVolumeClaim, bool, error) {
	pvObj, err := pv.NewKubeClient().
		Get(pvName, metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}
	pvcName := pvObj.Spec.ClaimRef.Name
	pvcNamespace := pvObj.Spec.ClaimRef.Namespace
	pvcObj, err := pvc.NewKubeClient().WithNamespace(pvcNamespace).
		Get(pvcName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, false, err
		}
	}
	if pvcObj.Annotations["volume.beta.kubernetes.io/storage-provisioner"] != cstorCSIDriver {
		csiPVC := &corev1.PersistentVolumeClaim{}
		csiPVC.Name = pvcName
		csiPVC.Namespace = pvcNamespace
		csiPVC.Annotations = map[string]string{
			"volume.beta.kubernetes.io/storage-provisioner": cstorCSIDriver,
		}
		csiPVC.Spec.AccessModes = pvObj.Spec.AccessModes
		csiPVC.Spec.Resources.Requests = pvObj.Spec.Capacity
		csiPVC.Spec.StorageClassName = &pvObj.Spec.StorageClassName
		csiPVC.Spec.VolumeMode = pvObj.Spec.VolumeMode
		csiPVC.Spec.VolumeName = pvObj.Name

		return csiPVC, true, nil
	}
	klog.Infof("pvc already migrated")
	return pvcObj, false, nil
}

func generateCSIPV(
	pvName string,
	pvcObj *corev1.PersistentVolumeClaim,
) (*corev1.PersistentVolume, bool, error) {
	pvObj, err := pv.NewKubeClient().
		Get(pvName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, false, err
		}
		if k8serrors.IsNotFound(err) {
			pvObj, err = generateCSIPVFromCV(pvName, pvcObj)
			if err != nil {
				return nil, false, err
			}
			return pvObj, true, nil
		}
	}
	if pvObj.Spec.PersistentVolumeSource.CSI == nil {
		csiPV := &corev1.PersistentVolume{}
		csiPV.Name = pvObj.Name
		csiPV.Annotations = map[string]string{
			"pv.kubernetes.io/provisioned-by": cstorCSIDriver,
		}
		csiPV.Spec.AccessModes = pvObj.Spec.AccessModes
		csiPV.Spec.ClaimRef = &corev1.ObjectReference{
			APIVersion: pvcObj.APIVersion,
			Kind:       pvcObj.Kind,
			Name:       pvcObj.Name,
			Namespace:  pvcObj.Namespace,
		}
		csiPV.Spec.Capacity = pvObj.Spec.Capacity
		csiPV.Spec.PersistentVolumeSource = corev1.PersistentVolumeSource{
			CSI: &corev1.CSIPersistentVolumeSource{
				Driver:       cstorCSIDriver,
				FSType:       pvObj.Spec.PersistentVolumeSource.ISCSI.FSType,
				VolumeHandle: pvObj.Name,
				VolumeAttributes: map[string]string{
					"openebs.io/cas-type":                          "cstor",
					"storage.kubernetes.io/csiProvisionerIdentity": "1574675355213-8081-cstor.csi.openebs.io",
				},
			},
		}
		csiPV.Spec.PersistentVolumeReclaimPolicy = corev1.PersistentVolumeReclaimDelete
		csiPV.Spec.StorageClassName = pvObj.Spec.StorageClassName
		csiPV.Spec.VolumeMode = pvObj.Spec.VolumeMode
		return csiPV, true, nil
	}
	klog.Infof("PV %s already in csi form", pvObj.Name)
	return pvObj, false, nil
}

func generateCSIPVFromCV(
	cvName string,
	pvcObj *corev1.PersistentVolumeClaim,
) (*corev1.PersistentVolume, error) {
	cvObj, err := cv.NewKubeclient().WithNamespace(cvNamespace).
		Get(cvName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	csiPV := &corev1.PersistentVolume{}
	csiPV.Name = cvObj.Name
	csiPV.Spec.AccessModes = pvcObj.Spec.AccessModes
	csiPV.Spec.ClaimRef = &corev1.ObjectReference{
		APIVersion: pvcObj.APIVersion,
		Kind:       pvcObj.Kind,
		Name:       pvcObj.Name,
		Namespace:  pvcObj.Namespace,
	}
	csiPV.Spec.Capacity = corev1.ResourceList{
		corev1.ResourceStorage: cvObj.Spec.Capacity,
	}
	csiPV.Spec.PersistentVolumeSource = corev1.PersistentVolumeSource{
		CSI: &corev1.CSIPersistentVolumeSource{
			Driver:       cstorCSIDriver,
			FSType:       cvObj.Annotations["openebs.io/fs-type"],
			VolumeHandle: cvObj.Name,
			VolumeAttributes: map[string]string{
				"openebs.io/cas-type":                          "cstor",
				"storage.kubernetes.io/csiProvisionerIdentity": "1574675355213-8081-cstor.csi.openebs.io",
			},
		},
	}
	csiPV.Spec.PersistentVolumeReclaimPolicy = corev1.PersistentVolumeReclaimDelete
	csiPV.Spec.StorageClassName = storageClass.Name
	csiPV.Spec.VolumeMode = pvcObj.Spec.VolumeMode
	return csiPV, nil
}

func createCVC(pvName string) error {
	var (
		err    error
		cvcObj *apis.CStorVolumeClaim
		cvObj  *apis.CStorVolume
	)
	cvcObj, err = cvc.NewKubeclient().WithNamespace(openebsNamespace).
		Get(pvName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			cvObj, err = cv.NewKubeclient().WithNamespace(cvNamespace).
				Get(pvName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			annotations := map[string]string{
				"openebs.io/volumeID": pvName,
			}
			labels := map[string]string{
				"openebs.io/cstor-pool-cluster": storageClass.Parameters["cstorPoolCluster"],
			}
			finalizers := []string{"cvc.openebs.io/finalizer"}
			cvcObj, err = cvc.NewBuilder().
				WithName(cvObj.Name).
				WithNamespace(openebsNamespace).
				WithAnnotations(annotations).
				WithLabelsNew(labels).
				WithFinalizers(finalizers).
				WithCapacityQty(cvObj.Spec.Capacity).
				WithReplicaCount(storageClass.Parameters["replicaCount"]).
				WithStatusPhase(apis.CStorVolumeClaimPhasePending).
				Build()
			if err != nil {
				return err
			}
			_, err = cvc.NewKubeclient().WithNamespace(openebsNamespace).
				Create(cvcObj)
			if err != nil {
				return err
			}
		}
		return err
	}
	klog.Infof("Updating OwnerRefs")
	err = updateOwnerRefs(cvcObj)
	return err
}

// updateStorageClass recreates a new storageclass with the csi provisioner
// the older annotations with the casconfig are also preserved for information
// as the information about the storageclass cannot be gathered from other
// resources a temporary storageclass is created before deleting the original
func updateStorageClass(pvName, scName string) error {
	scObj, err := sc.NewKubeClient().Get(scName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}
	if scObj == nil || scObj.Provisioner != cstorCSIDriver {
		tmpSCObj, err := createTmpSC(scName)
		if err != nil {
			return err
		}
		replicaCount, err := getReplicaCount(pvName)
		if err != nil {
			return err
		}
		cspcName, err := getCSPCName(pvName)
		if err != nil {
			return err
		}
		csiSC := tmpSCObj.DeepCopy()
		csiSC.ObjectMeta = metav1.ObjectMeta{
			Name:        scName,
			Annotations: tmpSCObj.Annotations,
		}
		csiSC.Provisioner = cstorCSIDriver
		csiSC.AllowVolumeExpansion = &trueBool
		csiSC.Parameters = map[string]string{
			"cas-type":         "cstor",
			"replicaCount":     replicaCount,
			"cstorPoolCluster": cspcName,
		}
		if scObj != nil {
			err = sc.NewKubeClient().Delete(scName, &metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
		scObj, err = sc.NewKubeClient().Create(csiSC)
		if err != nil {
			return err
		}
		storageClass = scObj
		err = sc.NewKubeClient().Delete(tmpSCObj.Name, &metav1.DeleteOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to delete temporary storageclass")
		}
	}
	return nil
}

func createTmpSC(scName string) (*storagev1.StorageClass, error) {
	tmpSCName := "tmp-migrate-" + scName
	tmpSCObj, err := sc.NewKubeClient().Get(tmpSCName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}
		scObj, err := sc.NewKubeClient().Get(scName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		tmpSCObj = scObj.DeepCopy()
		tmpSCObj.ObjectMeta = metav1.ObjectMeta{
			Name:        tmpSCName,
			Annotations: scObj.Annotations,
		}
		tmpSCObj, err = sc.NewKubeClient().Create(tmpSCObj)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create temporary storageclass")
		}
	}
	return tmpSCObj, nil
}

func getReplicaCount(pvName string) (string, error) {
	cvObj, err := cv.NewKubeclient().WithNamespace(cvNamespace).
		Get(pvName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return strconv.Itoa(cvObj.Spec.ReplicationFactor), nil
}

// the cv can be in the pvc namespace or openebs namespace
func populateCVNamespace(cvName string) error {
	cvList, err := cv.NewKubeclient().WithNamespace("").
		List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, cvObj := range cvList.Items {
		if cvObj.Name == cvName {
			cvNamespace = cvObj.Namespace
			return nil
		}
	}
	return errors.Errorf("cv %s not found for given pv", cvName)
}

func getCSPCName(pvName string) (string, error) {
	cvrList, err := cvr.NewKubeclient().WithNamespace(openebsNamespace).
		List(metav1.ListOptions{
			LabelSelector: "openebs.io/persistent-volume=" + pvName,
		})
	if err != nil {
		return "", err
	}
	if len(cvrList.Items) == 0 {
		return "", errors.Errorf("no cvr found for pv %s", pvName)
	}
	cspiName := cvrList.Items[0].Labels["cstorpoolinstance.openebs.io/name"]
	if cspiName == "" {
		return "", errors.Errorf("no cspi label found on cvr %s", cvrList.Items[0].Name)
	}
	lastIndex := strings.LastIndex(cspiName, "-")
	return cspiName[:lastIndex], nil
}

func updateOwnerRefs(cvcObj *apis.CStorVolumeClaim) error {
	cvcOwnerRef := *metav1.NewControllerRef(cvcObj,
		apis.SchemeGroupVersion.WithKind(cvcKind))
	cvObj, err := updateCVOwnerRef(cvcOwnerRef)
	if err != nil {
		return err
	}
	err = updateTargetSVCOwnerRef(cvcOwnerRef)
	if err != nil {
		return err
	}
	cvOwnerRef := *metav1.NewControllerRef(cvObj,
		apis.SchemeGroupVersion.WithKind(cvKind))
	err = updateCVROwnerRef(cvOwnerRef)
	if err != nil {
		return err
	}
	err = updateTargetDeployOwnerRef(cvOwnerRef)
	return err
}

func updateCVOwnerRef(cvcOwnerRef metav1.OwnerReference) (*apis.CStorVolume, error) {
	cvObj, err := cv.NewKubeclient().WithNamespace(cvNamespace).
		Get(cvcOwnerRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	newCVObj := cvObj.DeepCopy()
	newCVObj.OwnerReferences = []metav1.OwnerReference{
		cvcOwnerRef,
	}
	patchBytes, err := util.GetPatchData(cvObj, newCVObj)
	if err != nil {
		return nil, err
	}
	cvObj, err = cv.NewKubeclient().WithNamespace(cvNamespace).
		Patch(cvObj.Name, cvObj.Namespace, types.MergePatchType, patchBytes)
	if err != nil {
		return nil, err
	}
	return cvObj, nil
}

func updateTargetSVCOwnerRef(cvcOwnerRef metav1.OwnerReference) error {
	svcObj, err := svc.NewKubeClient().WithNamespace(cvNamespace).
		Get(cvcOwnerRef.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	newSVCObj := svcObj.DeepCopy()
	newSVCObj.OwnerReferences = []metav1.OwnerReference{
		cvcOwnerRef,
	}
	patchBytes, err := util.GetPatchData(svcObj, newSVCObj)
	if err != nil {
		return err
	}
	_, err = svc.NewKubeClient().WithNamespace(cvNamespace).
		Patch(svcObj.Name, types.MergePatchType, patchBytes)
	return err
}

func updateCVROwnerRef(cvOwnerRef metav1.OwnerReference) error {
	cvrList, err := cvr.NewKubeclient().WithNamespace(openebsNamespace).
		List(metav1.ListOptions{
			LabelSelector: "openebs.io/persistent-volume=" + cvOwnerRef.Name,
		})
	if err != nil {
		return err
	}
	for _, cvrObj := range cvrList.Items {
		newCVRObj := cvrObj.DeepCopy()
		newCVRObj.OwnerReferences = []metav1.OwnerReference{
			cvOwnerRef,
		}
		patchBytes, err := util.GetPatchData(cvrObj, newCVRObj)
		if err != nil {
			return err
		}
		_, err = cvr.NewKubeclient().WithNamespace(openebsNamespace).
			Patch(cvrObj.Name, cvrObj.Namespace, types.MergePatchType, patchBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateTargetDeployOwnerRef(cvOwnerRef metav1.OwnerReference) error {
	targetName := cvOwnerRef.Name + "-target"
	targetObj, err := deploy.NewKubeClient().WithNamespace(cvNamespace).
		Get(targetName)
	if err != nil {
		return err
	}
	newTargetObj := targetObj.DeepCopy()
	newTargetObj.OwnerReferences = []metav1.OwnerReference{
		cvOwnerRef,
	}
	patchBytes, err := util.GetPatchData(targetObj, newTargetObj)
	if err != nil {
		return err
	}
	_, err = deploy.NewKubeClient().WithNamespace(cvNamespace).
		Patch(targetObj.Name, types.MergePatchType, patchBytes)
	return err
}

// validatePVName checks whether there exist any pvc for given pv name
// this is required in case the pv gets deleted and only pvc is left
func validatePVName(pvName string) (*corev1.PersistentVolumeClaim, bool, error) {
	var pvcObj *corev1.PersistentVolumeClaim
	_, err := pv.NewKubeClient().Get(pvName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return pvcObj, false, err
		}
		pvcList, err := pvc.NewKubeClient().WithNamespace("").
			List(metav1.ListOptions{})
		if err != nil {
			return pvcObj, false, err
		}
		for _, pvcItem := range pvcList.Items {
			pvcItem := pvcItem // pin it
			if pvcItem.Spec.VolumeName == pvName {
				pvcObj = &pvcItem
				return pvcObj, false, nil
			}
		}
		return pvcObj, false, errors.Errorf("No PVC found for the given PV")
	}
	return pvcObj, true, nil
}
