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

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func fakeGetClientsetOk() (cli *kubernetes.Clientset, err error) {
	return &kubernetes.Clientset{}, nil
}

func fakeGetOk(cli *kubernetes.Clientset, name, namespace string, opts metav1.GetOptions) (*corev1.PersistentVolumeClaim, error) {
	return &corev1.PersistentVolumeClaim{}, nil
}

func fakeListOk(cli *kubernetes.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error) {
	return &corev1.PersistentVolumeClaimList{}, nil
}

func fakeDeleteOk(cli *kubernetes.Clientset, name, namespace string, opts *metav1.DeleteOptions) error {
	return nil
}

func fakeListErr(cli *kubernetes.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error) {
	return &corev1.PersistentVolumeClaimList{}, errors.New("some error")
}

func fakeGetErr(cli *kubernetes.Clientset, name, namespace string, opts metav1.GetOptions) (*corev1.PersistentVolumeClaim, error) {
	return &corev1.PersistentVolumeClaim{}, errors.New("some error")
}

func fakeDeleteErr(cli *kubernetes.Clientset, name, namespace string, opts *metav1.DeleteOptions) error {
	return errors.New("some error")
}

func fakeSetClientset(k *Kubeclient) {
	k.clientset = &kubernetes.Clientset{}
}

func fakeSetNilClientset(k *Kubeclient) {
	k.clientset = nil
}

func fakeGetNilErrClientSet() (clientset *kubernetes.Clientset, err error) {
	return nil, nil
}

func fakeGetClientSetErr() (clientset *kubernetes.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeClientSet(k *Kubeclient) {}

func fakeCreateFnOk(cli *kubernetes.Clientset, namespace string, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	return &corev1.PersistentVolumeClaim{}, nil
}

func fakeCreateFnErr(cli *kubernetes.Clientset, namespace string, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	return nil, errors.New("failed to create PVC")
}

func TestWithDefaultOptions(t *testing.T) {
	tests := map[string]struct {
		KubeClient *Kubeclient
	}{
		"When all are nil": {&Kubeclient{}},
		"When clientset is nil": {&Kubeclient{
			clientset:     nil,
			getClientset:  fakeGetClientsetOk,
			list:          fakeListOk,
			get:           fakeGetOk,
			create:        fakeCreateFnOk,
			del:           fakeDeleteOk,
			delCollection: nil,
		}},
		"When listFn nil": {&Kubeclient{
			getClientset:  fakeGetClientsetOk,
			list:          nil,
			get:           fakeGetOk,
			create:        fakeCreateFnOk,
			del:           fakeDeleteOk,
			delCollection: nil,
		}},
		"When getClientsetFn nil": {&Kubeclient{
			getClientset:  nil,
			list:          fakeListOk,
			get:           fakeGetOk,
			create:        fakeCreateFnOk,
			del:           fakeDeleteOk,
			delCollection: nil,
		}},
		"When getFn and CreateFn are nil": {&Kubeclient{
			getClientset:  fakeGetClientsetOk,
			list:          fakeListOk,
			get:           nil,
			create:        nil,
			del:           fakeDeleteOk,
			delCollection: nil,
		}},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			mock.KubeClient.withDefaults()
			if mock.KubeClient.getClientset == nil {
				t.Fatalf("test %q failed: expected getClientset not to be empty", name)
			}
			if mock.KubeClient.list == nil {
				t.Fatalf("test %q failed: expected list not to be empty", name)
			}
			if mock.KubeClient.get == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
			if mock.KubeClient.create == nil {
				t.Fatalf("test %q failed: expected create not to be empty", name)
			}
			if mock.KubeClient.del == nil {
				t.Fatalf("test %q failed: expected del not to be empty", name)
			}
			if mock.KubeClient.delCollection == nil {
				t.Fatalf("test %q failed: expected delCollection not to be empty", name)
			}
		})
	}
}

