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
	bd "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstor/pool/v1alpha3"
	cspc "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	cspi "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	cv "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	cvc "github.com/openebs/maya/pkg/cstorvolumeclaim/v1alpha1"
	kubeclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	node "github.com/openebs/maya/pkg/kubernetes/node/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	snap "github.com/openebs/maya/pkg/kubernetes/snapshot/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	templatefuncs "github.com/openebs/maya/pkg/templatefuncs/v1alpha1"
	unstruct "github.com/openebs/maya/pkg/unstruct/v1alpha2"
	result "github.com/openebs/maya/pkg/upgrade/result/v1alpha1"
	"github.com/openebs/maya/pkg/version"
	"github.com/openebs/maya/tests/artifacts"
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	maxRetry         = 30
	openebsNamespace = "openebs"
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
	NodeClient     *node.Kubeclient
	PodClient      *pod.KubeClient
	SCClient       *sc.Kubeclient
	PVCClient      *pvc.Kubeclient
	NSClient       *ns.Kubeclient
	SnapClient     *snap.Kubeclient
	CSPClient      *csp.Kubeclient
	CSPIClient     *cspi.Kubeclient
	SPCClient      *spc.Kubeclient
	SVCClient      *svc.Kubeclient
	CVClient       *cv.Kubeclient
	CVCClient      *cvc.Kubeclient
	CSPCClient     *cspc.Kubeclient
	CVRClient      *cvr.Kubeclient
	URClient       *result.Kubeclient
	UnstructClient *unstruct.Kubeclient
	DeployClient   *deploy.Kubeclient
	BDClient       *bd.Kubeclient
	BDCClient      *bdc.Kubeclient
	KubeConfigPath string
	NameSpace      string
	Config         interface{}
}

// SPCConfig provides config to create cstor pools
type SPCConfig struct {
	Name      string
	DiskType  string
	PoolType  string
	PoolCount int
	// OverProvisioning field is deprecated and not honoured
	IsOverProvisioning bool

	IsThickProvisioning bool
}

// SCConfig provides config to create storage class
type SCConfig struct {
	Name        string
	Annotations map[string]string
	Provisioner string
}

// PVCConfig provides config to create PersistentVolumeClaim
type PVCConfig struct {
	Name        string
	Namespace   string
	SCName      string
	Capacity    string
	AccessModes []corev1.PersistentVolumeAccessMode
}

// CVRConfig provides config to create CStorVolumeReplica
type CVRConfig struct {
	PoolObj    *apis.CStorPool
	VolumeName string
	Namespace  string
	Capacity   string
	Phase      string
	TargetIP   string
	ReplicaID  string
}

// ServiceConfig provides config to create Service
type ServiceConfig struct {
	Name        string
	Namespace   string
	Selectors   map[string]string
	ServicePort []corev1.ServicePort
}

var (
	pvLabel            = "openebs.io/persistent-volume="
	poolLabel          = "openebs.io/storagepoolclaim="
	cstorPoolNameLabel = "cstorpool.openebs.io/name"
)

// OperationsOptions abstracts creating an
// instance of operations
type OperationsOptions func(*Operations)

