package v1alpha1

import (
	"errors"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func fakeGetClientsetOk() (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeListfnOk(cli *clientset.Clientset, opts metav1.ListOptions) (*corev1.NodeList, error) {
	return &corev1.NodeList{}, nil
}

func fakeListErr(cli *clientset.Clientset, opts metav1.ListOptions) (*corev1.NodeList, error) {
	return nil, errors.New("some error")
}

func fakeGetClientSetNil() (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetClientSetErr() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func TestWithDefaultOptions(t *testing.T) {
	tests := map[string]struct {
		KubeClient *Kubeclient
	}{
		"When both listFn and getClientsetFn are error": {&Kubeclient{nil, fakeGetClientSetErr, fakeListErr}},
		"When both listFn and getClientsetFn are nil":   {&Kubeclient{}},
		"When listFn nil":                               {&Kubeclient{nil, fakeGetClientsetOk, nil}},
		"When getClientsetFn nil":                       {&Kubeclient{nil, nil, fakeListfnOk}},
		"When getClientsetFn and listFn are ok":         {&Kubeclient{nil, fakeGetClientsetOk, fakeListfnOk}},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mock.KubeClient.withDefaults()
			if mock.KubeClient.getClientset == nil {
				t.Fatalf("test %q failed: expected getClientset not to be empty", name)
			}
			if mock.KubeClient.list == nil {
				t.Fatalf("test %q failed: expected list not to be empty", name)
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
		"Positive 1": {false, &Kubeclient{nil, fakeGetClientSetNil, fakeListfnOk}},
		"Positive 2": {false, &Kubeclient{&clientset.Clientset{}, fakeGetClientSetNil, fakeListfnOk}},

		// Negative tests
		"Negative 1": {true, &Kubeclient{nil, fakeGetClientSetErr, fakeListfnOk}},
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

func TestKubenetesNodeList(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		list         listFn
		expectErr    bool
	}{
		"Test 1": {fakeGetClientSetErr, fakeListfnOk, true},
		"Test 2": {fakeGetClientsetOk, fakeListfnOk, false},
		"Test 3": {fakeGetClientsetOk, fakeListErr, true},
	}

	for name, mock := range tests {
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
