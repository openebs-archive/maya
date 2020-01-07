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

	errors "github.com/pkg/errors"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func fakeGetClientSetOk() (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeListFnOk(cli *clientset.Clientset, opts metav1.ListOptions) (*storagev1.StorageClassList, error) {
	return &storagev1.StorageClassList{}, nil
}

func fakeListFnErr(cli *clientset.Clientset, opts metav1.ListOptions) (*storagev1.StorageClassList, error) {
	return nil, errors.New("some error occured to get storageclass list")
}

func fakeGetClientSetNil() (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetClientSetErr() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeGetFnOk(cli *clientset.Clientset, name string, opts metav1.GetOptions) (*storagev1.StorageClass, error) {
	return &storagev1.StorageClass{}, nil
}

func fakeGetFnErr(cli *clientset.Clientset, name string, opts metav1.GetOptions) (*storagev1.StorageClass, error) {
	return nil, errors.New("failed to get storageclass")
}

func fakeCreateFnOk(cli *clientset.Clientset, sc *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	return &storagev1.StorageClass{}, nil
}

func fakeCreateFnErr(cli *clientset.Clientset, sc *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	return nil, errors.New("failed to create storageclass")
}

func fakeDeleteFnErr(cli *clientset.Clientset, name string, opts *metav1.DeleteOptions) error {
	return errors.New("failed to delete")
}

func fakeDeleteFnOk(cli *clientset.Clientset, name string, opts *metav1.DeleteOptions) error {
	return nil
}

func fakeGetClientSetForPathOk(fakeConfigPath string) (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeGetClientSetForPathErr(fakeConfigPath string) (cli *clientset.Clientset, err error) {
	return nil, errors.New("fake error")
}

func TestKubeClient(t *testing.T) {
	kubeclient := NewKubeClient()
	if reflect.DeepEqual(kubeclient, Kubeclient{}) {
		t.Fatalf("test failed: expect kubeclient not to be empty")
	}
}

func TestWithDefaultOptions(t *testing.T) {
	tests := map[string]struct {
		kubeClient *Kubeclient
	}{
		"T1": {&Kubeclient{}},
		"T2": {&Kubeclient{nil, "fake-path", fakeGetClientSetOk, fakeGetClientSetForPathOk, fakeListFnOk, fakeGetFnOk, fakeCreateFnOk, fakeDeleteFnOk}},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			mock.kubeClient.withDefaults()
			if mock.kubeClient.getClientset == nil {
				t.Fatalf("test %q failed: expected getClientset not to be empty", name)
			}
			if mock.kubeClient.list == nil {
				t.Fatalf("test %q failed: expected list not to be empty", name)
			}
			if mock.kubeClient.get == nil {
				t.Fatalf("test %q failed: expected get not to be emptu", name)
			}
			if mock.kubeClient.create == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
			if mock.kubeClient.del == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
			if mock.kubeClient.getClientsetForPath == nil {
				t.Fatalf("test %q failed: expected getClientset not to be nil", name)
			}
		})
	}
}

func TestGetClientSetForPathOrDirect(t *testing.T) {
	tests := map[string]struct {
		getClientSet        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		isErr               bool
	}{
		// Positive tests
		"Positive 1": {fakeGetClientSetNil, fakeGetClientSetForPathOk, "fake-path", false},
		"Positive 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", false},
		"Positive 3": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "fake-path", false},
		"Positive 4": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "", false},

		// Negative tests
		"Negative 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", true},
		"Negative 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", true},
		"Negative 3": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "fake-path", true},
		"Negative 4": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "", true},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientSet,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
			}
			_, err := fc.getClientsetForPathOrDirect()
			if mock.isErr && err == nil {
				t.Fatalf("test %q failed : expected error not to be nil but got %v", name, err)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test %q failed : expected error be nil but got %v", name, err)
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
		"Positive 1": {fakeGetClientSetNil, fakeGetClientSetForPathOk, "fake-path", false},
		"Positive 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", false},
		"Positive 3": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "fake-path", false},
		"Positive 4": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "", false},

		// Negative tests
		"Negative 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", true},
		"Negative 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", true},
		"Negative 3": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "fake-path", true},
		"Negative 4": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "", true},
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
			_, err := fc.getClientsetOrCached()
			if mock.expectErr && err == nil {
				t.Fatalf("test %q failed : expected error not to be nil but got %v", name, err)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("test %q failed : expected error be nil but got %v", name, err)
			}
		})
	}
}