// WithKubeConfigPath sets the kubeConfig path
// against operations instance
func WithKubeConfigPath(path string) OperationsOptions {
	return func(ops *Operations) {
		ops.KubeConfigPath = path
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
		ops.KubeClient = kubeclient.New(kubeclient.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.NSClient == nil {
		ops.NSClient = ns.NewKubeClient(ns.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.SCClient == nil {
		ops.SCClient = sc.NewKubeClient(sc.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.PodClient == nil {
		ops.PodClient = pod.NewKubeClient(pod.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.PVCClient == nil {
		ops.PVCClient = pvc.NewKubeClient(pvc.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.SnapClient == nil {
		ops.SnapClient = snap.NewKubeClient(snap.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.SPCClient == nil {
		ops.SPCClient = spc.NewKubeClient(spc.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.CSPClient == nil {
		ops.CSPClient, err = csp.KubeClient().WithKubeConfigPath(ops.KubeConfigPath)
		Expect(err).To(BeNil(), "while initilizing csp client")
	}
	if ops.CSPCClient == nil {
		ops.CSPCClient = cspc.NewKubeClient(cspc.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.CSPIClient == nil {
		ops.CSPIClient = cspi.NewKubeClient(cspi.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.CVClient == nil {
		ops.CVClient = cv.NewKubeclient(cv.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.CVCClient == nil {
		ops.CVCClient = cvc.NewKubeclient(cvc.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.CVRClient == nil {
		ops.CVRClient = cvr.NewKubeclient(cvr.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.URClient == nil {
		ops.URClient = result.NewKubeClient(result.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.UnstructClient == nil {
		ops.UnstructClient = unstruct.NewKubeClient(unstruct.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.DeployClient == nil {
		ops.DeployClient = deploy.NewKubeClient(deploy.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.BDClient == nil {
		ops.BDClient = bd.NewKubeClient(bd.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.NodeClient == nil {
		ops.NodeClient = node.NewKubeClient(node.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.BDCClient == nil {
		ops.BDCClient = bdc.NewKubeClient(bdc.WithKubeConfigPath(ops.KubeConfigPath))
	}
	if ops.SVCClient == nil {
		ops.SVCClient = svc.NewKubeClient(svc.WithKubeConfigPath(ops.KubeConfigPath))
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
func (ops *Operations) GetCstorVolumeCount(namespace, lselector string, expectedCVCount int, preds ...cv.Predicate) int {
	var cvCount int
	for i := 0; i < maxRetry; i++ {
		cvCount = ops.GetCVCount(namespace, lselector, preds...)
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
		cvCount := ops.GetCVCount(namespace, lselector, cv.IsHealthy())
		return cvCount
	},
		120, 10).Should(Equal(expectedCVCount))
}

// GetCstorVolumeReplicaCountEventually gives the count of cstorvolume based on
// selecter eventually
func (ops *Operations) GetCstorVolumeReplicaCountEventually(namespace, lselector string, expectedCVRCount int, pred ...cvr.Predicate) bool {
	return Eventually(func() int {
		cvCount := ops.GetCstorVolumeReplicaCount(namespace, lselector, pred...)
		return cvCount
	},
		120, 10).Should(Equal(expectedCVRCount))
}

// GetCstorVolumeClaimCountEventually gives the count of cstorvolume claim
func (ops *Operations) GetCstorVolumeClaimCountEventually(namespace, cvcName string, expectedCVCCount int) bool {
	return Eventually(func() int {
		cvcCount := ops.GetCstorVolumeClaimCount(namespace, cvcName)
		return cvcCount
	},
		120, 10).Should(Equal(expectedCVCCount))
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
func (ops *Operations) GetCVCount(namespace, lselector string, pred ...cv.Predicate) int {
	cvs, err := ops.CVClient.
		WithNamespace(namespace).
		List(metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return cv.
		NewListBuilder().
		WithAPIList(cvs).
		WithFilter(pred...).
		List().
		Len()
}

// GetCstorVolumeReplicaCount gives cstorvolumereplica healthy count currently based on selecter
func (ops *Operations) GetCstorVolumeReplicaCount(namespace, lselector string, pred ...cvr.Predicate) int {
	cvrs, err := ops.CVRClient.
		WithNamespace(namespace).
		List(metav1.ListOptions{LabelSelector: lselector})
	Expect(err).ShouldNot(HaveOccurred())
	return cvr.
		NewListBuilder().
		WithAPIList(cvrs).
		WithFilter(pred...).
		List().
		Len()
}

// GetCstorVolumeClaimCount gives cstorVolumeClaim healthy count currently
func (ops *Operations) GetCstorVolumeClaimCount(namespace, cvcName string) int {
	_, err := ops.CVCClient.WithNamespace(namespace).Get(cvcName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return 0
	}
	Expect(err).ShouldNot(HaveOccurred())
	return 1
}

// GetReadyNodes gives cstorvolumereplica healthy count currently based on selecter
func (ops *Operations) GetReadyNodes() *corev1.NodeList {
	nodes, err := ops.NodeClient.
		List(metav1.ListOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	return node.
		NewListBuilder().
		WithAPIList(nodes).
		WithFilter(node.IsReady()).
		List().
		ToAPIList()
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
		120, 10).
		Should(BeTrue())
}

// VerifyCapacity checks if the pvc capacity has been updated
func (ops *Operations) VerifyCapacity(pvcName, capacity string) bool {
	return Eventually(func() bool {
		volume, err := ops.PVCClient.
			Get(pvcName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		actualCapacity := volume.Status.Capacity[corev1.ResourceStorage]
		desiredCapacity, _ := resource.ParseQuantity(capacity)
		return (desiredCapacity.Cmp(actualCapacity) == 0)
	},
		120, 10).
		Should(BeTrue())
}

// PodDeleteCollection deletes all the pods in a namespace matched the given
// labelselector
func (ops *Operations) PodDeleteCollection(ns string, lopts metav1.ListOptions) error {
	deletePolicy := metav1.DeletePropagationForeground
	dopts := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	return ops.PodClient.WithNamespace(ns).DeleteCollection(lopts, dopts)
}

// IsPodRunningEventually return true if the pod comes to running state
func (ops *Operations) IsPodRunningEventually(namespace, podName string) bool {
	return Eventually(func() bool {
		p, err := ops.PodClient.
			WithNamespace(namespace).
			Get(podName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		return pod.NewForAPIObject(p).
			IsRunning()
	},
		150, 10).
		Should(BeTrue())
}

// ExecuteCMDEventually executes the command on pod container
// and returns stdout
func (ops *Operations) ExecuteCMDEventually(
	podObj *corev1.Pod,
	containerName,
	cmd string,
	expectStdout bool,
) string {
	var err error
	output := &pod.ExecOutput{}
	podName := podObj.Name
	namespace := podObj.Namespace
	status := ops.IsPodRunningEventually(namespace, podName)
	Expect(status).To(Equal(true),
		"while checking the status of pod {%s} in namespace {%s}",
		podName,
		namespace,
	)
	for i := 0; i < maxRetry; i++ {
		output, err = ops.PodClient.WithNamespace(namespace).
			Exec(
				podName,
				&corev1.PodExecOptions{
					Command: []string{
						"/bin/sh",
						"-c",
						cmd,
					},
					Container: containerName,
					Stdin:     false,
					Stdout:    true,
					Stderr:    true,
				},
			)
		Expect(err).ShouldNot(
			HaveOccurred(),
			"failed to execute command {%s} on pod {%s} namespace {%s}",
			cmd,
			podName,
			namespace,
		)
		// If caller pass expectStdout as false return from here
		if !expectStdout {
			return ""
		}
		if output.Stdout != "" {
			return output.Stdout
		}
		time.Sleep(5 * time.Second)
	}
	err = errors.Errorf(
		"failed to execute cmd %s on pod %s",
		cmd,
		podName,
	)
	Expect(err).To(BeNil(),
		"failed to execute cmd {%s} on pod {%s} in namespace {%s} stdout {%s}",
		cmd,
		podName,
		namespace,
		output.Stdout,
	)
	return ""
}

// RestartPodEventually restarts the pod and return
func (ops *Operations) RestartPodEventually(podObj *corev1.Pod) error {
	status := ops.IsPodRunningEventually(podObj.Namespace, podObj.Name)
	if !status {
		return errors.Errorf(
			"while checking the status of pod {%s} in namespace {%s} before restarting",
			podObj.Name,
			podObj.Namespace,
		)
	}

	err := ops.PodClient.WithNamespace(podObj.Namespace).
		Delete(podObj.Name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err,
			"failed to delete pod {%s} in namespace {%s}",
			podObj.Name,
			podObj.Namespace,
		)
	}

	status = ops.IsPodDeletedEventually(podObj.Namespace, podObj.Name)
	if !status {
		return errors.Errorf(
			"while checking termination of pod {%s} in namespace {%s}",
			podObj.Name,
			podObj.Namespace,
		)
	}
	return nil
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

// IsPVCDeletedEventually tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func (ops *Operations) IsPVCDeletedEventually(pvcName string) bool {
	return Eventually(func() bool {
		_, err := ops.PVCClient.
			Get(pvcName, metav1.GetOptions{})
		return isNotFound(err)
	},
		120, 10).
		Should(BeTrue())
}

// IsCSPCDeletedEventually tries to get the deleted cspc
// and returns true if cspc is not found
// else returns false
func (ops *Operations) IsCSPCDeletedEventually(cspcName string) bool {
	return Eventually(func() bool {
		_, err := ops.CSPCClient.
			Get(cspcName, metav1.GetOptions{})
		return isNotFound(err)
	},
		120, 10).
		Should(BeTrue())
}

// IsPodDeletedEventually checks if the pod is deleted or not eventually
func (ops *Operations) IsPodDeletedEventually(namespace, podName string) bool {
	return Eventually(func() bool {
		_, err := ops.PodClient.
			WithNamespace(namespace).
			Get(podName, metav1.GetOptions{})
		return isNotFound(err)
	},
		120, 10).
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

// DeleteCSPI deletes  ...
func (ops *Operations) DeleteCSPI(cspcName string, deleteCount int) {
	cspiAPIList, err := ops.CSPIClient.WithNamespace(ops.NameSpace).List(metav1.ListOptions{})
	Expect(err).To(BeNil())
	cspiList := cspi.
		ListBuilderFromAPIList(cspiAPIList).
		List().
		Filter(cspi.HasLabel(string(apis.CStorPoolClusterCPK), cspcName), cspi.IsStatus("ONLINE"))
	cspiCount := cspiList.Len()
	Expect(deleteCount).Should(BeNumerically("<=", cspiCount))

	for i := 0; i < deleteCount; i++ {
		err := ops.CSPIClient.WithNamespace(ops.NameSpace).Delete(cspiList.ObjectList.Items[i].Name, &metav1.DeleteOptions{})
		Expect(err).To(BeNil())
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

// GetCSPCountEventually get csp count based on the labelSelector and
// wait til csp count matches to provided argument count
func (ops *Operations) GetCSPCountEventually(labelSelector string, count int) bool {
	return Eventually(func() int {
		cspAPIList, err := ops.CSPClient.List(metav1.ListOptions{LabelSelector: labelSelector})
		Expect(err).To(BeNil())
		count := csp.
			ListBuilderForAPIObject(cspAPIList).
			List().
			Len()
		return count
	},
		120, 10).
		Should(Equal(count))
}

// GetCSPCount gets csp count based on spcName at that time
func (ops *Operations) GetCSPCount(labelSelector string) int {
	cspAPIList, err := ops.CSPClient.List(metav1.ListOptions{LabelSelector: labelSelector})
	Expect(err).To(BeNil())
	return len(cspAPIList.Items)
}

// VerifyDesiredCSPCount verifies whether count of CSP belongs to SPC in cluster
// matched with provided argument count
func (ops *Operations) VerifyDesiredCSPCount(spcObj *apis.StoragePoolClaim, poolCount int) {
	cspCount := ops.GetHealthyCSPCount(spcObj.Name, poolCount)
	Expect(cspCount).To(Equal(poolCount))

	// Check are there any extra csps
	cspCount = ops.GetCSPCount(getLabelSelector(spcObj))
	Expect(cspCount).To(Equal(poolCount), "Mismatch Of CSP Count")
}

// This function is local to this package
func getLabelSelector(spc *apis.StoragePoolClaim) string {
	return string(apis.StoragePoolClaimCPK) + "=" + spc.Name
}

// GetCSPICount gets cspi count based on cspc name at that time
func (ops *Operations) GetCSPICount(labelSelector string) int {
	cspiAPIList, err := ops.CSPIClient.WithNamespace(ops.NameSpace).List(metav1.ListOptions{LabelSelector: labelSelector})
	Expect(err).To(BeNil())
	return len(cspiAPIList.Items)
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

// GetCSPIResourceCountEventually gets  cspis based on cspcName
func (ops *Operations) GetCSPIResourceCountEventually(labelSelector string, expectedCSPICount int) int {
	var cspiCount int
	for i := 0; i < maxRetry; i++ {
		cspiCount = ops.GetCSPICount(labelSelector)
		if cspiCount == expectedCSPICount {
			return cspiCount
		}
		time.Sleep(5 * time.Second)
	}
	return cspiCount
}

// GetHealthyCSPICount gets healthy csp based on spcName
func (ops *Operations) GetHealthyCSPICount(cspcName string, expectedCSPICount int) int {
	var cspiCount int
	// as cspi deletion takes more time now for cleanup of its resources
	// for reconciled cspi to come up it can take additional time.
	for i := 0; i < (maxRetry + 60); i++ {
		cspiAPIList, err := ops.CSPIClient.WithNamespace(ops.NameSpace).List(metav1.ListOptions{})
		time.Sleep(5 * time.Second)
		Expect(err).To(BeNil())
		cspiCount = cspi.
			ListBuilderFromAPIList(cspiAPIList).
			List().
			Filter(cspi.HasLabel(string(apis.CStorPoolClusterCPK), cspcName), cspi.IsStatus("ONLINE")).
			Len()
		if cspiCount == expectedCSPICount {
			return cspiCount
		}
		time.Sleep(3 * time.Second)
	}
	return cspiCount
}

// GetBDCCountEventually gets BDC resource count based on provided list option.
func (ops *Operations) GetBDCCountEventually(listOptions metav1.ListOptions, expectedBDCCount int, namespace string) int {
	var bdcCount int
	for i := 0; i < maxRetry; i++ {
		bdcAPIList, err := ops.BDCClient.WithNamespace(namespace).List(listOptions)
		Expect(err).To(BeNil())
		bdcCount = len(bdcAPIList.Items)
		if bdcCount == expectedBDCCount {
			return bdcCount
		}
		time.Sleep(5 * time.Second)
	}
	return bdcCount
}

// IsCSPCNotExists returns true if the cspc with provided name does not exists.
func (ops *Operations) IsCSPCNotExists(cspcName string) bool {
	for i := 0; i < maxRetry; i++ {
		_, err := ops.CSPCClient.WithNamespace(ops.NameSpace).Get(cspcName, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return true
		}
		time.Sleep(5 * time.Second)
	}
	return false
}

// IsSPCNotExists returns true if the spc with provided name does not exists.
func (ops *Operations) IsSPCNotExists(spcName string) bool {
	for i := 0; i < maxRetry; i++ {
		_, err := ops.SCClient.Get(spcName, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return true
		}
		time.Sleep(5 * time.Second)
	}
	return false
}

// IsFinalizerExistsOnBDC returns true if the object with provided name contains the finalizer.
func (ops *Operations) IsFinalizerExistsOnBDC(bdcName, finalizer string) bool {
	for i := 0; i < maxRetry; i++ {
		bdcObj, err := ops.BDCClient.Get(bdcName, metav1.GetOptions{})
		Expect(err).To(BeNil())
		for _, f := range bdcObj.Finalizers {
			if f == finalizer {
				return true
			}
		}
		time.Sleep(5 * time.Second)
	}
	return false
}

// IsCSPCFinalizerExistsOnCSPC returns true if the cspc with provided name contains the cspc finalizer.
func (ops *Operations) IsCSPCFinalizerExistsOnCSPC(cspcName, cspcFinalizer string) bool {
	for i := 0; i < maxRetry; i++ {
		gotCSPC, err := ops.CSPCClient.WithNamespace(ops.NameSpace).Get(cspcName, metav1.GetOptions{})
		Expect(err).To(BeNil())
		for _, finalizer := range gotCSPC.Finalizers {

			if finalizer == cspcFinalizer {
				return true
			}
		}
		time.Sleep(5 * time.Second)
	}
	return false
}

// IsSPCFinalizerExistsOnSPC returns true if the spc with provided name contains the spc finalizer.
func (ops *Operations) IsSPCFinalizerExistsOnSPC(spcName, spcFinalizer string) bool {
	for i := 0; i < maxRetry; i++ {
		gotSPC, err := ops.SPCClient.Get(spcName, metav1.GetOptions{})
		Expect(err).To(BeNil())
		for _, finalizer := range gotSPC.Finalizers {
			if finalizer == spcFinalizer {
				return true
			}
		}
		time.Sleep(5 * time.Second)
	}
	return false
}

// IsCSPCFinalizerExistsOnBDCs returns true if the all the BDCs( selected by provided list options)
// has cspc finalizer on it.
func (ops *Operations) IsCSPCFinalizerExistsOnBDCs(listOptions metav1.ListOptions, cspcFinalizer string) bool {
	for i := 0; i < maxRetry; i++ {
		cspcFinalizerPresent := true
		gotBDCList, err := ops.BDCClient.WithNamespace(ops.NameSpace).List(listOptions)
		Expect(err).To(BeNil())
		for _, BDCObj := range gotBDCList.Items {
			BDCObj := BDCObj
			if !bdc.BuilderForAPIObject(&BDCObj).BDC.HasFinalizer(cspcFinalizer) {
				cspcFinalizerPresent = false
			}
		}
		if !cspcFinalizerPresent {
			time.Sleep(5 * time.Second)
		} else {
			return true
		}

	}
	return false
}

// IsSPCFinalizerExistsOnBDCs returns true if the all the BDCs( selected by provided list options)
// has spc finalizer on it.
func (ops *Operations) IsSPCFinalizerExistsOnBDCs(listOptions metav1.ListOptions, spcFinalizer string) bool {
	for i := 0; i < maxRetry; i++ {
		spcFinalizerPresent := true
		gotBDCList, err := ops.BDCClient.List(listOptions)
		Expect(err).To(BeNil())
		for _, BDCObj := range gotBDCList.Items {
			BDCObj := BDCObj
			if !bdc.BuilderForAPIObject(&BDCObj).BDC.HasFinalizer(spcFinalizer) {
				spcFinalizerPresent = false
			}
		}
		if !spcFinalizerPresent {
			time.Sleep(5 * time.Second)
		} else {
			return true
		}

	}
	return false
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
		120, 10).
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
	Expect(execOut.Len()).Should(BeNumerically(">=", 0), "while streaming the command in pod ", opts.podName, execErr.String(), execOut.String())
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

// GetBDCCount gets BDC resource count based on provided label selector
func (ops *Operations) GetBDCCount(lSelector, namespace string) int {
	bdcList, err := ops.BDCClient.
		WithNamespace(namespace).
		List(metav1.ListOptions{LabelSelector: lSelector})
	Expect(err).ShouldNot(HaveOccurred())
	return len(bdcList.Items)
}

// GetCSPCBDListForNode returns unclaimed block devices that can be used.
func (ops *Operations) GetCSPCBDListForNode(node *corev1.Node, blockDeviceCount int) []*apis.CStorPoolClusterBlockDevice {
	bdList, err := ops.BDClient.WithNamespace(ops.NameSpace).List(metav1.ListOptions{LabelSelector: string(apis.HostNameCPK) + "=" + node.GetLabels()[string(apis.HostNameCPK)]})
	Expect(err).To(BeNil())
	Expect(len(bdList.Items)).Should(BeNumerically(">=", blockDeviceCount))
	// TODO : Filter Unclaimed BDs
	bdList = bd.ListBuilderFromAPIList(bdList).BDL.Filter(bd.IsUnclaimed()).ObjectList

	var cspcBDs []*apis.CStorPoolClusterBlockDevice
	for i := 0; i < blockDeviceCount; i++ {
		cspcBD := &apis.CStorPoolClusterBlockDevice{}
		cspcBD.BlockDeviceName = bdList.Items[i].Name
		cspcBDs = append(cspcBDs, cspcBD)
	}

	return cspcBDs
}

// BuildAndCreateSPC builds and creates StoragePoolClaim in cluster
func (ops *Operations) BuildAndCreateSPC() *apis.StoragePoolClaim {
	spcConfig := ops.Config.(*SPCConfig)
	spcObj := spc.NewBuilder().
		WithGenerateName(spcConfig.Name).
		WithDiskType(spcConfig.DiskType).
		WithMaxPool(spcConfig.PoolCount).
		WithThickProvisioning(spcConfig.IsThickProvisioning).
		WithPoolType(spcConfig.PoolType).
		Build().Object
	newSPCObj, err := ops.SPCClient.Create(spcObj)
	Expect(err).To(BeNil())
	return newSPCObj
}

// CreateStorageClass builds and creates storageclass in cluster
func (ops *Operations) CreateStorageClass() *storagev1.StorageClass {
	scConfig := ops.Config.(*SCConfig)
	scObj, err := sc.NewBuilder().
		WithGenerateName(scConfig.Name).
		WithAnnotations(scConfig.Annotations).
		WithProvisioner(scConfig.Provisioner).Build()
	Expect(err).ShouldNot(
		HaveOccurred(),
		"while building storageclass {%s}", scConfig.Name,
	)
	scObj, err = ops.SCClient.Create(scObj)
	Expect(err).To(BeNil(), "while creating storageclass {%s}", scConfig.Name)
	return scObj
}

// BuildAndCreatePVC builds and creates PersistentVolumeClaim in cluster
func (ops *Operations) BuildAndCreatePVC() *corev1.PersistentVolumeClaim {
	pvcConfig := ops.Config.(*PVCConfig)
	pvcObj, err := pvc.NewBuilder().
		WithName(pvcConfig.Name).
		WithNamespace(pvcConfig.Namespace).
		WithStorageClass(pvcConfig.SCName).
		WithAccessModes(pvcConfig.AccessModes).
		WithCapacity(pvcConfig.Capacity).Build()
	Expect(err).ShouldNot(
		HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcConfig.Name,
		pvcConfig.Namespace,
	)
	pvcObj, err = ops.PVCClient.WithNamespace(pvcConfig.Namespace).Create(pvcObj)
	Expect(err).To(
		BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcConfig.Name,
		pvcConfig.Namespace,
	)
	return pvcObj
}

// BuildAndCreateCVR builds and creates CVR in cluster
func (ops *Operations) BuildAndCreateCVR() *apis.CStorVolumeReplica {
	cvrConfig := ops.Config.(*CVRConfig)
	cvrObj, err := cvr.NewBuilder().
		WithName(cvrConfig.VolumeName + "-" + cvrConfig.PoolObj.Name).
		WithNamespace(cvrConfig.Namespace).
		WithLabelsNew(getCVRLabels(cvrConfig.PoolObj, cvrConfig.VolumeName)).
		WithAnnotationsNew(getCVRAnnotations(cvrConfig.PoolObj)).
		WithFinalizers([]string{cvr.CStorVolumeReplicaFinalizer}).
		WithCapacity(cvrConfig.Capacity).
		WithTargetIP(cvrConfig.TargetIP).
		WithStatusPhase(apis.CStorVolumeReplicaPhase(cvrConfig.Phase)).
		WithReplicaID(cvrConfig.ReplicaID).
		Build()
	Expect(err).To(BeNil())
	cvrObj, err = ops.CVRClient.
		WithNamespace(cvrConfig.Namespace).
		Create(cvrObj)
	Expect(err).To(BeNil())
	return cvrObj
}

// BuildAndCreateService builds and creates Service in cluster
func (ops *Operations) BuildAndCreateService() *corev1.Service {
	svcConfig := ops.Config.(*ServiceConfig)
	buildSVCObj, err := svc.NewBuilder().
		WithGenerateName(svcConfig.Name).
		WithNamespace(svcConfig.Namespace).
		WithSelectorsNew(svcConfig.Selectors).
		WithPorts(svcConfig.ServicePort).
		WithType(corev1.ServiceTypeNodePort).
		Build()
	Expect(err).To(BeNil())
	svcObj, err := ops.SVCClient.
		WithNamespace(svcConfig.Namespace).
		Create(buildSVCObj)
	Expect(err).To(BeNil())
	return svcObj
}

// DeletePersistentVolumeClaim deletes PVC from cluster based on provided
// argument
func (ops *Operations) DeletePersistentVolumeClaim(name, namespace string) {
	err := ops.PVCClient.WithNamespace(namespace).Delete(name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil())
}

// VerifyVolumeResources verifies whether volume resource exist or not
func (ops *Operations) VerifyVolumeResources(pvName, namespace string, cvrPred cvr.PredicateList, cvPred cv.PredicateList) {
	volumeLabel := pvLabel + pvName
	targetPodCount := ops.GetPodRunningCountEventually(namespace, volumeLabel, 0)
	Expect(targetPodCount).To(Equal(0), "when pvc is deleted target pod should be deleted")

	cvCount := ops.GetCstorVolumeCount(openebsNamespace, volumeLabel, 0, cvPred...)
	Expect(cvCount).To(Equal(0), "when pvc is deleted cstorvolume should be deleted")

	IsCVRDeleted := ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, volumeLabel, 0, cvrPred...)
	Expect(IsCVRDeleted).To(Equal(true), "when pvc is deleted cstorvolumereplica should be deleted")
}

// DeleteStoragePoolClaim deletes StoragePoolClaim based on pool name
func (ops *Operations) DeleteStoragePoolClaim(spcName string) {
	err := ops.SPCClient.Delete(spcName, &metav1.DeleteOptions{})
	Expect(err).To(BeNil())
}

// VerifyPoolResources verifies whether pool resource exist or not
func (ops *Operations) VerifyPoolResources(spcName string) {
	labelSelector := poolLabel + spcName
	isCSPDeleted := ops.GetCSPCountEventually(labelSelector, 0)
	Expect(isCSPDeleted).To(Equal(true))
}

// VerifyVolumeStatus checks multiple resources related to volume
// 1. Verifies whether PVC is bound to pv or not
// 2. Verifies whether CStorVolume is in Healthy or not
// 3. Verifies whether specified number of CVR's are healthy or not
func (ops *Operations) VerifyVolumeStatus(
	pvcObj *corev1.PersistentVolumeClaim, replicaCount int, cvrPred cvr.PredicateList, cvPred cv.PredicateList) {
	status := ops.IsPVCBoundEventually(pvcObj.Name)
	Expect(status).To(Equal(true), "while checking status equal to bound")

	// GetLatest PVC object
	updatedPVCObj, err := ops.PVCClient.
		WithNamespace(pvcObj.Namespace).
		Get(pvcObj.Name, metav1.GetOptions{})
	Expect(err).To(BeNil())

	volumeLabel := pvLabel + updatedPVCObj.Spec.VolumeName
	cvCount := ops.GetCstorVolumeCount(openebsNamespace, volumeLabel, 1, cvPred...)
	Expect(cvCount).To(Equal(1), "while checking cstorvolume count")

	cvrCount := ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, volumeLabel, replicaCount, cvrPred...)
	Expect(cvrCount).To(Equal(true), "while checking cstorvolume replica count")
}

// DeleteVolumeResources deletes pvc and storageclass from cluster
func (ops *Operations) DeleteVolumeResources(
	pvcObj *corev1.PersistentVolumeClaim,
	scObj *storagev1.StorageClass) {
	ops.DeletePersistentVolumeClaim(pvcObj.Name, pvcObj.Namespace)
	ops.VerifyVolumeResources(pvcObj.Spec.VolumeName, openebsNamespace, cvr.PredicateList{}, cv.PredicateList{})
	err := ops.SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil())
}

// GetUnUsedCStorPool returns the csp object where the volume replica doesn't exist
func (ops *Operations) GetUnUsedCStorPool(
	cvrList *apis.CStorVolumeReplicaList,
	poolLabel string) *apis.CStorPool {
	usedPools := map[string]bool{}
	for _, cvrObj := range cvrList.Items {
		poolName, ok := cvrObj.GetLabels()[cstorPoolNameLabel]
		if ok {
			usedPools[poolName] = true
		}
	}
	cspList, err := ops.CSPClient.
		List(metav1.ListOptions{LabelSelector: poolLabel})
	Expect(err).To(BeNil())
	for _, obj := range cspList.Items {
		obj := obj
		if !usedPools[obj.Name] {
			return &obj
		}
	}
	err = errors.Errorf("pools are not available to migrate storage replica")
	Expect(err).To(BeNil())
	return nil
}

// getCVRAnnotations get the annotations for cstorvolumereplica
func getCVRAnnotations(pool *apis.CStorPool) map[string]string {
	return map[string]string{
		"cstorpool.openebs.io/hostname": pool.Labels["kubernetes.io/hostname"],
	}
}

// getCVRLabels get the labels for cstorvolumereplica
func getCVRLabels(pool *apis.CStorPool, volumeName string) map[string]string {
	return map[string]string{
		"cstorpool.openebs.io/name":    pool.Name,
		"cstorpool.openebs.io/uid":     string(pool.UID),
		"cstorvolume.openebs.io/name":  volumeName,
		"openebs.io/persistent-volume": volumeName,
		"openebs.io/version":           version.GetVersion(),
	}
}
