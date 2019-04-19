package v1alpha1

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
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

func TestKubeClient(t *testing.T) {
	kubeclient := KubeClient()
	if reflect.DeepEqual(kubeclient, Kubeclient{}) {
		t.Fatalf("test failed: expect kubeclient not to be empty")
	}
}

func TestWithDefaultOptions(t *testing.T) {
	tests := map[string]struct {
		KubeClient *Kubeclient
	}{
		"When all getClientsetFn, listFn and getFn are error": {&Kubeclient{nil, fakeGetClientSetErr, fakeListFnErr, fakeGetFnErr}},
		"When all are nil":                         {&Kubeclient{}},
		"When getClientSet is error":               {&Kubeclient{nil, fakeGetClientSetErr, nil, nil}},
		"When ListFn is error":                     {&Kubeclient{nil, nil, fakeListFnErr, nil}},
		"When GetFn is error":                      {&Kubeclient{nil, nil, nil, fakeGetFnErr}},
		"When listFn and getFn are error":          {&Kubeclient{nil, fakeGetClientSetOk, fakeListFnErr, fakeGetFnErr}},
		"When getClientsetFn and listFn are error": {&Kubeclient{nil, fakeGetClientSetErr, fakeListFnErr, fakeGetFnOk}},
		"When getClientsetFn and getFn are error":  {&Kubeclient{nil, fakeGetClientSetErr, fakeListFnOk, fakeGetFnErr}},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			mock.KubeClient.withDefaults()
			if mock.KubeClient.getClientset == nil {
				t.Fatalf("test %q failed: expected getClientset not to be empty", name)
			}
			if mock.KubeClient.list == nil {
				t.Fatalf("test %q failed: expected list not to be empty", name)
			}
			if mock.KubeClient.get == nil {
				t.Fatalf("test %q failed: expected get not to be emptu", name)
			}
		})
	}
}

func TestGetClientOrCached(t *testing.T) {
	tests := map[string]struct {
		KubeClient *Kubeclient
		expectErr  bool
	}{
		// Positive tests
		"Positive 1": {&Kubeclient{nil, fakeGetClientSetNil, fakeListFnOk, fakeGetFnErr}, false},
		"Positive 2": {&Kubeclient{&clientset.Clientset{}, fakeGetClientSetOk, fakeListFnOk, fakeGetFnOk}, false},

		// Negative tests
		"Negative 1": {&Kubeclient{nil, fakeGetClientSetErr, fakeListFnOk, fakeGetFnOk}, true},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
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

func TestKubenetesStorageClassList(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		list         listFn
		expectErr    bool
	}{
		// Negative tests
		"When GetClientSetErr": {fakeGetClientSetErr, fakeListFnOk, true},
		"When ListFnErr":       {fakeGetClientSetOk, fakeListFnErr, true},
		// Positive tests
		"When GetClientSetNil": {fakeGetClientSetNil, fakeListFnOk, false},
		"When both are ok":     {fakeGetClientSetOk, fakeListFnOk, false},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, list: mock.list}
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

func TestKubenetesStorageClassGet(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		get          getFn
		name         string
		expectErr    bool
	}{
		// Negative tests
		"When GetClientSetErr": {fakeGetClientSetErr, fakeGetFnOk, "SC1", true},
		"When GetFnErr":        {fakeGetClientSetOk, fakeGetFnErr, "SC2", true},
		// Positive tests
		"When GetClientSetNil": {fakeGetClientSetNil, fakeGetFnOk, "SC3", false},
		"When both are ok":     {fakeGetClientSetOk, fakeGetFnOk, "SC4", false},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, get: mock.get}
			_, err := k.Get(name, metav1.GetOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
