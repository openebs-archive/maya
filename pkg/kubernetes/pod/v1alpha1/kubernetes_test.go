package v1alpha1

import (
	"errors"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
)

func fakeGetClientset() (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeListfn(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*v1.PodList, error) {
	return &v1.PodList{}, nil
}

func fakeListErrfn(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*v1.PodList, error) {
	return &v1.PodList{}, errors.New("some error")
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

func TestWithDefaultOptions(t *testing.T) {
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

func TestWithClientsetBuildOption(t *testing.T) {
	tests := map[string]struct {
		Clientset             *client.Clientset
		expectKubeClientEmpty bool
	}{
		"Clientset is empty":     {nil, true},
		"Clientset is not empty": {&client.Clientset{}, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			h := WithClientSet(mock.Clientset)
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

func TestKubeClientBuildOption(t *testing.T) {
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

func TestGetClientOrCached(t *testing.T) {
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

func TestKubenetesPodList(t *testing.T) {
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
			k := kubeclient{getClientset: mock.getClientset, namespace: "", list: mock.list}
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
		"Test 2": {"namespace 1"},
		"Test 3": {"namespace 2"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := KubeClient(WithNamespace(mock.namespace))
			if k.namespace != mock.namespace {
				t.Fatalf("Test %q failed: expected %v got %v", name, mock.namespace, k.namespace)
			}
		})
	}
}
