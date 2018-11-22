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

package debug

import (
	env "github.com/openebs/maya/pkg/env/v1alpha1"
)

// EnableCPUProfiling returns true if ENV OPENEBS_IO_DEBUG_PROFILE
// contains cpu profile mentioned
//
func EnableCPUProfiling() (truth bool) {
	return env.Matches(env.DebugProfileENVK, env.DebugProfileCPUENVKV)
}

// ProfilePath returns path value passed via ENV OPENEBS_IO_DEBUG_PROFILE_PATH
// Defaults to "/tmp"
//
func GetProfilePath() string {
	path := env.LookupOrFalse(env.DebugProfilePathENVK)
	if len(path) == 0 || path == "false" {
		path = "/tmp"
	}
	return path
}
