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
	"strings"
)

// MultiYamlFetcher abstracts aggregating and
// returning multiple yaml documents as a string
type MultiYamlFetcher interface {
	FetchYamls() string
}

// ArtifactListPredicate abstracts evaluating a
// condition against the provided artifact list
type ArtifactListPredicate func() bool

// ParseArtifactListFromMultipleYamlsIf generates a
// list of Artifacts from yaml documents if predicate
// evaluation succeeds
func ParseArtifactListFromMultipleYamlsIf(
	m MultiYamlFetcher,
	p ArtifactListPredicate,
) (artifacts []*Artifact) {
	if p() {
		return ParseArtifactListFromMultipleYamls(m)
	}
	return
}

// ParseArtifactListFromMultipleYamls generates a list of
// Artifacts from the yaml documents.
//
// NOTE:
//  Each YAML document is assumed to be separated via "---"
func ParseArtifactListFromMultipleYamls(m MultiYamlFetcher) (artifacts []*Artifact) {
	docs := strings.Split(m.FetchYamls(), "---")
	for _, doc := range docs {
		doc = strings.TrimSpace(doc)
		if len(doc) == 0 {
			continue
		}
		artifacts = append(artifacts, &Artifact{Doc: doc})
	}
	return
}

// RegisteredArtifacts returns the list of latest
// Artifacts that will get installed
func RegisteredArtifacts() (list artifactList) {
	// Note: CRDs need to be installed first
	// Keep this at top of the list
	list.Items = append(list.Items, OpenEBSCRDArtifacts().Items...)

	list.Items = append(list.Items, JivaVolumeArtifacts().Items...)

	// Contains read/list/delete CAST for supporting older volumes
	// CAST defined here are provided as fallback options to latest CAST
	list.Items = append(list.Items, JivaVolumeArtifactsFor060().Items...)

	list.Items = append(list.Items, JivaPoolArtifacts().Items...)

	list.Items = append(list.Items, CstorPoolArtifacts().Items...)
	list.Items = append(list.Items, CstorVolumeArtifacts().Items...)
	list.Items = append(list.Items, CstorSnapshotArtifacts().Items...)
	list.Items = append(list.Items, CstorSparsePoolArtifacts().Items...)

	// Contains SC to help with provisioning from clone
	// This is generic for release till K8s supports native
	// way of cloning
	list.Items = append(list.Items, SnapshotPromoterSCArtifacts().Items...)

	// snapshots
	list.Items = append(list.Items, JivaSnapshotArtifacts().Items...)
	list.Items = append(list.Items, StoragePoolArtifacts().Items...)

	list.Items = append(list.Items, VolumeStatsArtifacts().Items...)

	// Local PV Artifacts
	list.Items = append(list.Items, LocalPVArtifacts().Items...)
	return
}
