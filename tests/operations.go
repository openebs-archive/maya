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
	"bytes"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstorpool/v1alpha3"
	cv "github.com/openebs/maya/pkg/cstorvolume/v1alpha1"
	cvr "github.com/openebs/maya/pkg/cstorvolumereplica/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	kubeclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	snap "github.com/openebs/maya/pkg/kubernetes/snapshot/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	templatefuncs "github.com/openebs/maya/pkg/templatefuncs/v1alpha1"
	unstruct "github.com/openebs/maya/pkg/unstruct/v1alpha2"
	result "github.com/openebs/maya/pkg/upgrade/result/v1alpha1"
	"github.com/openebs/maya/tests/artifacts"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	maxRetry = 30
)

// Options holds the args used for exec'ing into the pod
type Options struct {
	podName   string
	container string
	namespace string
	cmd       []string
}

// Operations provides clients amd methods to perform operations
type Operations struct {
	KubeClient     *kubeclient.Client
	PodClient      *pod.KubeClient
	SCClient       *sc.Kubeclient
	PVCClient      *pvc.Kubeclient
	NSClient       *ns.Kubeclient
	SnapClient     *snap.Kubeclient
	CSPClient      *csp.Kubeclient
	SPCClient      *spc.Kubeclient
	SVCClient      *svc.Kubeclient
	CVClient       *cv.Kubeclient
	CVRClient      *cvr.Kubeclient
	URClient       *result.Kubeclient
	UnstructClient *unstruct.Kubeclient
	DeployClient   *deploy.Kubeclient
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
	ops.withDefaults()
	return ops
}

// NewOptions returns the new instance of Options
func NewOptions() *Options {
	return new(Options)
}

// WithPodName fills the podName field in Options struct
func (o *Options) WithPodName(name string) *Options {
	o.podName = name
	return o
}

// WithNamespace fills the namespace field in Options struct
func (o *Options) WithNamespace(ns string) *Options {
	o.namespace = ns
	return o
}

// WithContainer fills the container field in Options struct
func (o *Options) WithContainer(container string) *Options {
	o.container = container
	return o
}

// WithCommand fills the cmd field in Options struct
func (o *Options) WithCommand(cmd ...string) *Options {
	o.cmd = cmd
	return o
}

// withDefaults sets the default options
// of operations instance
func (ops *Operations) withDefaults() {
	var err error
	if ops.KubeClient == nil {
		ops.KubeClient = kubeclient.New(kubeclient.WithKubeConfigPath(ops.kubeConfigPath))
	}
	if ops.NSClient == nil {
		ops.NSClient = ns.NewKubeClient(ns.WithKubeConfigPath(ops.kubeConfigPath))
	}
	if ops.SCClient == nil {
		ops.SCClient = sc.NewKubeClient(sc.WithKubeConfigPath(ops.kubeConfigPath))
	}
	if ops.PodClient == nil {
		ops.PodClient = pod.NewKubeClient(pod.WithKubeConfigPath(ops.kubeConfigPath))
	}
	if ops.PVCClient == nil {
		ops.PVCClient = pvc.NewKubeClient(pvc.WithKubeConfigPath(ops.kubeConfigPath))
	}
	if ops.SnapClient == nil {
		ops.SnapClient = snap.NewKubeClient(snap.WithKubeConfigPath(ops.kubeConfigPath))
	}
	if ops.SPCClient == nil {
		ops.SPCClient = spc.NewKubeClient(spc.WithKubeConfigPath(ops.kubeConfigPath))
	}
	if ops.CSPClient == nil {
		ops.CSPClient, err = csp.KubeClient().WithKubeConfigPath(ops.kubeConfigPath)
		Expect(err).To(BeNil(), "while initilizing csp client")
	}
	if ops.CVClient == nil {
		ops.CVClient = cv.NewKubeclient(cv.WithKubeConfigPath(ops.kubeConfigPath))
	}
	if ops.CVRClient == nil {
		ops.CVRClient = cvr.NewKubeclient(cvr.WithKubeConfigPath(ops.kubeConfigPath))
	}
	if ops.URClient == nil {
		ops.URClient = result.NewKubeClient(result.WithKubeConfigPath(ops.kubeConfigPath))
	}
	if ops.UnstructClient == nil {
		ops.UnstructClient = unstruct.NewKubeClient(unstruct.WithKubeConfigPath(ops.kubeConfigPath))
	}
	if ops.DeployClient == nil {
		ops.DeployClient = deploy.NewKubeClient(deploy.WithKubeConfigPath(ops.kubeConfigPath))
	}
}

