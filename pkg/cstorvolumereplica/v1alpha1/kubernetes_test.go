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
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	client "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakeGetClientset() (clientset *clientset.Clientset, err error) {
	return &client.Clientset{}, nil
}

func fakeListfn(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.CStorVolumeReplicaList, error) {
	return &apis.CStorVolumeReplicaList{}, nil
}

func fakeListErrfn(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.CStorVolumeReplicaList, error) {
	return &apis.CStorVolumeReplicaList{}, errors.New("some error")
}

func fakeSetClientset(k *kubeclient) {
	k.clientset = &client.Clientset{}
}

func fakeSetNilClientset(k *kubeclient) {
	k.clientset = nil
}

func fakeGetNilErrClientSet() (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetErrClientSet() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeClientSet(k *kubeclient) {}

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
		t.Run(name, func(t *testing.T) {
			fc := &kubeclient{}
			if !mock.expectListFn {
				fc.list = fakeListfn
			}
			if !mock.expectGetClientset {
				fc.getClientset = fakeGetClientset
			}

			fc.withDefaults()
			if mock.expectListFn && fc.list == nil {
				t.Fatalf("test %q failed: expected fc.list not to be empty", name)
			}
			if mock.expectGetClientset && fc.getClientset == nil {
				t.Fatalf("test %q failed: expected fc.getClientset not to be empty", name)
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
		t.Run(name, func(t *testing.T) {
			h := WithKubeClient(mock.Clientset)
			fake := &kubeclient{}
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
		opts            []kubeclientBuildOption
	}{
		"Positive 1": {true, []kubeclientBuildOption{fakeSetClientset}},
		"Positive 2": {true, []kubeclientBuildOption{fakeSetClientset, fakeClientSet}},
		"Positive 3": {true, []kubeclientBuildOption{fakeSetClientset, fakeClientSet, fakeClientSet}},

		"Negative 1": {false, []kubeclientBuildOption{fakeSetNilClientset}},
		"Negative 2": {false, []kubeclientBuildOption{fakeSetNilClientset, fakeClientSet}},
		"Negative 3": {false, []kubeclientBuildOption{fakeSetNilClientset, fakeClientSet, fakeClientSet}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := KubeClient(mock.opts...)
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
		KubeClient *kubeclient
	}{
		// Positive tests
		"Positive 1": {false, &kubeclient{nil, "", fakeGetNilErrClientSet, fakeListfn}},
		"Positive 2": {false, &kubeclient{&client.Clientset{}, "", fakeGetNilErrClientSet, fakeListfn}},
		// Negative tests
		"Negative 1": {true, &kubeclient{nil, "", fakeGetErrClientSet, fakeListfn}},
	}

	for name, mock := range tests {
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

func TestKubenetesList(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		list         listFn
		expectErr    bool
	}{
		"Test 1": {fakeGetErrClientSet, fakeListfn, true},
		"Test 2": {fakeGetClientset, fakeListfn, false},
		"Test 3": {fakeGetClientset, fakeListErrfn, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := kubeclient{getClientset: mock.getClientset, list: mock.list}
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
