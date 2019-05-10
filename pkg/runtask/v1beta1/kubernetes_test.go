/*
Copyright 2019 The OpenEBS Authors

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

package v1beta1

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakeGetClientsetOk() (cs *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeGetOk(cs *clientset.Clientset, name, namespace string,
	opts metav1.GetOptions) (*apis.RunTask, error) {
	return &apis.RunTask{}, nil
}

func fakeGetErr(cs *clientset.Clientset, name, namespace string,
	opts metav1.GetOptions) (*apis.RunTask, error) {
	return &apis.RunTask{}, errors.New("some error")
}

func fakeSetClientsetOk(k *Kubeclient) {
	k.clientset = &clientset.Clientset{}
}

func fakeSetClientsetNil(k *Kubeclient) {
	k.clientset = nil
}

func fakeGetClientsetNil() (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetClientsetErr() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func TestWithDefaults(t *testing.T) {
	tests := map[string]struct {
		getFn              getFunc
		getClientsetFn     getClientsetFunc
		expectGet          bool
		expectGetClientset bool
	}{
		// The current implementation of WithDefaults method can be
		// tested using these two combinations only.
		"When mockclient is empty": {nil, nil, false, false},
		"When mockclient contains all of them": {fakeGetOk,
			fakeGetClientsetOk, false, false},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				get:          mock.getFn,
				getClientset: mock.getClientsetFn,
			}

			fc.withDefaults()
			if mock.expectGet && fc.get == nil {
				t.Fatalf(`test %s failed: expected non-nil fc.get
but got %v`, name, fc.get)
			}
			if mock.expectGetClientset && fc.getClientset == nil {
				t.Fatalf(`test %s failed: expected non-nil fc.getClientset
but got %v`, name, fc.getClientset)
			}
		})
	}
}

func TestWithClientset(t *testing.T) {
	tests := map[string]struct {
		clientset       *clientset.Clientset
		expectClientset bool
	}{
		"Clientset is empty":     {nil, false},
		"Clientset is not empty": {&clientset.Clientset{}, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			kc := NewKubeClient(WithClientset(mock.clientset))
			if mock.expectClientset && kc.clientset == nil {
				t.Fatalf(`test %s failed, expected non-nil fake.clientset
but got %v`, name, kc.clientset)
			}
			if !mock.expectClientset && kc.clientset != nil {
				t.Fatalf(`test %s failed, expected nil fake.clientset
but got %v`, name, kc.clientset)
			}
		})
	}
}

func TestWithNamespace(t *testing.T) {
	tests := map[string]struct {
		namespace       string
		expectNamespace bool
	}{
		"namespace is empty":     {"", false},
		"namespace is not empty": {"test-ns", true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			kc := NewKubeClient(WithNamespace(mock.namespace))
			if mock.expectNamespace && kc.namespace == "" {
				t.Fatalf(`test %s failed, expected non-nil namespace
but got %v`, name, kc.clientset)
			}
			if !mock.expectNamespace && kc.namespace != "" {
				t.Fatalf(`test %s failed, expected nil namespace
but got %v`, name, kc.clientset)
			}
		})
	}
}

func TestKubeClientWithClientset(t *testing.T) {
	tests := map[string]struct {
		opt             KubeclientBuildOption
		expectClientSet bool
	}{
		"When non-nil clientset is passed": {
			fakeSetClientsetOk,
			true},
		"When nil clientset is passed": {
			fakeSetClientsetNil,
			false},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			c := NewKubeClient(mock.opt)
			if !mock.expectClientSet && c.clientset != nil {
				t.Fatalf(`test %s failed, expected nil c.clientset
but got %v`, name, c.clientset)
			}
			if mock.expectClientSet && c.clientset == nil {
				t.Fatalf(`test %s failed expected non-nil c.clientset
but got %v`, name, c.clientset)
			}
		})
	}
}

func TestGetClientOrCached(t *testing.T) {

	tests := map[string]struct {
		clientset    *clientset.Clientset
		getClientset getClientsetFunc
		get          getFunc
		expectErr    bool
	}{
		// Positive tests
		"When clientset is nil":     {nil, fakeGetClientsetNil, fakeGetOk, false},
		"When clientset is not nil": {&clientset.Clientset{}, fakeGetClientsetNil, fakeGetOk, false},
		// Negative tests
		"When getting clientset throws error": {nil, fakeGetClientsetErr, fakeGetOk, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			kc := &Kubeclient{
				clientset:    mock.clientset,
				getClientset: mock.getClientset,
				get:          mock.get,
			}
			c, err := kc.getClientOrCached()
			if mock.expectErr && err == nil {
				t.Fatalf("test %s failed : expected error but got %v", name, err)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("test %s failed : expected nil error but got %v", name, err)
			}
			if !reflect.DeepEqual(c, kc.clientset) {
				t.Fatalf(`test %s failed : expected clientset %v
but got %v`, name, kc.clientset, c)
			}
		})
	}
}

func TestKubernetesGet(t *testing.T) {
	tests := map[string]struct {
		resourceName string
		getClientset getClientsetFunc
		get          getFunc
		expectErr    bool
	}{
		"When getting clientset throws error": {"ur1", fakeGetClientsetErr, fakeGetOk, true},
		"When getting resource throws error":  {"ur2", fakeGetClientsetOk, fakeGetErr, true},
		"When resource name is empty string":  {"", fakeGetClientsetOk, fakeGetOk, true},
		"When none of them throws error":      {"ur3", fakeGetClientsetOk, fakeGetOk, false},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, get: mock.get}
			_, err := k.Get(mock.resourceName, metav1.GetOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("test %s failed: expected error but got %v", name, err)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("test %s failed: expected nil but got %v", name, err)
			}
		})
	}
}
