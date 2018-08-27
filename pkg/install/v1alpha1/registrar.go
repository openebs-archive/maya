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
// artifact
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
func RegisteredArtifactsFor070() (finallist ArtifactList) {
	finallist.Items = append(finallist.Items, JivaPoolArtifactsFor070().Items...)
	finallist.Items = append(finallist.Items, CstorPoolSpc070().Items...)
	// TODO
	// Below commented code will uncommented selectively
	//finallist.Items = append(finallist.Items, CstorPoolArtifactsFor070().Items...)
	//finallist.Items = append(finallist.Items, CstorVolumeArtifactsFor070().Items...)
	//finallist.Items = append(finallist.Items, JivaVolumeArtifactsFor070().Items...)

	return
}
