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

func TestKubeClient(t *testing.T) {
	kubeclient := NewKubeClient()
	if reflect.DeepEqual(kubeclient, Kubeclient{}) {
		t.Fatalf("test failed: expect kubeclient not to be empty")
	}
}

func TestWithDefaultOptions(t *testing.T) {
	tests := map[string]struct {
		KubeClient *Kubeclient
	}{
		"When all getClientsetFn, listFn and getFn are error": {&Kubeclient{nil, fakeGetClientSetErr, fakeListFnErr, fakeGetFnErr, fakeCreateFnErr, fakeDeleteFnErr}},
		"When all are nil":                         {&Kubeclient{}},
		"When getClientSet is error":               {&Kubeclient{nil, fakeGetClientSetErr, nil, nil, nil, nil}},
		"When ListFn is error":                     {&Kubeclient{nil, nil, fakeListFnErr, nil, nil, nil}},
		"When GetFn is error":                      {&Kubeclient{nil, nil, nil, fakeGetFnErr, nil, nil}},
		"When listFn and getFn are error":          {&Kubeclient{nil, fakeGetClientSetOk, fakeListFnErr, fakeGetFnErr, fakeCreateFnOk, fakeDeleteFnOk}},
		"When getClientsetFn and listFn are error": {&Kubeclient{nil, fakeGetClientSetErr, fakeListFnErr, fakeGetFnOk, fakeCreateFnOk, fakeDeleteFnOk}},
		"When getClientsetFn and getFn are error":  {&Kubeclient{nil, fakeGetClientSetErr, fakeListFnOk, fakeGetFnErr, fakeCreateFnOk, fakeDeleteFnOk}},
		"When CreateFn and DeleteFn are error":     {&Kubeclient{nil, fakeGetClientSetErr, fakeListFnOk, fakeGetFnErr, fakeCreateFnErr, fakeDeleteFnErr}},
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
			if mock.KubeClient.create == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
			if mock.KubeClient.del == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
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
		"Positive 1": {
			KubeClient: &Kubeclient{nil, fakeGetClientSetNil, fakeListFnOk, fakeGetFnErr, fakeCreateFnErr, fakeDeleteFnErr},
			expectErr:  false,
		},
		"Positive 2": {
			KubeClient: &Kubeclient{&clientset.Clientset{}, fakeGetClientSetOk, fakeListFnOk, fakeGetFnOk, fakeCreateFnOk, fakeDeleteFnOk},
			expectErr:  false,
		},

		// Negative tests
		"Negative 1": {
			KubeClient: &Kubeclient{nil, fakeGetClientSetErr, fakeListFnOk, fakeGetFnOk, nil, nil},
			expectErr:  true,
		},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			c, err := mock.KubeClient.getClientsetOrCached()
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

func TestKubenetesStorageClassCreate(t *testing.T) {
	tests := map[string]struct {
		getClientSet getClientsetFn
		create       createFn
		sc           *storagev1.StorageClass
		expectErr    bool
	}{
		"Negative Test 1": {
			getClientSet: fakeGetClientSetErr,
			create:       fakeCreateFnOk,
			sc:           &storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "SC-1"}},
			expectErr:    true,
		},
		"Negative Test 2": {
			getClientSet: fakeGetClientSetOk,
			create:       fakeCreateFnErr,
			sc:           &storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "SC-2"}},
			expectErr:    true,
		},
		"Negative Test 3": {
			getClientSet: fakeGetClientSetOk,
			create:       fakeCreateFnErr,
			sc:           nil,
			expectErr:    true,
		},
		"Positive Test 4": {
			getClientSet: fakeGetClientSetOk,
			create:       fakeCreateFnOk,
			sc:           nil,
			expectErr:    false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientSet, create: mock.create}
			_, err := k.Create(mock.sc)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubenetesStorageClassDelete(t *testing.T) {
	tests := map[string]struct {
		getClientSet getClientsetFn
		del          deleteFn
		scName       string
		expectErr    bool
	}{
		"Negative Test 1": {
			getClientSet: fakeGetClientSetErr,
			del:          fakeDeleteFnOk,
			scName:       "SC1",
			expectErr:    true,
		},
		"Negative Test 2": {
			getClientSet: fakeGetClientSetOk,
			del:          fakeDeleteFnErr,
			scName:       "SC2",
			expectErr:    true,
		},
		"Negative Test 3": {
			getClientSet: fakeGetClientSetOk,
			del:          fakeDeleteFnErr,
			scName:       "",
			expectErr:    true,
		},
		"Positive Test 4": {
			getClientSet: fakeGetClientSetOk,
			del:          fakeDeleteFnOk,
			scName:       "",
			expectErr:    false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientSet, del: mock.del}
			err := k.Delete(mock.scName, &metav1.DeleteOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
