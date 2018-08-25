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
	"strings"
)

// ENVKey is a typed string to represent various environment keys
type ENVKey string

const (
	// EnvKeyForOpenEBSNamespace is the environment variable to get
	// openebs namespace
	EnvKeyForOpenEBSNamespace ENVKey = "OPENEBS_NAMESPACE"
	// EnvKeyForOpenEBSServiceAccount is the environment variable to get
	// openebs serviceaccount
	EnvKeyForOpenEBSServiceAccount ENVKey = "OPENEBS_SERVICE_ACCOUNT"
)

// EnvironmentGetter abstracts fetching value from an environment variable
type EnvironmentGetter func(envKey string) (value string)

// Get fetches value from the provided environment variable
//
// NOTE:
//  This is an implementation of EnvironmentGetter
func Get(envKey string) (value string) {
	return getEnv(envKey)
}

// EnvironmentLookup abstracts looking up an environment variable
type EnvironmentLookup func(envKey string) (value string, present bool)

// Lookup looks up an environment variable
//
// NOTE:
//  This is an implementation of EnvironmentLookup
func Lookup(envKey string) (value string, present bool) {
	return lookupEnv(envKey)
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
