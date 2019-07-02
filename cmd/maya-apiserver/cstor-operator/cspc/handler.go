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

package cspc

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebs "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

var (
	// supportedPool is a map holding the supported raid configurations.
	supportedPool = map[apis.CasPoolValString]bool{
		apis.PoolTypeStripedCPV:  true,
		apis.PoolTypeMirroredCPV: true,
		apis.PoolTypeRaidzCPV:    true,
		apis.PoolTypeRaidz2CPV:   true,
	}
)

type clientSet struct {
	oecs openebs.Interface
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
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the cspc resource with this namespace/name
	cspc, err := c.cspcLister.CStorPoolClusters(ns).Get(name)
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
func (c *Controller) syncSpc(cspc *apis.CStorPoolCluster) error {
	//err := validate(cspc)
	//if err != nil {
	//	glog.Errorf("Validation of cspc failed:%s", err)
	//	return nil
	//}
	pendingPoolCount, err := c.getPendingPoolCount(cspc)
	if err != nil {
		return err
	}
	if pendingPoolCount > 0 {
		err = c.create(pendingPoolCount, cspc)
		if err != nil {
			return err
		}
	}
	return nil
}

// create is a wrapper function that calls the actual function to create pool as many time
// as the number of pools need to be created.
func (c *Controller) create(pendingPoolCount int, cspc *apis.CStorPoolCluster) error {
	var newSpcLease Leaser
	newSpcLease = &Lease{cspc, SpcLeaseKey, c.clientset, c.kubeclientset}
	err := newSpcLease.Hold()
	if err != nil {
		return errors.Wrapf(err, "Could not acquire lease on cspc object")
	}
	glog.V(4).Infof("Lease acquired successfully on storagepoolclaim %s ", cspc.Name)
	defer newSpcLease.Release()
	for poolCount := 1; poolCount <= pendingPoolCount; poolCount++ {
		glog.Infof("Provisioning pool %d/%d for storagepoolclaim %s", poolCount, pendingPoolCount, cspc.Name)
		err = c.CreateStoragePool(cspc)
		if err != nil {
			runtime.HandleError(errors.Wrapf(err, "Pool provisioning failed for %d/%d for storagepoolclaim %s", poolCount, pendingPoolCount, cspc.Name))
		}
	}
	return nil
}

// TODO : Move following function to algorithm package

func (c *Controller) getPendingPoolCount(cspc *apis.CStorPoolCluster) (int, error) {
	currentPoolCount, err := c.getCurrentPoolCount(cspc)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to get pending pool count for cspc %s", cspc.Name)
	}
	desiredPoolCount := len(cspc.Spec.Pools)

	return (desiredPoolCount - currentPoolCount), nil
}

// getCurrentPoolCount give the current pool count for the given auto provisioned spc.
func (c *Controller) getCurrentPoolCount(cspc *apis.CStorPoolCluster) (int, error) {
	// Get the current count of provisioned pool for the storagepool claim
	cspList, err := c.clientset.OpenebsV1alpha1().NewTestCStorPools("openebs").List(metav1.ListOptions{LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + cspc.Name})
	if err != nil {
		return 0, errors.Errorf("unable to get current pool count:unable to list cstor pools: %v", err)
	}
	return len(cspList.Items), nil
}

func (c *Controller) isPoolPending(cspc *apis.CStorPoolCluster) bool {
	pc, err := c.getPendingPoolCount(cspc)
	if err != nil {
		glog.Errorf("unable to get pending pool count : %v", err)
		return false
	}
	if pc > 0 {
		return true
	}
	return false
}