// VerifyOpenebs verify running state of required openebs control plane components
func (ops *Operations) VerifyOpenebs(expectedPodCount int) *Operations {
	By("waiting for maya-apiserver pod to come into running state")
	podCount := ops.GetPodRunningCountEventually(
		string(artifacts.OpenebsNamespace),
		string(artifacts.MayaAPIServerLabelSelector),
		expectedPodCount,
	)
	Expect(podCount).To(Equal(expectedPodCount))

	By("waiting for openebs-provisioner pod to come into running state")
	podCount = ops.GetPodRunningCountEventually(
		string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSProvisionerLabelSelector),
		expectedPodCount,
	)
	Expect(podCount).To(Equal(expectedPodCount))

	By("Verifying 'admission-server' pod status as running")
	_ = ops.GetPodRunningCountEventually(string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSAdmissionServerLabelSelector),
		expectedPodCount,
	)

	Expect(podCount).To(Equal(expectedPodCount))

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

// GetCstorVolumeCount gives the count of cstorvolume based on
// selecter
func (ops *Operations) GetCstorVolumeCount(namespace, lselector string, expectedCVCount int) int {
	var cvCount int
	for i := 0; i < maxRetry; i++ {
		cvCount = ops.GetCVCount(namespace, lselector)
		if cvCount == expectedCVCount {
			return cvCount
		}
		time.Sleep(5 * time.Second)
	}
	return cvCount
}

// GetCstorVolumeCountEventually gives the count of cstorvolume based on
// selecter eventually
func (ops *Operations) GetCstorVolumeCountEventually(namespace, lselector string, expectedCVCount int) bool {
	return Eventually(func() int {
		cvCount := ops.GetCVCount(namespace, lselector)
		return cvCount
	},
		60, 10).Should(Equal(expectedCVCount))
}

