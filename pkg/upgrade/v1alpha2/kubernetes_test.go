// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha2

import (
	"encoding/json"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/upgrade/v1alpha1/clientset/internalclientset"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func fakeGetClientsetOk() (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeGetClientsetForPathOk(
	fakeConfigPath string,
) (cli *clientset.Clientset, err error) {
	return &clientset.Clientset{}, nil
}

func fakeGetClientsetForPathErr(
	fakeConfigPath string,
) (cli *clientset.Clientset, err error) {
	return nil, errors.New("fake error")
}

func fakeGetFnOk(
	cli *clientset.Clientset,
	name, namespace string,
	opts metav1.GetOptions,
) (*apis.UpgradeTask, error) {
	return &apis.UpgradeTask{}, nil
}

func fakeListFnOk(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apis.UpgradeTaskList, error) {
	return &apis.UpgradeTaskList{}, nil
}

func fakeDeleteFnOk(
	cli *clientset.Clientset,
	name, namespace string,
	opts *metav1.DeleteOptions,
) error {
	return nil
}

func fakeListFnErr(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apis.UpgradeTaskList, error) {
	return &apis.UpgradeTaskList{}, errors.New("some error")
}

func fakeGetFnErr(
	cli *clientset.Clientset,
	name, namespace string,
	opts metav1.GetOptions,
) (*apis.UpgradeTask, error) {
	return &apis.UpgradeTask{}, errors.New("some error")
}

func fakeDeleteFnErr(
	cli *clientset.Clientset,
	name, namespace string,
	opts *metav1.DeleteOptions,
) error {
	return errors.New("some error")
}

func fakeGetClientsetErr() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeCreateFnOk(
	cli *clientset.Clientset,
	namespace string,
	upgradeTask *apis.UpgradeTask) (*apis.UpgradeTask, error) {
	return &apis.UpgradeTask{}, nil
}

func fakeCreateErr(
	cli *clientset.Clientset,
	namespace string,
	upgradeTask *apis.UpgradeTask,
) (*apis.UpgradeTask, error) {
	return nil, errors.New("failed to create UpgradeTask")
}

func fakePatchFnOk(
	cli *clientset.Clientset,
	namespace, name string,
	pt types.PatchType,
	data []byte,
	subresources ...string,
) (*apis.UpgradeTask, error) {
	return &apis.UpgradeTask{}, nil
}

func fakePatchFnErr(
	cli *clientset.Clientset,
	namespace, name string,
	pt types.PatchType,
	data []byte,
	subresources ...string,
) (*apis.UpgradeTask, error) {
	return nil, errors.New("fake error")
}

func TestWithDefaultOptionsForClientset(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
	}{
		"TestCase1":               {nil, nil},
		"When clientset is error": {fakeGetClientsetErr, fakeGetClientsetForPathErr},
		"When clientset is ok":    {fakeGetClientsetOk, fakeGetClientsetForPathOk},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
			}
			fc.WithDefaults()
			if fc.getClientset == nil {
				t.Fatalf(
					"test %q failed: expected getClientset not to be empty",
					name,
				)
			}
			if fc.getClientsetForPath == nil {
				t.Fatalf(
					"test %q failed: expected getClientset not to be nil",
					name,
				)
			}
		})
	}
}

func TestGetClientsetForPathOrDirect(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		isErr               bool
	}{
		// Positive tests
		"Positive 1": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "", false},
		"Positive 2": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "fake-path", false},
		"Positive 3": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "", false},

		// Negative tests
		"Negative 1": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "", true},
		"Negative 2": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "fake-path", true},
		"Negative 3": {fakeGetClientsetErr, fakeGetClientsetForPathErr, "fake-path", true},
		"Negative 4": {fakeGetClientsetErr, fakeGetClientsetForPathErr, "", true},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
			}
			_, err := fc.getClientsetForPathOrDirect()
			if mock.isErr && err == nil {
				t.Fatalf(
					"test %q failed : expected error not to be nil but got %v",
					name,
					err,
				)
			}
			if !mock.isErr && err != nil {
				t.Fatalf(
					"test %q failed : expected error be nil but got %v",
					name,
					err,
				)
			}
		})
	}
}

func TestWithClientsetBuildOption(t *testing.T) {
	tests := map[string]struct {
		Clientset             *clientset.Clientset
		expectKubeClientEmpty bool
	}{
		"Clientset is empty":     {nil, true},
		"Clientset is not empty": {&clientset.Clientset{}, false},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			h := WithKubeClient(mock.Clientset)
			fake := &Kubeclient{}
			h(fake)
			if mock.expectKubeClientEmpty && fake.clientset != nil {
				t.Fatalf(
					"test %q failed expected fake.clientset to be empty",
					name,
				)
			}
			if !mock.expectKubeClientEmpty && fake.clientset == nil {
				t.Fatalf(
					"test %q failed expected fake.clientset not to be empty",
					name,
				)
			}
		})
	}
}

