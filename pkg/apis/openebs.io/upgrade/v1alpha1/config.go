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

package v1alpha1

// ExecutorConfig represents task executor config.
// This contains Versions, upgrade Options and castemplate name.
type ExecutorConfig struct {
	Versions
	Options
	// Castemplate contain castemplate name which task executor
	// will use to upgrade single unit of resources.
	Castemplate string `json:"casTemplate"`
	// RuntimeConfigs is used to provide some runtime config to
	// castemplate engine. Task executor will directly copy this
	// config to castemplate engine.
	RuntimeConfigs []RuntimeConfig `json:"runtimeConfig"`
	// JobId contains unique id for each job
	JobId string `json:"jobId"`
	// Resources contains list of resources which we are going to upgrade
	Resources []ResourceDetails `json:"Resources"`
}

// Versions contains different version that executor will use at upgrade time.
// It mainly contains base and target version
type Versions struct {
	Base   string `json:"base"`
	Target string `json:"target"`
}

// Options contains different options that executor will use at upgrade time.
// These are used to take decision on what to do if an error occurred during
// update
type Options struct {
	RetryOnError    bool `json:"retryOnError"`
	RollbackOnError bool `json:"rollbackOnError"`
}

// RuntimeConfig holds a runtime configuration for executor
type RuntimeConfig struct {
	// Name of the config
	Name string `json:"name"`
	// Value represents any specific value that is applicable
	// to this config
	Value string `json:"value"`
	// Data represents an arbitrary map of key value pairs
	Data map[string]string `json:"data"`
}
