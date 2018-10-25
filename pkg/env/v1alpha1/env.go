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

package v1alpha1

import (
	"os"
	"strconv"
	"strings"
)

// ENVKey is a typed string to represent various environment keys used by this
// binary
type ENVKey string

const (
	// OpenEBSVersion is the environment variable to get openebs version
	OpenEBSVersion ENVKey = "OPENEBS_IO_VERSION"

	// OpenEBSNamespace is the environment variable to get openebs namespace
	//
	// This environment variable is set via kubernetes downward API
	OpenEBSNamespace ENVKey = "OPENEBS_NAMESPACE"

	// OpenEBSMayaPodName is the environment variable to get maya-apiserver pod
	// name
	//
	// This environment variable is set via kubernetes downward API
	OpenEBSMayaPodName ENVKey = "OPENEBS_MAYA_POD_NAME"

	// OpenEBSServiceAccount is the environment variable to get openebs
	// serviceaccount
	//
	// This environment variable is set via kubernetes downward API
	OpenEBSServiceAccount ENVKey = "OPENEBS_SERVICE_ACCOUNT"

	// TODO:
	//
	// The constants present here should be moved to respective/relevant packages
	// This file will hold env related operations as well as environment variables
	// that is common.
	//
	// All these might be moved to
	//
	// pkg/<entity_name>/v1alpha1/env_<entity_name>_<introduced_at_version>.go
	//
	// Need to discuss with team on above !!!
	//
	// TODO:
	//
	// The names of these variables should also change. It need not be suffixed
	// with ENVK. Need to discuss with team.

	// CASTemplateFeatureGateENVK is the ENV key to fetch cas template feature gate
	// i.e. if cas template based provisioning is enabled or disabled
	CASTemplateFeatureGateENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_FEATURE_GATE"

	// CASTemplateToListVolumeENVK is the ENV key that specifies the CAS Template
	// to list cas volumes
	CASTemplateToListVolumeENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_LIST_VOLUME"

	// CASTemplateToCreateJivaVolumeENVK is the ENV key that specifies the CAS Template
	// to create jiva cas volumes
	CASTemplateToCreateJivaVolumeENVK ENVKey = "OPENEBS_IO_JIVA_CAS_TEMPLATE_TO_CREATE_VOLUME"

	// CASTemplateToReadJivaVolumeENVK is the ENV key that specifies the CAS Template
	// to read jiva cas volumes
	CASTemplateToReadJivaVolumeENVK ENVKey = "OPENEBS_IO_JIVA_CAS_TEMPLATE_TO_READ_VOLUME"

	// CASTemplateToDeleteJivaVolumeENVK is the ENV key that specifies the CAS Template
	// to delete jiva cas volumes
	CASTemplateToDeleteJivaVolumeENVK ENVKey = "OPENEBS_IO_JIVA_CAS_TEMPLATE_TO_DELETE_VOLUME"

	// CASTemplateToCreateCStorVolumeENVK is the ENV key that specifies the CAS Template
	// to create cstor cas volumes
	CASTemplateToCreateCStorVolumeENVK ENVKey = "OPENEBS_IO_CSTOR_CAS_TEMPLATE_TO_CREATE_VOLUME"

	// CASTemplateToReadCStorVolumeENVK is the ENV key that specifies the CAS Template
	// to read cstor cas volumes
	CASTemplateToReadCStorVolumeENVK ENVKey = "OPENEBS_IO_CSTOR_CAS_TEMPLATE_TO_READ_VOLUME"

	// CASTemplateToDeleteCStorVolumeENVK is the ENV key that specifies the CAS Template
	// to delete cstor cas volumes
	CASTemplateToDeleteCStorVolumeENVK ENVKey = "OPENEBS_IO_CSTOR_CAS_TEMPLATE_TO_DELETE_VOLUME"

	// CASTemplateToCreatePoolENVK is the ENV key that specifies the CAS Template
	// to create storage pool
	CASTemplateToCreatePoolENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_CREATE_POOL"

	// CASTemplateToDeletePoolENVK is the ENV key that specifies the CAS Template
	// to delete storage pool
	CASTemplateToDeletePoolENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_DELETE_POOL"

	// CASTemplateToListCStorSnapshotENVK is the ENV key that specifies the CAS Template
	// to list cstor cas snapshots
	CASTemplateToListCStorSnapshotENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_LIST_CSTOR_SNAPSHOT"

	// CASTemplateToListJivaSnapshotENVK is the ENV key that specifies the CAS Template
	// to list jiva cas snapshots
	CASTemplateToListJivaSnapshotENVK ENVKey = "OPENEBS_IO_CAS_TEMPLATE_TO_LIST_JIVA_SNAPSHOT"

	// CASTemplateToCreateJivaSnapshotENVK is the ENV key that specifies the CAS Template
	// to create jiva cas snapshot
	CASTemplateToCreateJivaSnapshotENVK ENVKey = "OPENEBS_IO_JIVA_CAS_TEMPLATE_TO_CREATE_SNAPSHOT"

	// CASTemplateToReadJivaSnapshotENVK is the ENV key that specifies the CAS Template
	// to read jiva cas snapshot
	CASTemplateToReadJivaSnapshotENVK ENVKey = "OPENEBS_IO_JIVA_CAS_TEMPLATE_TO_READ_SNAPSHOT"

	// CASTemplateToDeleteJivaSnapshotENVK is the ENV key that specifies the CAS Template
	// to delete jiva cas snapshot
	CASTemplateToDeleteJivaSnapshotENVK ENVKey = "OPENEBS_IO_JIVA_CAS_TEMPLATE_TO_DELETE_SNAPSHOT"

	// CASTemplateToCreateCStorSnapshotENVK is the ENV key that specifies the CAS Template
	// to create cstor cas snapshot
	CASTemplateToCreateCStorSnapshotENVK ENVKey = "OPENEBS_IO_CSTOR_CAS_TEMPLATE_TO_CREATE_SNAPSHOT"

	// CASTemplateToReadCStorSnapshotENVK is the ENV key that specifies the CAS Template
	// to read cstor cas snapshot
	CASTemplateToReadCStorSnapshotENVK ENVKey = "OPENEBS_IO_CSTOR_CAS_TEMPLATE_TO_READ_SNAPSHOT"

	// CASTemplateToDeleteCStorSnapshotENVK is the ENV key that specifies the CAS Template
	// to delete cstor cas snapshot
	CASTemplateToDeleteCStorSnapshotENVK ENVKey = "OPENEBS_IO_CSTOR_CAS_TEMPLATE_TO_DELETE_SNAPSHOT"
)

