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
	"github.com/pkg/errors"
)

// TODO : Improve following function
func (ac *Config) GetCSPSpec() (*apis.NewTestCStorPool, error) {
	poolSpec := ac.SelectNode()

	cspObj, err := apiscsp.NewBuilder().
		WithName(ac.CSPC.Name).
		WithNodeSelector(poolSpec.NodeSelector).
		WithPoolConfig(&poolSpec.PoolConfig).
		WithRaidGroups(poolSpec.RaidGroups).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build CSP object for node selector {%v}", poolSpec.NodeSelector)
	}

	err = ac.ClaimBDsForNode(ac.GetBDListForNode(poolSpec))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to claim block devices for node selector {%v}", poolSpec.NodeSelector)
	}
	return cspObj.Object, nil
}
