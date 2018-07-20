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

// cstorPool OperationOptions contains the options with respect to
// cstorPool related operations
type cstorPoolOperationOptions struct {
	// runNamespace is the namespace where cstorPool operation will happen
	//runNamespace string
	// k8sClient will make K8s API calls
	k8sClient *m_k8s_client.K8sClient
}

// cstorPoolOperation exposes methods with respect to cstorPool related operations
// e.g. read, create, delete.
type cstorPoolOperation struct {
	// cstorPoolOperationOptions has the options to various cstorPool related
	// operations
	cstorPoolOperationOptions
	// cstorPool to create or read or delete
	cstorPool *v1alpha1.CStorPool
}

// NewCstorPoolOperation returns a new instance of cstorPoolOperation
func NewCstorPoolOperation(cstorPool *v1alpha1.CStorPool) (*cstorPoolOperation, error) {
	if cstorPool == nil {
		return nil, fmt.Errorf("Failed to instantiate cstorpool operation: nil cstorpool was provided")
	}

	kc, err := m_k8s_client.NewK8sClient(cstorPool.Namespace)
	if err != nil {
		return nil, err
	}
	// Put cstor pool object inside cstorPoolOperation object
	return &cstorPoolOperation{
		cstorPool: cstorPool,
		cstorPoolOperationOptions: cstorPoolOperationOptions{
			k8sClient: kc,
		},
	}, nil
}

// Create provisions an OpenEBS cstorPool
func (v *cstorPoolOperation) Create() (*v1alpha1.CStorPool, error) {
	if v.k8sClient == nil {
		return nil, fmt.Errorf("Unable to create cstorpool: nil k8s client")
	}

	castName := v.cstorPool.Annotations[string(v1alpha1.SPCASTemplateCK)]
	if len(castName) == 0 {
		//return nil, fmt.Errorf("unable to create cstorPool: missing create cas template at '%s'", v1alpha1.CASTemplateCVK)
		return nil, fmt.Errorf("Unable to create cstorpool: missing create cas template")
	}

	// fetch CASTemplate specifications
	//cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	cast, err := v.k8sClient.GetOEV1alpha1CAST(castName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// provision cas cstorPool via cas template engine
	cc, err := NewCASStoragePoolEngine(
		cast,
		string(v1alpha1.CstorPoolTLP),
		map[string]interface{}{
			string(v1alpha1.OwnerCTP):    v.cstorPool.Name,
			string(v1alpha1.PoolTypeCTP):     v.cstorPool.Spec.PoolSpec.PoolType,
		},
	)
	if err != nil {
		return nil, err
	}

	// create the cstorPool
	data, err := cc.Create()
	if err != nil {
		return nil, err
	}

	// unmarshall into openebs cstorPool
	cstorPool := &v1alpha1.CStorPool{}
	err = yaml.Unmarshal(data, cstorPool)
	if err != nil {
		return nil, err
	}
	return cstorPool, nil
}

func (v *cstorPoolOperation) Delete() (*v1alpha1.CStorPool, error) {
	if len(v.cstorPool.Name) == 0 {
		return nil, fmt.Errorf("Unable to delete cstorpool: cstor pool name not provided")
	}

	// cas template to delete a cstor pool
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

	// delete cstor pool via cas template engine
	engine, err := engine.NewCASEngine(
		cast,
		string(v1alpha1.CstorPoolTLP),
		map[string]interface{}{
			string(v1alpha1.OwnerCTP):    v.cstorPool.Name,
		},
	)
	if err != nil {
		return nil, err
	}

	// delete the cas cstorPool
	data, err := engine.Delete()
	if err != nil {
		return nil, err
	}

	// unmarshall into openebs cstorPool
	cstorPool := &v1alpha1.CStorPool{}
	err = yaml.Unmarshal(data, cstorPool)
	if err != nil {
		return nil, err
	}
	return cstorPool, nil
}