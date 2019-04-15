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

func TestObjectMetaString(t *testing.T) {
	tests := map[string]struct {
		object              ObjectMeta
		expectedStringParts []string
	}{
		"objectmeta": {
			ObjectMeta{
				Object: &metav1.ObjectMeta{
					Name:      "fake-name",
					Namespace: "fake-namespace",
				},
			},
			[]string{"Object:", "name: fake-name", "namespace: fake-namespace"},
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

func TestObjectMetaGoString(t *testing.T) {
	tests := map[string]struct {
		object              ObjectMeta
		expectedStringParts []string
	}{
		"objectmeta": {
			ObjectMeta{
				Object: &metav1.ObjectMeta{
					Name:      "fake-name",
					Namespace: "fake-namespace",
				},
			},
			[]string{"Object:", "name: fake-name", "namespace: fake-namespace"},
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
		"new instance of ObjectMeta": {
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
		b := &ObjectMeta{
			Object: &metav1.ObjectMeta{},
		}
		b.WithName(mock.name)
		if b.Object.Name != mock.expectedName {
			t.Fatalf("test %s failed, expected name %s, but got : %s",
				name, mock.expectedName, b.Object.Name)
		}
	}
}

func TestWithNamespace(t *testing.T) {
	tests := map[string]struct {
		namespace         string
		expectedNamespace string
	}{
		"namespace present": {
			"fake-namespace",
			"fake-namespace",
		},
		"empty namespace present": {
			"",
			"",
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &ObjectMeta{
			Object: &metav1.ObjectMeta{},
		}
		b.WithNamespace(mock.namespace)
		if b.Object.Namespace != mock.expectedNamespace {
			t.Fatalf("test %s failed, expected namespace %s, but got : %s",
				name, mock.expectedNamespace, b.Object.Namespace)
		}
	}
}

func TestWithLabels(t *testing.T) {
	tests := map[string]struct {
		labels       map[string]string
		expectlabels bool
	}{
		"label present": {
			map[string]string{
				"key": "value",
			},
			true,
		},
		"nil label present": {
			nil,
			false,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &ObjectMeta{
			Object: &metav1.ObjectMeta{},
		}
		b.WithLabels(mock.labels)
		if (b.Object.Labels != nil) != mock.expectlabels {
			t.Fatalf("test %s failed, expect labels %t, but got : %t",
				name, mock.expectlabels, b.Object.Labels != nil)
		}
	}
}

func TestWithAnnotations(t *testing.T) {
	tests := map[string]struct {
		annotations       map[string]string
		expectannotations bool
	}{
		"annotation present": {
			map[string]string{
				"key": "value",
			},
			true,
		},
		"nil annotation present": {
			nil,
			false,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &ObjectMeta{
			Object: &metav1.ObjectMeta{},
		}
		b.WithAnnotations(mock.annotations)
		if (b.Object.Annotations != nil) != mock.expectannotations {
			t.Fatalf("test %s failed, expect annotation %t, but got : %t",
				name, mock.expectannotations, b.Object.Annotations != nil)
		}
	}
}

func TestWithOwnerReferences(t *testing.T) {
	tests := map[string]struct {
		ownerReferences       []metav1.OwnerReference
		expectOwnerReferences bool
	}{
		"owner references present": {
			[]metav1.OwnerReference{
				metav1.OwnerReference{},
			},
			true,
		},
		"owner references not present": {
			[]metav1.OwnerReference{},
			false,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &ObjectMeta{
			Object: &metav1.ObjectMeta{},
		}
		b.WithOwnerReferences(mock.ownerReferences...)
		if (len(b.Object.OwnerReferences) != 0) != mock.expectOwnerReferences {
			t.Fatalf("test %s failed, expect owner references %t, but got : %t",
				name, mock.expectOwnerReferences, len(b.Object.OwnerReferences) != 0)
		}
	}
}

func TestWithAPIObject(t *testing.T) {
	tests := map[string]struct {
		objectMeta       *metav1.ObjectMeta
		expectObjectMeta bool
	}{
		"valid objectmeta present": {
			&metav1.ObjectMeta{},
			true,
		},
		"nil objectmeta present": {
			nil,
			false,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &ObjectMeta{}
		b.WithAPIObject(mock.objectMeta)
		if (b.Object != nil) != mock.expectObjectMeta {
			t.Fatalf("test %s failed, expect objectmeta %t, but got : %t",
				name, mock.expectObjectMeta, b.Object != nil)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		objectMeta metav1.ObjectMeta
		expecterr  bool
	}{
		"name present": {
			metav1.ObjectMeta{
				Name: "fake-name",
			},
			false,
		},
		"empty name present": {
			metav1.ObjectMeta{
				Name: "",
			},
			true,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &ObjectMeta{
			Object: &mock.objectMeta,
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
		objectMeta metav1.ObjectMeta
		errors     []error
		expecterr  bool
	}{
		"name present": {
			metav1.ObjectMeta{
				Name: "fake-name",
			},
			[]error{},
			false,
		},
		"empty name not present": {
			metav1.ObjectMeta{
				Name: "",
			},
			[]error{},
			true,
		},
		"name and error present": {
			metav1.ObjectMeta{
				Name: "fake-name",
			},
			[]error{errors.New("")},
			true,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &ObjectMeta{
			Object: &mock.objectMeta,
			errors: mock.errors,
		}
		_, err := b.Build()
		if (err != nil) != mock.expecterr {
			t.Fatalf("test %s failed, expect error %t, but got : %t",
				name, mock.expecterr, err != nil)
		}
	}
}
