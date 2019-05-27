// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	client "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	_, _ = fakeGetErrClientSetForPath("")
	_, _ = fakeGetNilErrClientSetForPath("")
)

func fakeGetClientsetOk() (clientset *clientset.Clientset, err error) {
	return &client.Clientset{}, nil
}

func fakeListOk(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.CStorVolumeList, error) {
	return &apis.CStorVolumeList{}, nil
}

func fakeGetOk(cli *clientset.Clientset, name, namespace string, opts metav1.GetOptions) (*apis.CStorVolume, error) {
	return &apis.CStorVolume{}, nil
}

func fakeDeleteOk(cli *clientset.Clientset, name, namespace string, opts *metav1.DeleteOptions) error {
	return nil
}

func fakeListErr(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.CStorVolumeList, error) {
	return &apis.CStorVolumeList{}, errors.New("some error")
}

func fakeGetErr(cli *clientset.Clientset, name, namespace string, opts metav1.GetOptions) (*apis.CStorVolume, error) {
	return &apis.CStorVolume{}, errors.New("some error")
}

func fakeDeleteErr(cli *clientset.Clientset, name, namespace string, opts *metav1.DeleteOptions) error {
	return errors.New("some error")

}

func fakeSetClientsetOk(k *Kubeclient) {
	k.clientset = &client.Clientset{}
}

func fakeSetClientsetNil(k *Kubeclient) {
	k.clientset = nil
}

func fakeGetNilErrClientSet() (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetNilErrClientSetForPath(path string) (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetErrClientSetForPath(path string) (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeGetErrClientSet() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeClientSet(k *Kubeclient) {}

func TestWithDefaults(t *testing.T) {
	tests := map[string]struct {
		getFn              getFn
		getClientsetFn     getClientsetFn
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
				t.Fatalf(`test %s failed: expected non-nil fc.get but got %v`, name, fc.get)
			}
			if mock.expectGetClientset && fc.getClientset == nil {
				t.Fatalf(`test %s failed: expected non-nil fc.getClientset but got %v`, name, fc.getClientset)
			}
		})
	}
}

func TestKubernetesWithKubeClient(t *testing.T) {
	tests := map[string]struct {
		Clientset             *client.Clientset
		expectKubeClientEmpty bool
	}{
		"Clientset is empty":     {nil, true},
		"Clientset is not empty": {&client.Clientset{}, false},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			h := WithClientSet(mock.Clientset)
			fake := &Kubeclient{}
			h(fake)
			if mock.expectKubeClientEmpty && fake.clientset != nil {
				t.Fatalf("test %q failed expected fake.clientset to be empty", name)
			}
			if !mock.expectKubeClientEmpty && fake.clientset == nil {
				t.Fatalf("test %q failed expected fake.clientset not to be empty", name)
			}
		})
	}
}

func TestKubernetesKubeClient(t *testing.T) {
	tests := map[string]struct {
		expectClientSet bool
		opts            []KubeclientBuildOption
	}{
		"Positive 1": {true, []KubeclientBuildOption{fakeSetClientsetOk}},
		"Positive 2": {true, []KubeclientBuildOption{fakeSetClientsetOk, fakeClientSet}},
		"Positive 3": {true, []KubeclientBuildOption{fakeSetClientsetOk, fakeClientSet, fakeClientSet}},

		"Negative 1": {false, []KubeclientBuildOption{fakeSetClientsetNil}},
		"Negative 2": {false, []KubeclientBuildOption{fakeSetClientsetNil, fakeClientSet}},
		"Negative 3": {false, []KubeclientBuildOption{fakeSetClientsetNil, fakeClientSet, fakeClientSet}},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			c := NewKubeclient(mock.opts...)
			if !mock.expectClientSet && c.clientset != nil {
				t.Fatalf("test %q failed expected fake.clientset to be empty", name)
			}
			if mock.expectClientSet && c.clientset == nil {
				t.Fatalf("test %q failed expected fake.clientset not to be empty", name)
			}
		})
	}
}

func TesKubernetestGetClientOrCached(t *testing.T) {
	tests := map[string]struct {
		expectErr  bool
		KubeClient *Kubeclient
	}{
		// Positive tests
		"Positive 1": {false, &Kubeclient{nil, "", "", fakeGetNilErrClientSet, fakeGetNilErrClientSetForPath, fakeGetOk, fakeListOk, fakeDeleteOk}},
		"Positive 2": {false, &Kubeclient{&client.Clientset{}, "", "", fakeGetNilErrClientSet, fakeGetNilErrClientSetForPath, fakeGetOk, fakeListOk, fakeDeleteOk}},
		// Negative tests
		"Negative 1": {true, &Kubeclient{nil, "", "", fakeGetErrClientSet, fakeGetErrClientSetForPath, fakeGetOk, fakeListOk, fakeDeleteOk}},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			c, err := mock.KubeClient.getClientOrCached()
			if mock.expectErr && err == nil {
				t.Fatalf("test %q failed : expected error not to be nil but got %v", name, err)
			}
			if !reflect.DeepEqual(c, mock.KubeClient.clientset) {
				t.Fatalf("test %q failed : expected clientset %v but got %v", name, mock.KubeClient.clientset, c)
			}
		})
	}
}

func TestWithNamespaceBuildOption(t *testing.T) {
	tests := map[string]struct {
		namespace string
	}{
		"Test 1": {""},
		"Test 2": {"alpha"},
		"Test 3": {"beta"},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := NewKubeclient(WithNamespace(mock.namespace))
			if k.namespace != mock.namespace {
				t.Fatalf("Test %q failed: expected %v got %v", name, mock.namespace, k.namespace)
			}
		})
	}
}

func TestKubenetesList(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		list         listFn
		expectErr    bool
	}{
		"Test 1": {fakeGetErrClientSet, fakeListOk, true},
		"Test 2": {fakeGetClientsetOk, fakeListOk, false},
		"Test 3": {fakeGetClientsetOk, fakeListErr, true},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, list: mock.list}
			_, err := k.List(metav1.ListOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubenetesGet(t *testing.T) {
	tests := map[string]struct {
		getClientset    getClientsetFn
		get             getFn
		expectErr       bool
		name, namespace string
	}{
		"Test 1": {fakeGetErrClientSet, fakeGetOk, true, "testvol", "test-ns"},
		"Test 2": {fakeGetClientsetOk, fakeGetOk, false, "testvol", "test-ns"},
		"Test 3": {fakeGetClientsetOk, fakeGetErr, true, "testvol", ""},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, get: mock.get}
			_, err := k.Get(mock.name, metav1.GetOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubenetesDelete(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		del          delFn
		expectErr    bool
		name         string
	}{
		"Test 1": {fakeGetErrClientSet, fakeDeleteOk, true, "testvol"},
		"Test 2": {fakeGetClientsetOk, fakeDeleteOk, false, "testvol"},
		"Test 3": {fakeGetClientsetOk, fakeDeleteErr, true, "testvol"},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, del: mock.del}
			err := k.Delete(mock.name)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
