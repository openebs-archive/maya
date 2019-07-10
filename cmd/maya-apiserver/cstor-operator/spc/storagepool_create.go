/*
Copyright 2017 The OpenEBS Authors

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
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
	nodeselect "github.com/openebs/maya/pkg/algorithm/nodeselect/v1alpha1"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha1"
	cspc "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	cspcbd "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/cstorpoolblockdevice"
	poolspec "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/cstorpoolspecs"
	raidgroup "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/raidgroups"
	"github.com/openebs/maya/pkg/storagepool"
	spcv1alpha1 "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/kubernetes/pkg/util/slice"
)

// PoolCreateConfig is config object used to create a cstor pool.
type PoolCreateConfig struct {
	*nodeselect.Config
	*Controller
}

// CasPoolBuilder is the builder object for cas pool.
type CasPoolBuilder struct {
	CasPool *apis.CasPool
}

// Cas template is a custom resource which has a list of runTasks.

// runTasks are configmaps which has defined yaml templates for resources that needs
// to be created or deleted for a storagepool creation or deletion respectively.

// CreateOrUpdateCStorPoolCluster is a function that does following:
// 1. It receives storagepoolclaim object from the spc watcher event handler.
// 2. After successful validation, it will call a worker function to create or
//    update cstorpoolcluster based on availability of cspc object
func (c *Controller) CreateOrUpdateCStorPoolCluster(spcGot *apis.StoragePoolClaim) error {
	poolconfig := c.NewPoolCreateConfig(spcGot)
	maxPools := *spcGot.Spec.MaxPools
	cspcObjList, err := c.cspcLister.
		CStorPoolClusters(poolconfig.Namespace).
		List(
			klabels.SelectorFromSet(
				map[string]string{
					string(apis.StoragePoolClaimCPK): spcGot.Name,
				},
			),
		)
	if err != nil {
		return errors.Wrapf(err,
			"failed to list cspc in %s namespace using spc %s label",
			poolconfig.Namespace,
			spcGot.Name,
		)
	}
	if len(cspcObjList) == 0 {
		//create cstorpoolcluster
		glog.V(4).Infof(
			"Creating cstorpoolcluster in namespce %s for storagepoolclaim %s",
			poolconfig.Namespace,
			spcGot.Name,
		)
		// create cstorpoolcluster
		return poolconfig.createCStorPoolCluster(poolconfig.Namespace)
	}

	cspcObj := cspcObjList[0]
	if len(cspcObj.Spec.Pools) < maxPools {
		return poolconfig.updateCStorPoolCluster(cspcObj)
	}
	return nil
}

// createCStorPoolCluster creates cspc object and patch spc with finalizers
func (pc *PoolCreateConfig) createCStorPoolCluster(
	namespace string) error {
	var err error
	updatedSPC := pc.Spc
	if !slice.ContainsString(pc.Spc.ObjectMeta.Finalizers, spcv1alpha1.SPCFinalizer, nil) {
		updatedSPC, err = pc.patchSPCWithFinalizers()
		if err != nil {
			return errors.Wrapf(
				err,
				"failed to patch spc %s", pc.Spc.Name,
			)
		}
	}
	pc.Spc = updatedSPC
	spc := pc.Spc
	cspcObj, err := pc.buildCSPCFromSPC(spc, namespace)
	if err != nil {
		return errors.Wrapf(err,
			"failed to build cstorpoolcluster from spc %s",
			spc.Name,
		)
	}
	newCSPCObj, err := pc.clientset.
		OpenebsV1alpha1().
		CStorPoolClusters(cspcObj.Namespace).
		Create(cspcObj)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to create cspc %s in namespace %s",
			cspcObj.Name,
			cspcObj.Namespace,
		)
	}
	glog.Infof(
		"Successfully create cstorpoolcluster %s in namespace %s from spc %s",
		newCSPCObj.Name,
		newCSPCObj.Namespace,
		spc.Name,
	)
	return nil
}

// patchSPCWithFinalizers patches spc with spc finalizer
func (pc *PoolCreateConfig) patchSPCWithFinalizers() (*apis.StoragePoolClaim, error) {
	// Add finalizers on spc
	spcObj, err := pc.spcLister.
		Get(pc.Spc.Name)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get spc %s from lister",
			pc.Spc.Name,
		)
	}

	// make deepcopy of existing object to update it
	newSPCObj := spcObj.DeepCopy()
	spcBuilderObj := spcv1alpha1.BuilderForObject(
		&spcv1alpha1.SPC{
			Object: newSPCObj,
		},
	).
		WithFinalizersNew(spcv1alpha1.SPCFinalizer).
		Build()
	patchBytes, err := getPatchData(spcObj, spcBuilderObj.Object)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get patch bytes for spc %s",
			pc.Spc.Name,
		)
	}
	updatedSPCObj, err := pc.clientset.
		OpenebsV1alpha1().
		StoragePoolClaims().
		Patch(pc.Spc.Name, types.MergePatchType, patchBytes)
	if err != nil {
		//TODO: is deletetion of cspc is required?
		_ = pc.deleteCSPCResource(spcObj.Name, pc.Namespace)
		return nil, errors.Wrapf(
			err,
			"failed to update spc %s with finalizers",
			spcObj.Name,
		)
	}
	return updatedSPCObj, nil
}

// updateCStorPoolCluster updates the pool specs of cspc object
func (pc *PoolCreateConfig) updateCStorPoolCluster(
	cspcObj *apis.CStorPoolCluster) error {
	dupCSPCObj := cspcObj.DeepCopy()
	mapNodeBlockDeviceList, err := pc.NodeBlockDeviceSelector()
	if err != nil {
		return errors.Wrapf(err,
			"failed to update cspc %s in namespace %s",
			dupCSPCObj.Name,
			dupCSPCObj.Namespace,
		)
	}
	cspcBuilderObj := cspc.BuilderFromCSPC(
		cspc.NewForAPIObject(dupCSPCObj),
	)
	pendingPoolSpecCount := *pc.Spc.Spec.MaxPools - len(dupCSPCObj.Spec.Pools)
	customCSPCObj, err := pc.
		addPoolSpecToCSPCBuilder(cspcBuilderObj, mapNodeBlockDeviceList, pendingPoolSpecCount).
		Build()
	if err != nil {
		return err
	}

	patchBytes, err := getPatchData(cspcObj, customCSPCObj.ToAPI())
	if err != nil {
		return err
	}
	updatedCSPCObj, err := pc.clientset.
		OpenebsV1alpha1().
		CStorPoolClusters(dupCSPCObj.Namespace).
		Patch(dupCSPCObj.Name, types.MergePatchType, patchBytes)
	if err != nil {
		return err
	}
	glog.Infof(
		"Successfully updated cstorpoolcluster %s in namespace %s",
		updatedCSPCObj.Name,
		updatedCSPCObj.Namespace,
	)
	return nil
}

// buildCSPCFromSPC builds new cspc from spc
func (pc *PoolCreateConfig) buildCSPCFromSPC(
	spc *apis.StoragePoolClaim,
	namespace string) (*apis.CStorPoolCluster, error) {
	cpscFinalizers := []string{cspc.CSPCFinalizer}
	// get namespace where OpenEBS is installed
	cspcBuildObj := cspc.NewBuilder().
		WithName(spc.Name).
		WithLabelsNew(getSPCLabels(spc)).
		WithNamespace(namespace).
		WithFinalizersNew(cpscFinalizers)
	mapNodeBlockDeviceList, err := pc.NodeBlockDeviceSelector()
	if err != nil {
		return nil, err
	}
	noOfPoolSpecs := *spc.Spec.MaxPools
	customCSPCObj, err := pc.
		addPoolSpecToCSPCBuilder(cspcBuildObj, mapNodeBlockDeviceList, noOfPoolSpecs).
		Build()
	if err != nil {
		return nil, err
	}
	return customCSPCObj.ToAPI(), nil
}

// addPoolSpecToCSPCBuilder tries to add required numeber of pool specs to
// existing cspc builder
func (pc *PoolCreateConfig) addPoolSpecToCSPCBuilder(
	cspcBuilderObj *cspc.Builder,
	mapNodeBlockDeviceList map[string]*cspcbd.ListBuilder,
	reqPoolSpecCount int) *cspc.Builder {
	currentPoolSpecCount := 0
	for nodeName, customCSPCBDList := range mapNodeBlockDeviceList {
		if currentPoolSpecCount == reqPoolSpecCount {
			break
		}
		// blockdevice gives nodeName as a value of kubernetes.io/hostName label
		// using above label get node labels
		nodeObj, err := pc.nodeLister.Get(nodeName)
		if err != nil {
			glog.Errorf("failed to get node %s object", nodeName)
			continue
		}
		//Assumption kubernetes.io/hostName is the unique label available on node
		val, ok := nodeObj.GetLabels()[string(apis.HostNameCPK)]
		if !ok {
			glog.Errorf("failed to get value of label %s on node %s object",
				apis.HostNameCPK,
				nodeName,
			)
			continue
		}
		nodeSelector := map[string]string{string(apis.HostNameCPK): val}
		cacheFile := pc.Spc.Spec.PoolSpec.CacheFile
		poolType := pc.Spc.Spec.PoolSpec.PoolType
		poolSpecBuilder := poolspec.NewBuilder().
			WithNodeSelectorNew(nodeSelector).
			WithCompression("off").
			WithDefaultRaidGroupType(poolType).
			WithRaidGroupBuilder(
				raidgroup.NewBuilder().
					WithType(poolType).
					WithName("group-1").
					WithCSPCBlockDeviceList(customCSPCBDList),
			)
		if cacheFile != "" {
			poolSpecBuilder = poolSpecBuilder.WithCacheFilePath(cacheFile)
		}
		cspcBuilderObj = cspcBuilderObj.WithPoolSpecBuilder(poolSpecBuilder)
		currentPoolSpecCount++
	}
	return cspcBuilderObj
}

// getCasPool returns a configured cas pool object.
func (pc *PoolCreateConfig) getCasPool(spc *apis.StoragePoolClaim) (*apis.CasPool, error) {
	casPool := NewCasPoolBuilder().
		withSpcName(spc.Name).
		withCasTemplateName(spc.Annotations[string(v1alpha1.CreatePoolCASTemplateKey)]).
		withDiskType(spc.Spec.Type).
		withPoolType(spc.Spec.PoolSpec.PoolType).
		withAnnotations(spc.Annotations).
		withMaxPool(spc).
		Build()
	casPoolWithDisks, err := pc.withDisks(casPool, spc)
	if err != nil {
		return nil, err
	}
	return casPoolWithDisks, nil
}

// NewPoolCreateConfig returns an instance of pool create config.
func (c *Controller) NewPoolCreateConfig(spc *apis.StoragePoolClaim) *PoolCreateConfig {
	poolconfig := &PoolCreateConfig{
		nodeselect.NewConfig(spc),
		c,
	}
	return poolconfig
}

// NewCasPoolBuilder returns an empty instance of CasPoolBuilder.
func NewCasPoolBuilder() *CasPoolBuilder {
	return &CasPoolBuilder{
		CasPool: &apis.CasPool{},
	}
}

func (cb *CasPoolBuilder) withCasTemplateName(casTemplateName string) *CasPoolBuilder {
	//casTemplateName := spc.Annotations[string(v1alpha1.CreatePoolCASTemplateKey)]
	cb.CasPool.CasCreateTemplate = casTemplateName
	return cb
}

func (cb *CasPoolBuilder) withSpcName(name string) *CasPoolBuilder {
	cb.CasPool.StoragePoolClaim = name
	return cb
}

func (cb *CasPoolBuilder) withPoolType(poolType string) *CasPoolBuilder {
	cb.CasPool.PoolType = poolType
	return cb
}

func (cb *CasPoolBuilder) withMaxPool(spc *apis.StoragePoolClaim) *CasPoolBuilder {
	if isAutoProvisioning(spc) {
		cb.CasPool.MaxPools = *spc.Spec.MaxPools
	}
	return cb
}

func (cb *CasPoolBuilder) withDiskType(diskType string) *CasPoolBuilder {
	cb.CasPool.Type = diskType
	return cb
}

func (cb *CasPoolBuilder) withAnnotations(annotations map[string]string) *CasPoolBuilder {
	cb.CasPool.Annotations = annotations
	return cb
}

// Build returns an instance of cas pool object.
func (cb *CasPoolBuilder) Build() *apis.CasPool {
	return cb.CasPool
}

// poolCreateWorker is a worker function which will create a storagepool.
func poolCreateWorker(pool *apis.CasPool) error {

	glog.Infof("Creating storagepool for storagepoolclaim %s via CASTemplate", pool.StoragePoolClaim)

	storagepoolOps, err := storagepool.NewCasPoolOperation(pool)
	if err != nil {
		return errors.Wrapf(err, "NewCasPoolOperation failed error")
	}
	_, err = storagepoolOps.Create()
	if err != nil {
		return errors.Wrapf(err, "failed to create cas template based storagepool")

	}
	glog.Infof("Cas template based storagepool created successfully: name '%s'", pool.StoragePoolClaim)
	return nil
}

// withBlockDevices builds the CasPool object by filling details like
// blockDeviceList,nodeName etc.
// Some of the fields of the CasPool object is passed to CAS engine.
// CasPool object(type) is the contract on which CAS engine is instantiated for cStor pool creation.
func (pc *PoolCreateConfig) withDisks(casPool *apis.CasPool, spc *apis.StoragePoolClaim) (*apis.CasPool, error) {
	// getDiskList will hold node and block devices attached to it to be used for storagepool provisioning.
	_, err := pc.NodeBlockDeviceSelector()
	if err != nil {
		return nil, errors.Wrapf(err, "aborting storagepool create operation as no node qualified")
	}

	claimedNodeBDs, err := pc.ClaimBlockDevice(nil, spc)
	if err != nil {
		return nil, errors.Wrapf(err, "aborting storagepool create operation as no claimed block devices available")
	}

	// Fill the node name to the CasPool object.
	casPool.NodeName = claimedNodeBDs.NodeName
	//casPool.DiskList = nodeDisks.Disks.Items
	//TODO: Improve Following Code
	if spc.Spec.PoolSpec.PoolType == string(apis.PoolTypeStripedCPV) {
		for _, claimedBD := range claimedNodeBDs.BlockDeviceList {
			var bdList []apis.CspBlockDevice
			var group apis.BlockDeviceGroup
			blockDevice := apis.CspBlockDevice{
				Name:        claimedBD.BDName,
				InUseByPool: true,
				DeviceID:    claimedBD.DeviceID,
			}
			bdList = append(bdList, blockDevice)
			group = apis.BlockDeviceGroup{
				Item: bdList,
			}
			casPool.BlockDeviceList = append(casPool.BlockDeviceList, group)
		}
		return casPool, nil
	}
	count := blockdevice.DefaultDiskCount[spc.Spec.PoolSpec.PoolType]
	for i := 0; i < len(claimedNodeBDs.BlockDeviceList); i = i + count {
		var bdList []apis.CspBlockDevice
		var group apis.BlockDeviceGroup
		for j := 0; j < count; j++ {
			blockDevice := apis.CspBlockDevice{
				Name:        claimedNodeBDs.BlockDeviceList[i+j].BDName,
				InUseByPool: true,
				DeviceID:    claimedNodeBDs.BlockDeviceList[i+j].DeviceID,
			}
			bdList = append(bdList, blockDevice)
		}
		group = apis.BlockDeviceGroup{
			Item: bdList,
		}
		casPool.BlockDeviceList = append(casPool.BlockDeviceList, group)
	}
	return casPool, nil
}

// TODO: Move to block device package
func (pc *PoolCreateConfig) getDeviceID(blockDeviceName string) (string, error) {
	var deviceID string
	blockDevice, err := pc.BlockDeviceClient.Get(blockDeviceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if len(blockDevice.Spec.DevLinks) != 0 && len(blockDevice.Spec.DevLinks[0].Links) != 0 {
		deviceID = blockDevice.Spec.DevLinks[0].Links[0]
	} else {
		deviceID = blockDevice.Spec.Path
	}
	return deviceID, nil
}

func getSPCLabels(spc *apis.StoragePoolClaim) map[string]string {
	return map[string]string{string(apis.StoragePoolClaimCPK): spc.Name}
}

func getPatchData(oldObj, newObj interface{}) ([]byte, error) {
	oldData, err := json.Marshal(oldObj)
	if err != nil {
		return nil, fmt.Errorf("marshal old object failed: %v", err)
	}
	newData, err := json.Marshal(newObj)
	if err != nil {
		return nil, fmt.Errorf("mashal new object failed: %v", err)
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, oldObj)
	if err != nil {
		return nil, fmt.Errorf("CreateTwoWayMergePatch failed: %v", err)
	}
	return patchBytes, nil
}
