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

// MetaTaskIdentity will provide the required identity to a task
type MetaTaskIdentity struct {
	// Identity provides a unique identification to this
	// task.
	//
	// NOTE:
	//  Usage: There should not be two tasks with same identity
	// in a cas template engine run.
	Identity string `json:"id"`
	// Kind of the task
	Kind string `json:"kind"`
	// APIVersion of the task
	APIVersion string `json:"apiVersion"`
}

// taskIdentifier enables operations w.r.t a task's identity
type taskIdentifier struct {
	// identity identifies a task
	identity MetaTaskIdentity
}

func newTaskIdentifier(identity MetaTaskIdentity) (taskIdentifier, error) {
	if len(identity.Identity) == 0 {
		return taskIdentifier{}, fmt.Errorf("failed to create task identifier instance: task id is missing")
	}

	if len(identity.Kind) == 0 {
		return taskIdentifier{}, fmt.Errorf("failed to create task identifier instance: task kind is missing")
	}

	if len(identity.APIVersion) == 0 {
		return taskIdentifier{}, fmt.Errorf("failed to create task identifier instance: task apiVersion is missing")
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

func (i taskIdentifier) isStoragePoolClaim() bool {
	return i.identity.Kind == string(m_k8s_client.StroagePoolClaimCRKK)
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

func (i taskIdentifier) isDisk() bool {
	return i.identity.Kind == string(m_k8s_client.DiskCRKK)
}

func (i taskIdentifier) isCstorPool() bool {
	return i.identity.Kind == string(m_k8s_client.CStorPoolCRKK)
}

func (i taskIdentifier) isCstorVolume() bool {
	return i.identity.Kind == string(m_k8s_client.CStorVolumeCRKK)
}

func (i taskIdentifier) isCstorVolumeReplica() bool {
	return i.identity.Kind == string(m_k8s_client.CStorVolumeReplicaCRKK)
}

func (i taskIdentifier) isOEV1alpha1Disk() bool {
	return i.isOEV1alpha1() && i.isDisk()
}

func (i taskIdentifier) isOEV1alpha1CSP() bool {
	return i.isOEV1alpha1() && i.isCstorPool()
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
func (i taskIdentifier) isOEV1alpha1SPC() bool {
	return i.isOEV1alpha1() && i.isStoragePoolClaim()
}

func (i taskIdentifier) isOEV1alpha1SP() bool {
	return i.isOEV1alpha1() && i.isStoragePool()
}

func (i taskIdentifier) isOEV1alpha1CV() bool {
	return i.isOEV1alpha1() && i.isCstorVolume()
}

func (i taskIdentifier) isOEV1alpha1CVR() bool {
	return i.isOEV1alpha1() && i.isCstorVolumeReplica()
}
