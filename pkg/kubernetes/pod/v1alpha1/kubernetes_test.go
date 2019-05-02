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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
)

func fakeGetClientSetOk(kubeConfigPath string) (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeListFnOk(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
	return &corev1.PodList{}, nil
}

func fakeListFnErr(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
	return &corev1.PodList{}, errors.New("some error")
}

func fakeDeleteFnOk(cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error {
	return nil
}

func fakeDeleteFnErr(cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error {
	return errors.New("some error while delete")
}

func fakeGetFnOk(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*corev1.Pod, error) {
	return &corev1.Pod{}, nil
}

func fakeGetErrfn(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*corev1.Pod, error) {
	return &corev1.Pod{}, errors.New("Not found")
}

func fakeSetClientset(k *KubeClient) {
	k.clientset = &client.Clientset{}
}

func fakeSetNilClientset(k *KubeClient) {
	k.clientset = nil
}

func fakeGetClientSetNil(kubeConfigPath string) (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetClientSetErr(kubeConfigPath string) (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeClientSet(k *KubeClient) {}

func TestWithDefaultOptions(t *testing.T) {
	tests := map[string]struct {
		kubeClient *KubeClient
	}{
		"When all are nil": {&KubeClient{}},
		"When clientset is nil": {&KubeClient{
			clientset:    nil,
			getClientset: fakeGetClientSetOk,
			list:         fakeListFnOk,
			get:          fakeGetFnOk,
			del:          fakeDeleteFnOk,
		}},
		"When listFn nil": {&KubeClient{
			getClientset: fakeGetClientSetOk,
			list:         nil,
			get:          fakeGetFnOk,
			del:          fakeDeleteFnOk,
		}},
		"When getClientsetFn nil": {&KubeClient{
			getClientset: nil,
			list:         fakeListFnOk,
			get:          fakeGetFnOk,
			del:          fakeDeleteFnOk,
		}},
		"When getFn and CreateFn are nil": {&KubeClient{
			getClientset: fakeGetClientSetOk,
			list:         fakeListFnOk,
			get:          nil,
			del:          fakeDeleteFnOk,
		}},
		"When all are error": {&KubeClient{
			getClientset: fakeGetClientSetErr,
			list:         fakeListFnErr,
			get:          nil,
			del:          fakeDeleteFnErr,
		}},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			mock.kubeClient.withDefaults()
			if mock.kubeClient.get == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
			if mock.kubeClient.list == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
			if mock.kubeClient.del == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
			if mock.kubeClient.getClientset == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
		})
	}
}

func TestWithClientsetBuildOption(t *testing.T) {
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
			fake := &KubeClient{}
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

func TestKubeClientBuildOption(t *testing.T) {
	tests := map[string]struct {
		opts            []KubeClientBuildOption
		expectClientSet bool
	}{
		"Positive 1": {[]KubeClientBuildOption{fakeSetClientset}, true},
		"Positive 2": {[]KubeClientBuildOption{fakeSetClientset, fakeClientSet}, true},
		"Positive 3": {[]KubeClientBuildOption{fakeSetClientset, fakeClientSet, fakeClientSet}, true},

		"Negative 1": {[]KubeClientBuildOption{fakeSetNilClientset}, false},
		"Negative 2": {[]KubeClientBuildOption{fakeSetNilClientset, fakeClientSet}, false},
		"Negative 3": {[]KubeClientBuildOption{fakeSetNilClientset, fakeClientSet, fakeClientSet}, false},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			c := NewKubeClient(mock.opts...)
			if !mock.expectClientSet && c.clientset != nil {
				t.Fatalf("test %q failed expected fake.clientset to be empty", name)
			}
			if mock.expectClientSet && c.clientset == nil {
				t.Fatalf("test %q failed expected fake.clientset not to be empty", name)
			}
		})
	}
}

func TestGetClientOrCached(t *testing.T) {
	tests := map[string]struct {
		kubeClient *KubeClient
		expectErr  bool
	}{
		// Positive tests
		"Positive 1": {&KubeClient{nil, "", "fake-path", fakeGetClientSetNil, fakeListFnOk, fakeDeleteFnOk, fakeGetFnOk}, false},
		"Positive 2": {&KubeClient{&client.Clientset{}, "", "", fakeGetClientSetNil, fakeListFnOk, fakeDeleteFnOk, fakeGetFnOk}, false},

		// Negative tests
		"Negative 1": {&KubeClient{nil, "", "", fakeGetClientSetErr, fakeListFnOk, fakeDeleteFnOk, fakeGetFnOk}, true},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			c, err := mock.kubeClient.getClientsetOrCached()
			if mock.expectErr && err == nil {
				t.Fatalf("test %q failed : expected error not to be nil but got %v", name, err)
			}
			if !reflect.DeepEqual(c, mock.kubeClient.clientset) {
				t.Fatalf("test %q failed : expected clientset %v but got %v", name, mock.kubeClient.clientset, c)
			}
		})
	}
}

func TestKubernetesPodList(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		list         listFn
		expectErr    bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeListFnOk, true},
		"Test 2": {fakeGetClientSetOk, fakeListFnOk, false},
		"Test 3": {fakeGetClientSetOk, fakeListFnErr, true},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := KubeClient{getClientset: mock.getClientset, namespace: "", list: mock.list}
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

func TestKubernetesDeletePod(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		delete       deleteFn
		podName      string
		expectErr    bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeDeleteFnOk, "pod-1", true},
		"Test 2": {fakeGetClientSetOk, fakeDeleteFnOk, "pod-2", false},
		"Test 3": {fakeGetClientSetOk, fakeDeleteFnErr, "pod-3", true},
		"Test 4": {fakeGetClientSetOk, fakeDeleteFnOk, "", true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := KubeClient{getClientset: mock.getClientset, namespace: "", del: mock.delete}
			err := k.Delete(mock.podName, &metav1.DeleteOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubernetesGetPod(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		get          getFn
		podName      string
		expectErr    bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeGetFnOk, "pod-1", true},
		"Test 2": {fakeGetClientSetOk, fakeGetFnOk, "pod-2", false},
		"Test 3": {fakeGetClientSetOk, fakeGetErrfn, "pod-3", true},
		"Test 4": {fakeGetClientSetOk, fakeGetFnOk, "", true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := KubeClient{getClientset: mock.getClientset, namespace: "", get: mock.get}
			_, err := k.Get(mock.podName, metav1.GetOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestWithNamespaceBuildOption(t *testing.T) {
	tests := map[string]struct {
		namespace string
	}{
		"Test 1": {""},
		"Test 2": {"namespace 1"},
		"Test 3": {"namespace 2"},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := NewKubeClient(WithNamespace(mock.namespace))
			if k.namespace != mock.namespace {
				t.Fatalf("Test %q failed: expected %v got %v", name, mock.namespace, k.namespace)
			}
		})
	}
}
