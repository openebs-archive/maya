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
	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apiscsp "github.com/openebs/maya/pkg/cstor/newpool/v1alpha3"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPendingPoolCount returns the pending pool count that should be created for the
// given CStorPoolCluster.
func (c *Config) GetPendingPoolCount() (int, error) {
	currentPoolCount, err := c.GetCurrentPoolCount()
	if err != nil {
		return 0, errors.Wrapf(err, "unable to get pending pool count for cspc %s", c.CSPC.Name)
	}
	desiredPoolCount := len(c.CSPC.Spec.Pools)

	return (desiredPoolCount - currentPoolCount), nil
}

// GetCurrentPoolCount give the current pool count for the given CStorPoolCluster.
func (c *Config) GetCurrentPoolCount() (int, error) {
	cspList, err := apiscsp.NewKubeClient().WithNamespace(c.Namespace).List(metav1.ListOptions{LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + c.CSPC.Name})
	if err != nil {
		return 0, errors.Errorf("unable to get current pool count:unable to list cstor pools: %v", err)
	}
	return len(cspList.Items), nil
}

// IsPoolPending returns true if pool is pending for creation.
func (c *Config) IsPoolPending() bool {
	pc, err := c.GetPendingPoolCount()
	if err != nil {
		glog.Errorf("unable to get pending pool count : %v", err)
		return false
	}
	if pc > 0 {
		return true
	}
	return false
}
