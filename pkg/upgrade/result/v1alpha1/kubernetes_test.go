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

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/upgrade/v1alpha1/clientset/internalclientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func fakeGetClientset() (cs *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeGetClientsetForPath(kubeConfigPath string) (cs *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeListfn(cs *clientset.Clientset, namespace string,
	opts metav1.ListOptions) (*apis.UpgradeResultList, error) {
	return &apis.UpgradeResultList{}, nil
}

func fakeListErrfn(cs *clientset.Clientset, namespace string,
	opts metav1.ListOptions) (*apis.UpgradeResultList, error) {
	return &apis.UpgradeResultList{}, errors.New("some error")
}

func fakeGetfn(cs *clientset.Clientset, name string, namespace string,
	opts metav1.GetOptions) (*apis.UpgradeResult, error) {
	return &apis.UpgradeResult{}, nil
}

func fakeGetErrfn(cs *clientset.Clientset, name string, namespace string,
	opts metav1.GetOptions) (*apis.UpgradeResult, error) {
	return &apis.UpgradeResult{}, errors.New("some error")
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

func fakeCreateOk(cs *clientset.Clientset, upgradeResultObj *apis.UpgradeResult,
	namespace string) (*apis.UpgradeResult, error) {
	return &apis.UpgradeResult{}, nil
}

func fakeCreateErr(cs *clientset.Clientset, upgradeResultObj *apis.UpgradeResult,
	namespace string) (*apis.UpgradeResult, error) {
	return &apis.UpgradeResult{}, errors.New("some error")
}

func fakePatchOk(cs *clientset.Clientset, name string, pt types.PatchType, patchObj []byte,
	namespace string) (*apis.UpgradeResult, error) {
	return &apis.UpgradeResult{}, nil
}

func fakePatchErr(cs *clientset.Clientset, name string, pt types.PatchType, patchObj []byte,
	namespace string) (*apis.UpgradeResult, error) {
	return &apis.UpgradeResult{}, errors.New("some error")
}

func fakeUpdateOk(cs *clientset.Clientset, updateObj *apis.UpgradeResult,
	namespace string) (*apis.UpgradeResult, error) {
	return &apis.UpgradeResult{}, nil
}

func fakeUpdateErr(cs *clientset.Clientset, updateObj *apis.UpgradeResult,
	namespace string) (*apis.UpgradeResult, error) {
	return &apis.UpgradeResult{}, errors.New("some error")
}

func TestWithDefaults(t *testing.T) {
	tests := map[string]struct {
		listFn             listFunc
		getFn              getFunc
		getClientsetFn     getClientsetFunc
		createFn           createFunc
		patchFn            patchFunc
		updateFn           updateFunc
		expectList         bool
		expectGet          bool
		expectGetClientset bool
		expectCreate       bool
		expectPatch        bool
		expectUpdate       bool
	}{
		// The current implementation of WithDefaults method can be
		// tested using these two combinations only.
		"When mockclient is empty": {nil, nil, nil, nil, nil, nil, false, false, false, false, false, false},
		"When mockclient contains all of them": {fakeListfn, fakeGetfn,
			fakeGetClientset, fakeCreateOk, fakePatchOk, fakeUpdateOk, false, false, false, false, false, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{}
			fc.list = mock.listFn
			fc.get = mock.getFn
			fc.getClientset = mock.getClientsetFn
			fc.create = mock.createFn
			fc.patch = mock.patchFn

			fc.withDefaults()
			list := (fc.list == nil)
			if list != mock.expectList {
				t.Fatalf(`test %s failed: expected non-nil fc.list
but got %v`, name, fc.list)
			}
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
			create := (fc.create == nil)
			if create != mock.expectCreate {
				t.Fatalf(`test %s failed: expected non-nil fc.create
but got %v`, name, fc.create)
			}
			patch := (fc.patch == nil)
			if patch != mock.expectPatch {
				t.Fatalf(`test %s failed: expected non-nil fc.patch
but got %v`, name, fc.patch)
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
func TestNewKubeClientWithClientset(t *testing.T) {
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
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			c := NewKubeClient(mock.opts...)
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
func TestWithNamespace(t *testing.T) {
	tests := map[string]struct {
		namespace         string
		expectedNamespace string
	}{
		"Namespace is empty":     {"", ""},
		"Namespace is not empty": {"abc", "abc"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			fake := &Kubeclient{}
			fake.WithNamespace(mock.namespace)
			if fake.namespace != mock.expectedNamespace {
				t.Fatalf(`test %s failed, expected %v got %v`,
					name, mock.expectedNamespace, fake.namespace)
			}
		})
	}
}
func TestGetClientOrCached(t *testing.T) {
	tests := map[string]struct {
		Kubeclient *Kubeclient
		expectErr  bool
	}{
		// Positive tests
		"When clientset is nil": {&Kubeclient{nil, "default", "",
			fakeGetNilErrClientSet, fakeListfn, fakeGetfn, fakeCreateOk,
			fakePatchOk, fakeUpdateOk, fakeGetClientsetForPath}, false},
		"When clientset is not nil": {&Kubeclient{&clientset.Clientset{},
			"", "", fakeGetNilErrClientSet, fakeListfn, fakeGetfn, fakeCreateOk,
			fakePatchOk, fakeUpdateOk, fakeGetClientsetForPath}, false},
		// Negative tests
		"When getting clientset throws error": {&Kubeclient{nil, "", "",
			fakeGetErrClientSet, fakeListfn, fakeGetfn, fakeCreateOk,
			fakePatchOk, fakeUpdateOk, fakeGetClientsetForPath}, true},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			c, err := mock.Kubeclient.getClientOrCached()
			if mock.expectErr && err == nil {
				t.Fatalf("test %s failed : expected error but got %v", name, err)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("test %s failed : expected nil error but got %v", name, err)
			}
			if !reflect.DeepEqual(c, mock.Kubeclient.clientset) {
				t.Fatalf(`test %s failed : expected clientset %v
but got %v`, name, mock.Kubeclient.clientset, c)
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
		"When getting clientset throws error": {fakeGetErrClientSet, fakeListfn, true},
		"When listing resource throws error":  {fakeGetClientset, fakeListErrfn, true},
		"When none of them throws error":      {fakeGetClientset, fakeListfn, false},
	}

	for name, mock := range tests {
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

func TestKubernetesCreate(t *testing.T) {
	var upgradeResultObject = &apis.UpgradeResult{
		ObjectMeta: metav1.ObjectMeta{Name: "upgradeResult1"}}
	tests := map[string]struct {
		upgradeResultObj *apis.UpgradeResult
		getClientset     getClientsetFunc
		create           createFunc
		expectErr        bool
	}{
		"When getting clientset throws error": {
			&apis.UpgradeResult{},
			fakeGetErrClientSet,
			fakeCreateOk, true},
		"When creating resource throws error": {
			&apis.UpgradeResult{},
			fakeGetClientset,
			fakeCreateErr,
			true},
		"When upgradeResult object is nil": {
			nil,
			fakeGetClientset,
			fakeCreateOk,
			false},
		"When an empty upgradeResult struct is given": {
			&apis.UpgradeResult{},
			fakeGetClientset,
			fakeCreateOk,
			false},
		"When non-empty upgradeResult struct is given": {
			upgradeResultObject,
			fakeGetClientset,
			fakeCreateOk,
			false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, create: mock.create}
			_, err := k.Create(mock.upgradeResultObj)
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
	var patchObjStr = "{status:{actualCount:611,desiredCount:611}}"
	tests := map[string]struct {
		resourceName     string
		patchType        types.PatchType
		upgradeResultObj []byte
		getClientset     getClientsetFunc
		patch            patchFunc
		expectErr        bool
	}{
		"When get clientset throws error": {
			"ur1", "application/merge-patch+json", []byte{},
			fakeGetErrClientSet,
			fakePatchOk,
			true},
		"When patch resource throws error": {
			"ur2", "application/json-patch+json", []byte{},
			fakeGetClientset,
			fakePatchErr,
			true},
		"When patch object name is empty string": {
			"", "application/merge-patch+json", nil,
			fakeGetClientset,
			fakePatchOk,
			true},
		"When patch object is nil": {
			"ur3", "application/merge-patch+json", nil,
			fakeGetClientset,
			fakePatchOk,
			false},
		"When non-empty patch obj is given": {
			"ur5", "application/strategic-merge-patch+json", []byte(patchObjStr),
			fakeGetClientset,
			fakePatchOk,
			false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, patch: mock.patch}
			_, err := k.Patch(mock.resourceName, mock.patchType, mock.upgradeResultObj)
			if mock.expectErr && err == nil {
				t.Fatalf("test %s failed: expected error but got %v", name, err)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("test %s failed: expected nil but got %v", name, err)
			}
		})
	}
}
