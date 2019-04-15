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
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestOwnerReferenceString(t *testing.T) {
	tests := map[string]struct {
		object              OwnerReference
		expectedStringParts []string
	}{
		"objectmeta": {
			OwnerReference{
				Object: &metav1.OwnerReference{
					Name:       "fake-name",
					Kind:       "fake-kind",
					APIVersion: "fake-apiversion",
					UID:        "fake-uid",
				},
			},
			[]string{"Object:", "name: fake-name", "kind: fake-kind",
				"apiVersion: fake-apiversion", "uid: fake-uid"},
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

func TestOwnerReferenceGoString(t *testing.T) {
	tests := map[string]struct {
		object              OwnerReference
		expectedStringParts []string
	}{
		"objectmeta": {
			OwnerReference{
				Object: &metav1.OwnerReference{
					Name:       "fake-name",
					Kind:       "fake-kind",
					APIVersion: "fake-apiversion",
					UID:        "fake-uid",
				},
			},
			[]string{"Object:", "name: fake-name", "kind: fake-kind",
				"apiVersion: fake-apiversion", "uid: fake-uid"},
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

func TestNew(t *testing.T) {
	tests := map[string]struct {
		expectObject bool
		expectErrs   bool
	}{
		"new instance of OwnerReference": {
			true,
			false,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := New()
		if (b.Object != nil) != mock.expectObject {
			t.Fatalf("test %s failed, expect object %t, but got : %t",
				name, mock.expectObject, b.Object != nil)
		}
		if (len(b.errors) != 0) != mock.expectErrs {
			t.Fatalf("test %s failed, expect error %t, but got : %t",
				name, mock.expectErrs, len(b.errors) != 0)
		}
	}
}

func TestWithName(t *testing.T) {
	tests := map[string]struct {
		name         string
		expectedName string
	}{
		"name present": {
			"fake-name",
			"fake-name",
		},
		"empty name present": {
			"",
			"",
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &OwnerReference{
			Object: &metav1.OwnerReference{},
		}
		b.WithName(mock.name)
		if b.Object.Name != mock.expectedName {
			t.Fatalf("test %s failed, expected name %s, but got : %s",
				name, mock.expectedName, b.Object.Name)
		}
	}
}

func TestWithKind(t *testing.T) {
	tests := map[string]struct {
		kind         string
		expectedKind string
	}{
		"kind present": {
			"fake-namespace",
			"fake-namespace",
		},
		"empty kind present": {
			"",
			"",
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &OwnerReference{
			Object: &metav1.OwnerReference{},
		}
		b.WithKind(mock.kind)
		if b.Object.Kind != mock.expectedKind {
			t.Fatalf("test %s failed, expected kind %s, but got : %s",
				name, mock.expectedKind, b.Object.Kind)
		}
	}
}

func TestWithAPIVersion(t *testing.T) {
	tests := map[string]struct {
		apiVersion         string
		expectedAPIVersion string
	}{
		"apiVersion present": {
			"apps/v1",
			"apps/v1",
		},
		"empty apiVersion present": {
			"",
			"",
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &OwnerReference{
			Object: &metav1.OwnerReference{},
		}
		b.WithAPIVersion(mock.apiVersion)
		if b.Object.APIVersion != mock.expectedAPIVersion {
			t.Fatalf("test %s failed, expected apiVersion %s, but got : %s",
				name, mock.expectedAPIVersion, b.Object.APIVersion)
		}
	}
}

func TestWithUID(t *testing.T) {
	tests := map[string]struct {
		uid         types.UID
		expectedUID types.UID
	}{
		"uid present": {
			"apps/v1",
			"apps/v1",
		},
		"empty uid present": {
			"",
			"",
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &OwnerReference{
			Object: &metav1.OwnerReference{},
		}
		b.WithUID(mock.uid)
		if b.Object.UID != mock.uid {
			t.Fatalf("test %s failed, expected uid %s, but got : %s",
				name, mock.expectedUID, b.Object.UID)
		}
	}
}

func TestWithAPIObject(t *testing.T) {
	tests := map[string]struct {
		ownerReference       *metav1.OwnerReference
		expectOwnerReference bool
	}{
		"valid ownerReference present": {
			&metav1.OwnerReference{},
			true,
		},
		"nil ownerReference present": {
			nil,
			false,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &OwnerReference{}
		b.WithAPIObject(mock.ownerReference)
		if (b.Object != nil) != mock.expectOwnerReference {
			t.Fatalf("test %s failed, expect ownerReference %t, but got : %t",
				name, mock.expectOwnerReference, b.Object != nil)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		ownerReference metav1.OwnerReference
		expecterr      bool
	}{
		"valid owner reference": {
			metav1.OwnerReference{
				Name:       "fake-name",
				APIVersion: "fake-api-version",
				Kind:       "fake-kind",
				UID:        "fake-uid",
			},
			false,
		},
		"name not present": {
			metav1.OwnerReference{
				APIVersion: "fake-api-version",
				Kind:       "fake-kind",
				UID:        "fake-uid",
			},
			true,
		},
		"api version not present": {
			metav1.OwnerReference{
				Name: "fake-name",
				Kind: "fake-kind",
				UID:  "fake-uid",
			},
			true,
		},
		"kind not present": {
			metav1.OwnerReference{
				Name:       "fake-name",
				APIVersion: "fake-api-version",
				UID:        "fake-uid",
			},
			true,
		},
		"uid not present": {
			metav1.OwnerReference{
				Name:       "fake-name",
				APIVersion: "fake-api-version",
				Kind:       "fake-kind",
			},
			true,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &OwnerReference{
			Object: &mock.ownerReference,
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
		ownerReference metav1.OwnerReference
		errors         []error
		expecterr      bool
	}{
		"valid owner reference": {
			metav1.OwnerReference{
				Name:       "fake-name",
				APIVersion: "fake-api-version",
				Kind:       "fake-kind",
				UID:        "fake-uid",
			},
			[]error{},
			false,
		},
		"valid owner reference but error present": {
			metav1.OwnerReference{
				Name:       "fake-name",
				APIVersion: "fake-api-version",
				Kind:       "fake-kind",
				UID:        "fake-uid",
			},
			[]error{errors.New("")},
			true,
		},
		"name not present": {
			metav1.OwnerReference{
				APIVersion: "fake-api-version",
				Kind:       "fake-kind",
				UID:        "fake-uid",
			},
			[]error{},
			true,
		},
		"api version not present": {
			metav1.OwnerReference{
				Name: "fake-name",
				Kind: "fake-kind",
				UID:  "fake-uid",
			},
			[]error{},
			true,
		},
		"kind not present": {
			metav1.OwnerReference{
				Name:       "fake-name",
				APIVersion: "fake-api-version",
				UID:        "fake-uid",
			},
			[]error{},
			true,
		},
		"uid not present": {
			metav1.OwnerReference{
				Name:       "fake-name",
				APIVersion: "fake-api-version",
				Kind:       "fake-kind",
			},
			[]error{},
			true,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &OwnerReference{
			Object: &mock.ownerReference,
			errors: mock.errors,
		}
		_, err := b.Build()
		if (err != nil) != mock.expecterr {
			t.Fatalf("test %s failed, expect error %t, but got : %t",
				name, mock.expecterr, err != nil)
		}
	}
}
