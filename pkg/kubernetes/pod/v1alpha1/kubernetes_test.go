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

func fakeGetClientSetOk() (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeListFnOk(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*v1.PodList, error) {
	return &v1.PodList{}, nil
}

func fakeListFnErr(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*v1.PodList, error) {
	return &v1.PodList{}, errors.New("some error")
}

func fakeDeleteFnOk(cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error {
	return nil
}

func fakeDeleteFnErr(cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error {
	return errors.New("some error while delete")
}

func fakeGetFnOk(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*v1.Pod, error) {
	return &v1.Pod{}, nil
}

func fakeGetErrfn(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*v1.Pod, error) {
	return &v1.Pod{}, errors.New("Not found")
}

func fakeSetClientset(k *Kubeclient) {
	k.clientset = &client.Clientset{}
}

func fakeSetNilClientset(k *Kubeclient) {
	k.clientset = nil
}

func fakeGetNilErrClientSet() (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetClientSetErr() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeClientSet(k *Kubeclient) {}

func TestWithDefaultOptions(t *testing.T) {
	tests := map[string]struct {
		kubeClient *Kubeclient
	}{
		"When all are nil": {&Kubeclient{}},
		"When clientset is nil": {&Kubeclient{
			clientset:    nil,
			getClientset: fakeGetClientSetOk,
			list:         fakeListFnOk,
			get:          fakeGetFnOk,
			del:          fakeDeleteFnOk,
		}},
		"When listFn nil": {&Kubeclient{
			getClientset: fakeGetClientSetOk,
			list:         nil,
			get:          fakeGetFnOk,
			del:          fakeDeleteFnOk,
		}},
		"When getClientsetFn nil": {&Kubeclient{
			getClientset: nil,
			list:         fakeListFnOk,
			get:          fakeGetFnOk,
			del:          fakeDeleteFnOk,
		}},
		"When getFn and CreateFn are nil": {&Kubeclient{
			getClientset: fakeGetClientSetOk,
			list:         fakeListFnOk,
			get:          nil,
			del:          fakeDeleteFnOk,
		}},
		"When all are error": {&Kubeclient{
			getClientset: fakeGetClientSetErr,
			list:         fakeListFnErr,
			get:          nil,
			del:          fakeDeleteFnErr,
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
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
			if mock.kubeClient.del == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
			}
			if mock.kubeClient.getClientset == nil {
				t.Fatalf("test %q failed: expected get not to be empty", name)
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
		name, mock := name, mock
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
		KubeClient *Kubeclient
	}{
		// Positive tests
		"Positive 1": {false, &Kubeclient{nil, "", fakeGetNilErrClientSet, fakeListFnOk, fakeDeleteFnOk, fakeGetFnOk}},
		"Positive 2": {false, &Kubeclient{&client.Clientset{}, "", fakeGetNilErrClientSet, fakeListFnOk, fakeDeleteFnOk, fakeGetFnOk}},

		// Negative tests
		"Negative 1": {true, &Kubeclient{nil, "", fakeGetClientSetErr, fakeListFnOk, fakeDeleteFnOk, fakeGetFnOk}},
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

func TestKubenetesPodList(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		list         listFn
		expectErr    bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeListFnOk, true},
		"Test 2": {fakeGetClientSetOk, fakeListFnOk, false},
		"Test 3": {fakeGetClientSetOk, fakeListFnErr, true},
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

func TestKubenetesDeletePod(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		delete       deleteFn
		podName      string
		expectErr    bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeDeleteFnOk, "pod-1", true},
		"Test 2": {fakeGetClientSetOk, fakeDeleteFnOk, "pod-2", false},
		"Test 3": {fakeGetClientSetOk, fakeDeleteFnErr, "pod-3", true},
		"Test 4": {fakeGetClientSetOk, fakeDeleteFnOk, "", true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, namespace: "", del: mock.delete}
			err := k.Delete(mock.podName, &metav1.DeleteOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubenetesGetPod(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		get          getFn
		podName      string
		expectErr    bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeGetFnOk, "pod-1", true},
		"Test 2": {fakeGetClientSetOk, fakeGetFnOk, "pod-2", false},
		"Test 3": {fakeGetClientSetOk, fakeGetErrfn, "pod-3", true},
		"Test 4": {fakeGetClientSetOk, fakeGetFnOk, "", true},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := Kubeclient{getClientset: mock.getClientset, namespace: "", get: mock.get}
			_, err := k.Get(mock.podName, metav1.GetOptions{})
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
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			k := KubeClient(WithNamespace(mock.namespace))
			if k.namespace != mock.namespace {
				t.Fatalf("Test %q failed: expected %v got %v", name, mock.namespace, k.namespace)
			}
		})
	}
}
