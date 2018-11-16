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
	"github.com/openebs/maya/pkg/engine"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
	v1_storage "k8s.io/api/storage/v1"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// OperationOptions contains the options with respect to
// volume related operations
type OperationOptions struct {
	// k8sClient will make K8s API calls
	k8sClient *m_k8s_client.K8sClient
}

// Operation exposes methods with respect to volume related operations
// e.g. read, create, delete.
type Operation struct {
	// OperationOptions has the options to various volume related
	// operations
	OperationOptions
	// volume to create or read or delete
	volume *v1alpha1.CASVolume
}

// NewOperation returns a new instance of volumeOperation
func NewOperation(volume *v1alpha1.CASVolume) (*Operation, error) {
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

	return &Operation{
		volume: volume,
		OperationOptions: OperationOptions{
			k8sClient: kc,
		},
	}, nil
}

// getCloneLabels returns a map of clone specific configuration
func (v *Operation) getCloneLabels() (map[string]interface{}, error) {
	// Initially all the values are set to their defaults
	cloneLabels := map[string]interface{}{
		string(v1alpha1.SnapshotNameVTP):         "",
		string(v1alpha1.SourceVolumeTargetIPVTP): "",
		string(v1alpha1.IsCloneEnableVTP):        "false",
		string(v1alpha1.StorageClassVTP):         "",
		string(v1alpha1.SourceVolumeVTP):         "",
	}

	// if volume is clone enabled then update cloneLabels map
	if v.volume.CloneSpec.IsClone {
		// fetch source PV using client go
		pv, err := v.k8sClient.GetPV(v.volume.CloneSpec.SourceVolume, mach_apis_meta_v1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("source volume %q for clone volume %q could not be retrieved", v.volume.CloneSpec.SourceVolume, v.volume.Name)
		}
		// Set isCloneEnable to true
		cloneLabels[string(v1alpha1.IsCloneEnableVTP)] = "true"

		// extract and assign relevant clone spec fields to cloneLabels
		cloneLabels[string(v1alpha1.SnapshotNameVTP)] = v.volume.CloneSpec.SnapshotName
		cloneLabels[string(v1alpha1.SourceVolumeTargetIPVTP)] = strings.TrimSpace(strings.Split(pv.Spec.ISCSI.TargetPortal, ":")[0])
		cloneLabels[string(v1alpha1.StorageClassVTP)] = v.volume.Labels[string(v1alpha1.StorageClassKey)]
		cloneLabels[string(v1alpha1.SourceVolumeVTP)] = v.volume.CloneSpec.SourceVolume
	}
	return cloneLabels, nil
}

