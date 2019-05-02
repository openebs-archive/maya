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

package v1alpha1

import (
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	caspool "github.com/openebs/maya/pkg/caspool/v1alpha1"
	"github.com/pkg/errors"
)

/*
Following is a sample manual CSPC YAML:

apiVersion: openebs.io/v1alpha1
kind: CStorPoolCluster
metadata:
  name: cstor-sparse-pool-test
spec:
  nodes:
  - name: gke-cstor-it-default-pool-569eb31d-3l88
    poolSpec:
      poolType: striped
    groups:
    - name: group1
      disks:
      - name: sparse-3c1fc7491f9e4cf50053730740647318
        id: disk1
  name: cstor-sparse-pool-test
  type: sparse
*/

// GetCasPoolForManualProvisioning returns a CasPool object for manual provisioned pool.
func (op *Operations) getCasPoolForManualProvisioning() (*apisv1alpha1.CasPool, error) {
	err := op.validateManual()
	if err != nil {
		return nil, errors.Wrapf(err, "validation failed")
	}

	casPool, err := op.buildCasPoolForManualProvisioning()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build cas pool object for spc %s", op.CspcObject.Object.Name)
	}
	return casPool, nil
}

// buildCasPoolForManualProvisioning builds a CasPool object for pool creation.
func (op *Operations) buildCasPoolForManualProvisioning() (*apisv1alpha1.CasPool, error) {
	nodeNames := op.CspcObject.GetNodeNames()
	usableNode, diskGroups, err := op.getUsableNode(nodeNames)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get usable node for spc %s", op.CspcObject.Object.Name)
	}

	diskDeviceIDMap, err := op.getDiskDeviceIDMap()
	if err != nil {
		return nil, errors.Wrapf(err, "could not form disk device ID map for %s", op.CspcObject.Object.Name)
	}

	spcObject := op.CspcObject.Object
	cp := caspool.NewBuilder().
		WithDiskType(spcObject.Spec.Type).
		WithPoolType(op.CspcObject.GetPoolTypeForNode(usableNode)).
		WithAnnotations(op.CspcObject.GetAnnotations()).
		WithDiskGroup(diskGroups).
		WithCasTemplateName(op.CspcObject.GetCASTName()).
		WithCspcName(op.CspcObject.Object.Name).
		WithNodeName(usableNode).
		WithDiskDeviceIDMap(diskDeviceIDMap).
		Build().Object
	return cp, nil
}

// getUsableNode returns a node and disks attached to it where pool can be possibly provisioned.
func (op *Operations) getUsableNode(nodes []string) (string, []apisv1alpha1.CStorPoolClusterDiskGroups, error) {
	usedNodes, err := op.getUsedNode()
	if err != nil {
		return "", []apisv1alpha1.CStorPoolClusterDiskGroups{}, err
	}
	for _, node := range nodes {
		if !usedNodes[node] {
			return node, op.CspcObject.GetDiskGroupListForNode(node), nil
		}
	}
	return "", []apisv1alpha1.CStorPoolClusterDiskGroups{}, errors.Errorf("no usable node found for spc %s", op.CspcObject.Object.Name)
}

// Validate does validations for nodes and disks present in the spc.
func (op *Operations) validateManual() error {
	err := NewOperationsBuilderForObject(op).
		WithCheckf(IsTypeValid(), "disk type is not valid").
		WithCheckf(IsDiskActive(), "non active disk found in cspc").
		WithCheckf(IsNodeNotRepeated(), "duplicate node entry in cspc").
		WithCheckf(IsDiskNotRepeated(), "duplicate disk entry in cspc").
		WithCheckf(IsPoolTypeOnNodeValid(), "invalid pool type in node spec").
		WithCheckf(IsDiskCountValid(), "invalid numbers of disk for given pool type on node").
		WithCheckf(IsNodeDiskRelationValid(), "some disk(s) does not belong to the specified node").
		Validate()
	if err != nil {
		return errors.Wrapf(err, "validation for cstorpoolcluster %s failed", op.CspcObject.Object.Name)
	}
	return nil
}
