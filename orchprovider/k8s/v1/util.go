package v1

import (
	"fmt"
	"strings"

	"github.com/openebs/maya/types/v1"
	orchProfile "github.com/openebs/maya/types/v1/profile/orchestrator"
	volProfile "github.com/openebs/maya/volume/profiles"
	"k8s.io/client-go/kubernetes"
	k8sCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	k8sExtnsV1Beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
	k8sApiV1 "k8s.io/client-go/pkg/api/v1"
	k8sApisExtnsBeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"k8s.io/client-go/rest"
)

// K8sUtilGetter is an abstraction to fetch instances of K8sUtilInterface
type K8sUtilGetter interface {
	GetK8sUtil(volProfile.VolumeProvisionerProfile) K8sUtilInterface
}

// K8sUtilInterface is an abstraction over communicating with K8s APIs
type K8sUtilInterface interface {
	// Name of K8s utility
	Name() string

	// K8sClient fetches an instance of K8sClients. Will return
	// false if the util does not support providing K8sClients instance.
	K8sClient() (K8sClient, bool)

	// K8sClientV2 fetches an instance of K8sClientV2.
	K8sClientV2() (K8sClientV2, bool, error)
}

// TODO Deprecate in favour of K8sClientV2
// K8sClient is an abstraction to operate on various k8s entities.
type K8sClient interface {
	// IsInCluster indicates whether the operation is within cluster or in a
	// different cluster
	IsInCluster() (bool, error)

	// NS provides the namespace where operations will be executed
	NS() (string, error)

	// TODO
	//    Rename to PodOps
	//
	// Pods provides all the CRUD operations associated w.r.t a POD
	Pods() (k8sCoreV1.PodInterface, error)

	// TODO
	//    Rename to ServiceOps
	//
	// Services provides all the CRUD operations associated w.r.t a Service
	Services() (k8sCoreV1.ServiceInterface, error)

	// DeploymentOps provides all the CRUD operations associated w.r.t a Deployment
	DeploymentOps() (k8sExtnsV1Beta1.DeploymentInterface, error)
}

// K8sClientV2 is an abstraction to operate on various k8s entities.
type K8sClientV2 interface {
	// IsInClusterV2 indicates whether the operation is within cluster or in a
	// different cluster
	IsInClusterV2() (bool, error)

	// NSV2 provides the namespace where operations will be executed
	NSV2() (string, error)

	// StorageClassOps provides all the CRUD & more operations associated
	// w.r.t a StorageClass
	StorageClassOps() (storagev1.StorageClassInterface, error)
}

// k8sUtil provides the concrete implementation for below interfaces:
//
// 1. k8s.K8sUtilInterface interface
// 2. k8s.K8sClients interface
type k8sUtil struct {

	// namespace refers to K8s namespace where this operation
	// will be performed
	namespace string

	// inCS refers to the ClientSet capable of communicating
	// within the current K8s cluster i.e. where this binary is
	// running
	inCS *kubernetes.Clientset

	// outCS refers to the ClientSet capable of communicating
	// outside of current K8s cluster i.e. where this binary is
	// running
	outCS *kubernetes.Clientset

	caCert     string
	caPath     string
	clientCert string
	clientKey  string
	insecure   bool

	// TODO Deprecate in favour of volume
	// volProfile has context related information w.r.t k8s
	volProfile volProfile.VolumeProvisionerProfile

	// volume represents an OpenEBS volume which will be
	// placed/updated in K8s
	volume *v1.Volume
}

// This is a plain k8s utility & hence the name
func (k *k8sUtil) Name() string {
	ns, _ := k.NS()
	return fmt.Sprintf("k8sutil @ '%s'", ns)
}

// k8sUtil implements K8sClient interface. Hence it returns
// self
func (k *k8sUtil) K8sClient() (K8sClient, bool) {
	return k, true
}

// NS provides the namespace where operations will be executed
func (k *k8sUtil) NS() (string, error) {
	if nil == k.volProfile {
		return "", fmt.Errorf("Volume provisioner profile not initialized at '%s'", k.Name())
	}

	// Fetch vol from volume provisioner profile
	vol, err := k.volProfile.Volume()
	if err != nil {
		return "", err
	}

	// Get orchestrator provider profile from vol
	oPrfle, err := orchProfile.GetOrchProviderProfile(vol)
	if err != nil {
		return "", err
	}

	// Get the namespace which will be queried
	ns, err := oPrfle.NS()
	if err != nil {
		return "", err
	}

	return ns, nil
}

