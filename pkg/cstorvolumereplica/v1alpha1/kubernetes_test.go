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
	"errors"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	client "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fakeGetClientset gets the cvr clientset
func fakeGetClientset() (clientset *clientset.Clientset, err error) {
	return &client.Clientset{}, nil
}

func fakeListOk(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apis.CStorVolumeReplicaList, error) {
	return &apis.CStorVolumeReplicaList{}, nil
}

func fakeListErrfn(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apis.CStorVolumeReplicaList, error) {
	return &apis.CStorVolumeReplicaList{}, errors.New("some error")
}

func fakeGetOk(
	cli *clientset.Clientset,
	name, namespace string,
	opts metav1.GetOptions,
) (*apis.CStorVolumeReplica, error) {
	return &apis.CStorVolumeReplica{}, nil
}

func fakeGetErr(
	cli *clientset.Clientset,
	name, namespace string,
	opts metav1.GetOptions,
) (*apis.CStorVolumeReplica, error) {
	return &apis.CStorVolumeReplica{}, errors.New("some error")
}

func fakeCreateOk(
	cli *clientset.Clientset,
	namespace string,
	volr *apis.CStorVolumeReplica,
) (*apis.CStorVolumeReplica, error) {
	return &apis.CStorVolumeReplica{}, nil
}

func fakeCreateErr(
	cli *clientset.Clientset,
	namespace string,
	volr *apis.CStorVolumeReplica,
) (*apis.CStorVolumeReplica, error) {
	return &apis.CStorVolumeReplica{}, errors.New("some error")
}

func fakeDeleteOk(
	cli *clientset.Clientset,
	name, namespace string,
	opts *metav1.DeleteOptions,
) error {
	return nil
}

func fakeDeleteErr(
	cli *clientset.Clientset,
	name, namespace string,
	opts *metav1.DeleteOptions,
) error {
	return errors.New("some error")
}

func fakeSetClientset(k *Kubeclient) {
	k.clientset = &client.Clientset{}
}

func fakeSetNilClientset(k *Kubeclient) {
	k.clientset = nil
}

func fakeGetErrClientSet() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeClientSet(k *Kubeclient) {}

func fakeGetClientSetForPathOk(
	fakeConfigPath string,
) (*clientset.Clientset, error) {
	return &client.Clientset{}, nil
}

func fakeGetClientSetForPathErr(
	fakeConfigPath string,
) (cli *clientset.Clientset, err error) {
	return nil, errors.New("fake error")
}

func TestKubernetesWithDefaults(t *testing.T) {
	tests := map[string]struct {
		expectListFn, expectGetClientset bool
	}{
		"When mockclient is empty":                {true, true},
		"When mockclient contains getClientsetFn": {false, true},
		"When mockclient contains ListFn":         {true, false},
		"When mockclient contains both":           {true, true},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{}
			if !mock.expectListFn {
				fc.list = fakeListOk
			}
			if !mock.expectGetClientset {
				fc.getClientset = fakeGetClientset
			}

			fc.withDefaults()
			if mock.expectListFn && fc.list == nil {
				t.Fatalf(
					"test %q failed: expected fc.list not to be empty",
					name,
				)
			}
			if mock.expectGetClientset && fc.getClientset == nil {
				t.Fatalf(
					"test %q failed: expected fc.getClientset not to be empty",
					name,
				)
			}
		})
	}
}

