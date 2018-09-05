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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
	"strings"
	"testing"
)

func mockLabels(labels ...string) (l map[string]string) {
	if len(labels) == 0 {
		return
	}
	l = map[string]string{}
	for _, lbl := range labels {
		if len(lbl) == 0 {
			continue
		}
		if !strings.Contains(lbl, ":") {
			continue
		}
		lPair := strings.Split(lbl, ":")
		l[lPair[0]] = lPair[1]
	}
	return
}

func mockUnstructOptionsFromLabels(labels []string) (o UnstructuredOptions) {
	o.Labels = mockLabels(labels...)
	return
}

func mockUnstructOptionsFromNSAndLabels(ns string, labels []string) (o UnstructuredOptions) {
	o.Labels = mockLabels(labels...)
	o.Namespace = ns
	return
}

func mockUnstructFromLabels(labels []string) (u *unstructured.Unstructured) {
	u = &unstructured.Unstructured{
		Object: map[string]interface{}{"metadata": map[string]interface{}{}},
	}
	u.SetLabels(mockLabels(labels...))
	return
}

func mockUnstructFromOptions(o UnstructuredOptions) (u *unstructured.Unstructured) {
	u = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": o.Namespace,
				"labels":    o.Labels,
			},
		},
	}
	return
}

func mockUnstructFromKind(kind string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{"kind": kind},
	}
}

func mockUnstructFromNS(ns string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": ns,
			},
		},
	}
}

func TestKindToResource(t *testing.T) {
	tests := map[string]struct {
		kind     string
		expected string
	}{
		"pod to pods":                           {kind: "pod", expected: "pods"},
		"storageclass to storageclasses":        {kind: "storageclass", expected: "storageclasses"},
		"castemplate to castemplates":           {kind: "castemplate", expected: "castemplates"},
		"runtask to runtasks":                   {kind: "runtask", expected: "runtasks"},
		"storagepool to storagepools":           {kind: "storagepool", expected: "storagepools"},
		"storagepoolclaim to storagepoolclaims": {kind: "storagepoolclaim", expected: "storagepoolclaims"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := kind(mock.kind)
			actual := k.resource()
			if actual != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%s' actual '%s'", name, mock.expected, actual)
			}
		})
	}
}

