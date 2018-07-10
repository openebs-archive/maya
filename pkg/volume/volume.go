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

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// volumeOperationOptions contains the options with respect to
// volume related operations
type volumeOperationOptions struct {
	// k8sClient will make K8s API calls
	k8sClient *m_k8s_client.K8sClient
}

// VolumeOperation exposes methods with respect to volume related operations
// e.g. read, create, delete.
type VolumeOperation struct {
	// volumeOperationOptions has the options to various volume related
	// operations
	volumeOperationOptions
	// volume to create or read or delete
	volume *v1alpha1.CASVolume
}

// NewVolumeOperation returns a new instance of volumeOperation
func NewVolumeOperation(volume *v1alpha1.CASVolume) (*VolumeOperation, error) {
	if volume == nil {
		return nil, fmt.Errorf("failed to instantiate volume operation: nil volume was provided")
	}

	if len(volume.Namespace) == 0 {
		return nil, fmt.Errorf("failed to instantiate volume operation: missing run namespace")
	}

	kc, err := m_k8s_client.NewK8sClient(volume.Namespace)
	if err != nil {
		return nil, err
	}

	return &VolumeOperation{
		volume: volume,
		volumeOperationOptions: volumeOperationOptions{
			k8sClient: kc,
		},
	}, nil
}

// Create provisions an OpenEBS volume
func (v *VolumeOperation) Create() (*v1alpha1.CASVolume, error) {
	if v.k8sClient == nil {
		return nil, fmt.Errorf("unable to create volume: nil k8s client")
	}

	capacity := v.volume.Spec.Capacity
	if len(capacity) == 0 {
		capacity = v.volume.Labels[string(v1alpha1.CapacityCVDK)]
	}

	if len(capacity) == 0 {
		return nil, fmt.Errorf("unable to create volume: missing volume capacity")
	}

	// TODO
	//
	// UnComment below once provisioner is able to send name of PVC
	//
	// pvc name corresponding to this volume
	//pvcName := v.volume.Labels[string(v1alpha1.PersistentVolumeClaimCVK)]
	//if len(pvcName) == 0 {
	//	return nil, fmt.Errorf("unable to create volume: missing persistent volume claim")
	//}

	// fetch the pvc specifications
	//pvc, err := v.k8sClient.GetPVC(pvcName, mach_apis_meta_v1.GetOptions{})
	//if err != nil {
	//	return nil, err
	//}

	// extract the cas volume config from pvc
	//casConfigPVC := pvc.Annotations[string(v1alpha1.CASConfigCVK)]

	// TODO
	//
	// Remove below two lines once provisioner is able to send name of PVC
	pvcName := ""
	casConfigPVC := ""

	// get the storage class name corresponding to this volume
	scName := v.volume.Labels[string(v1alpha1.StorageClassCVK)]
	if len(scName) == 0 {
		return nil, fmt.Errorf("unable to create volume: missing storage class")
	}

	// fetch the storage class specifications
	sc, err := v.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// extract the cas volume config from storage class
	casConfigSC := sc.Annotations[string(v1alpha1.CASConfigCVK)]

	// cas template to create a cas volume
	castName := sc.Annotations[string(v1alpha1.CASTemplateCVK)]
	if len(castName) == 0 {
		return nil, fmt.Errorf("unable to create volume: missing create cas template at '%s'", v1alpha1.CASTemplateCVK)
	}

	// fetch CASTemplate specifications
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// provision cas volume via cas template engine
	cc, err := NewCASCreate(
		casConfigPVC,
		casConfigSC,
		cast,
		map[string]string{
			string(v1alpha1.OwnerVTP):                 v.volume.Name,
			string(v1alpha1.CapacityVTP):              capacity,
			string(v1alpha1.RunNamespaceVTP):          v.volume.Namespace,
			string(v1alpha1.PersistentVolumeClaimVTP): pvcName,
		},
	)
	if err != nil {
		return nil, err
	}

	// create the volume
	data, err := cc.create()
	if err != nil {
		return nil, err
	}

	// unmarshall into openebs volume
	vol := &v1alpha1.CASVolume{}
	err = yaml.Unmarshal(data, vol)
	if err != nil {
		return nil, err
	}

	return vol, nil
}

