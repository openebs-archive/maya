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

package upgrader

// ResourcePatch has all the patches required to upgrade a resource
type ResourcePatch struct {
	Name              string
	OpenebsNamespace  string
	From, To          string
	ImageTag, BaseURL string
	// UpgradeTask       *utask.UpgradeTask
}

// ResourcePatchOptions ...
type ResourcePatchOptions func(*ResourcePatch)

// WithName ...
func WithName(name string) ResourcePatchOptions {
	return func(r *ResourcePatch) {
		r.Name = name
	}
}

// FromVersion ...
func FromVersion(from string) ResourcePatchOptions {
	return func(r *ResourcePatch) {
		r.From = from
	}
}

// ToVersion ...
func ToVersion(to string) ResourcePatchOptions {
	return func(r *ResourcePatch) {
		r.To = to
	}
}

// WithOpenebsNamespace ...
func WithOpenebsNamespace(namespace string) ResourcePatchOptions {
	return func(r *ResourcePatch) {
		r.OpenebsNamespace = namespace
	}
}

// WithImageTag ...
func WithImageTag(imagetag string) ResourcePatchOptions {
	return func(r *ResourcePatch) {
		r.ImageTag = imagetag
	}
}

// WithBaseURL ...
func WithBaseURL(url string) ResourcePatchOptions {
	return func(r *ResourcePatch) {
		r.BaseURL = url
	}
}

// NewResourcePatch returns a new instance of ResourcePatch
func NewResourcePatch(opts ...ResourcePatchOptions) *ResourcePatch {
	r := &ResourcePatch{}
	for _, o := range opts {
		o(r)
	}
	return r
}
