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

// MultiYamlFetcher abstracts aggregating and returning multiple yaml documents
// as a string
type MultiYamlFetcher func() string

// ArtifactListPredicate abstracts evaluating a condition against the provided
// artifact list
type ArtifactListPredicate func() bool

// ParseArtifactListFromMultipleYamlConditional will help in adding a list of yamls that should be installed
// by the installer
// ParseArtifactListFromMultipleYamlConditional acts on ArtifactListPredicate return value, if true the yaml
// gets added to installation list else it is not added.
func ParseArtifactListFromMultipleYamlConditional(multipleYamls MultiYamlFetcher, p ArtifactListPredicate) (artifacts []*Artifact) {
	if p() {
		return ParseArtifactListFromMultipleYamls(multipleYamls)
	}
	return
}

// ParseArtifactListFromMultipleYamls generates a list of Artifacts from the
// yaml documents.
//
// NOTE:
//  Each YAML document is assumed to be separated via "---"
func ParseArtifactListFromMultipleYamls(multipleYamls MultiYamlFetcher) (artifacts []*Artifact) {
	docs := strings.Split(multipleYamls(), "---")
	for _, doc := range docs {
		doc = strings.TrimSpace(doc)
		if len(doc) == 0 {
			continue
		}

		artifacts = append(artifacts, &Artifact{Doc: doc})
	}
	return
}

// RegisteredArtifactsFor070 returns the list of 0.7.0 Artifacts that will get
// installed
func RegisteredArtifactsFor070() (list ArtifactList) {

	//Note: CRDs have to be installed first. Keep this at top of the list.
	list.Items = append(list.Items, OpenEBSCRDArtifactsFor070().Items...)

	list.Items = append(list.Items, JivaVolumeArtifactsFor070().Items...)
	//Contains the read/list/delete CAST for supporting older volumes
	//The CAST defined here are provided as fallback options to corresponding
	//0.7.0 CAST
	list.Items = append(list.Items, JivaVolumeArtifactsFor060().Items...)
	list.Items = append(list.Items, JivaPoolArtifactsFor070().Items...)

	list.Items = append(list.Items, CstorPoolArtifactsFor070().Items...)
	list.Items = append(list.Items, CstorVolumeArtifactsFor070().Items...)
	list.Items = append(list.Items, CstorSparsePoolSpcArtifactsFor070().Items...)

	//Contains the SC to help with provisioning from clone.
	//This is generic for release till K8s supports native way of cloning.
	list.Items = append(list.Items, SnapshotPromoterSCArtifacts().Items...)

	return
}
