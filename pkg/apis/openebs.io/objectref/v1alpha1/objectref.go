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

// ObjectRef refers to an object
//
// NOTE:
//  This is typically used to refer an
// object from another object
type ObjectRef struct {
	Kind       string `json:"kind"`       // Kind of object that is being refered to
	Name       string `json:"name"`       // Name of object that is being refered to
	Namespace  string `json:"namespace"`  // Namespace of object that is being refered to
	APIVersion string `json:"apiVersion"` // APIVersion of object that is being refered to
}

// ControllerRef refers to a object that plays
// the role of controller
//
// NOTE:
//  Controller reference sets the controller
// should exclusively control the object
//
// NOTE:
//  The controlled object will typically
// refer to its controller via this
// property
type ControllerRef struct {
	ObjectRef
}

// SubResourceHash represent a sub resource within
// any resource. It has properties that can identify
// a sub resource as well its last saved hash value.
type SubResourceHash struct {
	Name string `json:"name"` // identify the sub resource by name
	Path string `json:"path"` // identify the sub resource by path
	Hash string `json:"hash"` // last saved hash value of this sub resource
}

// HashRef refers to an object and
// that object's hash value that was last
// saved
//
// NOTE:
//  A resource can make use of the
// refered object and save the refered
// object's hash value within itself. This
// is very helpful during the resource's
// next reconcile action where it can decide
// if the refered object had any changes
// that should be taken in by this resource.
type HashRef struct {
	ObjectRef
	SubResource []SubResourceHash `json:subResources`
}

// HashRefList defines a list of
// HashRef
type HashRefList []HashRef

// CatalogRef refers to a catalog object
//
// NOTE:
//  Catalog reference sets the catalog
// that was used to create the resource
//
// NOTE:
//  A resource object will typically
// refer to its catalog via this
// property
type CatalogRef struct {
	HashRef
}
