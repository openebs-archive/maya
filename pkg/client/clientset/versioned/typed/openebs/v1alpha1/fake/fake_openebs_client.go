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
package fake

import (
	v1alpha1 "github.com/openebs/maya/pkg/client/clientset/versioned/typed/openebs/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeOpenebsV1alpha1 struct {
	*testing.Fake
}

func (c *FakeOpenebsV1alpha1) CStorPools() v1alpha1.CStorPoolInterface {
	return &FakeCStorPools{c}
}

func (c *FakeOpenebsV1alpha1) CStorVolumeReplicas() v1alpha1.CStorVolumeReplicaInterface {
	return &FakeCStorVolumeReplicas{c}
}

func (c *FakeOpenebsV1alpha1) StoragePools() v1alpha1.StoragePoolInterface {
	return &FakeStoragePools{c}
}

func (c *FakeOpenebsV1alpha1) StoragePoolClaims() v1alpha1.StoragePoolClaimInterface {
	return &FakeStoragePoolClaims{c}
}

func (c *FakeOpenebsV1alpha1) VolumeParameterGroups() v1alpha1.VolumeParameterGroupInterface {
	return &FakeVolumeParameterGroups{c}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeOpenebsV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
