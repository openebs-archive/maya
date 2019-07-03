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
	// HostName holds the hostname key for kubernetes node.
	HostName = "kubernetes.io/hostname"
)

// Config embeds CSPC object and namespace where openebs is installed.
type Config struct {
	// CSPC is the CStorPoolCluster object.
	CSPC *apis.CStorPoolCluster
	// Namespace is the namespace where openebs is installed.
	Namespace string
}

// NewConfig returns an instance of Config based on CSPC object.
func NewConfig(cspc *apis.CStorPoolCluster, ns string) *Config {
	return &Config{CSPC: cspc, Namespace: ns}
}
