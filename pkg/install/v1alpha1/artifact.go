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
	"fmt"
)

// Artifact has the JSON compatible artifact that will be installed or applied
type Artifact struct {
	// Doc represents the JSON compatible artifact
	Doc string
}

// ArtifactMiddleware abstracts updating a given artifact
type ArtifactMiddleware func(given *Artifact) (updated *Artifact)

// ArtifactList is the list of artifacts that will be installed or uninstalled
type ArtifactList struct {
	Items []*Artifact
}

// VersionArtifactLister abstracts fetching a list of artifacts based on
// provided version
type VersionArtifactLister func(version string) (ArtifactList, error)

// ListArtifactsByVersion returns artifacts based on the provided version
func ListArtifactsByVersion(version string) (ArtifactList, error) {
	switch version {
	case "0.7.0":
		return RegisteredArtifactsFor070(), nil
	default:
		return ArtifactList{}, fmt.Errorf("invalid version '%s': failed to list artifacts by version", version)
	}
}
