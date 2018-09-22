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
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	template "github.com/openebs/maya/pkg/template/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Installer abstracts installation
type Installer interface {
	Install() (errors []error)
}

// simpleInstaller installs artifacts by making use of install config
//
// NOTE:
//  This is an implementation of Installer
type simpleInstaller struct {
	configProvider    ConfigProvider
	artifactLister    VersionArtifactLister
	artifactTemplater ArtifactMiddleware
	namespaceUpdater  k8s.UnstructuredMiddleware
	envLister         EnvLister
	errorList
}

func (i *simpleInstaller) preInstall() (errs []error) {
	// set the env for installer to work
	elist := envInstallConfig().SetP("installer", isEnvNotPresent)
	glog.Infof("%+v", elist.Infos())

	return elist.Errors()
}

// Install the resources specified in the install config
//
// NOTE:
//  This is an implementation of Installer interface
func (i *simpleInstaller) Install() []error {
	var (
		allUnstructured []*unstructured.Unstructured
	)

	errs := i.preInstall()
	if len(errs) != 0 {
		return errs
	}

	if i.configProvider == nil {
		return i.addError(fmt.Errorf("nil config provider: simple installer failed"))
	}

	config, err := i.configProvider.Provide()
	if err != nil {
		return i.addError(errors.Wrap(err, "simple installer failed"))
	}

	for _, install := range config.Spec.Install {
		// fetch the environments required for this install version
		elist, err := i.envLister(Version(install.Version))
		if err != nil {
			i.addError(err)
		}

		// set the environments required for this install version
		eslist := elist.SetP(install.Version, isEnvNotPresent)
		glog.Infof("%+v", eslist.Infos())
		i.addErrors(eslist.Errors())

		// list the artifacts w.r.t install version
		list, err := i.artifactLister(Version(install.Version))
		if err != nil {
			i.addError(errors.Wrapf(err, "simple installer failed to list artifacts for version '%s'", install.Version))
			continue
		}

		// run the list of artifacts through templating if it is a CASTemplate
		list, errs := list.MapIf(i.artifactTemplater, IsNotRunTask)
		if len(errs) != 0 {
			i.addErrors(errs)
		}

		// get list of unstructured instances from list of artifacts
		ulist, errs := list.UnstructuredList()
		if len(errs) != 0 {
			i.addErrors(errs)
		}

		ulist = ulist.MapAll([]k8s.UnstructuredMiddleware{
			i.namespaceUpdater,
			k8s.UpdateLabels(k8s.UnstructuredOptions{
				Labels: map[string]string{"openebs.io/version": install.Version},
			}),
		})

		allUnstructured = append(allUnstructured, ulist.Items...)
	}

	for _, unstruct := range allUnstructured {
		cu := k8s.CreateOrUpdate(k8s.GroupVersionResourceFromGVK(unstruct), unstruct.GetNamespace())
		u, err := cu.Apply(unstruct)
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
	openebsNS := menv.Get(menv.OpenEBSNamespace)

	// config provider to fetch install config i.e. a config map
	p := Config(openebsNS, menv.Get(InstallerConfigName))

	// templater to template the artifacts before installation
	t := ArtifactTemplater(NewTemplateKeyValueList().Values(), template.TextTemplate)

	// a condition based namespace updater
	uOpts := k8s.UnstructuredOptions{Namespace: openebsNS}
	nu := k8s.UpdateNamespaceP(uOpts, k8s.IsNamespaceScoped)

	// lister to list artifacts for install
	l := ListArtifactsByVersion

	// env lister to list environment objects
	e := EnvList

	return &simpleInstaller{
		configProvider:    p,
		artifactLister:    l,
		artifactTemplater: t,
		namespaceUpdater:  nu,
		envLister:         e,
	}
}