func TestWithClientsetBuildOption(t *testing.T) {
	tests := map[string]struct {
		Clientset             *kubernetes.Clientset
		expectKubeClientEmpty bool
	}{
		"Clientset is empty":     {nil, true},
		"Clientset is not empty": {&kubernetes.Clientset{}, false},
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

func TestKubeClientBuildOption(t *testing.T) {
	tests := map[string]struct {
		expectClientSet bool
		opts            []KubeclientBuildOption
	}{
		"Positive 1": {true, []KubeclientBuildOption{fakeSetClientset}},
		"Positive 2": {true, []KubeclientBuildOption{fakeSetClientset, fakeClientSet}},
		"Positive 3": {true, []KubeclientBuildOption{fakeSetClientset, fakeClientSet, fakeClientSet}},

		"Negative 1": {false, []KubeclientBuildOption{fakeSetNilClientset}},
		"Negative 2": {false, []KubeclientBuildOption{fakeSetNilClientset, fakeClientSet}},
		"Negative 3": {false, []KubeclientBuildOption{fakeSetNilClientset, fakeClientSet, fakeClientSet}},
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
		expectErr  bool
		KubeClient *Kubeclient
	}{
		// Positive tests
		"Positive 1": {false, &Kubeclient{nil, "", fakeGetNilErrClientSet, fakeListOk, nil, nil, nil, nil}},
		"Positive 2": {false, &Kubeclient{&kubernetes.Clientset{}, "", fakeGetNilErrClientSet, fakeListOk, nil, nil, nil, nil}},

		// Negative tests
		"Negative 1": {true, &Kubeclient{nil, "", fakeGetClientSetErr, fakeListOk, nil, nil, nil, nil}},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			c, err := mock.KubeClient.getClientSetOrCached()
			if mock.expectErr && err == nil {
				t.Fatalf("test %q failed : expected error not to be nil but got %v", name, err)
			}
			if !reflect.DeepEqual(c, mock.KubeClient.clientset) {
				t.Fatalf("test %q failed : expected clientset %v but got %v", name, mock.KubeClient.clientset, c)
			}
		})
	}
}

func TestKubernetesPVCList(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		list         listFn
		expectErr    bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeListOk, true},
		"Test 2": {fakeGetClientsetOk, fakeListOk, false},
		"Test 3": {fakeGetClientsetOk, fakeListErr, true},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, namespace: "", list: mock.list}
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
			k := NewKubeClient(WithNamespace(mock.namespace))
			if k.namespace != mock.namespace {
				t.Fatalf("Test %q failed: expected %v got %v", name, mock.namespace, k.namespace)
			}
		})
	}
}

func TestKubenetesGetPVC(t *testing.T) {
	tests := map[string]struct {
		getClientset    getClientsetFn
		get             getFn
		name, namespace string
		expectErr       bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeGetOk, "testvol", "test-ns", true},
		"Test 2": {fakeGetClientsetOk, fakeGetOk, "testvol", "test-ns", false},
		"Test 3": {fakeGetClientsetOk, fakeGetErr, "testvol", "", true},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, get: mock.get}
			_, err := k.Get(mock.name, metav1.GetOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil, got %v", name, err)
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
		del          deleteFn
		name         string
		expectErr    bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeDeleteOk, "testvol", true},
		"Test 2": {fakeGetClientsetOk, fakeDeleteOk, "testvol", false},
		"Test 3": {fakeGetClientsetOk, fakeDeleteErr, "testvol", true},
		"Test 4": {fakeGetClientsetOk, fakeDeleteErr, "", true},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, del: mock.del}
			err := k.Delete(mock.name, &metav1.DeleteOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubernetesPVCCreate(t *testing.T) {
	tests := map[string]struct {
		getClientSet getClientsetFn
		create       createFn
		pvc          *v1.PersistentVolumeClaim
		expectErr    bool
	}{
		"Negative Test 1": {
			getClientSet: fakeGetClientSetErr,
			create:       fakeCreateFnOk,
			pvc:          &v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "PVC-1"}},
			expectErr:    true,
		},
		"Negative Test 2": {
			getClientSet: fakeGetClientsetOk,
			create:       fakeCreateFnErr,
			pvc:          &v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "PVC-2"}},
			expectErr:    true,
		},
		"Negative Test 3": {
			getClientSet: fakeGetClientsetOk,
			create:       fakeCreateFnErr,
			pvc:          nil,
			expectErr:    true,
		},
		"Positive Test 4": {
			getClientSet: fakeGetClientsetOk,
			create:       fakeCreateFnOk,
			pvc:          nil,
			expectErr:    false,
		},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientSet, create: mock.create}
			_, err := k.Create(mock.pvc)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
