/*
Copyright 2018 The OpenEBS Authors

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

package v1alpha1

import (
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"testing"
)

func fakeGetClientsetOk(c *rest.Config) (*kubernetes.Clientset, error) {
	return &kubernetes.Clientset{}, nil
}

func fakeGetClientsetErr(c *rest.Config) (*kubernetes.Clientset, error) {
	return nil, errors.New("fake error")
}

func fakeInClusterConfigOk() (*rest.Config, error) {
	return &rest.Config{}, nil
}

func fakeInClusterConfigErr() (*rest.Config, error) {
	return nil, errors.New("fake error")
}

func fakeBuildConfigFromFlagsOk(kubemaster string, kubeconfig string) (*rest.Config, error) {
	return &rest.Config{}, nil
}

func fakeBuildConfigFromFlagsErr(kubemaster string, kubeconfig string) (*rest.Config, error) {
	return nil, errors.New("fake error")
}

func fakeGetKubeConfigPathOk(e env.ENVKey) string {
	return "fake"
}

func fakeGetKubeConfigPathNil(e env.ENVKey) string {
	return ""
}

func fakeGetKubeMasterIPOk(e env.ENVKey) string {
	return "fake"
}

func fakeGetKubeMasterIPNil(e env.ENVKey) string {
	return ""
}

func TestNewInCluster(t *testing.T) {
	c := New(InCluster())
	if !c.IsInCluster {
		t.Fatalf("test failed: expected IsInCluster as 'true' actual '%t'", c.IsInCluster)
	}
}

func TestConfig(t *testing.T) {
	tests := map[string]struct {
		isInCluster        bool
		getInClusterConfig getInClusterConfigFunc
		getKubeMasterIP    getKubeMasterIPFunc
		getKubeConfigPath  getKubeConfigPathFunc
		getConfigFromENV   buildConfigFromFlagsFunc
		isErr              bool
	}{
		"t1": {true, fakeInClusterConfigOk, nil, nil, nil, false},
		"t2": {true, fakeInClusterConfigErr, nil, nil, nil, true},
		"t3": {false, fakeInClusterConfigErr, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, true},
		"t4": {false, fakeInClusterConfigOk, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, false},
		"t5": {false, nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathNil, fakeBuildConfigFromFlagsOk, false},
		"t6": {false, nil, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
		"t7": {false, nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{
				IsInCluster:          mock.isInCluster,
				getInClusterConfig:   mock.getInClusterConfig,
				getKubeMasterIP:      mock.getKubeMasterIP,
				getKubeConfigPath:    mock.getKubeConfigPath,
				buildConfigFromFlags: mock.getConfigFromENV,
			}
			_, err := c.Config()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestGetConfigFromENV(t *testing.T) {
	tests := map[string]struct {
		getKubeMasterIP   getKubeMasterIPFunc
		getKubeConfigPath getKubeConfigPathFunc
		getConfigFromENV  buildConfigFromFlagsFunc
		isErr             bool
	}{
		"t1": {fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, true},
		"t2": {fakeGetKubeMasterIPNil, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
		"t3": {fakeGetKubeMasterIPOk, fakeGetKubeConfigPathNil, fakeBuildConfigFromFlagsOk, false},
		"t4": {fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
		"t5": {fakeGetKubeMasterIPNil, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, true},
		"t6": {fakeGetKubeMasterIPOk, fakeGetKubeConfigPathNil, fakeBuildConfigFromFlagsErr, true},
		"t7": {fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, true},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{
				getKubeMasterIP:      mock.getKubeMasterIP,
				getKubeConfigPath:    mock.getKubeConfigPath,
				buildConfigFromFlags: mock.getConfigFromENV,
			}
			_, err := c.getConfigFromENV()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestClientset(t *testing.T) {
	tests := map[string]struct {
		isInCluster            bool
		getInClusterConfig     getInClusterConfigFunc
		getKubeMasterIP        getKubeMasterIPFunc
		getKubeConfigPath      getKubeConfigPathFunc
		getConfigFromENV       buildConfigFromFlagsFunc
		getKubernetesClientset getKubernetesClientsetFunc
		isErr                  bool
	}{
		"t10": {true, fakeInClusterConfigOk, nil, nil, nil, fakeGetClientsetOk, false},
		"t11": {true, fakeInClusterConfigOk, nil, nil, nil, fakeGetClientsetErr, true},
		"t12": {true, fakeInClusterConfigErr, nil, nil, nil, fakeGetClientsetOk, true},

		"t21": {false, nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathNil, fakeBuildConfigFromFlagsOk, fakeGetClientsetOk, false},
		"t22": {false, nil, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, fakeGetClientsetOk, false},
		"t23": {false, nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, fakeGetClientsetOk, false},
		"t24": {false, nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, fakeGetClientsetOk, true},
		"t25": {false, nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, fakeGetClientsetErr, true},

		"t30": {false, fakeInClusterConfigOk, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, fakeGetClientsetOk, false},
		"t31": {false, fakeInClusterConfigOk, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, fakeGetClientsetErr, true},
		"t32": {false, fakeInClusterConfigErr, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, nil, true},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{
				IsInCluster:            mock.isInCluster,
				getInClusterConfig:     mock.getInClusterConfig,
				getKubeMasterIP:        mock.getKubeMasterIP,
				getKubeConfigPath:      mock.getKubeConfigPath,
				buildConfigFromFlags:   mock.getConfigFromENV,
				getKubernetesClientset: mock.getKubernetesClientset,
			}
			_, err := c.Clientset()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}