func TestKindIsNamespaced(t *testing.T) {
	tests := map[string]struct {
		kind     string
		expected bool
	}{
		"is configmap namespaced?":        {kind: "configmap", expected: true},
		"is deployment namespaced?":       {kind: "deployment", expected: true},
		"is pod namespaced?":              {kind: "pod", expected: true},
		"is storageclass namespaced?":     {kind: "storageclass", expected: false},
		"is castemplate namespaced?":      {kind: "castemplate", expected: false},
		"is runtask namespaced?":          {kind: "runtask", expected: true},
		"is storagepool namespaced?":      {kind: "storagepool", expected: false},
		"is storagepoolclaim namespaced?": {kind: "storagepoolclaim", expected: false},
		"is persistentvolume namespaced?": {kind: "persistentvolume", expected: false},
		"is cstorpool namespaced?":        {kind: "cstorpool", expected: false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := kind(mock.kind)
			actual := k.isNamespaced()
			if actual != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestIsNamespaceScoped(t *testing.T) {
	tests := map[string]struct {
		kind     string
		expected bool
	}{
		"is configmap namespaced?":        {kind: "configmap", expected: true},
		"is deployment namespaced?":       {kind: "deployment", expected: true},
		"is pod namespaced?":              {kind: "pod", expected: true},
		"is storageclass namespaced?":     {kind: "storageclass", expected: false},
		"is castemplate namespaced?":      {kind: "castemplate", expected: false},
		"is runtask namespaced?":          {kind: "runtask", expected: true},
		"is storagepool namespaced?":      {kind: "storagepool", expected: false},
		"is storagepoolclaim namespaced?": {kind: "storagepoolclaim", expected: false},
		"is persistentvolume namespaced?": {kind: "persistentvolume", expected: false},
		"is cstorpool namespaced?":        {kind: "cstorpool", expected: false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := mockUnstructFromKind(mock.kind)
			actual := IsNamespaceScoped(u)
			if actual != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestUpdateNamespace(t *testing.T) {
	tests := map[string]struct {
		original string
		update   string
		expected string
	}{
		"default to default": {original: "default", update: "default", expected: "default"},
		"default to empty":   {original: "default", update: "", expected: ""},
		"default to openebs": {original: "default", update: "openebs", expected: "openebs"},
		"empty to empty":     {original: "", update: "", expected: ""},
		"empty to default":   {original: "", update: "default", expected: "default"},
		"openebs to default": {original: "openebs", update: "default", expected: "default"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := mockUnstructFromNS(mock.original)
			o := UnstructuredOptions{Namespace: mock.update}
			actual := UpdateNamespace(o)(u)
			if actual.GetNamespace() != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%s' actual '%s'", name, mock.expected, actual.GetNamespace())
			}
		})
	}
}

func TestUpdateNamespaceIfNamespaceScoped(t *testing.T) {
	tests := map[string]struct {
		original string
		update   string
		kind     string
		expected string
	}{
		// pod
		"default to default for pod": {original: "default", update: "default", kind: "pod", expected: "default"},
		"default to empty for pod":   {original: "default", update: "", kind: "pod", expected: ""},
		"default to openebs for pod": {original: "default", update: "openebs", kind: "pod", expected: "openebs"},
		"empty to empty for pod":     {original: "", update: "", kind: "pod", expected: ""},
		"empty to default for pod":   {original: "", update: "default", kind: "pod", expected: "default"},
		"openebs to default for pod": {original: "openebs", update: "default", kind: "pod", expected: "default"},
		// castemplate
		"default to default for castemplate": {original: "default", update: "default", kind: "castemplate", expected: "default"},
		"default to empty for castemplate":   {original: "default", update: "", kind: "castemplate", expected: "default"},
		"default to openebs for castemplate": {original: "default", update: "openebs", kind: "castemplate", expected: "default"},
		"empty to empty for castemplate":     {original: "", update: "", kind: "castemplate", expected: ""},
		"empty to default for castemplate":   {original: "", update: "default", kind: "castemplate", expected: ""},
		"openebs to default for castemplate": {original: "openebs", update: "default", kind: "castemplate", expected: "openebs"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := mockUnstructFromNS(mock.original)
			u.SetKind(mock.kind)
			o := UnstructuredOptions{Namespace: mock.update}
			actual := UpdateNamespaceP(o, IsNamespaceScoped)(u)
			if actual.GetNamespace() != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%s' actual '%s'", name, mock.expected, actual.GetNamespace())
			}
		})
	}
}

func TestUpdateLabels(t *testing.T) {
	tests := map[string]struct {
		original string
		update   string
		expected []string
	}{
		"k:v to k1:v1":              {"k:v", "k1:v1", []string{"k:v", "k1:v1"}},
		"nothing to k1:v1":          {"", "k1:v1", []string{"k1:v1"}},
		"openebs.io/k:v to k1:v1":   {"openebs.io/k:v", "k1:v1", []string{"openebs.io/k:v", "k1:v1"}},
		"openebs.io/k:v to nothing": {"openebs.io/k:v", "", []string{"openebs.io/k:v"}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := mockUnstructFromLabels([]string{mock.original})
			o := mockUnstructOptionsFromLabels([]string{mock.update})
			actual := UpdateLabels(o)(u)
			actualLbls := actual.GetLabels()
			expectedLbls := mockLabels(mock.expected...)
			if !reflect.DeepEqual(actualLbls, expectedLbls) {
				t.Fatalf("Test '%s' failed: expected '%+v' actual '%+v'", name, expectedLbls, actualLbls)
			}
		})
	}
}

func UnstructuredMiddlewareListUpdate(t *testing.T) {
	tests := map[string]struct {
		origNS       string
		origLbl      string
		updateNS     string
		updateLbl    string
		expectedNS   string
		expectedLbls []string
	}{
		"test 1": {"", "k0:v0", "", "k2:v2", "", []string{"k0:v0", "k2:v2"}},
		"test 2": {"default", "k1:v1", "openebs", "k2:v2", "openebs", []string{"k1:v1", "k2:v2"}},
		"test 3": {"default", "k1:v1", "", "k3:v3", "", []string{"k1:v1", "k3:v3"}},
		"test 4": {"", "", "openebs", "k4:v4", "openebs", []string{"k4:v4"}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			origOpt := mockUnstructOptionsFromNSAndLabels(mock.origNS, []string{mock.origLbl})
			updateOpt := mockUnstructOptionsFromNSAndLabels(mock.updateNS, []string{mock.updateLbl})
			origU := mockUnstructFromOptions(origOpt)
			ml := UnstructuredMiddlewareList{}
			ml = append(ml, UpdateNamespaceP(updateOpt, IsNamespaceScoped))
			ml = append(ml, UpdateLabels(updateOpt))
			// this is what we are Unit Testing
			actual := ml.Update(origU)
			expectedOpt := mockUnstructOptionsFromNSAndLabels(mock.expectedNS, mock.expectedLbls)
			if actual.GetNamespace() != expectedOpt.Namespace {
				t.Fatalf("Test '%s' failed: expected '%s' actual '%s'", name, expectedOpt.Namespace, actual.GetNamespace())
			}
			if !reflect.DeepEqual(actual.GetLabels(), expectedOpt.Labels) {
				t.Fatalf("Test '%s' failed: expected '%+v' actual '%+v'", name, expectedOpt.Labels, actual.GetLabels())
			}
		})
	}
}

