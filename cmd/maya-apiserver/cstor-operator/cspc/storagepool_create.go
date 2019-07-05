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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apiscsp "github.com/openebs/maya/pkg/cstor/newpool/v1alpha3"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/extnv1beta1/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	extnv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateStoragePool creates the required resource to provision a cStor pool
func (pc *PoolConfig) CreateStoragePool(cspcGot *apis.CStorPoolCluster) error {
	cspObj, err := pc.AlgorithConfig.GetCSPSpec()
	if err != nil {
		return errors.Wrap(err, "failed to get CSP spec")
	}
	err = pc.createCSP(cspObj)

	if err != nil {
		return errors.Wrapf(err, "failed to create csp for cspc {%s}", cspcGot.Name)
	}

	poolDeployObj := pc.GetPoolDeploySpec(cspObj)
	pc.createPoolDeployment(poolDeployObj)
	return nil
}

func (pc *PoolConfig) createCSP(csp *apis.NewTestCStorPool) error {
	_, err := apiscsp.NewKubeClient().WithNamespace(pc.AlgorithConfig.Namespace).Create(csp)
	return err
}

func (pc *PoolConfig) GetPoolDeploySpec(csp *apis.NewTestCStorPool) *extnv1beta1.Deployment {

	deployObj, _ := deploy.NewBuilder().
		WithName(csp.Name).
		WithNameSpace(csp.Namespace).
		WithAnnotations(map[string]string{}).
		WithLabels(map[string]string{}).
		WithOwnerReferences(csp).
		WithReplicaCount(getReplicaCount()).
		WithSelector(getSelector()).
		WithDeploymentStrategy(extnv1beta1.RecreateDeploymentStrategyType).
		WithPodTemplateSpec(getPodTemplateSpec()).Build()
	return deployObj.Object
}

func getReplicaCount() *int32 {
	var count int32 = 1
	return &count
}

// TODO: Fix following function -- ( currently only mocked)
func getSelector() *metav1.LabelSelector {
	return &metav1.LabelSelector{}
}

// TODO: Fix following function -- ( currently only mocked)
func getPodTemplateSpec() *v1.PodTemplateSpec {
	return &v1.PodTemplateSpec{}
}

func (pc *PoolConfig) createPoolDeployment(poolDeployObj *extnv1beta1.Deployment) error {
	_, err := deploy.KubeClient(deploy.WithNamespace(poolDeployObj.Namespace)).Create(poolDeployObj)
	return err
}
