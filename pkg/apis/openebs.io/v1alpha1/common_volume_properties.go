/*
Copyright 2018 The OpenEBS Authors.

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

// CapacitySpec wraps the Capacity property
//
// NOTE: It is intended to be used as embedded struct
type CapacitySpec struct {
	// Capacity will hold the capacity of this Volume e.g. 5GB
	Capacity string `json:"capacity"`
}

// TargetIPSpec wraps the TargetIP property
//
// NOTE: It is intended to be used as embedded struct
type TargetIPSpec struct {
	// TargetIP will hold the targetIP for this Volume
	TargetIP string `json:"targetIP"`
}

// TargetPortalSpec wraps the TargetPortal property
//
// NOTE: It is intended to be used as embedded struct
type TargetPortalSpec struct {
	// TargetPortal will hold the target portal for this volume
	TargetPortal string `json:"targetPortal"`
}

// TargetPortSpec wraps the TargetPort property
//
// NOTE: It is intended to be used as embedded struct
type TargetPortSpec struct {
	// TargetPort will hold the targetPort for this Volume eg. 3260
	TargetPort int `json:"targetPort"`
}

// IQNSpec wraps the iqn property
//
// NOTE: It is intended to be used as embedded struct
type IQNSpec struct {
	Iqn string `json:"iqn"`
}
