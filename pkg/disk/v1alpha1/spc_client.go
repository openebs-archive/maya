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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Get is spc client implementation to get disk.
func (s *SpcObjectClient) Get(name string) (*Disk, error) {
	spcDiskList := s.Spc.Spec.Disks.DiskList
	var diskName string
	for _, disk := range spcDiskList {
		if name == disk {
			diskName = name
		}
	}
	if diskName == "" {
		return nil, errors.Errorf("Disk %s not found in the given SPC %s", diskName, s.Spc.Name)
	}
	d, err := s.NDMClientset.OpenebsV1alpha1().Disks().Get(diskName, v1.GetOptions{})
	return &Disk{d, nil}, err
}

// List is spc client implementation to list disk.
func (s *SpcObjectClient) List(opts v1.ListOptions) (*DiskList, error) {
	diskL := &DiskList{
		DiskList: &apis.DiskList{},
		errs:     nil,
	}
	var err error
	spcDiskList := s.Spc.Spec.Disks.DiskList
	if len(spcDiskList) == 0 {
		return nil, errors.Errorf("No disk found in the given SPC %s", s.Spc.Name)
	}
	spcDiskMap := make(map[string]int)
	for _, diskName := range spcDiskList {
		spcDiskMap[diskName]++
	}
	getAllDisk, err := s.NDMClientset.OpenebsV1alpha1().Disks().List(opts)
	if getAllDisk.Items == nil {
		return nil, errors.Wrapf(err, "Could not get disk from kube apiserver")
	}
	for _, disk := range getAllDisk.Items {
		if spcDiskMap[disk.Name] > 0 {
			diskL.DiskList.Items = append(diskL.DiskList.Items, disk)
		}
	}
	return diskL, err
}

// Create is kubernetes client implementation to create disk.
func (s *SpcObjectClient) Create(diskObj *apis.Disk) (*Disk, error) {
	return nil, errors.New("Disk object creation is not supported through spc client")
}