func TestGetClientOrCached(t *testing.T) {
	tests := map[string]struct {
		getClientSet        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		expectErr           bool
	}{
		// Positive tests
		"Positive 1": {
			fakeGetClientset,
			fakeGetClientSetForPathOk,
			"",
			false,
		},
		"Positive 2": {
			fakeGetErrClientSet,
			fakeGetClientSetForPathOk,
			"fake-path",
			false,
		},
		"Positive 3": {
			fakeGetClientset,
			fakeGetClientSetForPathErr,
			"",
			false,
		},

		// Negative tests
		"Negative 1": {
			fakeGetErrClientSet,
			fakeGetClientSetForPathOk,
			"",
			true,
		},
		"Negative 2": {
			fakeGetClientset,
			fakeGetClientSetForPathErr,
			"fake-path",
			true,
		},
		"Negative 3": {
			fakeGetErrClientSet,
			fakeGetClientSetForPathErr,
			"fake-path",
			true,
		},
		"Negative 4": {
			fakeGetErrClientSet,
			fakeGetClientSetForPathErr,
			"",
			true,
		},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientSet,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
			}
			_, err := fc.getClientOrCached()
			if mock.expectErr && err == nil {
				t.Fatalf(
					"test %q failed : expected error not to be nil but got %v",
					name,
					err,
				)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf(
					"test %q failed : expected error be nil but got %v",
					name,
					err,
				)
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
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			h := WithKubeClient(mock.Clientset)
			fake := &Kubeclient{}
			h(fake)
			if mock.expectKubeClientEmpty && fake.clientset != nil {
				t.Fatalf(
					"test %q failed expected fake.clientset to be empty",
					name,
				)
			}
			if !mock.expectKubeClientEmpty && fake.clientset == nil {
				t.Fatalf(
					"test %q failed expected fake.clientset not to be empty",
					name,
				)
			}
		})
	}
}

func TestKubernetesKubeClient(t *testing.T) {
	tests := map[string]struct {
		expectClientSet bool
		opts            []KubeclientBuildOption
	}{
		"Positive 1": {
			true,
			[]KubeclientBuildOption{
				fakeSetClientset,
			},
		},
		"Positive 2": {
			true,
			[]KubeclientBuildOption{
				fakeSetClientset,
				fakeClientSet,
			},
		},
		"Positive 3": {
			true,
			[]KubeclientBuildOption{
				fakeSetClientset,
				fakeClientSet,
				fakeClientSet,
			},
		},

		"Negative 1": {
			false,
			[]KubeclientBuildOption{
				fakeSetNilClientset,
			},
		},
		"Negative 2": {
			false,
			[]KubeclientBuildOption{
				fakeSetNilClientset,
				fakeClientSet,
			},
		},
		"Negative 3": {
			false,
			[]KubeclientBuildOption{
				fakeSetNilClientset,
				fakeClientSet,
				fakeClientSet,
			},
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			c := NewKubeclient(mock.opts...)
			if !mock.expectClientSet && c.clientset != nil {
				t.Fatalf(
					"test %q failed expected fake.clientset to be empty",
					name,
				)
			}
			if mock.expectClientSet && c.clientset == nil {
				t.Fatalf(
					"test %q failed expected fake.clientset not to be empty",
					name,
				)
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
		"Test 2": {fakeGetClientset, fakeListOk, false},
		"Test 3": {fakeGetClientset, fakeListErrfn, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
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
		"Test 2": {fakeGetClientset, fakeGetOk, false, "testvol", "test-ns"},
		"Test 3": {fakeGetClientset, fakeGetErr, true, "testvol", ""},
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
		getClientset    getClientsetFn
		del             delFn
		expectErr       bool
		name, namespace string
	}{
		"Test 1": {fakeGetErrClientSet, fakeDeleteOk, true, "testvol", "test-ns"},
		"Test 2": {fakeGetClientset, fakeDeleteOk, false, "testvol", "test-ns"},
		"Test 3": {fakeGetClientset, fakeDeleteErr, true, "testvol", ""},
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

func TestKubenetesCreate(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		create       createFn
		expectErr    bool
		namespace    string
		cvr          *apis.CStorVolumeReplica
	}{
		"Test 1": {
			fakeGetErrClientSet,
			fakeCreateOk,
			true,
			"test-ns",
			&apis.CStorVolumeReplica{},
		},
		"Test 2": {
			fakeGetClientset,
			fakeCreateOk,
			false,
			"test-ns",
			&apis.CStorVolumeReplica{},
		},
		"Test 3": {
			fakeGetClientset,
			fakeCreateErr,
			true,
			"",
			&apis.CStorVolumeReplica{},
		},
		"Test 4": {
			fakeGetClientset,
			fakeCreateErr,
			true,
			"test-ns",
			nil,
		},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{
				getClientset: mock.getClientset,
				create:       mock.create,
			}
			_, err := k.Create(mock.cvr)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
