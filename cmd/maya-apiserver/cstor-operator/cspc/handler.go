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
	nodeselect "github.com/openebs/maya/pkg/algorithm/nodeselect/v1alpha2"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebs "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// TODO: Following will be used in future PRs.
//var (
//	// supportedPool is a map holding the supported raid configurations.
//	supportedPool = map[apis.CasPoolValString]bool{
//		apis.PoolTypeStripedCPV:  true,
//		apis.PoolTypeMirroredCPV: true,
//		apis.PoolTypeRaidzCPV:    true,
//		apis.PoolTypeRaidz2CPV:   true,
//	}
//)

type clientSet struct {
	oecs openebs.Interface
}

// PoolConfig embeds nodeselect config from algorithm package and Controller object.
type PoolConfig struct {
	AlgorithmConfig *nodeselect.Config
	Controller      *Controller
}

// NewPoolConfig returns a poolconfig object
func (c *Controller) NewPoolConfig(cspc *apis.CStorPoolCluster, namespace string) (*PoolConfig, error) {
	pc, err := nodeselect.
		NewBuilder().
		WithCSPC(cspc).
		WithNameSpace(namespace).
		Build()
	if err != nil {
		return nil, errors.Wrap(err, "could not get algorithm config for provisioning")
	}
	return &PoolConfig{AlgorithmConfig: pc, Controller: c}, nil

}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the cspcPoolUpdated resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	startTime := time.Now()
	glog.V(4).Infof("Started syncing cstorpoolcluster %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing cstorpoolcluster %q (%v)", key, time.Since(startTime))
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
	err = c.syncCSPC(cspcGot)
	return err
}

// enqueueCSPC takes a CSPC resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CSPC.
func (c *Controller) enqueueCSPC(cspc interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(cspc); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// synSpc is the function which tries to converge to a desired state for the cspc.
func (c *Controller) syncCSPC(cspc *apis.CStorPoolCluster) error {
	//err := validate(cspc)
	//if err != nil {
	//	glog.Errorf("Validation of cspc failed:%s", err)
	//	return nil
	//}
	openebsNameSpace := env.Get(env.OpenEBSNamespace)
	if openebsNameSpace == "" {
		message := fmt.Sprint("Could not sync CSPC: got empty namespace for openebs from env variable")
		c.recorder.Event(cspc, corev1.EventTypeWarning, "Getting Namespace", message)
		glog.Errorf("Could not sync CSPC {%s}: got empty namespace for openebs from env variable", cspc.Name)
		return nil
	}

	pc, err := c.NewPoolConfig(cspc, openebsNameSpace)
	if err != nil {
		message := fmt.Sprintf("Could not sync CSPC : failed to get pool config: {%s}", err.Error())
		c.recorder.Event(cspc, corev1.EventTypeWarning, "Creating Pool Config", message)
		glog.Errorf("Could not sync CSPC {%s}: failed to get pool config: {%s}", cspc.Name, err.Error())
		return nil
	}

	pendingPoolCount, err := pc.AlgorithmConfig.GetPendingPoolCount()
	if err != nil {
		message := fmt.Sprintf("Could not sync CSPC : failed to get pending pool count: {%s}", err.Error())
		c.recorder.Event(cspc, corev1.EventTypeWarning, "Getting Pending Pool(s) ", message)
		glog.Errorf("Could not sync CSPC {%s}: failed to get pending pool count:{%s}", cspc.Name, err.Error())
		return nil
	}

	if pendingPoolCount > 0 {
		err = pc.create(pendingPoolCount, cspc)
		if err != nil {
			message := fmt.Sprintf("Could not create pool(s) for CSPC: %s", err.Error())
			c.recorder.Event(cspc, corev1.EventTypeWarning, "Pool Create", message)
			glog.Errorf("Could not create pool(s) for CSPC {%s}:{%s}", cspc.Name, err.Error())
			return nil
		}
	}

	cspList, err := pc.AlgorithmConfig.GetCSPWithoutDeployment()
	if err != nil {
		// Note: CSP for which pool deployment does not exists are known as orphaned.
		message := fmt.Sprintf("Error in getting orphaned CSP :{%s}", err.Error())
		c.recorder.Event(cspc, corev1.EventTypeWarning, "Pool Create", message)
		glog.Errorf("Error in getting orphaned CSP for CSPC {%s}:{%s}", cspc.Name, err.Error())
		return nil
	}

	if len(cspList) > 0 {
		pc.createDeployForCSPList(cspList)
	}

	return nil
}

// create is a wrapper function that calls the actual function to create pool as many time
// as the number of pools need to be created.
func (pc *PoolConfig) create(pendingPoolCount int, cspc *apis.CStorPoolCluster) error {
	newSpcLease := &Lease{cspc, CSPCLeaseKey, pc.Controller.clientset, pc.Controller.kubeclientset}
	err := newSpcLease.Hold()
	if err != nil {
		return errors.Wrapf(err, "Could not acquire lease on cspc object")
	}
	glog.V(4).Infof("Lease acquired successfully on cstorpoolcluster %s ", cspc.Name)
	defer newSpcLease.Release()
	for poolCount := 1; poolCount <= pendingPoolCount; poolCount++ {
		err = pc.CreateStoragePool()
		if err != nil {
			message := fmt.Sprintf("Pool provisioning failed for %d/%d ", poolCount, pendingPoolCount)
			pc.Controller.recorder.Event(cspc, corev1.EventTypeWarning, "Create", message)
			runtime.HandleError(errors.Wrapf(err, "Pool provisioning failed for %d/%d for cstorpoolcluster %s", poolCount, pendingPoolCount, cspc.Name))
		} else {
			message := fmt.Sprintf("Pool Provisioned %d/%d ", poolCount, pendingPoolCount)
			pc.Controller.recorder.Event(cspc, corev1.EventTypeNormal, "Create", message)
			glog.Infof("Pool provisioned successfully %d/%d for cstorpoolcluster %s", poolCount, pendingPoolCount, cspc.Name)
		}
	}
	return nil
}

func (pc *PoolConfig) createDeployForCSPList(cspList []apis.NewTestCStorPool) {
	for _, cspObj := range cspList {
		cspObj := cspObj
		err := pc.createDeployForCSP(&cspObj)
		if err != nil {
			message := fmt.Sprintf("Failed to create pool deployment for CSP %s: %s", cspObj.Name, err.Error())
			pc.Controller.recorder.Event(pc.AlgorithmConfig.CSPC, corev1.EventTypeWarning, "PoolDeploymentCreate", message)
			runtime.HandleError(errors.Errorf("Failed to create pool deployment for CSP %s: %s", cspObj.Name, err.Error()))
		}
	}
}

func (pc *PoolConfig) createDeployForCSP(csp *apis.NewTestCStorPool) error {
	deployObj, err := pc.GetPoolDeploySpec(csp)
	if err != nil {
		return errors.Wrapf(err, "could not get deployment spec for csp {%s}", csp.Name)
	}
	err = pc.createPoolDeployment(deployObj)
	if err != nil {
		return errors.Wrapf(err, "could not create deployment for csp {%s}", csp.Name)
	}
	return nil
}
