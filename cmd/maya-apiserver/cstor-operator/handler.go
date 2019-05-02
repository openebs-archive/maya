/*
Copyright 2018 The OpenEBS Authors

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

package spc

import (
	"fmt"
	"github.com/golang/glog"
	nodeselectv1alpha2 "github.com/openebs/maya/pkg/algorithm/caspool/v1alpha1"
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebs "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"time"
)

// PoolOperation is used to create a cstor pool.
type PoolOperation struct {
	*nodeselectv1alpha2.Operations
	*Controller
}

type clientSet struct {
	oecs openebs.Interface
}

// NewPoolOperation returns an instance of PoolOperation
func (c *Controller) NewPoolOperation(cspc *apisv1alpha1.CStorPoolCluster) *PoolOperation {
	ops := nodeselectv1alpha2.NewOperationsBuilder().
		WithCStorPoolCluster(cspc).
		WithDefaults().
		Build()
	// TODO: Add maxPool nil check
	return &PoolOperation{ops, c}
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the cspcPoolUpdated resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	startTime := time.Now()
	glog.V(4).Infof("Started syncing storagepoolclaim %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing storagepoolclaim %q (%v)", key, time.Since(startTime))
	}()

	// Convert the namespace/name string into a distinct namespace and name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the cspc resource with this namespace/name
	cspc, err := c.cspcLister.Get(name)
	if k8serror.IsNotFound(err) {
		runtime.HandleError(fmt.Errorf("cspc '%s' has been deleted", key))
		return nil
	}
	if err != nil {
		return err
	}

	// Deep-copy otherwise we are mutating our cache.
	// TODO: Deep-copy only when needed.
	cspcGot := cspc.DeepCopy()
	err = c.syncSpc(cspcGot)
	return err
}

// enqueueSpc takes a SPC resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than SPC.
func (c *Controller) enqueueSpc(cspc interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(cspc); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// synSpc is the function which tries to converge to a desired state for the cspc.
func (c *Controller) syncSpc(cspc *apisv1alpha1.CStorPoolCluster) error {
	po := c.NewPoolOperation(cspc)
	pendingPoolCount, err := po.GetPendingPoolCount()
	if err != nil {
		// Do not return error -- as it will be tried again causing log flooding.
		// TODO: Think of backoff period
		// Instead log and return nil error, it will enqueued again for processing in resync period
		glog.Errorf("failed to get pending pool count for cspc %s:{%v}", cspc.Name, err)
		return err
	}

	if po.IsPoolCreationPending() {
		err = po.create(pendingPoolCount, cspc)
		if err != nil {
			glog.Errorf("failed to create pool for cspc %s:{%v}", cspc.Name, err)
			return err
		}
	}
	return nil
}

// create is a wrapper function that calls the actual function to create pool as many time
// as the number of pools need to be created.
func (po *PoolOperation) create(pendingPoolCount int, cspc *apisv1alpha1.CStorPoolCluster) error {
	var newSpcLease Leaser
	newSpcLease = &Lease{cspc, SpcLeaseKey, po.clientset, po.kubeclientset}
	err := newSpcLease.Hold()
	if err != nil {
		return errors.Wrapf(err, "Could not acquire lease on cspc object")
	}
	glog.V(4).Infof("Lease acquired successfully on storagepoolclaim %s ", cspc.Name)
	defer newSpcLease.Release()
	for poolCount := 1; poolCount <= pendingPoolCount; poolCount++ {
		glog.Infof("Provisioning pool %d/%d for storagepoolclaim %s", poolCount, pendingPoolCount, cspc.Name)
		err = po.CreateStoragePool(cspc)
		if err != nil {
			runtime.HandleError(errors.Wrapf(err, "Pool provisioning failed for %d/%d for storagepoolclaim %s", poolCount, pendingPoolCount, cspc.Name))
		}
	}
	return nil
}