func TestGetClientOrCached(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		expectErr           bool
	}{
		// Positive tests
		"Positive 1": {fakeGetClientsetOk, fakeGetClientsetForPathOk, "", false},
		"Positive 2": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "fake-path", false},
		"Positive 3": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "", false},

		// Negative tests
		"Negative 1": {fakeGetClientsetErr, fakeGetClientsetForPathOk, "", true},
		"Negative 2": {fakeGetClientsetOk, fakeGetClientsetForPathErr, "fake-path", true},
		"Negative 3": {fakeGetClientsetErr, fakeGetClientsetForPathErr, "fake-path", true},
		"Negative 4": {fakeGetClientsetErr, fakeGetClientsetForPathErr, "", true},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
			}
			_, err := fc.getClientsetOrCached()
			if mock.expectErr && err == nil {
				t.Fatalf(
					"test %q failed : expected error not to be nil but got %v",
					name,
					err,
				)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf(
					"test %q failed : expected error be nil but got %v",
					name,
					err,
				)
			}
		})
	}
}

func TestUpgradeTaskList(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		list                listFn
		expectedErr         bool
	}{
		// Positive tests
		"Positive 1": {
			nil,
			fakeGetClientsetForPathOk,
			"fake-path",
			fakeListFnOk,
			false,
		},
		"Positive 2": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathOk,
			"",
			fakeListFnOk,
			false,
		},
		"Positive 3": {
			fakeGetClientsetErr,
			fakeGetClientsetForPathOk,
			"fake-path",
			fakeListFnOk,
			false,
		},
		"Positive 4": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathErr,
			"",
			fakeListFnOk,
			false,
		},

		// Negative tests
		"Negative 1": {
			fakeGetClientsetErr,
			fakeGetClientsetForPathOk,
			"",
			fakeListFnOk,
			true,
		},
		"Negative 2": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathErr,
			"fake-path",
			fakeListFnOk,
			true,
		},
		"Negative 3": {
			fakeGetClientsetErr,
			fakeGetClientsetForPathErr,
			"fake-path",
			fakeListFnOk,
			true,
		},
		"Negative 4": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathOk,
			"",
			fakeListFnErr,
			true,
		},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				list:                mock.list,
			}
			_, err := fc.List(metav1.ListOptions{})
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestUpgradeTaskGet(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		get                 getFn
		UpgradeTaskName     string
		expectErr           bool
	}{
		"Test 1": {
			fakeGetClientsetErr,
			fakeGetClientsetForPathOk,
			"",
			fakeGetFnOk,
			"UpgradeTask-1",
			true,
		},
		"Test 2": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathErr,
			"fake-path",
			fakeGetFnOk,
			"UpgradeTask-1",
			true,
		},
		"Test 3": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathOk,
			"",
			fakeGetFnOk,
			"UpgradeTask-2",
			false,
		},
		"Test 4": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathOk,
			"fp",
			fakeGetFnErr,
			"UpgradeTask-3",
			true,
		},
		"Test 5": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathOk,
			"fakepath",
			fakeGetFnOk,
			"",
			true,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				namespace:           "default",
				get:                 mock.get,
			}
			_, err := k.Get(mock.UpgradeTaskName, metav1.GetOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestUpgradeTaskDelete(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		UpgradeTaskName     string
		delete              deleteFn
		expectErr           bool
	}{
		"Test 1": {
			fakeGetClientsetErr,
			fakeGetClientsetForPathOk,
			"",
			"UpgradeTask-1",
			fakeDeleteFnOk,
			true,
		},
		"Test 2": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathOk,
			"fake-path2",
			"UpgradeTask-2",
			fakeDeleteFnOk,
			false,
		},
		"Test 3": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathOk,
			"",
			"UpgradeTask-3",
			fakeDeleteFnErr,
			true,
		},
		"Test 4": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathOk,
			"fakepath",
			"",
			fakeDeleteFnOk,
			true,
		},
		"Test 5": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathErr,
			"fake-path2",
			"UpgradeTask1",
			fakeDeleteFnOk,
			true,
		},
		"Test 6": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathErr,
			"fake-path2",
			"UpgradeTask1",
			fakeDeleteFnErr,
			true,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				namespace:           "",
				del:                 mock.delete,
			}
			err := k.Delete(mock.UpgradeTaskName, &metav1.DeleteOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestUpgradeTaskCreate(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		create              createFn
		upgradeTask         *apis.UpgradeTask
		expectErr           bool
	}{
		"Test 1": {
			getClientset:        fakeGetClientsetErr,
			getClientsetForPath: fakeGetClientsetForPathErr,
			kubeConfigPath:      "",
			create:              fakeCreateFnOk,
			upgradeTask: &apis.UpgradeTask{
				ObjectMeta: metav1.ObjectMeta{Name: "UpgradeTask-1"},
			},
			expectErr: true,
		},
		"Test 2": {
			getClientset:        fakeGetClientsetOk,
			getClientsetForPath: fakeGetClientsetForPathOk,
			kubeConfigPath:      "",
			create:              fakeCreateErr,
			upgradeTask: &apis.UpgradeTask{
				ObjectMeta: metav1.ObjectMeta{Name: "UpgradeTask-2"},
			},
			expectErr: true,
		},
		"Test 3": {
			getClientset:        fakeGetClientsetOk,
			getClientsetForPath: fakeGetClientsetForPathOk,
			kubeConfigPath:      "fake-path",
			create:              fakeCreateErr,
			upgradeTask:         nil,
			expectErr:           true,
		},
		"Test 4": {
			getClientset:        fakeGetClientsetErr,
			getClientsetForPath: fakeGetClientsetForPathOk,
			kubeConfigPath:      "fake-path",
			create:              fakeCreateFnOk,
			upgradeTask:         nil,
			expectErr:           true,
		},
		"Test 5": {
			getClientset:        fakeGetClientsetOk,
			getClientsetForPath: fakeGetClientsetForPathErr,
			kubeConfigPath:      "fake-path",
			create:              fakeCreateFnOk,
			upgradeTask:         nil,
			expectErr:           true,
		},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				create:              mock.create,
			}
			_, err := fc.Create(mock.upgradeTask)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestUpgradeTaskPatch(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientsetForPath getClientsetForPathFn
		kubeConfigPath      string
		patch               patchFn
		UpgradeTaskName     string
		expectErr           bool
	}{
		"Test 1": {
			fakeGetClientsetErr,
			fakeGetClientsetForPathOk,
			"",
			fakePatchFnOk,
			"upgradeTask-1", true},
		"Test 2": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathErr,
			"fake-path",
			fakePatchFnOk,
			"upgradeTask-1",
			true,
		},
		"Test 3": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathOk,
			"",
			fakePatchFnOk,
			"upgradeTask-2",
			false,
		},
		"Test 4": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathOk,
			"fp",
			fakePatchFnErr,
			"upgradeTask-3",
			true,
		},
		"Test 5": {
			fakeGetClientsetOk,
			fakeGetClientsetForPathOk,
			"fakepath",
			fakePatchFnOk,
			"",
			true,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			k := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientsetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				patch:               mock.patch,
			}
			//fake data
			data, _ := json.Marshal(mock)
			_, err := k.Patch(mock.UpgradeTaskName, types.MergePatchType, data)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestGetUpgradeDetailedStatuses(t *testing.T) {
	tests := map[string]struct {
		status       apis.UpgradeDetailedStatuses
		expectOutput bool
	}{
		"Test 1": {
			apis.UpgradeDetailedStatuses{
				Step: apis.PreUpgrade,
				Status: apis.Status{
					Phase: apis.StepWaiting,
				},
			},
			true,
		},
		"Test 2": {
			apis.UpgradeDetailedStatuses{
				Step: apis.PreUpgrade,
				Status: apis.Status{
					Phase:   apis.StepCompleted,
					Message: "fake-message",
				},
			},
			true,
		},
		"Test 3": {
			apis.UpgradeDetailedStatuses{
				Step: apis.PreUpgrade,
				Status: apis.Status{
					Phase:   apis.StepErrored,
					Message: "fake-message",
					Reason:  "fake-reason",
				},
			},
			true,
		},
		// negative test
		"Test 4": {
			apis.UpgradeDetailedStatuses{},
			false,
		},
		"Test 5": {
			apis.UpgradeDetailedStatuses{
				Step: apis.PreUpgrade,
			},
			false,
		},
		"Test 6": {
			apis.UpgradeDetailedStatuses{
				Step: apis.PreUpgrade,
				Status: apis.Status{
					Phase: apis.StepCompleted,
				},
			},
			false,
		},
		"Test 7": {
			apis.UpgradeDetailedStatuses{
				Step: apis.PreUpgrade,
				Status: apis.Status{
					Phase:   apis.StepErrored,
					Message: "fake-message",
				},
			},
			false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {

			output := IsValidStatus(mock.status)
			if mock.expectOutput != output {
				t.Fatalf(
					"Test %q failed: expected %v not to be %v",
					name,
					output,
					mock.expectOutput,
				)
			}
		})
	}
}
