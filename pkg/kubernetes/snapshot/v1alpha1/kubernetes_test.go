// Copyright © 2018-2019 The OpenEBS Authors
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

package v1alpha1

import (
	"errors"
	"testing"

	snapshot "github.com/openebs/maya/pkg/apis/openebs.io/snapshot/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/snapshot/v1alpha1/clientset/internalclientset/typed/snapshot/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakeGetClientSetOk() (*clientset.OpenebsV1alpha1Client, error) {
	return &clientset.OpenebsV1alpha1Client{}, nil
}

func fakeGetClientSetErr() (*clientset.OpenebsV1alpha1Client, error) {
	return nil, errors.New("Some error")
}

func fakeListFnOk(cli *clientset.OpenebsV1alpha1Client, opts metav1.ListOptions) (*snapshot.VolumeSnapshotList, error) {
	return &snapshot.VolumeSnapshotList{}, nil
}

func fakeListFnErr(cli *clientset.OpenebsV1alpha1Client, opts metav1.ListOptions) (*snapshot.VolumeSnapshotList, error) {
	return nil, errors.New("some error")
}

func fakeGetClientSetForPathOk(fakeConfigPath string) (*clientset.OpenebsV1alpha1Client, error) {
	return &clientset.OpenebsV1alpha1Client{}, nil
}

func fakeGetClientSetForPathErr(fakeConfigPath string) (*clientset.OpenebsV1alpha1Client, error) {
	return nil, errors.New("fake error")
}

func TestSnapshotList(t *testing.T) {
	tests := map[string]struct {
		getClientset        getClientsetFn
		getClientSetForPath getClientsetForPathFn
		kubeConfigPath      string
		list                listFn
		expectErr           bool
	}{
		// Positive tests
		"Positive 1": {nil, fakeGetClientSetForPathOk, "fake-path", fakeListFnOk, false},
		"Positive 2": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeListFnOk, false},
		"Positive 3": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "fake-path", fakeListFnOk, false},
		"Positive 4": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "", fakeListFnOk, false},

		// Negative tests
		"Negative 1": {fakeGetClientSetErr, fakeGetClientSetForPathOk, "", fakeListFnOk, true},
		"Negative 2": {fakeGetClientSetOk, fakeGetClientSetForPathErr, "fake-path", fakeListFnOk, true},
		"Negative 3": {fakeGetClientSetErr, fakeGetClientSetForPathErr, "fake-path", fakeListFnOk, true},
		"Negative 4": {fakeGetClientSetOk, fakeGetClientSetForPathOk, "", fakeListFnErr, true},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			fc := &Kubeclient{
				getClientset:        mock.getClientset,
				getClientsetForPath: mock.getClientSetForPath,
				kubeConfigPath:      mock.kubeConfigPath,
				list:                mock.list,
			}
			_, err := fc.List(metav1.ListOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
