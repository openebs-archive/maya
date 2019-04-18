package v1alpha1

import (
	"errors"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func fakeGetClientset() (cli *kubernetes.Clientset, err error) {
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

func fakeGetErrClientSet() (clientset *kubernetes.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeClientSet(k *Kubeclient) {}

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
		name, mock := name, mock
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
		"Positive 1": {false, &Kubeclient{nil, "", fakeGetNilErrClientSet, fakeListOk, nil, nil, nil}},
		"Positive 2": {false, &Kubeclient{&kubernetes.Clientset{}, "", fakeGetNilErrClientSet, fakeListOk, nil, nil, nil}},

		// Negative tests
		"Negative 1": {true, &Kubeclient{nil, "", fakeGetErrClientSet, fakeListOk, nil, nil, nil}},
	}

	for name, mock := range tests {
		name, mock := name, mock
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

func TestKubenetesPVCList(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		list         listFn
		expectErr    bool
	}{
		"Test 1": {fakeGetErrClientSet, fakeListOk, true},
		"Test 2": {fakeGetClientset, fakeListOk, false},
		"Test 3": {fakeGetClientset, fakeListErr, true},
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
		"Test 1": {fakeGetErrClientSet, fakeGetOk, "testvol", "test-ns", true},
		"Test 2": {fakeGetClientset, fakeGetOk, "testvol", "test-ns", false},
		"Test 3": {fakeGetClientset, fakeGetErr, "testvol", "", true},
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
		"Test 1": {fakeGetErrClientSet, fakeDeleteOk, "testvol", true},
		"Test 2": {fakeGetClientset, fakeDeleteOk, "testvol", false},
		"Test 3": {fakeGetClientset, fakeDeleteErr, "testvol", true},
		"Test 4": {fakeGetClientset, fakeDeleteErr, "", true},
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
