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
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	template "github.com/openebs/maya/pkg/template/v1alpha1"
	"github.com/pkg/errors"
	"strings"
)

// ArtifactIdentifier is a typed string to help identify the type of artifact
type ArtifactIdentifier string

const (
	// CASTemplateArtifact helps in identifying a CAS Template based artifact
	CASTemplateArtifact ArtifactIdentifier = "kind: CASTemplate"
	// RunTaskArtifact helps in identifying a RunTask based artifact
	RunTaskArtifact ArtifactIdentifier = "kind: RunTask"
)

// Artifact represents a YAML compatible artifact that will be installed,
// applied, etc
type Artifact struct {
	// Doc represents the YAML compatible artifact
	Doc string
}

// ArtifactMiddleware abstracts updating a given artifact
type ArtifactMiddleware func(given *Artifact) (updated *Artifact, err error)

// ArtifactPredicate abstracts evaluating a condition against the provided
// artifact
type ArtifactPredicate func(given *Artifact) bool

// IsCASTemplate flags if the provided artifact is of type CASTemplate
//
// NOTE:
//  This is an implementation of ArtifactPredicate
func IsCASTemplate(given *Artifact) bool {
	return given != nil && strings.Contains(given.Doc, string(CASTemplateArtifact))
}

// IsNotRunTask flags if the provided artifact is not of type RunTask
//
// NOTE:
//  This is an implementation of ArtifactPredicate
func IsNotRunTask(given *Artifact) bool {
	return given != nil && !strings.Contains(given.Doc, string(RunTaskArtifact))
}

// ArtifactTemplater updates an artifact instance by templating it and returns
// the resulting templated instance
func ArtifactTemplater(values map[string]interface{}, t template.Templater) ArtifactMiddleware {
	return func(given *Artifact) (updated *Artifact, err error) {
		if given == nil {
			err = fmt.Errorf("nil artifact instance: failed to template the artifact")
			return
		}

		if t == nil {
			err = fmt.Errorf("nil templater instance: failed to template the artifact '%s'", given.Doc)
			return
		}

		templated, err := t("artifact", given.Doc, values)
		if err != nil {
			err = errors.Wrapf(err, "failed to template the artifact '%s' with values '%+v'", given.Doc, values)
			return
		}

		updated = &Artifact{Doc: templated}
		return
	}
}

// ArtifactList is the list of artifacts that will be installed or uninstalled
type ArtifactList struct {
	Items []*Artifact
}

// VersionArtifactLister abstracts fetching a list of artifacts based on
// provided version
type VersionArtifactLister func(v version) (ArtifactList, error)

// ListArtifactsByVersion returns artifacts based on the provided version
func ListArtifactsByVersion(v version) (ArtifactList, error) {
	switch v {
	case version070:
		return RegisteredArtifactsFor070(), nil
	default:
		return ArtifactList{}, fmt.Errorf("invalid version '%+v': failed to list artifacts by version", v)
	}
}

// MapIf will execute the ArtifactMiddleware conditionally based on
// ArtifactPredicate
func (l ArtifactList) MapIf(m ArtifactMiddleware, p ArtifactPredicate) (filtered ArtifactList, errs []error) {
	var err error
	for _, artifact := range l.Items {
		err = nil
		if p(artifact) {
			artifact, err = m(artifact)
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}
		filtered.Items = append(filtered.Items, artifact)
	}
	return
}

// UnstructuredList transforms this ArtifactList to corresponding list of
// unstructured instances
func (l ArtifactList) UnstructuredList() (ul k8s.UnstructList, errs []error) {
	return l.UnstructuredListC(k8s.CreateUnstructuredFromYaml)
}

// UnstructuredListC transforms this ArtifactList to corresponding list of
// unstructured instances by making use of unstructured creator instance
func (l ArtifactList) UnstructuredListC(c k8s.UnstructuredCreator) (ul k8s.UnstructList, errs []error) {
	for _, artifact := range l.Items {
		unstructured, err := c(artifact.Doc)
		if err != nil {
			errs = append(errs, errors.Wrap(err, "failed to transform artifact into unstructured instance"))
			continue
		}
		ul = append(ul, unstructured)
	}
	return
}
