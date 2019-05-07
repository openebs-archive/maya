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

package v1alpha2

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakeListOk(cs *clientset.Clientset,
	opts metav1.ListOptions) (*apis.CStorPoolList, error) {
	return &apis.CStorPoolList{}, nil
}

func fakeListErr(cs *clientset.Clientset,
	opts metav1.ListOptions) (*apis.CStorPoolList, error) {
	return nil, errors.New("some error")
}

func fakeGetOk(cs *clientset.Clientset, name string,
	opts metav1.GetOptions) (*apis.CStorPool, error) {
	return &apis.CStorPool{}, nil
}

func fakeGetErr(cs *clientset.Clientset, name string,
	opts metav1.GetOptions) (*apis.CStorPool, error) {
	return nil, errors.New("some error")
}

func fakeCreateOk(cs *clientset.Clientset,
	obj *apis.CStorPool) (*apis.CStorPool, error) {
	return &apis.CStorPool{}, nil
}

func fakeCreateErr(cs *clientset.Clientset,
	obj *apis.CStorPool) (*apis.CStorPool, error) {
	return nil, errors.New("some error")
}

func fakePatchOk(cs *clientset.Clientset, name string,
	pt types.PatchType, patchObj []byte) (*apis.CStorPool, error) {
	return &apis.CStorPool{}, nil
}

func fakePatchErr(cs *clientset.Clientset, name string, pt types.PatchType,
	patchObj []byte) (*apis.CStorPool, error) {
	return nil, errors.New("some error")
}

func fakeDeleteOk(cs *clientset.Clientset, name string,
	opt *metav1.DeleteOptions) error {
	return nil
}

func fakeDeleteErr(cs *clientset.Clientset, name string,
	opt *metav1.DeleteOptions) error {
	return errors.New("some error")
}

func fakeGetClientsetOk() (cs *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeGetClientsetNil() (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetClientsetErr() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func TestWithDefaults(t *testing.T) {
	tests := map[string]struct {
		listFn             listFunc
		getFn              getFunc
		getClientsetFn     getClientsetFunc
		createFn           createFunc
		patchFn            patchFunc
		deleteFn           delFn
		expectList         bool
		expectGet          bool
		expectGetClientset bool
		expectCreate       bool
		expectPatch        bool
		expectDelete       bool
	}{
		"When mockclient is empty": {nil, nil, nil, nil, nil, nil, true, true, true, true, true, true},
		"When mockclient contains all of them": {fakeListOk, fakeGetOk,
			fakeGetClientsetOk, fakeCreateOk, fakePatchOk, fakeDeleteOk, true, true, true, true, true, true},
	}

	for name, mock := range tests {
		name := name //pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{}
			fc.list = mock.listFn
			fc.get = mock.getFn
			fc.getClientset = mock.getClientsetFn
			fc.create = mock.createFn
			fc.patch = mock.patchFn

			fc.withDefaults()

			if (fc.list != nil) != mock.expectList {
				t.Fatalf("test %s failed: expected non-nil fc.list but got %v", name, fc.list)
			}

			if (fc.get != nil) != mock.expectGet {
				t.Fatalf("test %s failed: expected non-nil fc.get but got %v", name, fc.get)
			}

			if (fc.getClientset != nil) != mock.expectGetClientset {
				t.Fatalf("test %s failed: expected non-nil fc.getClientset but got %v", name, fc.getClientset)
			}

			if (fc.create != nil) != mock.expectCreate {
				t.Fatalf("test %s failed: expected non-nil fc.create but got %v", name, fc.create)
			}

			if (fc.patch != nil) != mock.expectPatch {
				t.Fatalf("test %s failed: expected non-nil fc.patch but got %v", name, fc.patch)
			}

			if (fc.del != nil) != mock.expectDelete {
				t.Fatalf("test %s failed: expected non-nil fc.patch but got %v", name, fc.del)
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
		name := name //pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			h := WithClientset(mock.clientSet)
			fake := &Kubeclient{}
			h(fake)

			if mock.isKubeClient && fake.clientset == nil {
				t.Fatalf("test %s failed, expected non-nil fake.clientset but got %v", name, fake.clientset)
			}

			if !mock.isKubeClient && fake.clientset != nil {
				t.Fatalf("test %s failed, expected nil fake.clientset but got %v", name, fake.clientset)
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
			fakeGetClientsetNil, fakeListOk, fakeGetOk, fakeCreateOk, fakePatchOk, fakeDeleteOk}, false},
		"When clientset is not nil": {&Kubeclient{&clientset.Clientset{},
			fakeGetClientsetOk, fakeListOk, fakeGetOk, fakeCreateOk, fakePatchOk, fakeDeleteOk}, false},
		// Negative tests
		"When getting clientset throws error": {&Kubeclient{nil,
			fakeGetClientsetErr, fakeListOk, fakeGetOk, fakeCreateOk, fakePatchOk, fakeDeleteOk}, true},
	}

	for name, mock := range tests {
		name := name //pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			c, err := mock.kubeClient.getClientsetOrCached()

			if mock.expectErr && err == nil {
				t.Fatalf("test %s failed : expected error but got %v", name, err)
			}

			if !mock.expectErr && err != nil {
				t.Fatalf("test %s failed : expected nil error but got %v", name, err)
			}

			if !reflect.DeepEqual(c, mock.kubeClient.clientset) {
				t.Fatalf("test %s failed : expected clientset %v but got %v", name, mock.kubeClient.clientset, c)
			}
		})
	}
}

