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

package volume

import (
	"fmt"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	"github.com/openebs/maya/types/v1"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO
// Should there be VolumeCreator, VolumeDeleter,
// VolumeReader, VolumeUpdater, etc ??
// volumeOperation or volumeOperator should
// be composed of these Interfaces. Good for mocking
// the callers which use these public methods.

type VolumeOperation struct {
	// volume to be provisioned
	volume *v1.Volume
	// k8sClient will make K8s API calls
	// This is useful for mocking purposes
	k8sClient *m_k8s_client.K8sClient
}

// VolumeOperation returns a new instance of volumeOperation
func NewVolumeOperation(volume *v1.Volume) (*VolumeOperation, error) {
	if volume == nil {
		return nil, fmt.Errorf("Nil volume: Can not instantiate 'Volume Operation'")
	}

	runNS := volume.Labels.K8sNamespace
	if len(runNS) == 0 {
		return nil, fmt.Errorf("Missing run namespace: Can not instantiate 'Volume Operation'")
	}

	// TODO
	// Check if k8sclient needs a namespace as its being used to
	// query StorageClass
	kc, err := m_k8s_client.NewK8sClient(runNS)
	if err != nil {
		return nil, err
	}

	return &VolumeOperation{
		volume:    volume,
		k8sClient: kc,
	}, nil
}

// Create provisions an OpenEBS volume
func (v *VolumeOperation) Create() (*v1.Volume, error) {
	if v.k8sClient == nil {
		return nil, fmt.Errorf("Nil k8s client: Can not create volume")
	}

	capacity := v.volume.Capacity
	if len(capacity) == 0 {
		capacity = v.volume.Labels.CapacityOld
	}

	if len(capacity) == 0 {
		return nil, fmt.Errorf("Missing volume capacity: Can not create volume")
	}

  // get the run namespace
	ns := v.volume.Namespace
	if len(ns) == 0 {
		ns = v.volume.Labels.K8sNamespace
	}

	// get the storage class name corresponding to this volume
	scName := v.volume.Labels.K8sStorageClass
	if len(scName) == 0 {
		return nil, fmt.Errorf("Missing k8s storage class: Can not create volume")
	}

	// fetch the storage class specifications
	sc, err := v.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// TODO
	// verify the provisioner & its
	// version that are set in the storage class

	// extract the volume policy name from storage class
	vpName := sc.Parameters[string(v1.VolumePolicyVK)]
	if len(vpName) == 0 {
		return nil, fmt.Errorf("Missing volume policy: Can not create volume")
	}

	// fetch the volume policy specifications
	vp, err := v.k8sClient.GetOEV1alpha1VP(vpName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// provision the volume by using the volume policy engine
	engine, err := PolicyEngine(vp, map[string]string{
		string(v1alpha1.OwnerVTP):                 v.volume.Name,
		string(v1alpha1.CapacityVTP):              capacity,
		string(v1alpha1.RunNamespaceVTP):          ns,
		string(v1alpha1.PersistentVolumeClaimVTP): v.volume.Labels.K8sPersistentVolumeClaim,
	})
	if err != nil {
		return nil, err
	}

	// create the volume
	anns, err := engine.execute()
	if err != nil {
		return nil, err
	}

	v.volume.Annotations = anns
	return v.volume, nil
}

func (v *VolumeOperation) Delete() {
	// ??
}

func (v *VolumeOperation) Read() {
	// ??
}

func (v *VolumeOperation) List() {
	// ??
}
