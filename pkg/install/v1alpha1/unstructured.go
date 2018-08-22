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
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
)

// ArtifactToUnstructuredListTransformer abstracts transforming a list of
// Artifacts to corresponding list of unstructured instances
type ArtifactToUnstructuredListTransformer func(list ArtifactList) ([]*unstructured.Unstructured, []error)

// TransformArtifactToUnstructuredList transforms a list of Artifacts to
// corresponding list of unstructured instances
//
// NOTE:
//  This is an implementation of ArtifactToUnstructuredListTransformer
func TransformArtifactToUnstructuredList(list ArtifactList) (unstructuredList []*unstructured.Unstructured, errs []error) {
	for _, artifact := range list.Items {

		unstructured, err := k8s.CreateUnstructuredFromYaml(artifact.Doc)
		if err != nil {
			errs = append(errs, errors.Wrap(err, "failed to transform artifact into unstructured instance"))
			continue
		}

		unstructuredList = append(unstructuredList, unstructured)
	}
	return
}

// WithInstallUnstructuredUpdater abstracts updating Unstructured instance based
// on install specs
type WithInstallUnstructuredUpdater func(install Install) k8s.UnstructuredMiddleware

// updateUnstructuredNamespace updates the unstructured's namespace
//
// NOTE:
//  This is an implementation of WithInstallUnstructuredUpdater
func updateUnstructuredNamespace(install Install) k8s.UnstructuredMiddleware {
	return func(unstructured *unstructured.Unstructured) *unstructured.Unstructured {
		if unstructured == nil {
			return unstructured
		}

		namespace := strings.TrimSpace(install.SetOptions.Namespace)
		if len(namespace) == 0 {
			return unstructured
		}

		unstructured.SetNamespace(namespace)
		return unstructured
	}
}

// updateUnstructuredLabels updates the unstructured's labels
//
// NOTE:
//  This is an implementation of WithInstallUnstructuredUpdater
func updateUnstructuredLabels(install Install) k8s.UnstructuredMiddleware {
	return func(unstructured *unstructured.Unstructured) *unstructured.Unstructured {
		if unstructured == nil {
			return unstructured
		}

		if install.SetOptions.Labels == nil || len(install.SetOptions.Labels) == 0 {
			return unstructured
		}

		unstructured.SetLabels(install.SetOptions.Labels)
		return unstructured
	}
}

// updateUnstructuredAnnotations updates the unstructured's annotations
//
// NOTE:
//  This is an implementation of WithInstallUnstructuredUpdater
func updateUnstructuredAnnotations(install Install) k8s.UnstructuredMiddleware {
	return func(unstructured *unstructured.Unstructured) *unstructured.Unstructured {
		if unstructured == nil {
			return unstructured
		}

		if install.SetOptions.Annotations == nil || len(install.SetOptions.Annotations) == 0 {
			return unstructured
		}

		unstructured.SetAnnotations(install.SetOptions.Annotations)
		return unstructured
	}
}

// WithInstallUnstructuredUpdaterList returns a list of unstructured updaters
// based on install specs
func WithInstallUnstructuredUpdaterList(install Install, updaters []WithInstallUnstructuredUpdater) []k8s.UnstructuredMiddleware {
	var unstructMiddlewares []k8s.UnstructuredMiddleware

	for _, updater := range updaters {
		unstructMiddlewares = append(unstructMiddlewares, updater(install))
	}

	return unstructMiddlewares
}
