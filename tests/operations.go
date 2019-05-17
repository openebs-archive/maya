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

package tests

import (
	"time"

	. "github.com/onsi/gomega"

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	snap "github.com/openebs/maya/pkg/kubernetes/snapshot/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	templatefuncs "github.com/openebs/maya/pkg/templatefuncs/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	maxRetry = 30
)

// Operations provides clients amd methods to perform operations
type Operations struct {
	PodClient      *pod.KubeClient
	ScClient       *sc.Kubeclient
	PvcClient      *pvc.Kubeclient
	NsClient       *ns.Kubeclient
	SnapClient     *snap.Kubeclient
	kubeConfigPath string
}

// OperationsOptions abstracts creating an
// instance of operations
type OperationsOptions func(*Operations)

// WithKubeConfigPath sets the kubeConfig path
// against operations instance
func WithKubeConfigPath(path string) OperationsOptions {
	return func(ops *Operations) {
		ops.kubeConfigPath = path
	}
}

// NewOperations returns a new instance of kubeclient meant for
// cstor volume replica operations
func NewOperations(opts ...OperationsOptions) *Operations {
	ops := &Operations{}
	for _, o := range opts {
		o(ops)
	}
	ops.newPodClient()
	ops.newNsClient()
	ops.newSCClient()
	ops.newPVCClient()
	ops.newSnapClient()
	return ops
}

func (ops *Operations) newPodClient() *Operations {
	newPodClient := pod.NewKubeClient(pod.WithKubeConfigPath(ops.kubeConfigPath))
	ops.PodClient = newPodClient
	return ops
}

func (ops *Operations) newNsClient() *Operations {
	newNsClient := ns.NewKubeClient(ns.WithKubeConfigPath(ops.kubeConfigPath))
	ops.NsClient = newNsClient
	return ops
}

func (ops *Operations) newSCClient() *Operations {
	newSCClient := sc.NewKubeClient(sc.WithKubeConfigPath(ops.kubeConfigPath))
	ops.ScClient = newSCClient
	return ops
}

func (ops *Operations) newPVCClient() *Operations {
	newPVCClient := pvc.NewKubeClient(pvc.WithKubeConfigPath(ops.kubeConfigPath))
	ops.PvcClient = newPVCClient
	return ops
}

func (ops *Operations) newSnapClient() *Operations {
	newSnapClient := snap.NewKubeClient(snap.WithKubeConfigPath(ops.kubeConfigPath))
	ops.SnapClient = newSnapClient
	return ops
}

// GetPodRunningCountEventually gives the number of pods running eventually
func (ops *Operations) GetPodRunningCountEventually(namespace, lselector string, expectedPodCount int) int {
	var podCount int
	for i := 0; i < maxRetry; i++ {
		podCount = ops.GetPodRunningCount(namespace, lselector)
		if podCount == expectedPodCount {
			return podCount
		}
		time.Sleep(5 * time.Second)
	}
	return podCount
}

// GetPodRunningCount gives number of pods running currently
func (ops *Operations) GetPodRunningCount(namespace, lselector string) int {
	pods, err := ops.PodClient.
		WithNamespace(namespace).
		List(metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return pod.
		ListBuilderForAPIList(pods).
		WithFilter(pod.IsRunning()).
		List().
		Len()
}

// IsPVCBound checks if the pvc is bound or not
func (ops *Operations) IsPVCBound(pvcName string) bool {
	volume, err := ops.PvcClient.
		Get(pvcName, metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	return pvc.NewForAPIObject(volume).IsBound()
}

// GetSnapshotTypeEventually returns type of snapshot eventually
func (ops *Operations) GetSnapshotTypeEventually(snapName string) string {
	var snaptype string
	for i := 0; i < maxRetry; i++ {
		snaptype = ops.GetSnapshotType(snapName)
		if snaptype == "Ready" {
			return snaptype
		}
		time.Sleep(5 * time.Second)
	}
	return snaptype
}

// GetSnapshotType returns type of snapshot currently
func (ops *Operations) GetSnapshotType(snapName string) string {
	snap, err := ops.SnapClient.
		Get(snapName, metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	if len(snap.Status.Conditions) > 0 {
		return string(snap.Status.Conditions[0].Type)
	}
	return "NotReady"
}

// IsSnapshotDeleted checks if the snapshot is deleted or not
func (ops *Operations) IsSnapshotDeleted(snapName string) bool {
	for i := 0; i < maxRetry; i++ {
		_, err := ops.SnapClient.
			Get(snapName, metav1.GetOptions{})
		if err != nil {
			return true
		}
		time.Sleep(5 * time.Second)
	}
	return false
}

// IsPVCDeleted tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func (ops *Operations) IsPVCDeleted(pvcName string) bool {
	_, err := ops.PvcClient.
		Get(pvcName, metav1.GetOptions{})
	if isNotFound(err) {
		return true
	}
	return false
}

// isNotFound returns true if the original
// cause of error was due to castemplate's
// not found error or kubernetes not found
// error
func isNotFound(err error) bool {
	switch err := errors.Cause(err).(type) {
	case *templatefuncs.NotFoundError:
		return true
	default:
		return k8serrors.IsNotFound(err)
	}
}
