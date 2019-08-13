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
	"github.com/golang/glog"
	nodeselect "github.com/openebs/maya/pkg/algorithm/nodeselect/v1alpha2"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apiscsp "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DownScalePool deletes the required pool.
func (pc *PoolConfig) DownScalePool() error {
	orphanedCSP, err := pc.getOrphanedCStorPools()
	if err != nil {
		pc.Controller.recorder.Event(pc.AlgorithmConfig.CSPC, corev1.EventTypeWarning,
			"DownScale", "Pool downscale failed "+err.Error())
		return errors.Wrap(err, "could not get orphaned CSP(s)")
	}
	for _, cspName := range orphanedCSP {
		pc.Controller.recorder.Event(pc.AlgorithmConfig.CSPC, corev1.EventTypeNormal,
			"DownScale", "De-provisioning pool "+cspName)

		// TODO : As part of deleting a CSP, do we need to delete associated BDCs ?

		err := apiscsp.NewKubeClient().WithNamespace(pc.AlgorithmConfig.Namespace).Delete(cspName, &metav1.DeleteOptions{})
		if err != nil {
			pc.Controller.recorder.Event(pc.AlgorithmConfig.CSPC, corev1.EventTypeWarning,
				"DownScale", "De-provisioning pool "+cspName+"failed")
			glog.Errorf("De-provisioning pool %s failed: %s", cspName, err)
		}
	}
	return nil
}

// getOrphanedCStorPools returns a list CSP names that should be deleted.
// TODO : Move to algorithm package
func (pc *PoolConfig) getOrphanedCStorPools() ([]string, error) {
	var orphanedCSP []string
	nodePresentOnCSPC, err := pc.getNodePresentOnCSPC()
	if err != nil {
		return []string{}, errors.Wrap(err, "could not get node names of pool config present on CSPC")
	}
	cspList, err := apiscsp.NewKubeClient().WithNamespace(pc.AlgorithmConfig.Namespace).List(
		metav1.ListOptions{LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + pc.AlgorithmConfig.CSPC.Name})

	if err != nil {
		return []string{}, errors.Wrap(err, "could not list CSP(s)")
	}

	for _, cspObj := range cspList.Items {
		cspObj := cspObj
		if nodePresentOnCSPC[cspObj.Spec.HostName] {
			continue
		}
		orphanedCSP = append(orphanedCSP, cspObj.Name)
	}
	return orphanedCSP, nil
}

// getNodePresentOnCSPC returns a map of node names where pool shoul
// be present.
// TODO: Improve method name.
// TODO: Move to CSPC package
func (pc *PoolConfig) getNodePresentOnCSPC() (map[string]bool, error) {
	nodeMap := make(map[string]bool)
	for _, pool := range pc.AlgorithmConfig.CSPC.Spec.Pools {
		nodeName, err := nodeselect.GetNodeFromLabelSelector(pool.NodeSelector)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not get node name for node selector {%v} "+
					"from cspc %s", pool.NodeSelector, pc.AlgorithmConfig.CSPC.Name)
		}
		nodeMap[nodeName] = true
	}
	return nodeMap, nil
}
