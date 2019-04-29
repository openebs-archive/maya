// Copyright © 2017-2019 The OpenEBS Authors
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

package v1

const (
	// BytesToGB used to convert bytes to GB
	BytesToGB = 1073741824
	// BytesToMB used to convert bytes to MB
	BytesToMB = 1048567
	// BytesToKB used to convert bytes to KB
	BytesToKB = 1024
	// MicSec used to convert to microsec to second
	MicSec = 1000000
	// MinWidth used in tabwriter
	MinWidth = 0
	// MaxWidth used in tabwriter
	MaxWidth = 0
	// Padding used in tabwriter
	Padding = 4

	// ControllerPort : Jiva volume controller listens on this for various api
	// requests.
	ControllerPort string = ":9501"
	// InfoAPI is the api for getting the volume access modes.
	InfoAPI string = "/replicas"
	// ReplicaPort : Jiva volume replica listens on this for various api
	// requests
	ReplicaPort string = ":9502"
	// StatsAPI is api to query about the volume stats from both replica and
	// controller.
	StatsAPI string = "/stats"
)
