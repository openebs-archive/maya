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

// TODO
// Make use of pkg/msg instead of errorList

package v1alpha1

import (
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	template "github.com/openebs/maya/pkg/template/v1alpha1"
	"github.com/openebs/maya/pkg/version"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Installer abstracts installation
type Installer interface {
	Install() (errors []error)
	GetInstalled() string
}

// simpleInstaller installs artifacts by making use of install config
//
// NOTE:
//  This is an implementation of Installer
type simpleInstaller struct {
	artifactTemplater ArtifactMiddleware
	envLister         EnvLister
	errorList
}

func (i *simpleInstaller) prepareResources() k8s.UnstructedList {
	elist, err := i.envLister.List()
	if err != nil {
		i.addError(err)
	}

	// set the environments conditionally required for install
	eslist := elist.SetIf(version.Current(), isEnvNotPresent)
	glog.Infof("%+v", eslist.Infos())
	i.addErrors(eslist.Errors())

	// list the artifacts w.r.t latest version
	al := RegisteredArtifacts()
	// run the list of artifacts through templating if it is not a RunTask
	al, errs := al.MapIf(i.artifactTemplater, IsNotRunTask)
	if len(errs) != 0 {
		i.addErrors(errs)
	}

	// get list of unstructured instances from list of artifacts
	ulist, errs := al.ToUnstructuredList()
	if len(errs) != 0 {
		i.addErrors(errs)
	}
	return ulist
}

// setRules sets the install rules against the artifacts
func (i *simpleInstaller) setRules(ulist k8s.UnstructedList) (ul []*unstructured.Unstructured) {
	// order of list of middlewares is crucial as each middleware will mutate the
	// unstructured instance
	nlist := ulist.MapAllIfAny([]k8s.UnstructuredMiddleware{
		k8s.UnstructuredMap(
			k8s.UpdateLabels(map[string]string{
				string(v1alpha1.VersionKey): menv.Get(menv.OpenEBSVersion),
			}, false),
			k8s.IsNameUnVersioned,
		),
		k8s.SuffixNameWithVersion(),
		k8s.UnstructuredMapAll([]k8s.UnstructuredMiddleware{
			k8s.SuffixWithVersionAtPath("spec.run.tasks"),
			k8s.SuffixWithVersionAtPath("spec.output"),
		},
			k8s.IsCASTemplate,
		),
		k8s.UnstructuredMap(
			k8s.UpdateNamespace(menv.Get(menv.OpenEBSNamespace)),
			k8s.IsNamespaceScoped,
		),
		k8s.UnstructuredMap(
			k8s.AddNameToLabels(string(v1alpha1.CASTNameKey), false),
			k8s.IsCASTemplate,
		),
	}, k8s.IsCASTemplate, k8s.IsRunTask)

	ul = append(ul, nlist.Items...)
	return
}

// Install the resources specified in the install config
//
// NOTE:
//  This is an implementation of Installer interface
func (i *simpleInstaller) Install() []error {
	ulist := i.prepareResources()
	ul := i.setRules(ulist)
	for _, unstruct := range ul {
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

// GetInstalled get the installed resources by the maya installer.
// If all the resources are created it returns 'Completed' status,
// if there are some resources which are not created yet returns
// 'InProgress'
func (i *simpleInstaller) GetInstalled() string {
	ulist := i.prepareResources()
	ul := i.setRules(ulist)
	for _, unstruct := range ul {
		cu := k8s.CreateOrUpdate(k8s.GroupVersionResourceFromGVK(unstruct), unstruct.GetNamespace())
		_, err := cu.Get(unstruct.GetName(), metav1.GetOptions{})
		if err != nil && apierrors.IsNotFound(errors.Cause(err)) {
			return "InProgress"
		}
	}
	return "Completed"
}

// SimpleInstaller returns a new instance of simpleInstaller
func SimpleInstaller() Installer {
	// templater to template the artifacts before installation
	t := ArtifactTemplater(map[string]interface{}{}, template.TextTemplate)

	// env variables required for install
	e := EnvInstall()

	return &simpleInstaller{
		artifactTemplater: t,
		envLister:         e,
	}
}
