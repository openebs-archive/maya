/*
Copyright 2017 The OpenEBS Authors.

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
package v1alpha1

// CasPool is a type which will be utilised by CAS engine to perform
// storagepool related operation

type CasPool struct {
	// StoragePoolClaim is the name of the storagepoolclaim object
	StoragePoolClaim string

	// CasCreateTemplate is the cas template that will be used for storagepool create
	// operation
	CasCreateTemplate string

	// CasDeleteTemplate is the cas template that will be used for storagepool delete
	// operation
	CasDeleteTemplate string

	// Namespace can be passed via storagepoolclaim as labels to decide on the
	// execution of namespaced resources with respect to storagepool
	Namespace string
}
