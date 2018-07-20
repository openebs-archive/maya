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
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/openebs/maya/pkg/engine"
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

	castName := v.pool.CasCreateTemplate
	if len(castName) == 0 {
		return nil, fmt.Errorf("Unable to create storagepool: missing create cas template")
	}

	// fetch CASTemplate specifications
	//cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// provision cas storagepool via cas template engine
	cc, err := NewCASStoragePoolEngine(
		cast,
		string(v1alpha1.StoragePoolTLP),
		map[string]interface{}{
			string(v1alpha1.OwnerCTP):    v.pool.StoragePoolClaim,
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
	// Need to decide on correct way of passing it.
	castName := "cstor-pool-delete-cast"
	if len(castName) == 0 {
		// use the default delete cas template otherwise
		//castName = string(v1alpha1.CASTemplateForDeleteCVD)
		fmt.Println("No CAS template for delete")
	}

	// fetch delete cas template specifications
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// delete storagepool via cas template engine
	engine, err := engine.NewCASEngine(
		cast,
		string(v1alpha1.StoragePoolTLP),
		map[string]interface{}{
			string(v1alpha1.OwnerCTP):    v.pool.StoragePoolClaim,
		},
	)
	if err != nil {
		return nil, err
	}

	// delete the CasPool
	data, err := engine.Delete()
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