// Create provisions an OpenEBS volume
func (v *Operation) Create() (*v1alpha1.CASVolume, error) {
	if v.k8sClient == nil {
		return nil, fmt.Errorf("unable to create volume: nil k8s client")
	}

	capacity := v.volume.Spec.Capacity

	if len(capacity) == 0 {
		return nil, fmt.Errorf("unable to create volume: missing volume capacity")
	}

	pvcName := v.volume.Labels[string(v1alpha1.PersistentVolumeClaimKey)]
	if len(pvcName) == 0 {
		return nil, fmt.Errorf("unable to create volume: missing persistent volume claim")
	}

	// fetch the pvc specifications
	pvc, err := v.k8sClient.GetPVC(pvcName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// extract the cas volume config from pvc
	casConfigPVC := pvc.Annotations[string(v1alpha1.CASConfigKey)]

	cloneLabels, err := v.getCloneLabels()
	if err != nil {
		return nil, err
	}
	scName := v.volume.Labels[string(v1alpha1.StorageClassKey)]

	if len(scName) == 0 {
		return nil, fmt.Errorf("unable to create volume: missing storage class")
	}

	// scName might not be initialized in getCloneLabels
	// assign the latest available scName
	cloneLabels[string(v1alpha1.StorageClassVTP)] = scName

	// fetch the storage class specifications
	sc, err := v.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// extract the cas volume config from storage class
	casConfigSC := sc.Annotations[string(v1alpha1.CASConfigKey)]

	// cas template to create a cas volume
	castName := getCreateCASTemplate("", sc)
	if len(castName) == 0 {
		return nil, fmt.Errorf("unable to create volume: missing create cas template at '%s'", v1alpha1.CASTemplateKeyForVolumeCreate)
	}

	// fetch CASTemplate specifications
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	volumeLabels := map[string]interface{}{
		string(v1alpha1.OwnerVTP):                 v.volume.Name,
		string(v1alpha1.CapacityVTP):              capacity,
		string(v1alpha1.RunNamespaceVTP):          v.volume.Namespace,
		string(v1alpha1.PersistentVolumeClaimVTP): pvcName,
	}

	runtimeVolumeValues := util.MergeMaps(volumeLabels, cloneLabels)

	// provision CAS volume via CAS volume specific CAS template engine
	engine, err := NewVolumeEngine(
		casConfigPVC,
		casConfigSC,
		cast,
		string(v1alpha1.VolumeTLP),
		runtimeVolumeValues,
	)
	if err != nil {
		return nil, err
	}

	// create the volume
	data, err := engine.Run()
	if err != nil {
		return nil, err
	}

	// unmarshall result into openebs volume
	vol := &v1alpha1.CASVolume{}
	err = yaml.Unmarshal(data, vol)
	if err != nil {
		return nil, err
	}

	return vol, nil
}

// Delete removes a CASVolume
func (v *Operation) Delete() (*v1alpha1.CASVolume, error) {
	if len(v.volume.Name) == 0 {
		return nil, fmt.Errorf("unable to delete volume: volume name not provided")
	}

	// pv details
	pv, err := v.k8sClient.GetPV(v.volume.Name, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// sc details
	scName := pv.Labels[string(v1alpha1.StorageClassKey)]
	if len(scName) == 0 {
		scName = pv.Spec.StorageClassName
	}
	if len(scName) == 0 {
		return nil, fmt.Errorf("failed to delete volume '%s': missing storage class in PV object", v.volume.Name)
	}
	sc, err := v.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	casConfigSC := sc.Annotations[string(v1alpha1.CASConfigKey)]

	// cas template corresponding to this volume
	casType := pv.Labels[string(v1alpha1.CASTypeKey)]
	castName := getDeleteCASTemplate(casType, sc)
	if len(castName) == 0 {
		return nil, fmt.Errorf("failed to delete volume '%s': missing cas template", v.volume.Name)
	}
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// delete cas volume via cas template engine
	engine, err := NewVolumeEngine(
		"",
		casConfigSC,
		cast,
		string(v1alpha1.VolumeTLP),
		map[string]interface{}{
			string(v1alpha1.OwnerVTP):        v.volume.Name,
			string(v1alpha1.RunNamespaceVTP): v.volume.Namespace,
		},
	)
	if err != nil {
		return nil, err
	}

	// delete volume
	data, err := engine.Run()
	if err != nil {
		return nil, err
	}

	// unmarshall result into openebs volume
	vol := &v1alpha1.CASVolume{}
	err = yaml.Unmarshal(data, vol)
	if err != nil {
		return nil, err
	}

	return vol, nil
}

// Get the openebs volume details
func (v *Operation) Read() (*v1alpha1.CASVolume, error) {
	if len(v.volume.Name) == 0 {
		return nil, fmt.Errorf("unable to read volume: volume name not provided")
	}

	// storage engine type
	storageEngine := ""

	// check if sc name is already present, if not then extract it
	scName := v.volume.Labels[string(v1alpha1.StorageClassKey)]
	if len(scName) == 0 {
		// fetch the pv specification
		pv, err := v.k8sClient.GetPV(v.volume.Name, mach_apis_meta_v1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// extract the sc name
		scName = strings.TrimSpace(pv.Spec.StorageClassName)

		// extract the storage engine
		storageEngine = pv.Labels[string(v1alpha1.CASTypeKey)]
	}

	if len(scName) == 0 {
		return nil, fmt.Errorf("unable to read volume '%s': missing storage class name", v.volume.Name)
	}

	// fetch the sc specification
	sc, err := v.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// extract read cas template name from sc annotation
	castName := getReadCASTemplate(storageEngine, sc)
	if len(castName) == 0 {
		return nil, fmt.Errorf("unable to read volume '%s': missing cas template for read '%s'", v.volume.Name, v1alpha1.CASTemplateKeyForVolumeRead)
	}

	// fetch read cas template specifications
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// read cas volume via cas template engine
	engine, err := engine.New(
		cast,
		string(v1alpha1.VolumeTLP),
		map[string]interface{}{
			string(v1alpha1.OwnerVTP):        v.volume.Name,
			string(v1alpha1.RunNamespaceVTP): v.volume.Namespace,
		},
	)
	if err != nil {
		return nil, err
	}

	// read volume details by executing engine
	data, err := engine.Run()
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

// ListOperation exposes methods to execute volume list operation
type ListOperation struct {
	// OperationOptions has the options to various volume related
	// operations
	OperationOptions
	// volumes to list operation
	volumes *v1alpha1.CASVolumeList
}

// NewListOperation returns a new instance of ListOperation that is
// capable of listing volumes
func NewListOperation(volumes *v1alpha1.CASVolumeList) (*ListOperation, error) {
	if volumes == nil {
		return nil, fmt.Errorf("failed to instantiate 'volume list operation': nil list options provided")
	}

	kc, err := m_k8s_client.NewK8sClient("")
	if err != nil {
		return nil, err
	}

	return &ListOperation{
		volumes: volumes,
		OperationOptions: OperationOptions{
			k8sClient: kc,
		},
	}, nil
}

// List returns a list of CASVolumeList
func (v *ListOperation) List() (*v1alpha1.CASVolumeList, error) {
	// cas template to list cas volumes
	castNames := menv.Get(menv.CASTemplateToListVolumeENVK)
	if len(castNames) == 0 {
		return nil, fmt.Errorf("failed to list volume: cas template to list volume is not set as environment variable")
	}
	vols := &v1alpha1.CASVolumeList{
		Items: []v1alpha1.CASVolume{},
	}

	for _, castName := range strings.Split(castNames, ",") {
		// fetch read cas template specifications
		cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// read cas volume via cas template engine
		engine, err := engine.New(
			cast,
			string(v1alpha1.VolumeTLP),
			map[string]interface{}{
				string(v1alpha1.RunNamespaceVTP): v.volumes.Namespace,
			},
		)
		if err != nil {
			return nil, err
		}

		// list volume details by executing engine
		data, err := engine.Run()
		if err != nil {
			return nil, err
		}

		// unmarshall into openebs volume
		tvols := &v1alpha1.CASVolumeList{}
		err = yaml.Unmarshal(data, tvols)
		if err != nil {
			return nil, err
		}

		vols.Items = append(vols.Items, tvols.Items...)
	}
	return vols, nil
}

func getCreateCASTemplate(defaultCasType string, sc *v1_storage.StorageClass) string {
	castName := sc.Annotations[string(v1alpha1.CASTemplateKeyForVolumeCreate)]
	// if cas template for the given operation is empty then fetch from environment variables
	if len(castName) == 0 {
		casType := strings.ToLower(sc.Annotations[string(v1alpha1.CASTypeKey)])
		// if casType is missing in sc annotation then use the default cas type
		if casType == "" {
			casType = strings.ToLower(defaultCasType)
		}
		// check for cas-type, if cstor, set create cas template to cstor,
		// if jiva or for jiva and if absent then default to jiva
		if casType == string(v1.CStorVolumeType) {
			castName = menv.Get(menv.CASTemplateToCreateCStorVolumeENVK)
		} else if casType == string(v1.JivaVolumeType) || casType == "" {
			castName = menv.Get(menv.CASTemplateToCreateJivaVolumeENVK)
		}
	}
	return castName
}

func getReadCASTemplate(defaultCasType string, sc *v1_storage.StorageClass) string {
	castName := sc.Annotations[string(v1alpha1.CASTemplateKeyForVolumeRead)]
	// if cas template for the given operation is empty then fetch from environment variables
	if len(castName) == 0 {
		casType := strings.ToLower(sc.Annotations[string(v1alpha1.CASTypeKey)])
		// if casType is missing in sc annotation then use the default cas type
		if casType == "" {
			casType = strings.ToLower(defaultCasType)
		}
		// check for cas-type, if cstor, set create cas template to cstor,
		// if jiva or for jiva and if absent then default to jiva
		if casType == string(v1.CStorVolumeType) {
			castName = menv.Get(menv.CASTemplateToReadCStorVolumeENVK)
		} else if casType == string(v1.JivaVolumeType) || casType == "" {
			castName = menv.Get(menv.CASTemplateToReadJivaVolumeENVK)
		}
	}
	return castName
}

func getDeleteCASTemplate(defaultCasType string, sc *v1_storage.StorageClass) string {
	castName := sc.Annotations[string(v1alpha1.CASTemplateKeyForVolumeDelete)]
	// if cas template for the given operation is empty then fetch from environment variables
	if len(castName) == 0 {
		casType := strings.ToLower(sc.Annotations[string(v1alpha1.CASTypeKey)])
		// if casType is missing in sc annotation then use the default cas type
		if casType == "" {
			casType = strings.ToLower(defaultCasType)
		}
		// check for cas-type, if cstor, set create cas template to cstor,
		// if jiva or for jiva and if absent then default to jiva
		if casType == string(v1.CStorVolumeType) {
			castName = menv.Get(menv.CASTemplateToDeleteCStorVolumeENVK)
		} else if casType == string(v1.JivaVolumeType) || casType == "" {
			castName = menv.Get(menv.CASTemplateToDeleteJivaVolumeENVK)
		}
	}
	return castName
}
