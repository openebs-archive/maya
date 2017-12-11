/*
Copyright 2017 The OpenEBS Authors

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

package k8s

import (
	"testing"

	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetConfigMap(t *testing.T) {
	tests := map[string]struct {
		name  string
		opts  mach_apis_meta_v1.GetOptions
		cm    *api_core_v1.ConfigMap
		isErr bool
	}{
		"valid configmap": {"", mach_apis_meta_v1.GetOptions{}, &api_core_v1.ConfigMap{}, false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			kc := &K8sClient{
				ConfigMap: test.cm,
			}

			_, err := kc.GetConfigMap(test.name, test.opts)

			if !test.isErr && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}
		})
	}
}

func TestGetPVC(t *testing.T) {
	tests := map[string]struct {
		name  string
		opts  mach_apis_meta_v1.GetOptions
		pvc   *api_core_v1.PersistentVolumeClaim
		isErr bool
	}{
		"valid volume claim": {"", mach_apis_meta_v1.GetOptions{}, &api_core_v1.PersistentVolumeClaim{}, false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			kc := &K8sClient{
				PVC: test.pvc,
			}

			_, err := kc.GetPVC(test.name, test.opts)

			if !test.isErr && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}
		})
	}
}

func TestGetService(t *testing.T) {
	tests := map[string]struct {
		name    string
		opts    mach_apis_meta_v1.GetOptions
		service *api_core_v1.Service
		isErr   bool
	}{
		"valid service": {"", mach_apis_meta_v1.GetOptions{}, &api_core_v1.Service{}, false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			kc := &K8sClient{
				Service: test.service,
			}

			_, err := kc.GetService(test.name, test.opts)

			if !test.isErr && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}
		})
	}
}

func TestGetPod(t *testing.T) {
	tests := map[string]struct {
		name  string
		opts  mach_apis_meta_v1.GetOptions
		pod   *api_core_v1.Pod
		isErr bool
	}{
		"valid pod": {"", mach_apis_meta_v1.GetOptions{}, &api_core_v1.Pod{}, false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			kc := &K8sClient{
				Pod: test.pod,
			}

			_, err := kc.GetPod(test.name, test.opts)

			if !test.isErr && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}
		})
	}
}

func TestGetDeployment(t *testing.T) {
	tests := map[string]struct {
		name   string
		opts   mach_apis_meta_v1.GetOptions
		deploy *api_extn_v1beta1.Deployment
		isErr  bool
	}{
		"valid deployment": {"", mach_apis_meta_v1.GetOptions{}, &api_extn_v1beta1.Deployment{}, false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			kc := &K8sClient{
				Deployment: test.deploy,
			}

			_, err := kc.GetDeployment(test.name, test.opts)

			if !test.isErr && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}
		})
	}
}
