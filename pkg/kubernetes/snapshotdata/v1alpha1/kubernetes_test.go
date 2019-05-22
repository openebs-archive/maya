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
	"encoding/json"
	"errors"
	"testing"

	"github.com/openebs/maya/pkg/apis/openebs.io/snapshot/v1alpha1"
	snapshot "github.com/openebs/maya/pkg/apis/openebs.io/snapshot/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/snapshot/v1alpha1/clientset/internalclientset/typed/snapshot/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func fakeGetClientSetOk() (*clientset.OpenebsV1alpha1Client, error) {
	return &clientset.OpenebsV1alpha1Client{}, nil
}

func fakeGetClientSetErr() (*clientset.OpenebsV1alpha1Client, error) {
	return nil, errors.New("Some error")
}

func fakeListFnOk(cli *clientset.OpenebsV1alpha1Client, opts metav1.ListOptions) (*snapshot.VolumeSnapshotDataList, error) {
	return &snapshot.VolumeSnapshotDataList{}, nil
}

func fakeListFnErr(cli *clientset.OpenebsV1alpha1Client, opts metav1.ListOptions) (*snapshot.VolumeSnapshotDataList, error) {
	return nil, errors.New("some error")
}

func fakeGetClientSetForPathOk(fakeConfigPath string) (*clientset.OpenebsV1alpha1Client, error) {
	return &clientset.OpenebsV1alpha1Client{}, nil
}

func fakeGetClientSetForPathErr(fakeConfigPath string) (*clientset.OpenebsV1alpha1Client, error) {
	return nil, errors.New("fake error")
}

func fakeDeleteFnOk(cli *clientset.OpenebsV1alpha1Client, name string, opts *metav1.DeleteOptions) error {
	return nil
}

func fakeDeleteFnErr(cli *clientset.OpenebsV1alpha1Client, name string, opts *metav1.DeleteOptions) error {
	return errors.New("some error while delete")
}

func fakeGetFnOk(cli *clientset.OpenebsV1alpha1Client, name string, opts metav1.GetOptions) (*snapshot.VolumeSnapshotData, error) {
	return &snapshot.VolumeSnapshotData{}, nil
}

func fakeGetErrfn(cli *clientset.OpenebsV1alpha1Client, name string, opts metav1.GetOptions) (*snapshot.VolumeSnapshotData, error) {
	return &snapshot.VolumeSnapshotData{}, errors.New("Not found")
}

func fakeSetClientset(k *Kubeclient) {
	k.clientset = &clientset.OpenebsV1alpha1Client{}
}

func fakeSetNilClientset(k *Kubeclient) {
	k.clientset = nil
}

func fakeGetClientSetNil() (clientset *clientset.OpenebsV1alpha1Client, err error) {
	return nil, nil
}

func fakePatchFnOk(cli *clientset.OpenebsV1alpha1Client, name string, pt types.PatchType, data []byte, subresources ...string) (*v1alpha1.VolumeSnapshotData, error) {
	return &snapshot.VolumeSnapshotData{}, nil
}

func fakePatchFnErr(cli *clientset.OpenebsV1alpha1Client, name string, pt types.PatchType, data []byte, subresources ...string) (*v1alpha1.VolumeSnapshotData, error) {
	return nil, errors.New("fake error")
}

func fakeClientSet(k *Kubeclient) {}

func TestWithDefaultOptions(t *testing.T) {
	tests := map[string]struct {
		kubeClient *Kubeclient
	}{
		"T1": {&Kubeclient{}},
		"T2": {&Kubeclient{
			clientset:    nil,
			getClientset: fakeGetClientSetOk,
			list:         fakeListFnOk,
			get:          fakeGetFnOk,
			del:          fakeDeleteFnOk,
			patch:        fakePatchFnOk,
		}},
		"T3": {&Kubeclient{
			getClientset: fakeGetClientSetOk,
			list:         nil,
			get:          fakeGetFnOk,
			del:          fakeDeleteFnOk,
		}},
		"T4": {&Kubeclient{
			getClientset: nil,
			list:         fakeListFnOk,
			get:          fakeGetFnOk,
			del:          fakeDeleteFnOk,
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
				t.Fatalf("test %q failed: expected list not to be empty", name)
			}
			if mock.kubeClient.del == nil {
				t.Fatalf("test %q failed: expected delete not to be empty", name)
			}
			if mock.kubeClient.patch == nil {
				t.Fatalf("test %q failed: expected patch not to be empty", name)
			}
			if mock.kubeClient.getClientset == nil {
				t.Fatalf("test %q failed: expected getClientset not to be empty", name)
			}
		})
	}
}

