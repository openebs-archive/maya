/*
Copyright 2019 The OpenEBS Authors

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

package v1alpha2

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apiscsp "github.com/openebs/maya/pkg/cstor/newpool/v1alpha3"
	"github.com/openebs/maya/pkg/version"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	// CasTypeCStor is the key for cas type cStor
	CasTypeCStor = "cstor"
)

// GetCSPSpec returns a CSP spec that should be created and claims all the
// block device present in the CSP spec
func (ac *Config) GetCSPSpec() (*apis.NewTestCStorPool, error) {
	poolSpec, nodeName, err := ac.SelectNode()
	if err != nil || nodeName == "" {
		return nil, errors.Wrap(err, "failed to select a node")
	}
	csplabels := ac.buildLabelsForCSP(nodeName)
	cspObj, err := apiscsp.NewBuilder().
		WithName(ac.CSPC.Name + "-" + rand.String(4)).
		WithNamespace(ac.Namespace).
		WithNodeSelectorByReference(poolSpec.NodeSelector).
		WithNodeName(nodeName).
		WithPoolConfig(&poolSpec.PoolConfig).
		WithRaidGroups(poolSpec.RaidGroups).
		WithCSPCOwnerReference(ac.CSPC).
		WithLabelsNew(csplabels).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build CSP object for node selector {%v}", poolSpec.NodeSelector)
	}

	err = ac.ClaimBDsForNode(ac.GetBDListForNode(poolSpec))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to claim block devices for node {%s}", nodeName)
	}
	return cspObj.Object, nil
}

// buildLabelsForCSP builds labels for CSP
// TODO : Improve following using builders
func (ac *Config) buildLabelsForCSP(nodeName string) map[string]string {
	labels := make(map[string]string)
	labels[HostName] = nodeName
	labels[string(apis.CStorPoolClusterCPK)] = ac.CSPC.Name
	labels[string(apis.OpenEBSVersionKey)] = version.GetVersion()
	labels[string(apis.CASTypeKey)] = CasTypeCStor
	return labels
}