func (v *VolumeOperation) Delete() (*v1alpha1.CASVolume, error) {
	if len(v.volume.Name) == 0 {
		return nil, fmt.Errorf("unable to delete volume: volume name not provided")
	}

	// cas template to delete a cas volume
	castName := v.volume.Annotations[string(v1alpha1.CASTemplateForDeleteCVK)]
	if len(castName) == 0 {
		// use the default delete cas template otherwise
		castName = string(v1alpha1.CASTemplateForDeleteCVD)
	}

	// fetch delete cas template specifications
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// delete cas volume via cas template engine
	engine, err := NewCASEngine(
		cast,
		map[string]string{
			string(v1alpha1.OwnerVTP):        v.volume.Name,
			string(v1alpha1.RunNamespaceVTP): v.volume.Namespace,
		},
	)
	if err != nil {
		return nil, err
	}

	// delete the cas volume
	data, err := engine.delete()
	if err != nil {
		return nil, err
	}

	// unmarshall into openebs volume
	vol := &v1alpha1.CASVolume{}
	err = yaml.Unmarshal(data, vol)
	if err != nil {
		return nil, err
	}

	return vol, nil
}

// Get the openebs volume details
func (v *VolumeOperation) Read() (*v1alpha1.CASVolume, error) {
	if len(v.volume.Name) == 0 {
		return nil, fmt.Errorf("unable to read volume: volume name not provided")
	}

	// cas template to read a cas volume
	castName := v.volume.Annotations[string(v1alpha1.CASTemplateForReadCVK)]
	if len(castName) == 0 {
		// use the default read cas template otherwise
		castName = string(v1alpha1.CASTemplateForReadCVD)
	}

	// fetch read cas template specifications
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// read cas volume via cas template engine
	engine, err := NewCASEngine(
		cast,
		map[string]string{
			string(v1alpha1.OwnerVTP):        v.volume.Name,
			string(v1alpha1.RunNamespaceVTP): v.volume.Namespace,
		},
	)
	if err != nil {
		return nil, err
	}

	// read the volume details
	data, err := engine.read()
	if err != nil {
		return nil, err
	}

	// unmarshall into openebs volume
	vol := &v1alpha1.CASVolume{}
	err = yaml.Unmarshal(data, vol)
	if err != nil {
		return nil, err
	}

	return vol, nil
}

// VolumeListOperation exposes methods to execute volume list operation
type VolumeListOperation struct {
	// volumeOperationOptions has the options to various volume related
	// operations
	volumeOperationOptions
	// volumes to list operation
	volumes *v1alpha1.CASVolumeList
}

// NewVolumeListOperation returns a new instance of VolumeListOperation that is
// capable of listing volumes
func NewVolumeListOperation(volumes *v1alpha1.CASVolumeList) (*VolumeListOperation, error) {
	if volumes == nil {
		return nil, fmt.Errorf("failed to instantiate 'volume list operation': nil list options provided")
	}

	kc, err := m_k8s_client.NewK8sClient("")
	if err != nil {
		return nil, err
	}

	return &VolumeListOperation{
		volumes: volumes,
		volumeOperationOptions: volumeOperationOptions{
			k8sClient: kc,
		},
	}, nil
}

func (v *VolumeListOperation) List() (*v1alpha1.CASVolumeList, error) {
	// cas template to list cas volumes
	castName := v.volumes.Annotations[string(v1alpha1.CASTemplateForListCVK)]
	if len(castName) == 0 {
		// use the default list cas template otherwise
		castName = string(v1alpha1.CASTemplateForListCVD)
	}

	// fetch read cas template specifications
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// read cas volume via cas template engine
	engine, err := NewCASEngine(
		cast,
		map[string]string{
			string(v1alpha1.RunNamespaceVTP): v.volumes.Namespace,
		},
	)
	if err != nil {
		return nil, err
	}

	// read the volume details
	data, err := engine.list()
	if err != nil {
		return nil, err
	}

	// unmarshall into openebs volume
	vols := &v1alpha1.CASVolumeList{}
	err = yaml.Unmarshal(data, vols)
	if err != nil {
		return nil, err
	}

	return vols, nil
}
