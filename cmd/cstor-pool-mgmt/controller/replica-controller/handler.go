/*
Copyright 2018 The OpenEBS Authors.

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

package replicacontroller

import (
	"fmt"
	"os"
	"reflect"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/volumereplica"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the CStorReplicaUpdated resource
// with the current status of the resource.
func (c *CStorVolumeReplicaController) syncHandler(key, operation string) error {
	cVRGot, err := c.getVolumeReplicaResource(key, operation)
	if err != nil {
		return err
	}
	status, err := c.cVREventHandler(operation, cVRGot)
	if status == "" {
		return nil
	}
	cVRGot.Status.Phase = apis.CStorVolumeReplicaPhase(status)
	if err != nil {
		glog.Errorf(err.Error())
		glog.Infof("cVR:%v, %v; Status: %v", cVRGot.Name,
			string(cVRGot.GetUID()), cVRGot.Status.Phase)
		_, err := c.clientset.OpenebsV1alpha1().CStorVolumeReplicas().Update(cVRGot)
		if err != nil {
			return err
		}
		return err
	}
	_, err = c.clientset.OpenebsV1alpha1().CStorVolumeReplicas().Update(cVRGot)
	if err != nil {
		return err
	}
	glog.Infof("cVR:%v, %v; Status: %v", cVRGot.Name,
		string(cVRGot.GetUID()), cVRGot.Status.Phase)
	return nil

}

func (c *CStorVolumeReplicaController) cVREventHandler(operation string, cVR *apis.CStorVolumeReplica) (string, error) {
	err := volumereplica.CheckValidVolumeReplica(cVR)
	if err != nil {
		c.recorder.Event(cVR, corev1.EventTypeWarning, common.FailureValidate, common.MessageResourceFailValidate)
		return string(apis.CVRStatusOffline), err
	}

	// PoolNameHandler tries to get pool name and blocks for
	// particular number of attempts.
	var noOfAttempts = 5
	if !common.PoolNameHandler(cVR, noOfAttempts) {
		return string(apis.CVRStatusOffline), fmt.Errorf("Pool not present")
	}
	if err != nil {
		return string(apis.CVRStatusOffline), err
	}

	fullVolName := "cstor-" + cVR.Labels["cstorpool.openebs.io/uid"] + "/" + cVR.Labels["cstorvolume.openebs.io/uid"]
	glog.Infof("fullVolName: %v", fullVolName)

	switch operation {
	case "add":
		glog.Infof("Processing cvr added event: %v, %v", cVR.ObjectMeta.Name, string(cVR.GetUID()))

		// To check if volume is already imported with pool.
		importedFlag := common.CheckForInitialImportedPoolVol(common.InitialImportedPoolVol, fullVolName)
		if importedFlag && !IsInitStatus(cVR) {
			glog.Infof("CStorVolumeReplica %v is already imported", string(cVR.ObjectMeta.UID))
			c.recorder.Event(cVR, corev1.EventTypeNormal, common.SuccessImported, common.MessageResourceImported)
			return string(apis.CVRStatusOnline), nil
		}

		// IsInitStatus is to check if initial status of cVR object is `init`.
		if IsInitStatus(cVR) {
			err = volumereplica.CreateVolume(cVR, fullVolName)
			if err != nil {
				glog.Errorf("cVr creation failure: %v", err.Error())
				return string(apis.CVRStatusOffline), err
			}
			c.recorder.Event(cVR, corev1.EventTypeNormal, common.SuccessCreated, common.MessageResourceCreated)
			glog.Infof("cVR creation successful: %v, %v", cVR.ObjectMeta.Name, string(cVR.GetUID()))
			return string(apis.CVRStatusOnline), nil
		}

		break

	case "destroy":
		glog.Infof("Processing cvr deleted event %v, %v", cVR.ObjectMeta.Name, string(cVR.GetUID()))

		err := volumereplica.DeleteVolume(fullVolName)
		if err != nil {
			c.recorder.Event(cVR, corev1.EventTypeWarning, common.FailureDestroy, common.MessageResourceFailDestroy)
			return string(apis.CVRStatusDeletionFailed), err
		}
		// removeFinalizer is to remove finalizer of cVR resource.
		err = c.removeFinalizer(cVR)
		if err != nil {
			return string(apis.CVRStatusOffline), err
		}
		return "", nil
	}
	return string(apis.CVRStatusInvalid), nil
}

// getVolumeReplicaResource returns object corresponding to the resource key
func (c *CStorVolumeReplicaController) getVolumeReplicaResource(key, operation string) (*apis.CStorVolumeReplica, error) {
	// Convert the key(namespace/name) string into a distinct name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil, nil
	}

	cStorVolumeReplicaUpdated, err := c.clientset.OpenebsV1alpha1().CStorVolumeReplicas().Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorPool resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("cStorVolumeReplicaUpdated '%s' in work queue no longer exists", key))
			return nil, nil
		}
		return nil, err
	}
	return cStorVolumeReplicaUpdated, nil
}

// removeFinalizer is to remove finalizer of CStorVolumeReplica resource.
func (c *CStorVolumeReplicaController) removeFinalizer(cVR *apis.CStorVolumeReplica) error {
	if len(cVR.Finalizers) > 0 {
		cVR.Finalizers = []string{}
	}
	_, err := c.clientset.OpenebsV1alpha1().CStorVolumeReplicas().Update(cVR)
	if err != nil {
		return err
	}
	glog.Infof("Removed Finalizer: %v, %v", cVR.ObjectMeta.Name, string(cVR.GetUID()))
	return nil
}

// IsRightCStorVolumeReplica is to check if the cvr request is for particular pod/application.
func IsRightCStorVolumeReplica(cVR *apis.CStorVolumeReplica) bool {
	if os.Getenv("cstorid") == string(cVR.ObjectMeta.Labels["cstorpool.openebs.io/uid"]) {
		return true
	}
	return false
}

// IsDestroyEvent is to check if the call is for CStorVolumeReplica destroy.
func IsDestroyEvent(cVR *apis.CStorVolumeReplica) bool {
	if cVR.ObjectMeta.DeletionTimestamp != nil {
		return true
	}
	return false
}

// IsOnlyStatusChange is to check only status change of cStorVolumeReplica object.
func IsOnlyStatusChange(oldCVR, newCVR *apis.CStorVolumeReplica) bool {
	if reflect.DeepEqual(oldCVR.Spec, newCVR.Spec) &&
		!reflect.DeepEqual(oldCVR.Status, newCVR.Status) {
		return true
	}
	return false
}

// IsInitStatus is to check if the status of cStorVolumeReplica object is `init`.
func IsInitStatus(cVR *apis.CStorVolumeReplica) bool {
	if string(cVR.Status.Phase) == string(apis.CVRStatusInit) {
		glog.Infof("cVR init: %v, %v", cVR.ObjectMeta.Name, string(cVR.ObjectMeta.UID))
		return true
	}
	return false
}

// IsDeletionFailedBefore is to make sure no other operation should happen if the
// status of CStorVolumeReplica is deletion-failed.
func IsDeletionFailedBefore(cVR *apis.CStorVolumeReplica) bool {
	if cVR.Status.Phase == "deletion-failed" {
		return true
	}
	return false
}