func TestKubenetesStorageClassList(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		list                listFn
		expectedErr         bool
	}{
		// Positive tests
		"Positive 1": {fakeGetClientSetNil, fakeGetClientSetForPathOk, "fake-path", fakeListFnOk, false},
		"Positive 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeListFnOk, false},
		"Positive 3": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "fake-path", fakeListFnOk, false},
		"Positive 4": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "", fakeListFnOk, false},

		// Negative tests
		"Negative 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", fakeListFnOk, true},
		"Negative 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", fakeListFnErr, true},
		"Negative 3": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "fake-path", fakeListFnOk, true},
		"Negative 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeListFnErr, true},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				list:                mock.list,
			}
			_, err := fc.List(metav1.ListOptions{})
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubenetesStorageClassGet(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		get                 getFn
		expectedErr         bool
	}{
		// Positive tests
		"Positive 1": {fakeGetClientSetNil, fakeGetClientSetForPathOk, "fake-path", fakeGetFnOk, false},
		"Positive 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeGetFnOk, false},
		"Positive 3": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "fake-path", fakeGetFnOk, false},
		"Positive 4": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "", fakeGetFnOk, false},

		// Negative tests
		"Negative 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", fakeGetFnOk, true},
		"Negative 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", fakeGetFnErr, true},
		"Negative 3": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "fake-path", fakeGetFnOk, true},
		"Negative 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeGetFnErr, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				get:                 mock.get,
			}
			_, err := fc.Get(name, metav1.GetOptions{})
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubenetesStorageClassCreate(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		create              createFn
		sc                  *storagev1.StorageClass
		expectedErr         bool
	}{
		"Test 1": {
			getClientset:        fakeGetClientSetErr,
			getClientSetForPath: fakeGetClientSetForPathOk,
			kubeConfigPath:      "",
			create:              fakeCreateFnOk,
			sc:                  &storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "SC-1"}},
			expectedErr:         true,
		},
		"Test 2": {
			getClientset:        fakeGetClientSetOk,
			getClientSetForPath: fakeGetClientSetForPathOk,
			kubeConfigPath:      "fake-path",
			create:              fakeCreateFnErr,
			sc:                  &storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "SC-2"}},
			expectedErr:         true,
		},
		"Test 3": {
			getClientset:        fakeGetClientSetOk,
			getClientSetForPath: fakeGetClientSetForPathOk,
			kubeConfigPath:      "fake-path",
			create:              fakeCreateFnErr,
			sc:                  nil,
			expectedErr:         true,
		},
		"Test 4": {
			getClientset:        fakeGetClientSetOk,
			getClientSetForPath: fakeGetClientSetForPathErr,
			kubeConfigPath:      "",
			create:              fakeCreateFnOk,
			sc:                  &storagev1.StorageClass{},
			expectedErr:         false,
		},
		"Test 5": {
			getClientset:        fakeGetClientSetOk,
			getClientSetForPath: fakeGetClientSetForPathErr,
			kubeConfigPath:      "fp",
			create:              fakeCreateFnOk,
			sc:                  nil,
			expectedErr:         true,
		},
		"Test 6": {
			getClientset:        fakeGetClientSetOk,
			getClientSetForPath: fakeGetClientSetForPathErr,
			kubeConfigPath:      "",
			create:              fakeCreateFnOk,
			sc:                  nil,
			expectedErr:         true,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				create:              mock.create,
			}
			_, err := fc.Create(mock.sc)
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubenetesStorageClassDelete(t *testing.T) {
	tests := map[string]struct {
		getClientSet        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		del                 deleteFn
		scName              string
		expectErr           bool
	}{
		"Negative Test 1": {
			getClientSet:        fakeGetClientSetErr,
			getClientSetForPath: fakeGetClientSetForPathOk,
			kubeConfigPath:      "",
			del:                 fakeDeleteFnOk,
			scName:              "SC1",
			expectErr:           true,
		},
		"Negative Test 2": {
			getClientSet:        fakeGetClientSetOk,
			getClientSetForPath: fakeGetClientSetForPathErr,
			kubeConfigPath:      "",
			del:                 fakeDeleteFnErr,
			scName:              "SC2",
			expectErr:           true,
		},
		"Negative Test 3": {
			getClientSet:        fakeGetClientSetOk,
			getClientSetForPath: fakeGetClientSetForPathOk,
			kubeConfigPath:      "",
			del:                 fakeDeleteFnErr,
			scName:              "",
			expectErr:           true,
		},
		"Negative Test 4": {
			getClientSet:        fakeGetClientSetOk,
			getClientSetForPath: fakeGetClientSetForPathErr,
			kubeConfigPath:      "fp",
			del:                 fakeDeleteFnErr,
			scName:              "",
			expectErr:           true,
		},
		"Positive Test 5": {
			getClientSet:        fakeGetClientSetOk,
			getClientSetForPath: fakeGetClientSetForPathOk,
			kubeConfigPath:      "",
			del:                 fakeDeleteFnOk,
			scName:              "",
			expectErr:           false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientSet,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				del:                 mock.del,
			}
			err := fc.Delete(mock.scName, &metav1.DeleteOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
