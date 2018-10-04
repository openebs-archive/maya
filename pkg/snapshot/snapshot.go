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
	"strings"

	yaml "github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	"github.com/openebs/maya/pkg/engine"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/openebs/maya/types/v1"
	"github.com/pkg/errors"
	v1_storage "k8s.io/api/storage/v1"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// snapshotOptions contains the options with respect to
// snapshot related operations
type snapshotOptions struct {
	// k8sClient will make K8s API calls
	k8sClient *m_k8s_client.K8sClient
}

// snapshot exposes methods with respect to snapshot related operations
// e.g. read, create, delete.
type snapshot struct {
	// snapshotOptions has the options to various snapshot related
	// operations
	snapshotOptions
	// snapshot to create or read or delete
	casSnapshot *v1alpha1.CASSnapshot
}

// snapshotList exposes methods to execute snapshot list operation
type snapshotList struct {
	// snapshotOptions has the options to various snapshot related
	// operations
	snapshotOptions
	// snapshots to list operation
	casSnapshots *v1alpha1.CASSnapshotList
}

// Snapshot returns a new instance of snapshot
func Snapshot(casSnapshot *v1alpha1.CASSnapshot) (*snapshot, error) {
	if casSnapshot == nil {
		return nil, errors.Errorf("failed to instantiate snapshot operation: nil snapshot was provided")
	}

	if len(casSnapshot.Namespace) == 0 {
		return nil, errors.Errorf("failed to instantiate snapshot operation: missing run namespace")
	}

	kc, err := m_k8s_client.NewK8sClient(casSnapshot.Namespace)
	if err != nil {
		return nil, err
	}

	return &snapshot{
		casSnapshot: casSnapshot,
		snapshotOptions: snapshotOptions{
			k8sClient: kc,
		},
	}, nil
}

// SnapshotList returns a new instance of snapshotList that is
// capable of listing snapshots
func SnapshotList(snapshots *v1alpha1.CASSnapshotList) (*snapshotList, error) {
	if snapshots == nil {
		return nil, errors.Errorf("failed to instantiate 'snapshot list operation': nil list options provided")
	}

	kc, err := m_k8s_client.NewK8sClient("")
	if err != nil {
		return nil, err
	}

	return &snapshotList{
		casSnapshots: snapshots,
		snapshotOptions: snapshotOptions{
			k8sClient: kc,
		},
	}, nil
}

