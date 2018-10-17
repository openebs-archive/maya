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

// options contains the options with respect to
// snapshot related operations
type options struct {
	// k8sClient will make K8s API calls
	k8sClient   *m_k8s_client.K8sClient
	snapOptions *v1alpha1.SnapshotOptions
}

// snapshot exposes methods with respect to snapshot related operations
// e.g. read, create, delete.
type snapshot struct {
	// options has the options to various snapshot related
	// operations
	options
}

// Snapshot returns a new instance of snapshot
func Snapshot(opts *v1alpha1.SnapshotOptions) (*snapshot, error) {
	if len(opts.Namespace) == 0 {
		return nil, errors.Errorf("failed to instantiate snapshot operation: missing run namespace")
	}

	kc, err := m_k8s_client.NewK8sClient(opts.Namespace)
	if err != nil {
		return nil, err
	}

	return &snapshot{
		options: options{
			k8sClient:   kc,
			snapOptions: opts,
		},
	}, nil
}

// Create creates an OpenEBS snapshot of a volume
func (s *snapshot) Create() (*v1alpha1.CASSnapshot, error) {
	if s.k8sClient == nil {
		return nil, errors.Errorf("unable to create snapshot: nil k8s client")
	}

	// fetch the pv specifications
	pv, err := s.k8sClient.GetPV(s.snapOptions.VolumeName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	scName := pv.Spec.StorageClassName
	if len(scName) == 0 {
		return nil, errors.Errorf("unable to create snapshot %s: missing storage class in PV %s", s.snapOptions.Name, s.snapOptions.VolumeName)
	}

	// fetch the storage class specifications
	sc, err := s.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	castName := GetCreateCASTemplate(sc)
	if len(castName) == 0 {
		return nil, errors.Errorf("unable to create snapshot %s: missing cas template for create snapshot", s.snapOptions.Name)
	}

	// fetch read cas template specifications
	cast, err := s.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	snapshotLables := map[string]interface{}{
		string(v1alpha1.OwnerVTP):        s.snapOptions.Name,
		string(v1alpha1.VolumeSTP):       s.snapOptions.VolumeName,
		string(v1alpha1.RunNamespaceVTP): s.snapOptions.Namespace,
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
	pv, err := s.k8sClient.GetPV(s.snapOptions.VolumeName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	scName := pv.Spec.StorageClassName
	if len(scName) == 0 {
		return nil, errors.Errorf("unable to read snapshot %s: missing storage class in PV %s", s.snapOptions.Name, s.snapOptions.VolumeName)
	}

	// fetch the storage class specifications
	sc, err := s.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	castName := GetReadCASTemplate(sc)
	if len(castName) == 0 {
		return nil, errors.Errorf("unable to read snapshot %s: missing cas template for read snapshot", s.snapOptions.Name)
	}

	// fetch read cas template specifications
	cast, err := s.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	snapshotLables := map[string]interface{}{
		string(v1alpha1.OwnerVTP):        s.snapOptions.Name,
		string(v1alpha1.RunNamespaceVTP): s.snapOptions.Namespace,
		string(v1alpha1.VolumeSTP):       s.snapOptions.VolumeName,
	}

	// read cas volume via cas template engine
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
	pv, err := s.k8sClient.GetPV(s.snapOptions.VolumeName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	scName := pv.Spec.StorageClassName
	if len(scName) == 0 {
		return nil, errors.Errorf("unable to delete snapshot %s: missing storage class in PV %s", s.snapOptions.Name, s.snapOptions.VolumeName)
	}

	// fetch the storage class specifications
	sc, err := s.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	castName := GetDeleteCASTemplate(sc)
	if len(castName) == 0 {
		return nil, errors.Errorf("unable to delete snapshot %s: missing cas template for delete snapshot", s.snapOptions.Name)
	}

	// fetch read cas template specifications
	cast, err := s.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	snapshotLables := map[string]interface{}{
		string(v1alpha1.OwnerVTP):        s.snapOptions.Name,
		string(v1alpha1.RunNamespaceVTP): s.snapOptions.Namespace,
		string(v1alpha1.VolumeSTP):       s.snapOptions.VolumeName,
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

func (s *snapshot) List() (*v1alpha1.CASSnapshotList, error) {
	if s.k8sClient == nil {
		return nil, errors.Errorf("unable to list snapshot: nil k8s client")
	}
	// fetch the pv specifications
	pv, err := s.k8sClient.GetPV(s.snapOptions.VolumeName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	scName := pv.Spec.StorageClassName
	if len(scName) == 0 {
		return nil, errors.Errorf("unable to list snapshot: missing storage class in PV %s", s.snapOptions.VolumeName)
	}

	// fetch the storage class specifications
	sc, err := s.k8sClient.GetStorageV1SC(scName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	castName := GetListCASTemplate(sc)
	if len(castName) == 0 {
		return nil, errors.Errorf("unable to list snapshots: missing cas template for list snapshot")
	}

	// fetch read cas template specifications
	cast, err := s.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	snapshotLables := map[string]interface{}{
		string(v1alpha1.RunNamespaceVTP): s.snapOptions.Namespace,
		string(v1alpha1.VolumeSTP):       s.snapOptions.VolumeName,
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

func GetReadCASTemplate(sc *v1_storage.StorageClass) string {
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

func GetCreateCASTemplate(sc *v1_storage.StorageClass) string {
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

func GetDeleteCASTemplate(sc *v1_storage.StorageClass) string {
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

func GetListCASTemplate(sc *v1_storage.StorageClass) string {
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
