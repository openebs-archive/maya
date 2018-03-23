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

package task

import (
	"fmt"

	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
)

// TaskIdentity will provide the required identity to a task
type TaskIdentity struct {
	// Identifier provides a unique identification of this
	// task. There should not be two tasks with same identity
	// in a workflow.
	//
	// NOTE:
	//  Identity will be provided by the workflow
	Identity string
	// Kind of the task
	Kind string `json:"kind"`
	// APIVersion of the task
	APIVersion string `json:"apiVersion"`
}

// taskIdentifier enables operations w.r.t a task's identity
type taskIdentifier struct {
	// identity identifies a task
	identity TaskIdentity
}

func newTaskIdentifier(identity TaskIdentity) (taskIdentifier, error) {
	if len(identity.Identity) == 0 {
		return taskIdentifier{}, fmt.Errorf("missing task identity: can not create task identifier instance")
	}

	if len(identity.Kind) == 0 {
		return taskIdentifier{}, fmt.Errorf("missing task kind: can not create task identifier instance")
	}

	if len(identity.APIVersion) == 0 {
		return taskIdentifier{}, fmt.Errorf("missing task apiVersion: can not create task identifier instance")
	}

	return taskIdentifier{
		identity: identity,
	}, nil
}

func (i taskIdentifier) isPod() bool {
	return i.identity.Kind == string(m_k8s_client.PodKK)
}

func (i taskIdentifier) isDeployment() bool {
	return i.identity.Kind == string(m_k8s_client.DeploymentKK)
}

func (i taskIdentifier) isService() bool {
	return i.identity.Kind == string(m_k8s_client.ServiceKK)
}

func (i taskIdentifier) isStoragePool() bool {
	return i.identity.Kind == string(m_k8s_client.StroagePoolCRKK)
}

func (i taskIdentifier) isConfigMap() bool {
	return i.identity.Kind == string(m_k8s_client.ConfigMapKK)
}

func (i taskIdentifier) isPVC() bool {
	return i.identity.Kind == string(m_k8s_client.PersistentVolumeClaimKK)
}

func (i taskIdentifier) isExtnV1B1() bool {
	return i.identity.APIVersion == string(m_k8s_client.ExtensionsV1Beta1KA)
}

func (i taskIdentifier) isAppsV1B1() bool {
	return i.identity.APIVersion == string(m_k8s_client.AppsV1B1KA)
}

func (i taskIdentifier) isCoreV1() bool {
	return i.identity.APIVersion == string(m_k8s_client.CoreV1KA)
}

func (i taskIdentifier) isOEV1alpha1() bool {
	return i.identity.APIVersion == string(m_k8s_client.OEV1alpha1KA)
}

func (i taskIdentifier) isExtnV1B1Deploy() bool {
	return i.isExtnV1B1() && i.isDeployment()
}

func (i taskIdentifier) isAppsV1B1Deploy() bool {
	return i.isAppsV1B1() && i.isDeployment()
}

func (i taskIdentifier) isCoreV1Pod() bool {
	return i.isCoreV1() && i.isPod()
}

func (i taskIdentifier) isCoreV1Service() bool {
	return i.isCoreV1() && i.isService()
}

func (i taskIdentifier) isCoreV1PVC() bool {
	return i.isCoreV1() && i.isPVC()
}

func (i taskIdentifier) isOEV1alpha1SP() bool {
	return i.isOEV1alpha1() && i.isStoragePool()
}
