/*
Copyright 2018 The OpenEBS Authors

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

// TODO
// Remove this file if no longer required by installer !!

package v1alpha1

import (
	ver "github.com/openebs/maya/pkg/version"
)

type version string

const (
	version070     version = "0.7.0"
	invalidVersion version = "invalid.version"
)

// CurrentVersion returns openebs version
func CurrentVersion() version {
	return version(ver.Current())
}

// Version returns the version in version type if present
func Version(version string) version {
	switch version {
	case "0.7.0":
		return version070
	default:
		return invalidVersion
	}
}
