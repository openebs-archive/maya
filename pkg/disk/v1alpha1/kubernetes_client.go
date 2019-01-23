/*
Copyright 2019 The OpenEBS Authors

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

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Get is Kubernetes client implementation to get disk.
func (k *KubernetesClient) Get(name string) (*Disk, error) {
	d, err := k.Clientset.OpenebsV1alpha1().Disks().Get(name, v1.GetOptions{})

	return &Disk{d, nil}, err
}

// List is kubernetes client implementation to list disk.
func (k *KubernetesClient) List(opts v1.ListOptions) (*DiskList, error) {
	dl, err := k.Clientset.OpenebsV1alpha1().Disks().List(opts)
	return &DiskList{dl, nil}, err
}

// Create is kubernetes client implementation to create disk.
func (k *KubernetesClient) Create(diskObj *apis.Disk) (*Disk, error) {
	d, err := k.Clientset.OpenebsV1alpha1().Disks().Create(diskObj)
	return &Disk{d, nil}, err
}
