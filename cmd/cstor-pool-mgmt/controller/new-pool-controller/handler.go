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

	"github.com/pkg/errors"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apis2 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	zpool "github.com/openebs/maya/pkg/cstor/pool/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the cStorPoolUpdated resource
// with the current status of the resource.
func (c *CStorPoolController) syncHandler(key string, operation common.QueueOperation) error {
	csp, err := c.getCSPObjFromKey(key)
	if err != nil {
		return err
	}

	return c.eventHandler(operation, csp)
}

// TODO remove event base handling
// cStorPoolEventHandler is to handle cstor pool related events.
func (c *CStorPoolController) eventHandler(operation common.QueueOperation, csp *apis2.CStorNPool) error {
	switch operation {
	case common.QOpAdd:
		return c.addEventHandler(csp)
	case common.QOpDestroy:
		return c.destroyEventHandler(csp)
	case common.QOpSync:
		return c.syncEventHandler(csp)
	case common.QOpModify:
		return c.modifyEventHandler(csp)
	}

	return errors.Errorf("Invalid event {%s} for cStorPool {%s}", operation, csp.Name)
}

func (c *CStorPoolController) addEventHandler(csp *apis2.CStorNPool) error {
	var status string
	var err error
	var isImported bool

	uid := string(csp.GetUID())
	importedPool := zpool.ImportedCStorPools[uid]

	// TODO
	// In which scenario it is possible??
	if importedPool != nil {
		c.recorder.Event(importedPool,
			corev1.EventTypeWarning,
			string(common.AlreadyPresent),
			string(common.MessageResourceAlreadyPresent))
		return nil
	}

	// take a lock for common
	// TODO move return to label and unlock
	common.SyncResources.Mux.Lock()

	status, isImported = zpool.IsPoolImported(csp, true /* wait if Zrepl is initializing */)
	if isImported {
		c.recorder.Event(csp,
			corev1.EventTypeNormal,
			string(common.AlreadyPresent),
			string(common.MessageResourceAlreadyPresent))
		goto updatestatus
	}

	// try to import pool
	status, isImported, err = zpool.Import(csp)
	if isImported {
		if err != nil {
			c.recorder.Event(csp,
				corev1.EventTypeWarning,
				string(common.FailureImported),
				string(common.FailureImportOperations))
			glog.Errorf("Failed to handle add event: import succeeded with failed operations %v", err.Error())
		} else {
			c.recorder.Event(csp,
				corev1.EventTypeNormal,
				string(common.SuccessImported),
				string(common.MessageResourceImported))
		}
		goto updatestatus
	}

	//TODO: Why pool validation? should we consider zpool status during sync op only?

	// IsInitStatus is to check if initial status of cstorpool object is `init`.
	if zpool.IsEmptyStatus(csp) || zpool.IsPendingStatus(csp) {
		err = zpool.Create(csp)
		if err != nil {
			glog.Errorf("Pool creation failure: %v", string(csp.GetUID()))
			c.recorder.Event(csp,
				corev1.EventTypeWarning,
				string(common.FailureCreate),
				string(common.MessageResourceFailCreate))
			status = string(apis.CStorPoolStatusOffline)
			goto updatestatus
		}
		glog.Infof("Pool creation successful: %v", string(csp.GetUID()))
		c.recorder.Event(csp,
			corev1.EventTypeNormal,
			string(common.SuccessCreated),
			string(common.MessageResourceCreated))
		status = string(apis.CStorPoolStatusOnline)
		goto updatestatus
	}
	//TODO else set errored status

updatestatus:
	common.SyncResources.Mux.Unlock()
	glog.Infof("TODO update status here  %s", status)
	return err
}

func (c *CStorPoolController) destroyEventHandler(csp *apis2.CStorNPool) error {
	// DeletePool is to delete cstor zpool.
	// It will also clear the label for relevant disk
	err := zpool.Delete(csp)
	if err != nil {
		c.recorder.Event(csp,
			corev1.EventTypeWarning,
			string(common.FailureDestroy),
			string(common.MessageResourceFailDestroy))
		// TODO
		// update status here
		// string(apis.CStorPoolStatusDeletionFailed), err
		return err
	}

	// removeFinalizer is to remove finalizer of cStorPool resource.
	err = c.removeFinalizer(csp)
	if err != nil {
		// TODO
		// Object will exist. Let's set status as offline
		return err
	}
	return nil
}

// syncEventHandler
func (c *CStorPoolController) syncEventHandler(csp *apis2.CStorNPool) error {
	//TODO add sync operation
	// Update pool status if changed
	newstatus, err := zpool.GetStatus(csp)
	if err != nil {
		c.recorder.Event(csp,
			corev1.EventTypeWarning,
			string(common.FailureStatusSync),
			string(common.MessageResourceFailStatusSync))
		return err
	}

	glog.Infof("got status %s for pool {%s}", newstatus, zpool.PoolName(csp))

	// TODO where to put pool status?
	return nil
}

// modifyEventHandler
func (c *CStorPoolController) modifyEventHandler(csp *apis2.CStorNPool) error {
	//TODO add modify
	err := zpool.Update(csp)
	if err != nil {
		glog.Errorf("Failed to modify pool {%s} .. {%v}", zpool.PoolName(csp), err)
		return err
	}

	// TODO where to put pool status
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
	if len(csp.Finalizers) > 0 {
		csp.Finalizers = []string{}
	}
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
