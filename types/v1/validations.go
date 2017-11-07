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

package v1

// IsVolumeType verifies if the provided VolumeType is a
// valid volume type or not.
//
// An empty vType is considered as valid whereas an invalid
// valued vType is considered as invalid.
func IsVolumeType(vType VolumeType) bool {
	if vType == JivaVolumeType {
		return true
	} //else if vType == CStorVolumeType {
	//return true
	//}

	if len(vType) != 0 {
		return false
	}

	// an empty volume type is valid as it may not have
	// been set
	return true
}

// IsOrchProvider verifies if the provided Orchestrator is a
// valid volume orchestrator provider or not.
//
// An empty op is considered as valid whereas an invalid
// valued op is considered as invalid.
func IsOrchProvider(op OrchProvider) bool {
	if op == K8sOrchProvider {
		return true
	}

	if len(op) != 0 {
		return false
	}

	// an empty volume type is valid as it may not have
	// been set
	return true
}