func TestWithDefaultsForClientSetPath(t *testing.T) {
	tests := map[string]struct {
		getClientSetForPath getClientsetForPathFn
	}{
		"T1": {nil},
		"T2": {fakeGetClientSetForPathOk},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientsetForPath: mock.getClientSetForPath,
			}
			fc.withDefaults()
			if fc.getClientsetForPath == nil {
				t.Fatalf("test %q failed: expected getClientsetForPath not to be nil", name)
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

func TestWithClientsetBuildOption(t *testing.T) {
	tests := map[string]struct {
		Clientset             *clientset.OpenebsV1alpha1Client
		expectKubeclientEmpty bool
	}{
		"Clientset is empty":     {nil, true},
		"Clientset is not empty": {&clientset.OpenebsV1alpha1Client{}, false},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			h := WithClientSet(mock.Clientset)
			fake := &Kubeclient{}
			h(fake)
			if mock.expectKubeclientEmpty && fake.clientset != nil {
				t.Fatalf("test %q failed expected fake.clientset to be empty", name)
			}
			if !mock.expectKubeclientEmpty && fake.clientset == nil {
				t.Fatalf("test %q failed expected fake.clientset not to be empty", name)
			}
		})
	}
}

func TestKubeclientBuildOption(t *testing.T) {
	tests := map[string]struct {
		opts            []KubeclientBuildOption
		expectClientSet bool
	}{
		"Positive 1": {[]KubeclientBuildOption{fakeSetClientset, WithKubeConfigPath("fake-path")}, true},
		"Positive 2": {[]KubeclientBuildOption{fakeSetClientset, fakeClientSet}, true},
		"Positive 3": {[]KubeclientBuildOption{fakeSetClientset, fakeClientSet, WithKubeConfigPath("fake-path")}, true},

		"Negative 1": {[]KubeclientBuildOption{fakeSetNilClientset, WithKubeConfigPath("fake-path")}, false},
		"Negative 2": {[]KubeclientBuildOption{fakeSetNilClientset, fakeClientSet}, false},
		"Negative 3": {[]KubeclientBuildOption{fakeSetNilClientset, fakeClientSet, WithKubeConfigPath("fake-path")}, false},
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

func TestVolumeSnapshotDataList(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		list                listFn
		expectErr           bool
	}{
		// Positive tests
		"Positive 1": {fakeGetClientSetNil, fakeGetClientSetForPathOk, "fake-path", fakeListFnOk, false},
		"Positive 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeListFnOk, false},
		"Positive 3": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "fake-path", fakeListFnOk, false},
		"Positive 4": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "", fakeListFnOk, false},

		// Negative tests
		"Negative 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", fakeListFnOk, true},
		"Negative 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", fakeListFnOk, true},
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
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestVolumeSnapshotDataDelete(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		snapName            string
		delete              deleteFn
		expectErr           bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", "vsd-1", fakeDeleteFnOk, true},
		"Test 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fake-path2", "vsd-2", fakeDeleteFnOk, false},
		"Test 3": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", "vsd-3", fakeDeleteFnErr, true},
		"Test 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fakepath", "", fakeDeleteFnOk, true},
		"Test 5": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path2", "vsd1", fakeDeleteFnOk, true},
		"Test 6": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path2", "vsd1", fakeDeleteFnErr, true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				del:                 mock.delete,
			}
			err := k.Delete(mock.snapName, &metav1.DeleteOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestVolumeSnapshtDataGet(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		get                 getFn
		snapName            string
		expectErr           bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", fakeGetFnOk, "vsd-1", true},
		"Test 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", fakeGetFnOk, "vsd-1", true},
		"Test 3": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeGetFnOk, "vsd-2", false},
		"Test 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fp", fakeGetErrfn, "vsd-3", true},
		"Test 5": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fakepath", fakeGetFnOk, "", true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				get:                 mock.get,
			}
			_, err := k.Get(mock.snapName, metav1.GetOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestWithBuildOption(t *testing.T) {
	tests := map[string]struct {
		kubeConfigPath string
	}{
		"Test 1": {""},
		"Test 2": {"fake-path"},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := NewKubeClient(WithKubeConfigPath(mock.kubeConfigPath))
			if k.kubeConfigPath != mock.kubeConfigPath {
				t.Fatalf("Test %q failed: expected %v got %v", name, mock.kubeConfigPath, k.kubeConfigPath)
			}
		})
	}
}

func TestVolumeSnapshtDataPatch(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		patch               patchFn
		snapName            string
		expectErr           bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", fakePatchFnOk, "vsd-1", true},
		"Test 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", fakePatchFnOk, "vsd-1", true},
		"Test 3": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakePatchFnOk, "vsd-2", false},
		"Test 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fp", fakePatchFnErr, "vsd-3", true},
		"Test 5": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "fakepath", fakePatchFnOk, "", true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				patch:               mock.patch,
			}
			//fake data
			data, _ := json.Marshal(mock)
			_, err := k.Patch(mock.snapName, types.MergePatchType, data)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
