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
	"github.com/golang/glog"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	commonenv "github.com/openebs/maya/pkg/env/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Installer abstracts installation
type Installer interface {
	Install() (errors []error)
}

type installErrors struct {
	errors []error
}

// addError adds an error to error list
func (i *installErrors) addError(err error) []error {
	if err == nil {
		return i.errors
	}

	i.errors = append(i.errors, err)
	return i.errors
}

// addErrors adds a list of errors to error list
func (i *installErrors) addErrors(errs []error) []error {
	for _, err := range errs {
		i.addError(err)
	}

	return i.errors
}

// simpleInstaller installs artifacts by making use of install config
//
// NOTE:
//  This is an implementation of Installer
type simpleInstaller struct {
	openebsNamespace     string
	configGetter         ConfigGetter
	artifactLister       VersionArtifactLister
	transformer          ArtifactToUnstructuredListTransformer
	unstructuredUpdaters []WithInstallUnstructuredUpdater
	namespaceUpdater     k8s.WithOptionsUnstructuredUpdater
	installErrors
}

// Install the resources specified in the install config
//
// NOTE:
//  This is an implementation of Installer interface
func (i *simpleInstaller) Install() []error {
	var (
		allUnstructured []*unstructured.Unstructured
	)

	if i.configGetter == nil {
		return i.addError(fmt.Errorf("nil config getter: simple installer failed"))
	}

	config, err := i.configGetter(env.Get(string(EnvKeyForInstallConfigName)))
	if err != nil {
		return i.addError(errors.Wrap(err, "simple installer failed"))
	}

	for _, install := range config.Spec.Install {
		list, err := i.artifactLister(install.Version)
		if err != nil {
			i.addError(errors.Wrapf(err, "simple installer failed to list artifacts for version '%s'", install.Version))
			continue
		}

		// transform list of artifacts to list of unstructured instances
		unstructs, errs := i.transformer(list)
		if len(errs) != 0 {
			i.addErrors(errs)
		}

		// override the unstructured instances from install set options
		for _, unstruct := range unstructs {
			// try updating the install artifact with namespace configured for openebs
			unstruct = i.namespaceUpdater(k8s.UnstructuredOptions{Namespace: i.openebsNamespace})(unstruct)

			// run the install artifact through install config based updaters
			installUpdaters := WithInstallUnstructuredUpdaterList(install, i.unstructuredUpdaters)
			finalUpdater := k8s.UnstructuredUpdater(installUpdaters)
			allUnstructured = append(allUnstructured, finalUpdater(unstruct))
		}
	}

	for _, unstruct := range allUnstructured {
		apply := k8s.NewResourceApplier(k8s.GroupVersionResourceFromGVK(unstruct), unstruct.GetNamespace())
		u, err := apply(unstruct)
		if err == nil {
			glog.Infof("'%s' '%s' installed successfully at namespace '%s'", u.GroupVersionKind(), u.GetName(), u.GetNamespace())
		} else {
			i.addError(err)
		}
	}

	return i.errors
}

// SimpleInstaller returns a new instance of simpleInstaller
func SimpleInstaller() Installer {
	// this is the namespace of the pod where this binary is running i.e.
	// namespace that is configured for this openebs component
	openebsNS := env.Get(string(commonenv.EnvKeyForOpenEBSNamespace))
	cmGetter := k8s.NewConfigMapGetter(openebsNS)

	return &simpleInstaller{
		openebsNamespace: openebsNS,
		configGetter:     WithConfigMapConfigGetter(cmGetter),
		artifactLister:   ListArtifactsByVersion,
		transformer:      TransformArtifactToUnstructuredList,
		unstructuredUpdaters: []WithInstallUnstructuredUpdater{
			updateUnstructuredNamespace,
			updateUnstructuredLabels,
			updateUnstructuredAnnotations,
		},
		namespaceUpdater: k8s.UpdateUnstructuredNamespace,
	}
}