func TestUnstructPredicateListIsNot(t *testing.T) {
	tests := map[string]struct {
		kind     string
		op       PredicateListOp
		p1       UnstructuredPredicate
		expected bool
	}{
		"test 1": {"runtask", AllPredicates, IsNamespaceScoped, false},
		"test 2": {"castemplate", AllPredicates, IsNamespaceScoped, true},
		"test 3": {"pod", AllPredicates, IsNamespaceScoped, false},
		"test 4": {"deployment", AllPredicates, IsNamespaceScoped, false},
		"test 5": {"storagepoolclaim", AllPredicates, IsNamespaceScoped, true},
		"test 6": {"storagepool", AllPredicates, IsNamespaceScoped, true},
		"test 7": {"persistentvolume", AllPredicates, IsNamespaceScoped, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := mockUnstructFromKind(mock.kind)
			pl := UnstructPredicateList{
				Op:    mock.op,
				Items: map[string]UnstructuredPredicate{"p1": mock.p1},
			}
			actual := pl.isNot(u)
			if actual != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestUnstructPredicateListIsAll(t *testing.T) {
	tests := map[string]struct {
		kind     string
		op       PredicateListOp
		p1       UnstructuredPredicate
		expected bool
	}{
		"test 1": {"runtask", AllPredicates, IsNamespaceScoped, true},
		"test 2": {"castemplate", AllPredicates, IsNamespaceScoped, false},
		"test 3": {"pod", AllPredicates, IsNamespaceScoped, true},
		"test 4": {"deployment", AllPredicates, IsNamespaceScoped, true},
		"test 5": {"storagepoolclaim", AllPredicates, IsNamespaceScoped, false},
		"test 6": {"storagepool", AllPredicates, IsNamespaceScoped, false},
		"test 7": {"persistentvolume", AllPredicates, IsNamespaceScoped, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := mockUnstructFromKind(mock.kind)
			pl := UnstructPredicateList{
				Op:    mock.op,
				Items: map[string]UnstructuredPredicate{"p1": mock.p1},
			}
			actual := pl.isAll(u)
			if actual != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestUnstructPredicateListIsAny(t *testing.T) {
	tests := map[string]struct {
		kind     string
		op       PredicateListOp
		p1       UnstructuredPredicate
		expected bool
	}{
		"test 1": {"runtask", AllPredicates, IsNamespaceScoped, true},
		"test 2": {"castemplate", AllPredicates, IsNamespaceScoped, false},
		"test 3": {"pod", AllPredicates, IsNamespaceScoped, true},
		"test 4": {"deployment", AllPredicates, IsNamespaceScoped, true},
		"test 5": {"storagepoolclaim", AllPredicates, IsNamespaceScoped, false},
		"test 6": {"storagepool", AllPredicates, IsNamespaceScoped, false},
		"test 7": {"persistentvolume", AllPredicates, IsNamespaceScoped, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := mockUnstructFromKind(mock.kind)
			pl := UnstructPredicateList{
				Op:    mock.op,
				Items: map[string]UnstructuredPredicate{"p1": mock.p1},
			}
			actual := pl.isAny(u)
			if actual != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t'", name, mock.expected, actual)
			}
		})
	}
}
