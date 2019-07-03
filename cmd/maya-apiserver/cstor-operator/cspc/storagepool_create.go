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

package cspc

import (
	nodeselect "github.com/openebs/maya/pkg/algorithm/nodeselect/v1alpha2"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apiscsp "github.com/openebs/maya/pkg/cstor/newpool/v1alpha3"
	"github.com/pkg/errors"
	"k8s.io/api/apps/v1"
)

func (c *Controller) CreateStoragePool(cspcGot *apis.CStorPoolCluster) error {
	newAlgorithmConfig := nodeselect.NewConfig(cspcGot, "openebs")
	cspObj, err := newAlgorithmConfig.GetCSPSpec()
	if err != nil {
		return errors.Wrap(err, "failed to get CSP spec")
	}
	err = c.createCSP(cspObj)

	if err != nil {
		return errors.Wrapf(err, "failed to create csp for cspc {%s}", cspcGot.Name)
	}

	poolDeployObj := c.GetPoolDeploySpec(cspObj)
	c.createPoolDeployment(poolDeployObj)
	return nil
}

func (c *Controller) createCSP(csp *apis.NewTestCStorPool) error {
	_, err := apiscsp.NewKubeClient().WithNamespace("openebs").Create(csp)
	if err != nil {
		return err
	}
	return nil
}

// TODO: Fix following function -- ( currently only mocked)
func (c *Controller) GetPoolDeploySpec(csp *apis.NewTestCStorPool) *v1.Deployment {
	return &v1.Deployment{}
}
func (c *Controller) createPoolDeployment(poolDeployObj *v1.Deployment) {

}