// EnvironmentSetter abstracts setting of environment variable
type EnvironmentSetter func(envKey ENVKey, value string) (err error)

// EnvironmentGetter abstracts fetching value from an environment variable
type EnvironmentGetter func(envKey ENVKey) (value string)

// EnvironmentLookup abstracts looking up an environment variable
type EnvironmentLookup func(envKey ENVKey) (value string, present bool)

// Set sets the provided environment variable
//
// NOTE:
//  This is an implementation of EnvironmentSetter
func Set(envKey ENVKey, value string) (err error) {
	return os.Setenv(string(envKey), value)
}

// Get fetches value from the provided environment variable
//
// NOTE:
//  This is an implementation of EnvironmentGetter
func Get(envKey ENVKey) (value string) {
	return getEnv(string(envKey))
}

// Lookup looks up an environment variable
//
// NOTE:
//  This is an implementation of EnvironmentLookup
func Lookup(envKey ENVKey) (value string, present bool) {
	return lookupEnv(string(envKey))
}

// Truthy returns boolean based on the environment variable's value
//
// The lookup value can be truthy (i.e. 1, t, TRUE, true) or falsy (0, false,
// etc) based on strconv.ParseBool logic
func Truthy(envKey ENVKey) (truth bool) {
	v, found := Lookup(envKey)
	if !found {
		return
	}
	truth, _ = strconv.ParseBool(v)
	return
}

// LookupOrFalse looks up an environment variable and returns a string "false"
// if environment variable is not present. It returns appropriate values for
// other cases.
func LookupOrFalse(envKey ENVKey) string {
	val, present := Lookup(envKey)
	if !present {
		return "false"
	}
	return strings.TrimSpace(val)
}

// getEnv fetches the provided environment variable's value
func getEnv(envKey string) (value string) {
	return strings.TrimSpace(os.Getenv(envKey))
}

// lookupEnv looks up the provided environment variable
func lookupEnv(envKey string) (value string, present bool) {
	value, present = os.LookupEnv(envKey)
	value = strings.TrimSpace(value)
	return
}
