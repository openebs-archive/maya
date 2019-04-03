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

package v1alpha1

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakeGetClientset() (cs *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeGetfn(cs *clientset.Clientset, name string,
	opts metav1.GetOptions) (*apis.CASTemplate, error) {
	return &apis.CASTemplate{}, nil
}

func fakeGetErrfn(cs *clientset.Clientset, name string,
	opts metav1.GetOptions) (*apis.CASTemplate, error) {
	return &apis.CASTemplate{}, errors.New("some error")
}

func fakeSetClientset(k *Kubeclient) {
	k.clientset = &clientset.Clientset{}
}

func fakeSetNilClientset(k *Kubeclient) {
	k.clientset = nil
}

func fakeGetNilErrClientSet() (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetErrClientSet() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeClientSet(k *Kubeclient) {}

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
		"When mockclient contains all of them": {fakeGetfn,
			fakeGetClientset, false, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{}
			fc.get = mock.getFn
			fc.getClientset = mock.getClientsetFn

			fc.withDefaults()
			get := (fc.get == nil)
			if get != mock.expectGet {
				t.Fatalf(`test %s failed: expected non-nil fc.get
but got %v`, name, fc.get)
			}
			getClientset := (fc.getClientset == nil)
			if getClientset != mock.expectGetClientset {
				t.Fatalf(`test %s failed: expected non-nil fc.getClientset
but got %v`, name, fc.getClientset)
			}
		})
	}
}
func TestWithClientset(t *testing.T) {
	tests := map[string]struct {
		clientSet    *clientset.Clientset
		isKubeClient bool
	}{
		"Clientset is empty":     {nil, false},
		"Clientset is not empty": {&clientset.Clientset{}, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			h := WithClientset(mock.clientSet)
			fake := &Kubeclient{}
			h(fake)
			if mock.isKubeClient && fake.clientset == nil {
				t.Fatalf(`test %s failed, expected non-nil fake.clientset
but got %v`, name, fake.clientset)
			}
			if !mock.isKubeClient && fake.clientset != nil {
				t.Fatalf(`test %s failed, expected nil fake.clientset
but got %v`, name, fake.clientset)
			}
		})
	}
}
func TestKubeClientWithClientset(t *testing.T) {
	tests := map[string]struct {
		expectClientSet bool
		opts            []KubeclientBuildOption
	}{
		"When non-nil clientset is passed": {true,
			[]KubeclientBuildOption{fakeSetClientset}},
		"When two options with a non-nil clientset are passed": {true,
			[]KubeclientBuildOption{fakeSetClientset, fakeClientSet}},
		"When three options with a non-nil clientset are passed": {true,
			[]KubeclientBuildOption{fakeSetClientset, fakeClientSet, fakeClientSet}},

		"When nil clientset is passed": {false,
			[]KubeclientBuildOption{fakeSetNilClientset}},
		"When two options with a nil clientset are passed": {false,
			[]KubeclientBuildOption{fakeSetNilClientset, fakeClientSet}},
		"When three options with a nil clientset are passed": {false,
			[]KubeclientBuildOption{fakeSetNilClientset, fakeClientSet, fakeClientSet}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := KubeClient(mock.opts...)
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
		kubeClient *Kubeclient
		expectErr  bool
	}{
		// Positive tests
		"When clientset is nil": {&Kubeclient{nil,
			fakeGetNilErrClientSet, fakeGetfn}, false},
		"When clientset is not nil": {&Kubeclient{&clientset.Clientset{},
			fakeGetNilErrClientSet, fakeGetfn}, false},
		// Negative tests
		"When getting clientset throws error": {&Kubeclient{nil,
			fakeGetErrClientSet, fakeGetfn}, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c, err := mock.kubeClient.getClientOrCached()
			if mock.expectErr && err == nil {
				t.Fatalf("test %s failed : expected error but got %v", name, err)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("test %s failed : expected nil error but got %v", name, err)
			}
			if !reflect.DeepEqual(c, mock.kubeClient.clientset) {
				t.Fatalf(`test %s failed : expected clientset %v
but got %v`, name, mock.kubeClient.clientset, c)
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
		"When getting clientset throws error": {"ur1", fakeGetErrClientSet, fakeGetfn, true},
		"When getting resource throws error":  {"ur2", fakeGetClientset, fakeGetErrfn, true},
		"When resource name is empty string":  {"", fakeGetClientset, fakeGetfn, true},
		"When none of them throws error":      {"ur3", fakeGetClientset, fakeGetfn, false},
	}

	for name, mock := range tests {
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