// InCluster indicates whether the operation is within cluster or in a
// different cluster
func (k *k8sUtil) IsInCluster() (bool, error) {
	if nil == k.volProfile {
		return false, fmt.Errorf("Volume provisioner profile not initialized at '%s'", k.Name())
	}

	// Fetch vol from volume provisioner profile
	vol, err := k.volProfile.Volume()
	if err != nil {
		return false, err
	}

	// Get orchestrator provider profile from vol
	oPrfle, err := orchProfile.GetOrchProviderProfile(vol)
	if err != nil {
		return false, err
	}

	// Which kind of request ? in-cluster or out-of-cluster ?
	isInCluster, err := oPrfle.InCluster()
	if err != nil {
		return false, err
	}

	return isInCluster, nil
}

// Pods is a utility function that provides a instance capable of
// executing various k8s pod related operations.
func (k *k8sUtil) Pods() (k8sCoreV1.PodInterface, error) {
	var cs *kubernetes.Clientset

	inC, err := k.IsInCluster()
	if err != nil {
		return nil, err
	}

	ns, err := k.NS()
	if err != nil {
		return nil, err
	}

	if inC {
		cs, err = k.getInClusterCS()
	} else {
		cs, err = k.getOutClusterCS()
	}

	if err != nil {
		return nil, err
	}

	return cs.CoreV1().Pods(ns), nil
}

// Services is a utility function that provides a instance capable of
// executing various k8s service related operations.
func (k *k8sUtil) Services() (k8sCoreV1.ServiceInterface, error) {
	var cs *kubernetes.Clientset

	inC, err := k.IsInCluster()
	if err != nil {
		return nil, err
	}

	ns, err := k.NS()
	if err != nil {
		return nil, err
	}

	if inC {
		cs, err = k.getInClusterCS()
	} else {
		cs, err = k.getOutClusterCS()
	}

	if err != nil {
		return nil, err
	}

	return cs.CoreV1().Services(ns), nil
}

// Services is a utility function that provides a instance capable of
// executing various k8s Deployment related operations.
func (k *k8sUtil) DeploymentOps() (k8sExtnsV1Beta1.DeploymentInterface, error) {
	var cs *kubernetes.Clientset

	inC, err := k.IsInCluster()
	if err != nil {
		return nil, err
	}

	ns, err := k.NS()
	if err != nil {
		return nil, err
	}

	if inC {
		cs, err = k.getInClusterCS()
	} else {
		cs, err = k.getOutClusterCS()
	}

	if err != nil {
		return nil, err
	}

	return cs.ExtensionsV1beta1().Deployments(ns), nil
}

// k8sUtil implements K8sClientV2 interface. Hence it returns
// self
func (k *k8sUtil) K8sClientV2() (K8sClientV2, bool, error) {

	if k.volume == nil {
		return nil, true, fmt.Errorf("Volume is not set")
	}

	return k, true, nil
}

// NSV2 provides the namespace where operations will be executed
func (k *k8sUtil) NSV2() (string, error) {
	if k.namespace != "" {
		return k.namespace, nil
	}

	k.namespace = k.volume.Namespace

	// error out if still empty
	if k.namespace == "" {
		return "", fmt.Errorf("Namespace is empty")
	}

	return k.namespace, nil
}

// InCluster indicates whether the operation is within cluster or in a
// different cluster
func (k *k8sUtil) IsInClusterV2() (bool, error) {
	// Which kind of request ? in-cluster or out-of-cluster ?
	outCluster := k.volume.Labels.K8sOutCluster
	if outCluster == "" {
		return true, nil
	}

	return false, nil
}

func (k *k8sUtil) StorageClassOps() (storagev1.StorageClassInterface, error) {
	cs, err := k.getClientSet()

	if err != nil {
		return nil, err
	}

	return cs.StorageV1().StorageClasses(), nil
}

// getClientSet is used to get a new http client capable
// of invoking K8s APIs.
func (k *k8sUtil) getClientSet() (*kubernetes.Clientset, error) {
	var cs *kubernetes.Clientset

	// Get if already available in current instance
	// NOTE: A new instance of k8sUtil is created per http request
	if k.inCS != nil {
		return k.inCS, nil
	}

	if k.outCS != nil {
		return k.outCS, nil
	}

	// Else get it fresh for this instance/http request
	inC, err := k.IsInClusterV2()
	if err != nil {
		return nil, err
	}

	// set based on in-cluster or out-of-cluster
	if inC {
		cs, err = k.getInClusterCS()
		// set it for future retrievals in same http request
		k.inCS = cs
	} else {
		cs, err = k.getOutClusterCS()
		// set it for future retrievals in same http request
		k.outCS = cs
	}

	if err != nil {
		return nil, err
	}

	return cs, nil
}

