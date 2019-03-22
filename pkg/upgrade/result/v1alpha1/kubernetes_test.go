package v1alpha1

import (
	"errors"
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/upgrade/v1alpha1/clientset/internalclientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakeGetClientset() (cs *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeListfn(cs *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.UpgradeResultList, error) {
	return &apis.UpgradeResultList{}, nil
}

func fakeListErrfn(cs *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.UpgradeResultList, error) {
	return &apis.UpgradeResultList{}, errors.New("some error")
}

func fakeGetfn(cs *clientset.Clientset, name string, namespace string, opts metav1.GetOptions) (*apis.UpgradeResult, error) {
	return &apis.UpgradeResult{}, nil
}

func fakeGetErrfn(cs *clientset.Clientset, name string, namespace string, opts metav1.GetOptions) (*apis.UpgradeResult, error) {
	return &apis.UpgradeResult{}, errors.New("some error")
}

func fakeSetClientset(k *kubeclient) {
	k.clientset = &clientset.Clientset{}
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

func TestWithDefaults(t *testing.T) {
	tests := map[string]struct {
		expectListFn, expectGetFn, expectGetClientset bool
	}{
		"When mockclient is empty":                           {false, false, false},
		"When mockclient contains getClientsetFn":            {false, false, true},
		"When mockclient contains ListFn":                    {true, false, false},
		"When mockclient contains GetFn":                     {false, true, false},
		"When mockclient contains ListFn and getClientsetFn": {true, false, true},
		"When mockclient contains GetFn and getClientsetFn ": {false, true, true},
		"When mockclient contains ListFn and GetFn":          {true, true, false},
		"When mockclient contains all of them":               {true, true, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			fc := &kubeclient{}
			if !mock.expectListFn {
				fc.list = fakeListfn
			}
			if !mock.expectGetFn {
				fc.get = fakeGetfn
			}
			if !mock.expectGetClientset {
				fc.getClientset = fakeGetClientset
			}

			fc.withDefaults()
			if mock.expectListFn && fc.list == nil {
				t.Fatalf("test %q failed: expected fc.list not to be empty", name)
			}
			if mock.expectGetFn && fc.get == nil {
				t.Fatalf("test %q failed: expected fc.get not to be empty", name)
			}
			if mock.expectGetClientset && fc.getClientset == nil {
				t.Fatalf("test %q failed: expected fc.getClientset not to be empty", name)
			}
		})
	}
}
func TestWithClientset(t *testing.T) {
	tests := map[string]struct {
		Clientset             *clientset.Clientset
		expectKubeClientEmpty bool
	}{
		"Clientset is empty":     {nil, true},
		"Clientset is not empty": {&clientset.Clientset{}, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			h := WithClientset(mock.Clientset)
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
func TestKubeClient(t *testing.T) {
	tests := map[string]struct {
		expectClientSet bool
		opts            []kubeclientBuildOption
	}{
		"When non-nil clientset is passed":                       {true, []kubeclientBuildOption{fakeSetClientset}},
		"When two options with a non-nil clientset are passed":   {true, []kubeclientBuildOption{fakeSetClientset, fakeClientSet}},
		"When three options with a non-nil clientset are passed": {true, []kubeclientBuildOption{fakeSetClientset, fakeClientSet, fakeClientSet}},

		"When nil clientset is passed":                       {false, []kubeclientBuildOption{fakeSetNilClientset}},
		"When two options with a nil clientset are passed":   {false, []kubeclientBuildOption{fakeSetNilClientset, fakeClientSet}},
		"When three options with a nil clientset are passed": {false, []kubeclientBuildOption{fakeSetNilClientset, fakeClientSet, fakeClientSet}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := KubeClient(mock.opts...)
			if !mock.expectClientSet && c.clientset != nil {
				t.Fatalf("test %q failed expected c.clientset to be empty", name)
			}
			if mock.expectClientSet && c.clientset == nil {
				t.Fatalf("test %q failed expected c.clientset not to be empty", name)
			}
		})
	}
}
func TestWithNamespace(t *testing.T) {
	tests := map[string]struct {
		namespace            string
		expectNamespaceEmpty bool
	}{
		"Namespace is empty":     {"", true},
		"Namespace is not empty": {"abc", false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			h := WithNamespace(mock.namespace)
			fake := &kubeclient{}
			h(fake)
			if mock.expectNamespaceEmpty && fake.namespace != "" {
				t.Fatalf("test %q failed expected fake.namespace to be empty", name)
			}
			if !mock.expectNamespaceEmpty && fake.namespace == "" {
				t.Fatalf("test %q failed expected fake.namespace not to be empty", name)
			}
		})
	}
}
func TestGetClientOrCached(t *testing.T) {
	tests := map[string]struct {
		expectErr  bool
		KubeClient *kubeclient
	}{
		// Positive tests
		"When clientset is nil":     {false, &kubeclient{nil, "default", fakeGetNilErrClientSet, fakeListfn, fakeGetfn}},
		"When clientset is not nil": {false, &kubeclient{&clientset.Clientset{}, "", fakeGetNilErrClientSet, fakeListfn, fakeGetfn}},
		// Negative tests
		"When getting clientset throws error": {true, &kubeclient{nil, "", fakeGetErrClientSet, fakeListfn, fakeGetfn}},
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
			k := kubeclient{getClientset: mock.getClientset, get: mock.get}
			_, err := k.Get(mock.resourceName, metav1.GetOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
