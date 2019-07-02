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

package v1alpha2

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

const (
	// DiskStateActive is the active state of the disks.
	DiskStateActive = "Active"
	// ProvisioningTypeManual is the manual provisioned SPC.
	ProvisioningTypeManual = "manual"
	// ProvisioningTypeAuto is the auto provisioned SPC.
	ProvisioningTypeAuto = "auto"
)

type blockDeviceList struct {
	Items []string
}

type nodeBlockDevice struct {
	NodeName string
}

// Config embeds clients for disk,csp and sp and contains, SPC object and ProvisioningType field which should tell
// provisioning type manual or auto.
type Config struct {
	// CSPC is the CStorPoolCluster object.
	CSPC *apis.CStorPoolCluster
	// Namespace is the namespace where openebs is installed
	Namespace string
}

// NewConfig returns an instance of Config based on SPC object.
func NewConfig(cspc *apis.CStorPoolCluster, ns string) *Config {
	return &Config{CSPC: cspc, Namespace: ns}
}