// getInClusterCS is used to initialize and return a new http client capable
// of invoking K8s APIs.
func (k *k8sUtil) getInClusterCS() (*kubernetes.Clientset, error) {

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// creates the in-cluster clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// getOutClusterCS is used to initialize and return a new http client capable
// of invoking outside the cluster K8s APIs.
func (k *k8sUtil) getOutClusterCS() (*kubernetes.Clientset, error) {
	return nil, fmt.Errorf("outClusterCS not supported in '%s'", k.Name())
}

//
func SetControllerIPs(cp k8sApiV1.Pod, annotations map[string]string) {
	current := strings.TrimSpace(cp.Status.PodIP)
	if current == "" {
		// Nothing to be done
		return
	}

	existing := strings.TrimSpace(annotations[string(v1.ControllerIPsAPILbl)])

	// Set the value or add to the existing values if not added earlier
	if existing == "" {
		annotations[string(v1.ControllerIPsAPILbl)] = current
	} else {
		annotations[string(v1.ControllerIPsAPILbl)] = existing + "," + current
	}
}

//
func SetReplicaIPs(rp k8sApiV1.Pod, annotations map[string]string) {
	current := strings.TrimSpace(rp.Status.PodIP)
	if current == "" {
		// Nothing to be done
		return
	}

	existing := strings.TrimSpace(annotations[string(v1.ReplicaIPsAPILbl)])

	// Set the value or add to the existing values if not added earlier
	if existing == "" {
		annotations[string(v1.ReplicaIPsAPILbl)] = current
	} else {
		annotations[string(v1.ReplicaIPsAPILbl)] = existing + "," + current
	}
}

//
func SetControllerStatuses(cp k8sApiV1.Pod, annotations map[string]string) {
	current := strings.TrimSpace(string(cp.Status.Phase))
	if current == "" {
		// Nothing to be done
		return
	}

	existing := strings.TrimSpace(annotations[string(v1.ControllerStatusAPILbl)])

	// Set the value or add to the existing values if not added earlier
	if existing == "" {
		annotations[string(v1.ControllerStatusAPILbl)] = current
	} else {
		annotations[string(v1.ControllerStatusAPILbl)] = existing + "," + current
	}
}

//
func SetReplicaStatuses(rp k8sApiV1.Pod, annotations map[string]string) {
	current := strings.TrimSpace(string(rp.Status.Phase))
	if current == "" {
		// Nothing to be done
		return
	}

	existing := strings.TrimSpace(annotations[string(v1.ReplicaStatusAPILbl)])

	// Set the value or add to the existing values if not added earlier
	if existing == "" {
		annotations[string(v1.ReplicaStatusAPILbl)] = current
	} else {
		annotations[string(v1.ReplicaStatusAPILbl)] = existing + "," + current
	}
}

// TODO
// Not sure !!
func SetServiceStatuses(svc k8sApiV1.Service, annotations map[string]string) {}

func SetReplicaCount(rd k8sApisExtnsBeta1.Deployment, annotations map[string]string) {
	annotations[string(v1.ReplicaCountAPILbl)] = fmt.Sprint(*rd.Spec.Replicas)
}

// TODO Get it from Pod
func SetReplicaVolSize(rd k8sApisExtnsBeta1.Deployment, annotations map[string]string) {
	// TODO
	// Set the size as labels in replica deployment & extract from the label
	// Current way of extraction is a very crude way !!
	con := rd.Spec.Template.Spec.Containers[0]
	size := con.Args[len(con.Args)-2]

	annotations[string(v1.VolumeSizeAPILbl)] = size
}

func SetIQN(vsm string, annotations map[string]string) {
	annotations[string(v1.IQNAPILbl)] = string(v1.JivaIqnFormatPrefix) + ":" + vsm
}

func SetControllerClusterIPs(svc k8sApiV1.Service, annotations map[string]string) {
	current := strings.TrimSpace(svc.Spec.ClusterIP)
	if current == "" {
		// Nothing to be done
		return
	}

	existing := strings.TrimSpace(annotations[string(v1.ClusterIPsAPILbl)])

	// Set the value or add to the existing values if not added earlier
	if existing == "" {
		annotations[string(v1.ClusterIPsAPILbl)] = current
	} else {
		annotations[string(v1.ClusterIPsAPILbl)] = existing + "," + current
	}
}

func SetISCSITargetPortals(svc k8sApiV1.Service, annotations map[string]string) {
	current := strings.TrimSpace(svc.Spec.ClusterIP)
	if current == "" {
		// Nothing to be done
		return
	}
	current = current + ":" + string(v1.JivaISCSIPortDef)

	existing := strings.TrimSpace(annotations[string(v1.TargetPortalsAPILbl)])

	// Set the value or add to the existing values if not added earlier
	if existing == "" {
		annotations[string(v1.TargetPortalsAPILbl)] = current
	} else {
		annotations[string(v1.TargetPortalsAPILbl)] = existing + "," + current
	}
}
