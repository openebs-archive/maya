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

package app

import (
	"fmt"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	"github.com/pkg/errors"
	"k8s.io/klog"

	zpool "github.com/openebs/maya/cmd/cstor-pool-mgmt/pool/v1alpha2"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// reconcile will ensure that pool for given
// key is created and running
func (c *CStorPoolInstanceController) reconcile(key string) error {
	var err error
	var isImported bool

	cspi, err := c.getCSPIObjFromKey(key)
	if err != nil || cspi == nil {
		return err
	}

	if IsReconcileDisabled(cspi) {
		c.recorder.Event(cspi,
			corev1.EventTypeWarning,
			fmt.Sprintf("reconcile is disabled via %q annotation", string(apis.OpenEBSDisableReconcileKey)),
			"Skipping reconcile")
		return nil
	}

	if IsDestroyed(cspi) {
		return c.destroy(cspi)
	}

	// take a lock for common package for updating variables
	common.SyncResources.Mux.Lock()

	// try to import pool
	if cspi.Annotations["cspuid"] != "" {
		isImported, err = zpool.Import(cspi, "cstor-"+cspi.Annotations["cspuid"])
	} else {
		isImported, err = zpool.Import(cspi, "")
	}
	if isImported {
		if err != nil {
			common.SyncResources.Mux.Unlock()
			c.recorder.Event(cspi,
				corev1.EventTypeWarning,
				string(common.FailureImported),
				fmt.Sprintf("Failed to import pool due to '%s'", err.Error()))
			return nil
		}
		zpool.CheckImportedPoolVolume()
		common.SyncResources.Mux.Unlock()
		delete(cspi.Annotations, "cspuid")
		err = c.update(cspi)
		if err != nil {
			c.recorder.Event(cspi,
				corev1.EventTypeWarning,
				string(common.FailedSynced),
				err.Error())
		}
		return nil
	}

	if IsEmptyStatus(cspi) || IsPendingStatus(cspi) {
		err = zpool.Create(cspi)
		if err != nil {
			// We will try to create it in next event
			c.recorder.Event(cspi,
				corev1.EventTypeWarning,
				string(common.FailureCreate),
				fmt.Sprintf("Failed to create pool due to '%s'", err.Error()))

			_ = zpool.Delete(cspi)
			common.SyncResources.Mux.Unlock()
			return nil
		}
		common.SyncResources.Mux.Unlock()

		c.recorder.Event(cspi,
			corev1.EventTypeNormal,
			string(common.SuccessCreated),
			fmt.Sprintf("Pool created successfully"))

		err = c.update(cspi)
		if err != nil {
			c.recorder.Event(cspi,
				corev1.EventTypeWarning,
				string(common.FailedSynced),
				err.Error())
		}
		return nil

	}
	common.SyncResources.Mux.Unlock()
	return nil
}

func (c *CStorPoolInstanceController) destroy(cspi *apis.CStorPoolInstance) error {
	var phase apis.CStorPoolPhase

	// DeletePool is to delete cstor zpool.
	// It will also clear the label for relevant disk
	err := zpool.Delete(cspi)
	if err != nil {
		c.recorder.Event(cspi,
			corev1.EventTypeWarning,
			string(common.FailureDestroy),
			fmt.Sprintf("Failed to delete pool due to '%s'", err.Error()))
		phase = apis.CStorPoolStatusDeletionFailed
		goto updatestatus
	}

	// removeFinalizer is to remove finalizer of cStorPoolInstance resource.
	err = c.removeFinalizer(cspi)
	if err != nil {
		// Object will exist. Let's set status as offline
		klog.Errorf("removeFinalizer failed %s", err.Error())
		phase = apis.CStorPoolStatusDeletionFailed
		goto updatestatus
	}
	klog.Infof("Pool %s deleted successfully", cspi.Name)
	return nil

updatestatus:
	cspi.Status.Phase = phase
	if _, er := zpool.OpenEBSClient.
		OpenebsV1alpha1().
		CStorPoolInstances(cspi.Namespace).
		Update(cspi); er != nil {
		klog.Errorf("Update failed %s", er.Error())
	}
	return err
}

func (c *CStorPoolInstanceController) update(cspi *apis.CStorPoolInstance) error {
	cspi, err := zpool.Update(cspi)
	if err != nil {
		return errors.Errorf("Failed to update pool due to %s", err.Error())
	}
	return c.updateStatus(cspi)
}

func (c *CStorPoolInstanceController) updateStatus(cspi *apis.CStorPoolInstance) error {
	var status apis.CStorPoolStatus
	var err error
	pool := zpool.PoolName(cspi)

	state, er := zpool.GetPropertyValue(pool, "health")
	if er != nil {
		err = zpool.ErrorWrapf(err, "Failed to fetch health")
	} else {
		status.Phase = apis.CStorPoolPhase(state)
	}

	freeSize, er := zpool.GetPropertyValue(pool, "free")
	if er != nil {
		err = zpool.ErrorWrapf(err, "Failed to fetch free size")
	} else {
		status.Capacity.Free = freeSize
	}

	usedSize, er := zpool.GetPropertyValue(pool, "allocated")
	if er != nil {
		err = zpool.ErrorWrapf(err, "Failed to fetch used size")
	} else {
		status.Capacity.Used = usedSize
	}

	totalSize, er := zpool.GetPropertyValue(pool, "size")
	if er != nil {
		err = zpool.ErrorWrapf(err, "Failed to fetch total size")
	} else {
		status.Capacity.Total = totalSize
	}

	if err != nil {
		return errors.Errorf("Failed to sync due to %s", err.Error())
	}

	if IsStatusChange(cspi.Status, status) {
		cspi.Status = status
		_, err = zpool.OpenEBSClient.
			OpenebsV1alpha1().
			CStorPoolInstances(cspi.Namespace).
			Update(cspi)
		if err != nil {
			return errors.Errorf("Failed to updateStatus due to '%s'", err.Error())
		}
	}
	return nil
}

// getCSPIObjFromKey returns object corresponding to the resource key
func (c *CStorPoolInstanceController) getCSPIObjFromKey(key string) (*apis.CStorPoolInstance, error) {
	// Convert the key(namespace/name) string into a distinct name
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil, nil
	}

	cspi, err := c.clientset.
		OpenebsV1alpha1().
		CStorPoolInstances(ns).
		Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorPoolInstance resource may no longer exist, in which case we stop
		// processing.
		if k8serror.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("CSPI '%s' in work queue no longer exists", key))
			return nil, nil
		}

		return nil, err
	}
	return cspi, nil
}

// removeFinalizer is to remove finalizer of cstorpoolinstance resource.
func (c *CStorPoolInstanceController) removeFinalizer(cspi *apis.CStorPoolInstance) error {
	if len(cspi.Finalizers) == 0 {
		return nil
	}
	cspi.Finalizers = []string{}
	_, err := c.clientset.
		OpenebsV1alpha1().
		CStorPoolInstances(cspi.Namespace).
		Update(cspi)
	if err != nil {
		return err
	}
	klog.Infof("Removed Finalizer: %v, %v",
		cspi.Name,
		string(cspi.GetUID()))
	return nil
}