func TestKubernetesList(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFunc
		list         listFunc
		expectErr    bool
	}{
		"When getting clientset throws error": {fakeGetClientsetErr, fakeListOk, true},
		"When listing resource throws error":  {fakeGetClientsetOk, fakeListErr, true},
		"When none of them throws error":      {fakeGetClientsetOk, fakeListOk, false},
	}

	for name, mock := range tests {
		name := name //pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, list: mock.list}
			_, err := k.List(metav1.ListOptions{})

			if mock.expectErr && err == nil {
				t.Fatalf("test %s failed: expected error but got %v", name, err)
			}

			if !mock.expectErr && err != nil {
				t.Fatalf("test %s failed: expected nil but got %v", name, err)
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
		"When getting clientset throws error": {"sp1", fakeGetClientsetErr, fakeGetOk, true},
		"When getting resource throws error":  {"sp2", fakeGetClientsetOk, fakeGetErr, true},
		"When resource name is empty string":  {"", fakeGetClientsetOk, fakeGetOk, true},
		"When none of them throws error":      {"sp3", fakeGetClientsetOk, fakeGetOk, false},
	}

	for name, mock := range tests {
		name := name //pin it
		mock := mock // pin it
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

func TestKubernetesCreate(t *testing.T) {
	tests := map[string]struct {
		obj          *apis.CStorPool
		getClientset getClientsetFunc
		create       createFunc
		expectErr    bool
	}{
		"When getting clientset throws error": {
			&apis.CStorPool{},
			fakeGetClientsetErr,
			fakeCreateOk, true},
		"When creating resource throws error": {
			&apis.CStorPool{},
			fakeGetClientsetOk,
			fakeCreateErr,
			true},
		"When CStorPool object is nil": {
			nil,
			fakeGetClientsetOk,
			fakeCreateOk,
			true},
		"When an empty CStorPool struct is given": {
			&apis.CStorPool{},
			fakeGetClientsetOk,
			fakeCreateOk,
			false},
		"When non-empty CStorPool struct is given": {
			&apis.CStorPool{
				ObjectMeta: metav1.ObjectMeta{Name: "csp-cjed"},
			},
			fakeGetClientsetOk,
			fakeCreateOk,
			false},
	}

	for name, mock := range tests {
		name := name //pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, create: mock.create}
			_, err := k.Create(mock.obj)

			if mock.expectErr && err == nil {
				t.Fatalf("test %s failed: expected error but got %v", name, err)
			}

			if !mock.expectErr && err != nil {
				t.Fatalf("test %s failed: expected nil but got %v", name, err)
			}
		})
	}
}

func TestKubernetesPatch(t *testing.T) {
	var patchObjStr = "{metadata:{labels:{openebs.io/version: 0.8.1}}}"
	tests := map[string]struct {
		resourceName string
		patchType    types.PatchType
		obj          []byte
		getClientset getClientsetFunc
		patch        patchFunc
		expectErr    bool
	}{
		"When get clientset throws error": {
			"ur1", "application/merge-patch+json", []byte{},
			fakeGetClientsetErr,
			fakePatchOk,
			true},
		"When patch resource throws error": {
			"ur2", "application/json-patch+json", []byte{},
			fakeGetClientsetOk,
			fakePatchErr,
			true},
		"When patch object name is empty string": {
			"", "application/merge-patch+json", nil,
			fakeGetClientsetOk,
			fakePatchOk,
			true},
		"When patch object is nil": {
			"ur3", "application/merge-patch+json", nil,
			fakeGetClientsetOk,
			fakePatchOk,
			false},
		"When non-empty patch obj is given": {
			"ur5", "application/strategic-merge-patch+json", []byte(patchObjStr),
			fakeGetClientsetOk,
			fakePatchOk,
			false},
	}

	for name, mock := range tests {
		name := name //pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, patch: mock.patch}
			_, err := k.Patch(mock.resourceName, mock.patchType, mock.obj)

			if mock.expectErr && err == nil {
				t.Fatalf("test %s failed: expected error but got %v", name, err)
			}

			if !mock.expectErr && err != nil {
				t.Fatalf("test %s failed: expected nil but got %v", name, err)
			}
		})
	}
}

func TestKubernetesDelete(t *testing.T) {
	tests := map[string]struct {
		csp          string
		getClientset getClientsetFunc
		del          delFn
		expectErr    bool
	}{
		"When getting clientset throws error": {
			"csp1",
			fakeGetClientsetErr,
			fakeDeleteOk, true},
		"When delete resource throws error": {
			"",
			fakeGetClientsetOk,
			fakeDeleteErr,
			true},
		"When non-empty StoragePool name is given": {
			"csp",
			fakeGetClientsetOk,
			fakeDeleteOk,
			false},
	}

	for name, mock := range tests {
		name := name //pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, del: mock.del}
			err := k.Delete(mock.csp, &metav1.DeleteOptions{})

			if mock.expectErr && err == nil {
				t.Fatalf("test %s failed: expected error but got %v", name, err)
			}

			if !mock.expectErr && err != nil {
				t.Fatalf("test %s failed: expected nil but got %v", name, err)
			}
		})
	}
}