// GetCstorVolumeReplicaCountEventually gives the count of cstorvolume based on
// selecter eventually
func (ops *Operations) GetCstorVolumeReplicaCountEventually(namespace, lselector string, expectedCVRCount int) bool {
	return Eventually(func() int {
		cvCount := ops.GetCstorVolumeReplicaCount(namespace, lselector)
		return cvCount
	},
		60, 10).Should(Equal(expectedCVRCount))
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

// GetCVCount gives cstorvolume healthy count currently based on selecter
func (ops *Operations) GetCVCount(namespace, lselector string) int {
	cvs, err := ops.CVClient.
		List(metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return cv.
		NewListBuilder().
		WithAPIList(cvs).
		WithFilter(cv.IsHealthy()).
		List().
		Len()
}

// GetCstorVolumeReplicaCount gives cstorvolumereplica healthy count currently based on selecter
func (ops *Operations) GetCstorVolumeReplicaCount(namespace, lselector string) int {
	cvrs, err := ops.CVRClient.
		List(metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return cvr.
		ListBuilder().
		WithAPIList(cvrs).
		WithFilter(cvr.IsHealthy()).
		List().
		Len()
}

// IsPVCBound checks if the pvc is bound or not
func (ops *Operations) IsPVCBound(pvcName string) bool {
	volume, err := ops.PVCClient.
		Get(pvcName, metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	return pvc.NewForAPIObject(volume).IsBound()
}

// IsPVCBoundEventually checks if the pvc is bound or not eventually
func (ops *Operations) IsPVCBoundEventually(pvcName string) bool {
	return Eventually(func() bool {
		volume, err := ops.PVCClient.
			Get(pvcName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		return pvc.NewForAPIObject(volume).IsBound()
	},
		60, 10).
		Should(BeTrue())
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
			return isNotFound(err)
		}
		time.Sleep(5 * time.Second)
	}
	return false
}

// IsPVCDeleted tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func (ops *Operations) IsPVCDeleted(pvcName string) bool {
	_, err := ops.PVCClient.
		Get(pvcName, metav1.GetOptions{})
	if isNotFound(err) {
		return true
	}
	return false
}

// IsPodDeletedEventually checks if the pod is deleted or not eventually
func (ops *Operations) IsPodDeletedEventually(namespace, podName string) bool {
	return Eventually(func() bool {
		_, err := ops.PodClient.
			WithNamespace(namespace).
			Get(podName, metav1.GetOptions{})
		return isNotFound(err)
	},
		60, 10).
		Should(BeTrue())
}

// GetPVNameFromPVCName gives the pv name for the given pvc
func (ops *Operations) GetPVNameFromPVCName(pvcName string) string {
	p, err := ops.PVCClient.
		Get(pvcName, metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	return p.Spec.VolumeName
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

// DeleteCSP ...
func (ops *Operations) DeleteCSP(spcName string, deleteCount int) {
	cspAPIList, err := ops.CSPClient.List(metav1.ListOptions{})
	Expect(err).To(BeNil())
	cspList := csp.
		ListBuilderForAPIObject(cspAPIList).
		List().
		Filter(csp.HasLabel(string(apis.StoragePoolClaimCPK), spcName), csp.IsStatus("Healthy"))
	cspCount := cspList.Len()
	Expect(deleteCount).Should(BeNumerically("<=", cspCount))

	for i := 0; i < deleteCount; i++ {
		_, err := ops.CSPClient.Delete(cspList.ObjectList.Items[i].Name, &metav1.DeleteOptions{})
		Expect(err).To(BeNil())

	}
}

// GetCSPCount gets csp count based on spcName
func (ops *Operations) GetCSPCount(spcName string, expectedCSPCount int) int {
	var cspCount int
	for i := 0; i < maxRetry; i++ {
		cspAPIList, err := ops.CSPClient.List(metav1.ListOptions{})
		Expect(err).To(BeNil())
		cspCount = csp.
			ListBuilderForAPIObject(cspAPIList).
			List().
			Len()
		if cspCount == expectedCSPCount {
			return cspCount
		}
		time.Sleep(5 * time.Second)
	}
	return cspCount
}

// GetHealthyCSPCount gets healthy csp based on spcName
func (ops *Operations) GetHealthyCSPCount(spcName string, expectedCSPCount int) int {
	var cspCount int
	for i := 0; i < maxRetry; i++ {
		cspAPIList, err := ops.CSPClient.List(metav1.ListOptions{})
		Expect(err).To(BeNil())
		cspCount = csp.
			ListBuilderForAPIObject(cspAPIList).
			List().
			Filter(csp.HasLabel(string(apis.StoragePoolClaimCPK), spcName), csp.IsStatus("Healthy")).
			Len()
		if cspCount == expectedCSPCount {
			return cspCount
		}
		time.Sleep(5 * time.Second)
	}
	return cspCount
}

// GetHealthyCSPCountEventually gets healthy csp based on spcName
func (ops *Operations) GetHealthyCSPCountEventually(spcName string, expectedCSPCount int) bool {
	return Eventually(func() int {
		cspAPIList, err := ops.CSPClient.List(metav1.ListOptions{})
		Expect(err).To(BeNil())
		count := csp.
			ListBuilderForAPIObject(cspAPIList).
			List().
			Filter(csp.HasLabel(string(apis.StoragePoolClaimCPK), spcName), csp.IsStatus("Healthy")).
			Len()
		return count
	},
		60, 10).
		Should(Equal(expectedCSPCount))
}

// ExecPod executes arbitrary command inside the pod
func (ops *Operations) ExecPod(opts *Options) ([]byte, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
		err     error
	)
	By("getting rest config")
	config, err := ops.KubeClient.GetConfigForPathOrDirect()
	Expect(err).To(BeNil(), "while getting config for exec'ing into pod")
	By("getting clientset")
	cset, err := ops.KubeClient.Clientset()
	Expect(err).To(BeNil(), "while getting clientset for exec'ing into pod")
	req := cset.
		CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Name(opts.podName).
		Namespace(opts.namespace).
		SubResource("exec").
		Param("container", opts.container).
		VersionedParams(&corev1.PodExecOptions{
			Container: opts.container,
			Command:   opts.cmd,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	By("creating a POST request for executing command")
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	Expect(err).To(BeNil(), "while exec'ing command in pod ", opts.podName)

	By("processing request")
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &execOut,
		Stderr: &execErr,
		Tty:    false,
	})
	Expect(err).To(BeNil(), "while streaming the command in pod ", opts.podName, execOut.String(), execErr.String())
	Expect(execOut.Len()).Should(BeNumerically(">", 0), "while streaming the command in pod ", opts.podName, execErr.String(), execOut.String())
	return execOut.Bytes(), nil
}

// GetPodCompletedCountEventually gives the number of pods running eventually
func (ops *Operations) GetPodCompletedCountEventually(namespace, lselector string, expectedPodCount int) int {
	var podCount int
	for i := 0; i < maxRetry; i++ {
		podCount = ops.GetPodCompletedCount(namespace, lselector)
		if podCount == expectedPodCount {
			return podCount
		}
		time.Sleep(5 * time.Second)
	}
	return podCount
}

// GetPodCompletedCount gives number of pods running currently
func (ops *Operations) GetPodCompletedCount(namespace, lselector string) int {
	pods, err := ops.PodClient.
		WithNamespace(namespace).
		List(metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return pod.
		ListBuilderForAPIList(pods).
		WithFilter(pod.IsCompleted()).
		List().
		Len()
}

// VerifyUpgradeResultTasksIsNotFail checks whether all the tasks in upgraderesult
// have success
func (ops *Operations) VerifyUpgradeResultTasksIsNotFail(namespace, lselector string) bool {
	urList, err := ops.URClient.
		WithNamespace(namespace).
		List(metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	for _, task := range urList.Items[0].Tasks {
		if task.Status == "Fail" {
			fmt.Printf("task : %v\n", task)
			return false
		}
	}
	return true
}