// Create creates an OpenEBS snapshot of a volume
func (s *snapshot) Create() (*v1alpha1.CASSnapshot, error) {
	if s.k8sClient == nil {
		return nil, errors.Errorf("unable to create snapshot: nil k8s client")
	}

	// fetch the pv specifications
	pv, err := s.k8sClient.GetPV(s.casSnapshot.Spec.VolumeName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	scName := pv.Spec.StorageClassName
	if len(scName) == 0 {
		return nil, errors.Errorf("unable to create snapshot %s: missing storage class in PV %s", s.casSnapshot.Name, s.casSnapshot.Spec.VolumeName)
	}

	// fetch the storage class specifications
	sc, err := s.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	castName := getCreateCASTemplate(sc)
	if len(castName) == 0 {
		return nil, errors.Errorf("unable to create snapshot %s: missing cas template for create snapshot", s.casSnapshot.Name)
	}

	// fetch read cas template specifications
	cast, err := s.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	snapshotLables := map[string]interface{}{
		string(v1alpha1.OwnerVTP):        s.casSnapshot.Name,
		string(v1alpha1.VolumeSTP):       s.casSnapshot.Spec.VolumeName,
		string(v1alpha1.RunNamespaceVTP): s.casSnapshot.Namespace,
	}

	// provision CAS snapshot via CAS snapshot specific CAS template engine
	cc, err := SnapshotEngine(
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

// Get the openebs snapshot details
func (s *snapshot) Read() (*v1alpha1.CASSnapshot, error) {
	if s.k8sClient == nil {
		return nil, errors.Errorf("unable to read snapshot: nil k8s client")
	}

	// fetch the pv specifications
	pv, err := s.k8sClient.GetPV(s.casSnapshot.Spec.VolumeName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	scName := pv.Spec.StorageClassName
	if len(scName) == 0 {
		return nil, errors.Errorf("unable to read snapshot %s: missing storage class in PV %s", s.casSnapshot.Name, s.casSnapshot.Spec.VolumeName)
	}

	// fetch the storage class specifications
	sc, err := s.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	castName := getReadCASTemplate(sc)
	if len(castName) == 0 {
		return nil, errors.Errorf("unable to read snapshot %s: missing cas template for read snapshot", s.casSnapshot.Name)
	}

	// fetch read cas template specifications
	cast, err := s.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	snapshotLables := map[string]interface{}{
		string(v1alpha1.OwnerVTP):        s.casSnapshot.Name,
		string(v1alpha1.RunNamespaceVTP): s.casSnapshot.Namespace,
		string(v1alpha1.VolumeSTP):       s.casSnapshot.Spec.VolumeName,
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

// Get the openebs snapshot details
func (s *snapshot) Delete() (*v1alpha1.CASSnapshot, error) {
	if s.k8sClient == nil {
		return nil, errors.Errorf("unable to delete snapshot: nil k8s client")
	}

	// fetch the pv specifications
	pv, err := s.k8sClient.GetPV(s.casSnapshot.Spec.VolumeName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	scName := pv.Spec.StorageClassName
	if len(scName) == 0 {
		return nil, errors.Errorf("unable to delete snapshot %s: missing storage class in PV %s", s.casSnapshot.Name, s.casSnapshot.Spec.VolumeName)
	}

	// fetch the storage class specifications
	sc, err := s.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	castName := getReadCASTemplate(sc)
	if len(castName) == 0 {
		return nil, errors.Errorf("unable to delete snapshot %s: missing cas template for delete snapshot", s.casSnapshot.Name)
	}

	// fetch read cas template specifications
	cast, err := s.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	snapshotLables := map[string]interface{}{
		string(v1alpha1.OwnerVTP):        s.casSnapshot.Name,
		string(v1alpha1.RunNamespaceVTP): s.casSnapshot.Namespace,
		string(v1alpha1.VolumeSTP):       s.casSnapshot.Spec.VolumeName,
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

func (sl *snapshotList) List() (*v1alpha1.CASSnapshotList, error) {
	if sl.k8sClient == nil {
		return nil, errors.Errorf("unable to list snapshot: nil k8s client")
	}
	// fetch the pv specifications
	pv, err := sl.k8sClient.GetPV(sl.casSnapshots.Options.VolumeName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	scName := pv.Spec.StorageClassName
	if len(scName) == 0 {
		return nil, errors.Errorf("unable to list snapshot: missing storage class in PV %s", sl.casSnapshots.Options.VolumeName)
	}

	// fetch the storage class specifications
	sc, err := sl.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	castName := getListCASTemplate(sc)
	if len(castName) == 0 {
		return nil, errors.Errorf("unable to list snapshots: missing cas template for list snapshot")
	}

	// fetch read cas template specifications
	cast, err := sl.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	snapshotLables := map[string]interface{}{
		string(v1alpha1.RunNamespaceVTP): sl.casSnapshots.Options.Namespace,
		string(v1alpha1.VolumeSTP):       sl.casSnapshots.Options.VolumeName,
	}

	// list cas volume via cas template engine
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

func getReadCASTemplate(sc *v1_storage.StorageClass) string {
	castName := sc.Annotations[string(v1alpha1.CASTemplateKeyForSnapshotRead)]
	// if cas template for the given operation is empty then fetch from environment variables
	if len(castName) == 0 {
		casType := strings.ToLower(sc.Annotations[string(v1alpha1.CASTypeKey)])
		// check for casType, if cstor, set read cas template to cstor,
		// if jiva or absent then default to jiva
		if casType == string(v1.CStorVolumeType) {
			castName = menv.Get(menv.CASTemplateToReadCStorSnapshotENVK)
		} else if casType == string(v1.JivaVolumeType) || casType == "" {
			castName = menv.Get(menv.CASTemplateToReadJivaSnapshotENVK)
		}
	}
	return castName
}

func getCreateCASTemplate(sc *v1_storage.StorageClass) string {
	castName := sc.Annotations[string(v1alpha1.CASTemplateKeyForSnapshotCreate)]
	// if cas template for the given operation is empty then fetch from environment variables
	if len(castName) == 0 {
		casType := strings.ToLower(sc.Annotations[string(v1alpha1.CASTypeKey)])
		// check for casType, if cstor, set create cas template to cstor,
		// if jiva or absent then default to jiva
		if casType == string(v1.CStorVolumeType) {
			castName = menv.Get(menv.CASTemplateToCreateCStorSnapshotENVK)
		} else if casType == string(v1.JivaVolumeType) || casType == "" {
			castName = menv.Get(menv.CASTemplateToCreateJivaSnapshotENVK)
		}
	}
	return castName
}

func getDeleteCASTemplate(sc *v1_storage.StorageClass) string {
	castName := sc.Annotations[string(v1alpha1.CASTemplateKeyForSnapshotDelete)]
	// if cas template for the given operation is empty then fetch from environment variables
	if len(castName) == 0 {
		casType := strings.ToLower(sc.Annotations[string(v1alpha1.CASTypeKey)])
		// check for casType, if cstor, set delete cas template to cstor,
		// if jiva or absent then default to jiva
		if casType == string(v1.CStorVolumeType) {
			castName = menv.Get(menv.CASTemplateToDeleteCStorSnapshotENVK)
		} else if casType == string(v1.JivaVolumeType) || casType == "" {
			castName = menv.Get(menv.CASTemplateToDeleteJivaSnapshotENVK)
		}
	}
	return castName
}

func getListCASTemplate(sc *v1_storage.StorageClass) string {
	castName := sc.Annotations[string(v1alpha1.CASTemplateKeyForSnapshotList)]
	// if cas template for the given operation is empty then fetch from environment variables
	if len(castName) == 0 {
		casType := strings.ToLower(sc.Annotations[string(v1alpha1.CASTypeKey)])
		// check for casType, if cstor, set list cas template to cstor,
		// if jiva or absent then default to jiva
		if casType == string(v1.CStorVolumeType) {
			castName = menv.Get(menv.CASTemplateToListCStorSnapshotENVK)
		} else if casType == string(v1.JivaVolumeType) || casType == "" {
			castName = menv.Get(menv.CASTemplateToListJivaSnapshotENVK)
		}
	}
	return castName
}
