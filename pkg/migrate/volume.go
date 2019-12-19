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
	"time"

	errors "github.com/pkg/errors"

	pv "github.com/openebs/maya/pkg/kubernetes/persistentvolume/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsVolumeMounted checks if the volume is mounted into any pod.
// This check is required as if mounted the pod will not allow
// deleting the pvc for recreation into csi volume.
func IsVolumeMounted(pvName string) (*corev1.PersistentVolume, error) {
	pvObj, err := pv.NewKubeClient().Get(pvName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	pvcName := pvObj.Spec.ClaimRef.Name
	pvcNamespace := pvObj.Spec.ClaimRef.Namespace
	podList, err := pod.NewKubeClient().WithNamespace(pvcNamespace).
		List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, podObj := range podList.Items {
		for _, volume := range podObj.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				if volume.PersistentVolumeClaim.ClaimName == pvcName {
					return nil, errors.Errorf(
						"the volume %s is mounted on %s, please scale down all apps before migrating",
						pvName,
						podObj.Name,
					)
				}
			}
		}
	}
	return pvObj, nil
}

// RetainPV sets the Retain policy on the PV.
// This operation is performed to prevent deletion of the OpenEBS
// resources while deleting the pvc to recreate with migrated spec.
func RetainPV(pvObj *corev1.PersistentVolume) error {
	pvObj.Spec.PersistentVolumeReclaimPolicy = corev1.PersistentVolumeReclaimRetain
	_, err := pv.NewKubeClient().Update(pvObj)
	if err != nil {
		return err
	}
	return nil
}

// RecreatePV recreates PV for the given PV object by first deleting
// the old PV with same name and creating a new PV having claimRef same
// as previous PV except for the uid to avoid any other PVC to claim it.
func RecreatePV(pvObj *corev1.PersistentVolume) (*corev1.PersistentVolume, error) {
	err := pv.NewKubeClient().Delete(pvObj.Name, &metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, err
	}
	err = isPVDeletedEventually(pvObj)
	if err != nil {
		return nil, err
	}
	pvObj, err = pv.NewKubeClient().Create(pvObj)
	if err != nil {
		return nil, err
	}
	return pvObj, nil
}

// RecreatePVC recreates PVC for the given PVC object by first deleting
// the old PVC with same name and creating a new PVC.
func RecreatePVC(pvcObj *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	err := pvc.NewKubeClient().WithNamespace(pvcObj.Namespace).
		Delete(pvcObj.Name, &metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, err
	}
	err = isPVCDeletedEventually(pvcObj)
	if err != nil {
		return nil, err
	}
	pvcObj, err = pvc.NewKubeClient().WithNamespace(pvcObj.Namespace).
		Create(pvcObj)
	if err != nil {
		return nil, err
	}
	return pvcObj, nil
}

// IsPVCDeletedEventually tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func isPVCDeletedEventually(pvcObj *corev1.PersistentVolumeClaim) error {
	for i := 1; i < 60; i++ {
		_, err := pvc.NewKubeClient().
			WithNamespace(pvcObj.Namespace).Get(pvcObj.Name, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return errors.Errorf("PVC %s still present", pvcObj.Name)
}

// IsPVDeletedEventually tries to get the deleted pv
// and returns true if pv is not found
// else returns false
func isPVDeletedEventually(pvObj *corev1.PersistentVolume) error {
	for i := 1; i < 60; i++ {
		_, err := pv.NewKubeClient().
			Get(pvObj.Name, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return errors.Errorf("PVC %s still present", pvObj.Name)
}
