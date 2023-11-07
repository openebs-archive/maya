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
	"context"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	template "github.com/openebs/maya/pkg/template/v1alpha1"
	"github.com/openebs/maya/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// A map of resources that is restricted to be updated by this
// installer. In other words, resources whose Kind matches the
// key of this map will not be considered for update by this
// installer
var restrictUpdates = map[string]bool{
	"StoragePoolClaim": true,
	"StorageClass":     true,
}

// Installer abstracts installation
type Installer interface {
	Install() (errors []error)
	Clean() error
}

// simpleInstaller installs artifacts by making use of install config
//
// NOTE:
//
//	This is an implementation of Installer
type simpleInstaller struct {
	artifactTemplater ArtifactMiddleware
	envLister         EnvLister

	// TODO use pkg/errors/v1alpha1
	errorList
}

// resourceMetadata holds the name and namespace of a resource
type resourceMetadata struct {
	name      string
	namespace string
}

func (i *simpleInstaller) prepareResources() k8s.UnstructedList {
	elist, err := i.envLister.List()
	if err != nil {
		i.addError(err)
	}

	// set the environments conditionally required for install
	eslist := elist.SetIf(version.Current(), isEnvNotPresent)
	klog.V(2).Infof("%+v", eslist.Infos())
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
	nlist := ulist.MapAllIfAny(
		[]k8s.UnstructuredMiddleware{
			k8s.UnstructuredMap(
				k8s.UpdateLabels(map[string]string{string(v1alpha1.OpenEBSVersionPlainKey): menv.Get(menv.OpenEBSVersion)}, false),
				k8s.IsNameUnVersioned,
			),
			k8s.SuffixNameWithVersion(),
			k8s.UnstructuredMapAll(
				[]k8s.UnstructuredMiddleware{
					k8s.SuffixWithVersionAtPath("spec.run.tasks"),
					k8s.SuffixWithVersionAtPath("spec.output"),
					k8s.AddNameToLabels(string(v1alpha1.CASTNamePlainKey), false),
					k8s.AddKubeServerVersionToLabels(false),
				},
				k8s.IsCASTemplate,
			),
			k8s.UnstructuredMap(
				k8s.UpdateNamespace(menv.Get(menv.OpenEBSNamespace)),
				k8s.IsNamespaceScoped,
			),
		},
		k8s.IsCASTemplate,
		k8s.IsRunTask,
	)

	ul = append(ul, nlist.Items...)
	return
}

// Install the resources specified in the install config
//
// NOTE:
//
//	This is an implementation of Installer interface
func (i *simpleInstaller) Install() []error {
	ulist := i.prepareResources()
	ul := i.setRules(ulist)
	for _, unstruct := range ul {
		applier := k8s.NewResourceCreateOrUpdater(
			k8s.GroupVersionResourceFromGVK(unstruct),
			unstruct.GetNamespace(),
			k8s.ResourceCreateOrUpdaterSkipUpdate(restrictUpdates[unstruct.GetKind()]),
		)
		u, err := applier.Apply(unstruct)
		if err == nil {
			klog.V(2).Infof(
				"Installer: %s %s/%s applied successfully",
				u.GroupVersionKind(),
				u.GetNamespace(),
				u.GetName(),
			)
		} else {
			i.addError(err)
		}
	}
	return i.errors
}

// cleanResources are the versioned openebs respurces that needs to cleaned
// after installing the latest versions of the respective resource
var cleanResources = []resourceMetadata{
	resourceMetadata{
		name: "castemplates",
	},
	resourceMetadata{
		name:      "runtasks",
		namespace: menv.Get(menv.OpenEBSNamespace),
	},
}

// Clean is the cleanup function that removes old version of openebs resources
func (i *simpleInstaller) Clean() error {
	for _, res := range cleanResources {
		gvr := schema.GroupVersionResource{
			Group:    "openebs.io",
			Version:  "v1alpha1",
			Resource: res.name,
		}
		k8sDynamic, err := k8s.Dynamic().Provide()
		if err != nil {
			return err
		}
		err = k8sDynamic.Resource(gvr).Namespace(res.namespace).DeleteCollection(
			context.TODO(),
			metav1.DeleteOptions{},
			metav1.ListOptions{
				LabelSelector: "version!=" + version.Current(),
			},
		)
		if err != nil {
			return err
		}
	}
	return nil
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
