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

package poolcontroller

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"

	zpool "github.com/openebs/maya/cmd/cstor-pool-mgmt/pool/v1alpha2"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apis2 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// reconcile will ensure that pool for given
// key is created and running
func (c *CStorPoolController) reconcile(key string) error {
	var err error
	var isImported bool

	csp, err := c.getCSPObjFromKey(key)
	if err != nil || csp == nil {
		return err
	}

	if IsReconcileDisabled(csp) {
		c.recorder.Event(csp,
			corev1.EventTypeWarning,
			fmt.Sprintf("reconcile is disabled via %q annotation", string(apis.OpenEBSDisableReconcileKey)),
			"Skipping reconcile")
		return nil
	}

	if IsDestroyed(csp) {
		return c.destroy(csp)
	}

	// take a lock for common
	common.SyncResources.Mux.Lock()

	// try to import pool
	isImported, err = zpool.Import(csp)
	if isImported {
		if err != nil {
			c.recorder.Event(csp,
				corev1.EventTypeWarning,
				string(common.FailureImported),
				string(common.FailureImportOperations))
			common.SyncResources.Mux.Unlock()
			return err
		}
		zpool.CheckImportedPoolVolume()
		common.SyncResources.Mux.Unlock()
		return c.update(csp)
	}

	if IsEmptyStatus(csp) || IsPendingStatus(csp) {
		err = zpool.Create(csp)
		if err != nil {
			// We will try to create it in next event
			_ = zpool.Delete(csp)
			c.recorder.Event(csp,
				corev1.EventTypeWarning,
				string(common.FailureCreate),
				fmt.Sprintf("%s : %s", string(common.MessageResourceFailCreate), err.Error()))
			common.SyncResources.Mux.Unlock()
			return err
		} else {
			c.recorder.Event(csp,
				corev1.EventTypeNormal,
				string(common.SuccessCreated),
				string(common.MessageResourceCreated))
		}
	}
	common.SyncResources.Mux.Unlock()

	return c.updateStatus(csp)
}

func (c *CStorPoolController) destroy(csp *apis2.CStorNPool) error {
	var phase apis2.CStorPoolPhase

	// DeletePool is to delete cstor zpool.
	// It will also clear the label for relevant disk
	err := zpool.Delete(csp)
	if err != nil {
		c.recorder.Event(csp,
			corev1.EventTypeWarning,
			string(common.FailureDestroy),
			string(common.MessageResourceFailDestroy))
		phase = apis2.CStorPoolStatusDeletionFailed
		goto updatestatus
	}

	// removeFinalizer is to remove finalizer of cStorPool resource.
	err = c.removeFinalizer(csp)
	if err != nil {
		// Object will exist. Let's set status as offline
		phase = apis2.CStorPoolStatusDeletionFailed
		goto updatestatus
	}
	return nil

updatestatus:
	csp.Status.Phase = phase
	_, _ = zpool.OpenEBSClient2.
		OpenebsV1alpha2().
		CStorNPools(csp.Namespace).
		Update(csp)
	return err
}

func (c *CStorPoolController) update(csp *apis2.CStorNPool) error {
	var err error

	err = zpool.Update(csp)
	if err != nil {
		c.recorder.Event(csp,
			corev1.EventTypeWarning,
			string(common.FailedSynced),
			fmt.Sprintf("Failed to update pool due to '%s'", err.Error()))
		return err
	}
	return c.updateStatus(csp)
}

func (c *CStorPoolController) updateStatus(csp *apis2.CStorNPool) error {
	var status apis2.CStorPoolStatus
	var err error
	pool := zpool.PoolName(csp)

	state, er := zpool.GetPropertyValue(pool, "health")
	if er != nil {
		err = zpool.ErrorWrapf(err, "Failed to fetch health")
	} else {
		status.Phase = apis2.CStorPoolPhase(state)
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
		c.recorder.Event(csp,
			corev1.EventTypeWarning,
			string(common.FailureStatusSync),
			string(common.MessageResourceFailStatusSync))
		return err
	}

	if IsStatusChange(csp.Status, status) {
		csp.Status = status
		_, err = zpool.OpenEBSClient2.
			OpenebsV1alpha2().
			CStorNPools(csp.Namespace).
			Update(csp)
		return err
	}
	return nil
}

// getCSPObjFromKey returns object corresponding to the resource key
func (c *CStorPoolController) getCSPObjFromKey(key string) (*apis2.CStorNPool, error) {
	// Convert the key(namespace/name) string into a distinct name
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil, nil
	}

	csp, err := c.clientset.
		OpenebsV1alpha2().
		CStorNPools(ns).
		Get(name, metav1.GetOptions{})
	if err != nil {
		// The cStorPool resource may no longer exist, in which case we stop
		// processing.
		if k8serror.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("CSP '%s' in work queue no longer exists", key))
			return nil, nil
		}

		return nil, err
	}
	return csp, nil
}

// removeFinalizer is to remove finalizer of cstorpool resource.
func (c *CStorPoolController) removeFinalizer(csp *apis2.CStorNPool) error {
	if len(csp.Finalizers) == 0 {
		return nil
	}

	csp.Finalizers = []string{}
	_, err := c.clientset.
		OpenebsV1alpha2().
		CStorNPools(csp.Namespace).
		Update(csp)
	if err != nil {
		return err
	}
	glog.Infof("Removed Finalizer: %v, %v",
		csp.Name,
		string(csp.GetUID()))
	return nil
}
