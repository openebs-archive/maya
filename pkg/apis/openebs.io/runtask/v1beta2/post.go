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

package v1beta2

// Post represents the desired specifications for the
// post field of runtask. It specifies the action or the
// commands that need to be executed post executing a
// runtask i.e. save the status or name of the current
// resource so that the next runtask can make use of it.
type Post struct {
	Operations []Operation `json:"operations"`
}

// Operation represents the desired details of a
// particular operation to be executed against a
// particular resource
type Operation struct {
	Run        string   `json:"run"`
	For        []string `json:"for"`
	WithFilter []string `json:"withFilter"`
	WithOutput []string `json:"withOutput"`
	As         string   `json:"as"`
}

// TopLevelProperty represents the top level property that
// is a starting point to represent a hierarchical chain of
// properties.
//
// e.g.
// Config.prop1.subprop1 = val1
// Config.prop1.subprop2 = val2
// In above example Config is a top level object
//
// NOTE:
//  The value of any hierarchical chain of properties
// can be parsed via dot notation
type TopLevelProperty string

const (
	// CurrentRuntimeObjectTLP is a top level property supported by CAS template engine
	// The runtime object of the current task's execution is stored in this top
	// level property.
	CurrentRuntimeObjectTLP TopLevelProperty = "RuntimeObject"
)
