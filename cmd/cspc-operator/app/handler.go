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

package app

import (
	"fmt"

	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	apiscspc "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	"github.com/openebs/maya/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"time"

	apiscsp "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	apispdb "github.com/openebs/maya/pkg/kubernetes/poddisruptionbudget"

	nodeselect "github.com/openebs/maya/pkg/algorithm/nodeselect/v1alpha2"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebs "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

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
	klog.V(4).Infof("Started syncing cstorpoolcluster %q (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing cstorpoolcluster %q (%v)", key, time.Since(startTime))
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
func (c *Controller) syncCSPC(cspcGot *apis.CStorPoolCluster) error {

	openebsNameSpace := env.Get(env.OpenEBSNamespace)
	if openebsNameSpace == "" {
		message := fmt.Sprint("Could not sync CSPC: got empty namespace for openebs from env variable")
		c.recorder.Event(cspcGot, corev1.EventTypeWarning, "Getting Namespace", message)
		klog.Errorf("Could not sync CSPC {%s}: got empty namespace for openebs from env variable", cspcGot.Name)
		return nil
	}

	cspcObj := cspcGot
	cspcObj, err := c.populateVersion(cspcObj)
	if err != nil {
		klog.Errorf("failed to add versionDetails to CSPC %s:%s", cspcGot.Name, err.Error())
		return nil
	}

	cspcGot = cspcObj
	pc, err := c.NewPoolConfig(cspcGot, openebsNameSpace)
	if err != nil {
		message := fmt.Sprintf("Could not sync CSPC : failed to get pool config: {%s}", err.Error())
		c.recorder.Event(cspcGot, corev1.EventTypeWarning, "Creating Pool Config", message)
		klog.Errorf("Could not sync CSPC {%s}: failed to get pool config: {%s}", cspcGot.Name, err.Error())
		return nil
	}

	// If CSPC is deleted -- handle the deletion.
	if !cspcGot.DeletionTimestamp.IsZero() {
		err = pc.handleCSPCDeletion()
		if err != nil {
			klog.Errorf("Failed to sync CSPC for deletion:%s", err.Error())
		}
		return nil
	}

	cspcBuilderObj, err := apiscspc.BuilderForAPIObject(cspcGot).Build()
	if err != nil {
		klog.Errorf("Failed to build CSPC api object %s", cspcGot.Name)
		return nil
	}

	cspc, err := cspcBuilderObj.AddFinalizer(apiscspc.CSPCFinalizer)
	if err != nil {
		klog.Errorf("Failed to add finalizer on CSPC %s:%s", cspcGot.Name, err.Error())
		return nil
	}

	// Create PDB for cStor pools only if user specified minAvalibility
	pc.HandlePDBForCSPC(cspc)

	pendingPoolCount, err := pc.AlgorithmConfig.GetPendingPoolCount()
	if err != nil {
		message := fmt.Sprintf("Could not sync CSPC : failed to get pending pool count: {%s}", err.Error())
		c.recorder.Event(cspc, corev1.EventTypeWarning, "Getting Pending Pool(s) ", message)
		klog.Errorf("Could not sync CSPC {%s}: failed to get pending pool count:{%s}", cspc.Name, err.Error())
		return nil
	}

	if pendingPoolCount < 0 {
		err = pc.DownScalePool()
		if err != nil {
			message := fmt.Sprintf("Could not downscale pool: %s", err.Error())
			c.recorder.Event(cspc, corev1.EventTypeWarning, "PoolDownScale", message)
			klog.Errorf("Could not downscale pool for CSPC %s: %s", cspc.Name, err.Error())
			return nil
		}
	}

	if pendingPoolCount > 0 {
		err = pc.create(pendingPoolCount, cspc)
		if err != nil {
			message := fmt.Sprintf("Could not create pool(s) for CSPC: %s", err.Error())
			c.recorder.Event(cspc, corev1.EventTypeWarning, "Pool Create", message)
			klog.Errorf("Could not create pool(s) for CSPC {%s}:{%s}", cspc.Name, err.Error())
			return nil
		}
	}

	cspList, err := pc.AlgorithmConfig.GetCSPIWithoutDeployment()
	if err != nil {
		// Note: CSP for which pool deployment does not exists are known as orphaned.
		message := fmt.Sprintf("Error in getting orphaned CSP :{%s}", err.Error())
		c.recorder.Event(cspc, corev1.EventTypeWarning, "Pool Create", message)
		klog.Errorf("Error in getting orphaned CSP for CSPC {%s}:{%s}", cspc.Name, err.Error())
		return nil
	}

	if len(cspList) > 0 {
		pc.createDeployForCSPList(cspList)
	}

	if pendingPoolCount == 0 {
		klog.V(2).Infof("Handling pool operations for CSPC %s if any", cspc.Name)
		pc.handleOperations()
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
	klog.V(4).Infof("Lease acquired successfully on cstorpoolcluster %s ", cspc.Name)
	for poolCount := 1; poolCount <= pendingPoolCount; poolCount++ {
		err = pc.CreateStoragePool()
		if err != nil {
			message := fmt.Sprintf("Pool provisioning failed for %d/%d ", poolCount, pendingPoolCount)
			pc.Controller.recorder.Event(cspc, corev1.EventTypeWarning, "Create", message)
			runtime.HandleError(errors.Wrapf(err, "Pool provisioning failed for %d/%d for cstorpoolcluster %s", poolCount, pendingPoolCount, cspc.Name))
		} else {
			message := fmt.Sprintf("Pool Provisioned %d/%d ", poolCount, pendingPoolCount)
			pc.Controller.recorder.Event(cspc, corev1.EventTypeNormal, "Create", message)
			klog.Infof("Pool provisioned successfully %d/%d for cstorpoolcluster %s", poolCount, pendingPoolCount, cspc.Name)
		}
	}
	return nil
}

func (pc *PoolConfig) createDeployForCSPList(cspList []apis.CStorPoolInstance) {
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

func (pc *PoolConfig) createDeployForCSP(csp *apis.CStorPoolInstance) error {
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

// handleCSPCDeletion handles deletion of a CSPC resource by deleting
// the associated CSP resource to it, removing the CSPC finalizer
// on BDC(s) used and then removing the CSPC finalizer on CSPC resource
// itself.

// It is necessary that CSPC resource has the CSPC finalizer on it in order to
// execute the handler.
func (pc *PoolConfig) handleCSPCDeletion() error {
	err := pc.deleteAssociatedCSPI()

	if err != nil {
		return errors.Wrapf(err, "failed to handle CSPC deletion")
	}

	cspcBuilderObj, err := apiscspc.BuilderForAPIObject(pc.AlgorithmConfig.CSPC).Build()
	if err != nil {
		klog.Errorf("Failed to build CSPC api object %s:%s", pc.AlgorithmConfig.CSPC.Name, err.Error())
		return nil
	}

	if cspcBuilderObj.HasFinalizer(apiscspc.CSPCFinalizer) {
		err := pc.removeCSPCFinalizer()
		if err != nil {
			return errors.Wrapf(err, "failed to handle CSPC %s deletion", pc.AlgorithmConfig.CSPC.Name)
		}
	}

	return nil
}

// deleteAssociatedCSPI deletes the CSPI resource(s) belonging to the given CSPC resource.
// If no CSPI resource exists for the CSPC, then a levelled info log is logged and function
// returns.
func (pc *PoolConfig) deleteAssociatedCSPI() error {
	err := apiscsp.NewKubeClient().WithNamespace(pc.AlgorithmConfig.Namespace).DeleteCollection(
		metav1.ListOptions{
			LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + pc.AlgorithmConfig.CSPC.Name,
		},
		&metav1.DeleteOptions{},
	)

	if k8serror.IsNotFound(err) {
		klog.V(2).Infof("Associated CSPI(s) of CSPC %s is already deleted:%s", pc.AlgorithmConfig.CSPC.Name, err.Error())
		return nil
	}

	if err != nil {
		return errors.Wrapf(err, "failed to delete associated CSPI(s):%s", err.Error())
	}
	klog.Infof("Associated CSPI(s) of CSPC %s deleted successfully ", pc.AlgorithmConfig.CSPC.Name)
	return nil
}

// removeSPCFinalizer removes CSPC finalizers on associated
// BDC resources and CSPC object itself.
func (pc *PoolConfig) removeCSPCFinalizer() error {
	cspList, err := apiscsp.NewKubeClient().List(metav1.ListOptions{
		LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + pc.AlgorithmConfig.CSPC.Name,
	})

	if err != nil {
		return errors.Wrap(err, "failed to remove CSPC finalizer on associated resources")
	}

	if len(cspList.Items) > 0 {
		return errors.Wrap(err, "failed to remove CSPC finalizer on associated resources as "+
			"CSPI(s) still exists for CSPC")
	}

	err = pc.removeSPCFinalizerOnAssociatedBDC()

	if err != nil {
		return errors.Wrap(err, "failed to remove CSPC finalizer on associated resources")
	}

	cspcBuilderObj, err := apiscspc.BuilderForAPIObject(pc.AlgorithmConfig.CSPC).Build()
	if err != nil {
		klog.Errorf("Failed to build CSPC api object %s", pc.AlgorithmConfig.CSPC.Name)
		return nil
	}

	err = cspcBuilderObj.RemoveFinalizer(apiscspc.CSPCFinalizer)

	if err != nil {
		return errors.Wrap(err, "failed to remove CSPC finalizer on associated resources")
	}
	return nil
}

// removeSPCFinalizerOnAssociatedBDC removes CSPC finalizer on associated BDC resource(s)
func (pc *PoolConfig) removeSPCFinalizerOnAssociatedBDC() error {
	bdcList, err := bdc.NewKubeClient().WithNamespace(pc.AlgorithmConfig.Namespace).List(
		metav1.ListOptions{
			LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + pc.AlgorithmConfig.CSPC.Name,
		})

	if err != nil {
		return errors.Wrapf(err, "failed to remove CSPC finalizer on BDC resources")
	}

	for _, bdcObj := range bdcList.Items {
		bdcObj := bdcObj
		_, err := bdc.BuilderForAPIObject(&bdcObj).BDC.RemoveFinalizer(apiscspc.CSPCFinalizer)
		if err != nil {
			return errors.Wrapf(err, "failed to remove CSPC finalizer on BDC %s", bdcObj.Name)
		}
	}

	return nil
}

// populateVersion assigns VersionDetails for old cspc object and newly created
// cspc
func (c *Controller) populateVersion(cspc *apis.CStorPoolCluster) (*apis.CStorPoolCluster, error) {
	if cspc.VersionDetails.Status.Current == "" {
		var err error
		var v string
		var obj *apis.CStorPoolCluster
		v, err = c.EstimateCSPCVersion(cspc)
		if err != nil {
			return nil, err
		}
		cspc.VersionDetails.Status.Current = v
		// For newly created spc Desired field will also be empty.
		cspc.VersionDetails.Desired = v
		cspc.VersionDetails.Status.DependentsUpgraded = true
		obj, err = c.clientset.OpenebsV1alpha1().
			CStorPoolClusters(env.Get(env.OpenEBSNamespace)).
			Update(cspc)

		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to update spc %s while adding versiondetails",
				cspc.Name,
			)
		}
		klog.Infof("Version %s added on spc %s", v, cspc.Name)
		return obj, nil
	}
	return cspc, nil
}

// EstimateCSPCVersion returns the cspi version if any cspi is present for the cspc or
// returns the maya version as the new cspi created will be of maya version
func (c *Controller) EstimateCSPCVersion(cspc *apis.CStorPoolCluster) (string, error) {

	cspiList, err := c.clientset.OpenebsV1alpha1().
		CStorPoolInstances(env.Get(env.OpenEBSNamespace)).
		List(
			metav1.ListOptions{
				LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + cspc.Name,
			})
	if err != nil {
		return "", errors.Wrapf(
			err,
			"failed to get the cstorpool instance list related to cspc : %s",
			cspc.Name,
		)
	}
	if len(cspiList.Items) == 0 {
		return version.Current(), nil
	}
	return cspiList.Items[0].Labels[string(apis.OpenEBSVersionKey)], nil
}

//TODO: Generate event

// HandlePDBForCSPC will create the PDB for CSPC based on minAvailable.
// HandlePDBForCSPC will does the following changes
// 1. If user updates the value to 0 then PDB corresponds to CSPC will be deleted.
// 2. If user updates the value other then existing PDB will be deleted and PDB
//    will be created with new value
func (pc *PoolConfig) HandlePDBForCSPC(cspc *apis.CStorPoolCluster) {
	pdbClient := apispdb.KubeClient().WithNamespace(cspc.Namespace)
	pdbList, err := pdbClient.
		List(metav1.ListOptions{LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + cspc.Name})
	if err != nil {
		klog.Errorf("failed to list poddisruptionbudget related to cspc: %s error: %v", cspc.Name, err)
		return
	}
	if len(pdbList.Items) > 1 {
		klog.Errorf("Invalid count of poddisruptionbudget instances: %d", len(pdbList.Items))
		return
	}
	// If there is any existing PDB and if MinAvailable in cspc got updated then
	// delete the existing PDB
	if len(pdbList.Items) == 1 &&
		cspc.Spec.PodDisruptionBudget.MinAvailable != pdbList.Items[0].Spec.MinAvailable.IntValue() {
		err = pdbClient.Delete(pdbList.Items[0].Name, &metav1.DeleteOptions{})
		if err != nil {
			klog.Errorf(
				"failed to delete poddisruptionbudget: %s related to cspc: %s error: %v",
				pdbList.Items[0].Name,
				cspc.Name,
				err,
			)
			return
		}
	}
	// create poddisruptionbudget with cspc minAvailable value
	if cspc.Spec.PodDisruptionBudget.MinAvailable > 0 {
		err = createPDBForCSPC(cspc)
	}
}

func createPDBForCSPC(cspc *apis.CStorPoolCluster) error {
	pdbObj := policy.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:   cspc.Name + "-" + fmt.Sprintf("%d", time.Now().UnixNano()),
			Labels: getPDBLabels(cspc),
		},
		Spec: policy.PodDisruptionBudgetSpec{
			MinAvailable: convertIntToIntStr(cspc.Spec.PodDisruptionBudget.MinAvailable),
			Selector:     getPDBSelector(cspc),
		},
	}
	_, err := apispdb.KubeClient().
		WithNamespace(cspc.Namespace).
		Create(&pdbObj)
	return err
}

func getPDBLabels(cspc *apis.CStorPoolCluster) map[string]string {
	return map[string]string{
		string(apis.CStorPoolClusterCPK): cspc.Name,
	}
}

func getPDBSelector(cspc *apis.CStorPoolCluster) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			string(apis.CStorPoolClusterCPK): cspc.Name,
			"app":                            "cstor-pool",
		},
	}
}

func convertIntToIntStr(val int) *intstr.IntOrString {
	intOrString := intstr.FromInt(val)
	return &intOrString
}
