/*
Copyright 2019 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/openebscluster/v1alpha1"
	"github.com/openebs/maya/pkg/client/generated/openebs.io/openebscluster/v1alpha1/clientset/internalclientset"
	kube "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"testing"
)

func fakeGetOpenebsClusterOk(k *kubeclient, cs *internalclientset.Clientset, name string) (*apis.OpenebsCluster, error) {
	return &apis.OpenebsCluster{}, nil
}

func fakeGetOpenebsClusterErr(k *kubeclient, cs *internalclientset.Clientset, name string) (*apis.OpenebsCluster, error) {
	return nil, errors.New("fake error")
}

func fakeGetConfigOk(c *kube.Client) (*rest.Config, error) {
	return &rest.Config{}, nil
}

func fakeGetConfigErr(c *kube.Client) (*rest.Config, error) {
	return nil, errors.New("fake error")
}

func fakeGetClientsetOk(c *rest.Config) (*internalclientset.Clientset, error) {
	return &internalclientset.Clientset{}, nil
}

func fakeGetClientsetErr(c *rest.Config) (*internalclientset.Clientset, error) {
	return nil, errors.New("fake error")
}

func TestKubeClient(t *testing.T) {
	tests := map[string]struct {
		opts []KubeClientOptionFunc
	}{
		"t1": {[]KubeClientOptionFunc{}},
		"t2": {[]KubeClientOptionFunc{WithNamespace("default"), InCluster()}},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := KubeClient(mock.opts...)
			if k == nil {
				t.Fatalf("test '%s' failed: expected not nil kubeclient actual nil", name)
			}
			if k.getConfig == nil {
				t.Fatalf("test '%s' failed: expected not nil getConfig actual nil", name)
			}
			if k.getClientset == nil {
				t.Fatalf("test '%s' failed: expected not nil getClientset actual nil", name)
			}
		})
	}
}

func TestKubeClientBuilder(t *testing.T) {
	k := KubeClientBuilder()
	if k == nil {
		t.Fatalf("test failed: expected not nil kubeclient builder actual nil")
	}
}

func TestKubeClientBuilderWithNamespace(t *testing.T) {
	tests := map[string]struct {
		namespace string
	}{
		"t1": {""},
		"t2": {"default"},
		"t3": {"openebs"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := KubeClientBuilder().WithNamespace(mock.namespace).Build()
			if k == nil {
				t.Fatalf("test '%s' failed: expected not nil kubeclient actual nil", name)
			}
			if k.getConfig == nil {
				t.Fatalf("test '%s' failed: expected not nil getConfig actual nil", name)
			}
			if k.getClientset == nil {
				t.Fatalf("test '%s' failed: expected not nil getClientset actual nil", name)
			}
		})
	}
}

func TestKubeClientBuilderInCluster(t *testing.T) {
	k := KubeClientBuilder().InCluster().Build()
	if k == nil {
		t.Fatalf("test failed: expected not nil kubeclient actual nil")
	}
	if k.getConfig == nil {
		t.Fatalf("test failed: expected not nil getConfig actual nil")
	}
	if k.getClientset == nil {
		t.Fatalf("test failed: expected not nil getClientset actual nil")
	}
}

func TestGetInternalClientset(t *testing.T) {
	tests := map[string]struct {
		getConfig    kube.GetConfigFunc
		getClientset getClientsetFunc
		isErr        bool
	}{
		"t10": {fakeGetConfigOk, fakeGetClientsetOk, false},
		"t11": {fakeGetConfigErr, fakeGetClientsetOk, true},
		"t12": {fakeGetConfigOk, fakeGetClientsetErr, true},
		"t13": {fakeGetConfigErr, fakeGetClientsetErr, true},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := &kubeclient{
				getConfig:    mock.getConfig,
				getClientset: mock.getClientset,
			}
			_, err := k.getInternalClientset()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := map[string]struct {
		name              string
		getConfig         kube.GetConfigFunc
		getClientset      getClientsetFunc
		getOpenebsCluster getOpenebsClusterFunc
		isErr             bool
	}{
		"t10": {"mytask", fakeGetConfigOk, fakeGetClientsetOk, fakeGetOpenebsClusterOk, false},
		"t11": {"", fakeGetConfigOk, fakeGetClientsetOk, fakeGetOpenebsClusterOk, true},
		"t12": {"mytask", fakeGetConfigErr, nil, nil, true},
		"t13": {"mytask", fakeGetConfigOk, fakeGetClientsetErr, nil, true},
		"t14": {"mytask", fakeGetConfigOk, fakeGetClientsetOk, fakeGetOpenebsClusterErr, true},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := &kubeclient{
				getConfig:         mock.getConfig,
				getClientset:      mock.getClientset,
				getOpenebsCluster: mock.getOpenebsCluster,
			}
			_, err := k.Get(mock.name)
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}
