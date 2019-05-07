// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

// Config struct fills the parameters of request or user input
type Config struct {
	// DriverName to be registered at CSI
	DriverName string
	// PluginType helps in specifying whether it is a node plugin or controller
	// Identity has to be run with both controller and node plugin
	// Same binry contains all the three plugins
	PluginType string
	// Version specifies the version of the CSI controller/node driver
	Version string
	// Endpoint on which requests are made by kubelet or external provisioner
	// Controller/node plugin will listen on this
	// This will be a unix based socket
	Endpoint string
	// NodeID helps in differentiating the nodes on which node deivers are
	// running. This is useful in case of topologies and publishing /
	// unpublishing volumes on nodes
	NodeID string
	// A REST Server is exposed on this URL for internal operations and Day2-ops
	RestURL string
}

//NewConfig returns config struct to initialize new driver
func NewConfig() *Config {
	return &Config{}
}
