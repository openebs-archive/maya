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

package snapshot

import (
	"fmt"

	yaml "github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/openebs/maya/types/v1"
	"github.com/pkg/errors"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// snapshotOperationOptions contains the options with respect to
// snapshot related operations
type snapshotOperationOptions struct {
	// k8sClient will make K8s API calls
	k8sClient *m_k8s_client.K8sClient
}

// SnapshotOperation exposes methods with respect to snapshot related operations
// e.g. read, create, delete.
type SnapshotOperation struct {
	// snapshotOperationOptions has the options to various snapshot related
	// operations
	snapshotOperationOptions
	// snapshot to create or read or delete
	snapshot *v1alpha1.CASSnapshot
}

// NewSnapshotOperation returns a new instance of snapshotOperation
func NewSnapshotOperation(snapshot *v1alpha1.CASSnapshot) (*SnapshotOperation, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("failed to instantiate snapshot operation: nil snapshot was provided")
	}

	if len(snapshot.Namespace) == 0 {
		return nil, fmt.Errorf("failed to instantiate snapshot operation: missing run namespace")
	}

	kc, err := m_k8s_client.NewK8sClient(snapshot.Namespace)
	if err != nil {
		return nil, err
	}

	return &SnapshotOperation{
		snapshot: snapshot,
		snapshotOperationOptions: snapshotOperationOptions{
			k8sClient: kc,
		},
	}, nil
}

// Create creates an OpenEBS snapshot of a volume
func (v *SnapshotOperation) Create() (*v1alpha1.CASSnapshot, error) {
	if v.k8sClient == nil {
		return nil, fmt.Errorf("unable to create snapshot: nil k8s client")
	}

	castName := getCreateCASTemplate(v.snapshot.Spec.CasType)
	if castName == "" {
		return nil, errors.Errorf("unable to create snapshot: could not find castemplate for engine type %q", v.snapshot.Spec.CasType)
	}

	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	snapshotLables := map[string]interface{}{
		string(v1alpha1.OwnerVTP):        v.snapshot.Name,
		string(v1alpha1.RunNamespaceVTP): v.snapshot.Namespace,
		string(v1alpha1.VolumeSTP):       v.snapshot.Spec.VolumeName,
	}

	// provision CAS snapshot via CAS snapshot specific CAS template engine
	cc, err := NewCASSnapshotEngine(
		cast,
		string(v1alpha1.SnapshotTLP),
		snapshotLables,
	)
	if err != nil {
		return nil, err
	}

	// create the snapshot
	data, err := cc.Create()
	if err != nil {
		return nil, err
	}

	// unmarshall into openebs snapshot
	snap := &v1alpha1.CASSnapshot{}
	err = yaml.Unmarshal(data, snap)
	if err != nil {
		return nil, err
	}

	return snap, nil
}

// TODO: uncomment and update when snapshot deleteion,read,list is to be supported

