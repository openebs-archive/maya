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

package storagepool

import (
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// casPoolOperationOptions contains the options with respect to
// storagepool related operations
type casPoolOperationOptions struct {
	// runNamespace is the namespace where storagepool operation will happen
	//runNamespace string
	// k8sClient will make K8s API calls
	k8sClient *m_k8s_client.K8sClient
}

// casPoolOperation exposes methods with respect to storagepool related operations
// e.g. read, create, delete.
type casPoolOperation struct {
	// casPoolOperationOptions has the options to various storagepool related
	// operations
	casPoolOperationOptions
	// pool to create or read or delete
	pool *v1alpha1.CasPool
}

// NewCasPoolOperation returns a new instance of casPoolOperation
func NewCasPoolOperation(pool *v1alpha1.CasPool) (*casPoolOperation, error) {
	if pool == nil {
		return nil, fmt.Errorf("Failed to instantiate storagepool operation: nil storagepool was provided")
	}

	kc, err := m_k8s_client.NewK8sClient(pool.Namespace)
	if err != nil {
		return nil, err
	}
	// Put pool object inside casPoolOperation object
	return &casPoolOperation{
		pool: pool,
		casPoolOperationOptions: casPoolOperationOptions{
			k8sClient: kc,
		},
	}, nil
}

// Create provisions an OpenEBS storagePool
func (v *casPoolOperation) Create() (*v1alpha1.CasPool, error) {
	if v.k8sClient == nil {
		return nil, fmt.Errorf("Unable to create storagepool: nil k8s client")
	}
	// cas template to create a storagepool
	castName := v.pool.CasCreateTemplate
	if len(castName) == 0 {
		// get default create CAS template to create storagepool from ENV variable
		castName = menv.Get(menv.CASTemplateToCreatePoolENVK)
	}
	if len(castName) == 0 {
		return nil, fmt.Errorf("Unable to create storagepool: missing create cas template")
	}
	// extract the cas openebs config from storagepoolclaim
	openebsConfig := v.pool.Annotations[string(v1alpha1.CASConfigKey)]
	// fetch CASTemplate specifications
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// provision cas storagepool via cas template engine
	cc, err := NewStoragePoolEngine(
		cast,
		openebsConfig,
		string(v1alpha1.StoragePoolTLP),
		map[string]interface{}{
			string(v1alpha1.OwnerCTP):             v.pool.StoragePoolClaim,
			string(v1alpha1.BlockDeviceListCTP):   v.pool.BlockDeviceList,
			string(v1alpha1.NodeNameCTP):          v.pool.NodeName,
			string(v1alpha1.PoolTypeCTP):          v.pool.PoolType,
			string(v1alpha1.BlockDeviceIDListCTP): v.pool.DeviceID,
		},
	)
	if err != nil {
		return nil, err
	}

	// create the storagePool
	data, err := cc.Create()
	if err != nil {
		return nil, err
	}

	// unmarshall into openebs storagepool
	pool := &v1alpha1.CasPool{}
	err = yaml.Unmarshal(data, pool)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func (v *casPoolOperation) Delete() (*v1alpha1.CasPool, error) {
	if len(v.pool.StoragePoolClaim) == 0 {
		return nil, fmt.Errorf("Unable to delete storagepool: storagepoolclaim name not provided")
	}

	// cas template to delete a storagepool
	castName := v.pool.CasDeleteTemplate
	if len(castName) == 0 {
		// get default delete CAS template to delete storagepool from ENV variable
		castName = menv.Get(menv.CASTemplateToDeletePoolENVK)
	}
	if len(castName) == 0 {
		return nil, fmt.Errorf("unable to delete storagepool: no cas template for delete found")
	}

	// fetch delete cas template specifications
	castObj, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// delete storagepool via cas template engine
	engine, err := cast.Engine(
		castObj,
		string(v1alpha1.StoragePoolTLP),
		map[string]interface{}{
			string(v1alpha1.OwnerCTP): v.pool.StoragePoolClaim,
		},
	)
	if err != nil {
		return nil, err
	}

	// delete CasPool by executing engine
	data, err := engine.Run()
	if err != nil {
		return nil, err
	}

	// unmarshall into openebs CasPool
	pool := &v1alpha1.CasPool{}
	err = yaml.Unmarshal(data, pool)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

// StoragePoolOperation holds the instance of StoragePool related operations
type StoragePoolOperation struct {
	poolName  string
	k8sClient *m_k8s_client.K8sClient
}

// NewStoragePoolOperation returns a new instance of StoragePoolOperation
func NewStoragePoolOperation(poolName string) (*StoragePoolOperation, error) {
	kc, err := m_k8s_client.NewK8sClient("")
	if err != nil {
		return nil, err
	}
	// Put pool object inside casPoolOperation object
	return &StoragePoolOperation{
		poolName:  poolName,
		k8sClient: kc,
	}, err
}

// List returns the list of storagepools
func (s *StoragePoolOperation) List() (*v1alpha1.CStorPoolList, error) {
	if s.k8sClient == nil {
		return nil, fmt.Errorf("unable to fetch K8s client")
	}

	// get CATemplate name from env
	castName := menv.Get(menv.CASTemplateToListStoragePoolENVK)

	// fetch read cas template specifications
	castObj, err := s.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// create new instance on CASEngine
	engine, err := cast.Engine(
		castObj,
		"",
		map[string]interface{}{},
	)
	if err != nil {
		return nil, err
	}

	// fetch data from engine execution
	data, err := engine.Run()
	if err != nil {
		return nil, err
	}

	// unmarshall into StoragePoolList
	sPool := &v1alpha1.CStorPoolList{}
	err = json.Unmarshal(data, sPool)
	if err != nil {
		return nil, err
	}
	return sPool, nil
}

// Read returns the list of storagepools
func (s *StoragePoolOperation) Read() (*v1alpha1.CStorPool, error) {
	if s.k8sClient == nil {
		return nil, fmt.Errorf("unable to fetch K8s client")
	}

	// get CATemplate name from env
	castName := menv.Get(menv.CASTemplateToReadStoragePoolENVK)

	// fetch read cas template specifications
	castObj, err := s.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// create new instance on CASEngine
	engine, err := cast.Engine(
		castObj,
		string(v1alpha1.StoragePoolTLP),
		map[string]interface{}{
			string(v1alpha1.OwnerCTP): s.poolName,
		},
	)
	if err != nil {
		return nil, err
	}

	// fetch data from engine execution
	data, err := engine.Run()
	if err != nil {
		return nil, err
	}

	// unmarshall into StoragePool
	sPool := &v1alpha1.CStorPool{}
	err = json.Unmarshal(data, sPool)
	if err != nil {
		return nil, err
	}
	return sPool, nil
}
