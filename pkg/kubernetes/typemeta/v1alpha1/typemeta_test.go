/*
Copyright 2019 The OpenEBS Authors.

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
	"testing"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTypeMetaString(t *testing.T) {
	tests := map[string]struct {
		object              TypeMeta
		expectedStringParts []string
	}{
		"typemeta": {
			TypeMeta{
				object: &metav1.TypeMeta{
					Kind:       "fake-kind",
					APIVersion: "fake-apiversion",
				},
			},
			[]string{"kind: fake-kind", "apiVersion: fake-apiversion"},
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			ymlstr := mock.object.String()
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}

func TestTypeMetaGoString(t *testing.T) {
	tests := map[string]struct {
		object              TypeMeta
		expectedStringParts []string
	}{
		"typemeta": {
			TypeMeta{
				object: &metav1.TypeMeta{
					Kind:       "fake-kind",
					APIVersion: "fake-apiversion",
				},
			},
			[]string{"kind: fake-kind", "apiVersion: fake-apiversion"},
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			ymlstr := mock.object.GoString()
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}

func TestNewBuilder(t *testing.T) {
	tests := map[string]struct {
		expectObject bool
		expectErrs   bool
	}{
		"new instance of TypeMeta": {
			true,
			false,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := NewBuilder()
		if (b.meta.object != nil) != mock.expectObject {
			t.Fatalf("test %s failed, expect object %t, but got : %t",
				name, mock.expectObject, b.meta.object != nil)
		}
		if (len(b.errors) != 0) != mock.expectErrs {
			t.Fatalf("test %s failed, expect error %t, but got : %t",
				name, mock.expectErrs, len(b.errors) != 0)
		}
	}
}

func TestWithKind(t *testing.T) {
	tests := map[string]struct {
		kind         string
		expectedKind string
	}{
		"kind present": {
			"fake-kind",
			"fake-kind",
		},
		"empty kind present": {
			"",
			"",
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &Builder{
			meta: &TypeMeta{
				object: &metav1.TypeMeta{},
			},
		}
		b.WithKind(mock.kind)
		if b.meta.object.Kind != mock.expectedKind {
			t.Fatalf("test %s failed, expected kind %s, but got : %s",
				name, mock.expectedKind, b.meta.object.Kind)
		}
	}
}

func TestWithAPIVersion(t *testing.T) {
	tests := map[string]struct {
		apiVersion         string
		expectedapiVersion string
	}{
		"api version present": {
			"fake-apiVersion",
			"fake-apiVersion",
		},
		"empty api version present": {
			"",
			"",
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &Builder{
			meta: &TypeMeta{
				object: &metav1.TypeMeta{},
			},
		}
		b.WithAPIVersion(mock.apiVersion)
		if b.meta.object.APIVersion != mock.apiVersion {
			t.Fatalf("test %s failed, expected api version %s, but got : %s",
				name, mock.apiVersion, b.meta.object.APIVersion)
		}
	}
}

func TestNewBuilderForAPIObject(t *testing.T) {
	tests := map[string]struct {
		typeMeta       *metav1.TypeMeta
		expectTypeMeta bool
	}{
		"valid typemeta present": {
			&metav1.TypeMeta{},
			true,
		},
		"nil typemeta present": {
			nil,
			true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		b := NewBuilderForAPIObject(mock.typeMeta)
		if (b.meta.object != nil) != mock.expectTypeMeta {
			t.Fatalf("test %s failed, expect typemeta %t, but got : %t",
				name, mock.expectTypeMeta, b.meta.object != nil)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		typeMeta  metav1.TypeMeta
		expecterr bool
	}{
		"only kind present": {
			metav1.TypeMeta{
				Kind: "fake-kind",
			},
			true,
		},
		"only api version present": {
			metav1.TypeMeta{
				APIVersion: "fake-api-version",
			},
			true,
		},
		"kind and api version present": {
			metav1.TypeMeta{
				Kind:       "fake-kind",
				APIVersion: "fake-api-version",
			},
			false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		b := &Builder{
			meta: &TypeMeta{
				object: &mock.typeMeta,
			},
		}
		err := b.validate()
		if (err != nil) != mock.expecterr {
			t.Fatalf("test %s failed, expect error %t, but got : %t",
				name, mock.expecterr, err != nil)
		}
	}
}

func TestBuild(t *testing.T) {
	tests := map[string]struct {
		typeMeta  metav1.TypeMeta
		errors    []error
		expecterr bool
	}{
		"only kind present": {
			metav1.TypeMeta{
				Kind: "fake-kind",
			},
			[]error{},
			true,
		},
		"only api version present": {
			metav1.TypeMeta{
				APIVersion: "fake-api-version",
			},
			[]error{},
			true,
		},
		"kind, api version and error present": {
			metav1.TypeMeta{
				Kind:       "fake-kind",
				APIVersion: "fake-api-version",
			},
			[]error{errors.New("")},
			true,
		},
		"kind and api version present": {
			metav1.TypeMeta{
				Kind:       "fake-kind",
				APIVersion: "fake-api-version",
			},
			[]error{},
			false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		b := &Builder{
			meta: &TypeMeta{
				object: &mock.typeMeta,
			},
			errors: mock.errors,
		}
		_, err := b.Build()
		if (err != nil) != mock.expecterr {
			t.Fatalf("test %s failed, expect error %t, but got : %t",
				name, mock.expecterr, err != nil)
		}
	}
}