/* func (v *SnapshotOperation) Delete() (*v1alpha1.CASSnapshot, error) {
	castName := getDeleteCASTemplate(v.snapshot.Spec.CasType)
	if len(castName) == 0 {
		return nil, fmt.Errorf("unable to delete snapshot %s: missing cas template for delete snapshot", v.snapshot.Name)
	}

	// fetch delete cas template specifications
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	snapshotLables := map[string]interface{}{
		string(v1alpha1.OwnerVTP):        v.snapshot.Name,
		string(v1alpha1.RunNamespaceVTP): v.snapshot.Namespace,
		string(v1alpha1.VolumeSTP):       v.snapshot.Spec.VolumeName,
	}

	// delete cas volume via cas template engine
	engine, err := engine.NewCASEngine(
		cast,
		string(v1alpha1.SnapshotTLP),
		snapshotLables,
	)
	if err != nil {
		return nil, err
	}

	// delete the cas volume
	data, err := engine.Delete()
	if err != nil {
		return nil, err
	}
	// unmarshall into openebs snapshot
	snap := &v1alpha1.CASSnapshot{}
	err = yaml.Unmarshal(data, snap)
	if err != nil {
		return nil, err
	}
	return snap, nil
}

// Get the openebs snapshot details
func (v *SnapshotOperation) Read() (*v1alpha1.CASSnapshot, error) {
	castName := getReadCASTemplate(v.snapshot.Spec.CasType)
	if len(castName) == 0 {
		return nil, fmt.Errorf("unable to read snapshot %s: missing cas template for read snapshot", v.snapshot.Name)
	}

	// fetch read cas template specifications
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	snapshotLables := map[string]interface{}{
		string(v1alpha1.OwnerVTP):        v.snapshot.Name,
		string(v1alpha1.RunNamespaceVTP): v.snapshot.Namespace,
		string(v1alpha1.VolumeSTP):       v.snapshot.Spec.VolumeName,
	}

	// delete cas volume via cas template engine
	engine, err := engine.NewCASEngine(
		cast,
		string(v1alpha1.SnapshotTLP),
		snapshotLables,
	)
	if err != nil {
		return nil, err
	}

	// read the cas snapshot
	data, err := engine.Read()
	if err != nil {
		return nil, err
	}
	// unmarshall into openebs snapshot
	snap := &v1alpha1.CASSnapshot{}
	err = yaml.Unmarshal(data, snap)
	if err != nil {
		return nil, err
	}
	return snap, nil
}

func (v *SnapshotListOperation) List() (*v1alpha1.CASSnapshotList, error) {
	castName := getListCASTemplate(v.snapshots.Spec.CasType)
	if len(castName) == 0 {
		return nil, fmt.Errorf("unable to list snapshots for volume %q: missing cas template for list snapshot", v.snapshots.Spec.VolumeName)
	}

	// fetch read cas template specifications
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	snapshotLables := map[string]interface{}{
		string(v1alpha1.RunNamespaceVTP): v.snapshots.Spec.Namespace,
		string(v1alpha1.VolumeSTP):       v.snapshots.Spec.VolumeName,
	}

	// delete cas volume via cas template engine
	engine, err := engine.NewCASEngine(
		cast,
		string(v1alpha1.SnapshotTLP),
		snapshotLables,
	)
	if err != nil {
		return nil, err
	}

	// list the cas snapshots
	data, err := engine.List()
	if err != nil {
		return nil, err
	}
	// unmarshall into openebs snapshot
	snapList := &v1alpha1.CASSnapshotList{}
	err = yaml.Unmarshal(data, snapList)
	if err != nil {
		return nil, err
	}
	return snapList, nil
}

// SnapshotListOperation exposes methods to execute snapshot list operation
type SnapshotListOperation struct {
	// snapshotOperationOptions has the options to various snapshot related
	// operations
	snapshotOperationOptions
	// snapshots to list operation
	snapshots *v1alpha1.CASSnapshotList
}

// NewSnapshotListOperation returns a new instance of SnapshotListOperation that is
// capable of listing snapshots
func NewSnapshotListOperation(snapshots *v1alpha1.CASSnapshotList) (*SnapshotListOperation, error) {
	if snapshots == nil {
		return nil, fmt.Errorf("failed to instantiate 'snapshot list operation': nil list options provided")
	}

	kc, err := m_k8s_client.NewK8sClient("")
	if err != nil {
		return nil, err
	}

	return &SnapshotListOperation{
		snapshots: snapshots,
		snapshotOperationOptions: snapshotOperationOptions{
			k8sClient: kc,
		},
	}, nil
}

func getDeleteCASTemplate(casType string) (castName string) {
	// check for casType, if cstor, set delete cas template to cstor,
	// if jiva or absent then default to jiva
	if casType == string(v1.CStorVolumeType) {
		castName = menv.Get(menv.CASTemplateToDeleteCStorSnapshotENVK)
	} else if casType == string(v1.JivaVolumeType) || casType == "" {
		castName = menv.Get(menv.CASTemplateToDeleteJivaSnapshotENVK)
	}
	return castName
}

func getReadCASTemplate(casType string) (castName string) {
	// check for casType, if cstor, set read cas template to cstor,
	// if jiva or absent then default to jiva
	if casType == string(v1.CStorVolumeType) {
		castName = menv.Get(menv.CASTemplateToReadCStorSnapshotENVK)
	} else if casType == string(v1.JivaVolumeType) || casType == "" {
		castName = menv.Get(menv.CASTemplateToReadJivaSnapshotENVK)
	}
	return castName
}

func getListCASTemplate(casType string) (castName string) {
	// check for casType, if cstor, set list cas template to cstor,
	// if jiva or absent then default to jiva
	if casType == string(v1.CStorVolumeType) {
		castName = menv.Get(menv.CASTemplateToListCStorSnapshotENVK)
	} else if casType == string(v1.JivaVolumeType) || casType == "" {
		castName = menv.Get(menv.CASTemplateToListJivaSnapshotENVK)
	}
	return castName
}
*/
func getCreateCASTemplate(casType string) (castName string) {
	// check for casType, if cstor, set create cas template to cstor,
	// if jiva or absent then default to jiva
	if casType == string(v1.CStorVolumeType) {
		castName = menv.Get(menv.CASTemplateToCreateCStorSnapshotENVK)
	} else if casType == string(v1.JivaVolumeType) || casType == "" {
		castName = menv.Get(menv.CASTemplateToCreateJivaSnapshotENVK)
	}
	return castName
}
