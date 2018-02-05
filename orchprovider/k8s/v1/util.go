package v1

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
	orchProfile "github.com/openebs/maya/types/v1/profile/orchestrator"
	volProfile "github.com/openebs/maya/volume/profiles"
	"k8s.io/client-go/kubernetes"
	k8sCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	k8sExtnsV1Beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
	//k8sApiV1 "k8s.io/client-go/pkg/api/v1"
	//k8sApisExtnsBeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"github.com/openebs/maya/pkg/client/clientset/versioned"
	oe_client_v1alpha1 "github.com/openebs/maya/pkg/client/clientset/versioned/typed/openebs/v1alpha1"
	k8sApiV1 "k8s.io/api/core/v1"
	k8sApisExtnsBeta1 "k8s.io/api/extensions/v1beta1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// OpenEBSImage represents an OpenEBS container image
type OpenEBSImage struct {
	// Image represents the image value
	// e.g. openebs/m-apiserver:latest & so on
	image string

	// envKey represents the ENV variable key
	// to fetch the image
	envKey v1.ENVKey
}

//
func NewOpenEBSImage(envKey v1.ENVKey) *OpenEBSImage {
	return &OpenEBSImage{
		envKey: envKey,
	}
}

//
func (o *OpenEBSImage) GetImage(useDefault bool) string {
	val := v1.GetEnv(o.envKey)

	if len(val) != 0 {
		return val
	}

	def := ""
	if useDefault {
		def = v1.ENVKeyToDefaults[o.envKey]
	}

	return def
}

// VolumeMarker represents a volume policy or property
// as key:value format
//
// NOTE:
//  A VolumeMarker can be transformed into a Label or into
// a Annotation
//
// NOTE:
//  It will be pointless to argue on whether to transform a
// marker into a Label or an Annotation. It will be
// fair to assume a Label to have actual value for a key
// & Annotation to have a referential value for a key. However,
// regidity in this assumption will not help us. This assumption
// will be challenged when Label & Annotation updates comes
// into picture.
type VolumeMarker struct {
	// Key is the annotation key
	Key string

	// Value is the marker value for a marker key
	Value string

	// IsMultiple flags if there is possibility to have multiple
	// values for a marker key
	IsMultiple bool

	// Values is an array of marker values for a marker key
	Values []string
}

//
func (a VolumeMarker) GetValuesAsCommaSep() string {
	if len(a.Values) == 0 {
		return ""
	}

	return strings.Join(a.Values, ",")
}

// VolumeMarkerBuilder builds all the volume markers
type VolumeMarkerBuilder struct {
	// Items represent all the volume markers
	Items []VolumeMarker
}

// NewVolumeMarkerBuilder will return a new instance of
// VolumeMarkerBuilder
func NewVolumeMarkerBuilder() *VolumeMarkerBuilder {
	return &VolumeMarkerBuilder{}
}

// GetVolumeMarkerBuilder will return a new instance of
// VolumeMarkerBuilder from the provided map
func GetVolumeMarkerBuilder(pairs map[string]string) *VolumeMarkerBuilder {
	var items []VolumeMarker
	for k, v := range pairs {
		m := VolumeMarker{
			Key: k,
		}

		if strings.Contains(v, ",") {
			m.IsMultiple = true
			m.Values = strings.Split(v, ",")
		} else {
			m.Value = v
		}

		items = append(items, m)
	}

	return &VolumeMarkerBuilder{
		Items: items,
	}
}

//
func (p *VolumeMarkerBuilder) AddMarkers(markers []VolumeMarker) {
	p.Items = append(p.Items, markers...)
}

// AddMultiples will add a new volume marker or append to an existing
// volume marker
func (p *VolumeMarkerBuilder) AddMultiples(key, value string, isMul bool) error {
	if len(key) == 0 {
		return fmt.Errorf("Marker key is missing")
	}

	if len(value) == 0 {
		// nil value(s) are possible
		value = string(v1.NilVV)
	}

	var m VolumeMarker
	var isPresent bool
	var mIndex int
	for i, a := range p.Items {
		if a.Key == key {
			if !isMul {
				return fmt.Errorf("Duplicate marker key '%s'", key)
			}
			m = a
			isPresent = true
			mIndex = i
			break
		}
	}

	if !isPresent {
		// create a new marker if not available earlier
		m = VolumeMarker{
			Key:        key,
			IsMultiple: isMul,
		}
	}

	// Append or Set the value based on isMul flag
	if isMul {
		m.Values = append(m.Values, value)
	} else {
		m.Value = value
	}

	if isPresent {
		// update as this is available already
		p.Items[mIndex] = m
	} else {
		// add
		items := append(p.Items, m)
		p.Items = items
	}

	return nil
}

// GetMarker returns the volume marker if available
func (p *VolumeMarkerBuilder) GetMarker(key string) (VolumeMarker, bool) {
	for _, a := range p.Items {
		if a.Key == key {
			return a, true
		}
	}

	return VolumeMarker{}, false
}

// Add will add a new volume marker
func (p *VolumeMarkerBuilder) Add(key, value string) error {
	return p.AddMultiples(key, value, false)
}

// IsPresent flags if a volume marker is already available
func (p *VolumeMarkerBuilder) IsPresent(key string) bool {
	for _, a := range p.Items {
		if a.Key == key {
			return true
		}
	}

	return false
}

// AddControllerIPs will add controller IP address(es) as a volume marker
func (p *VolumeMarkerBuilder) AddControllerIPs(pod k8sApiV1.Pod) {
	ip := pod.Status.PodIP

	// TODO
	// Deprecate
	// backward compatibility
	_ = p.AddMultiples(string(v1.ControllerIPsAPILbl), ip, true)

	// new key representation
	_ = p.AddMultiples(string(v1.JivaControllerIPsVK), ip, true)
}

// AddReplicaIPs will add replica IP address(es) as a volume marker
func (p *VolumeMarkerBuilder) AddReplicaIPs(pod k8sApiV1.Pod) {
	ip := pod.Status.PodIP

	// TODO
	// Deprecate
	// backward compatibility
	_ = p.AddMultiples(string(v1.ReplicaIPsAPILbl), ip, true)

	// new key representation
	_ = p.AddMultiples(string(v1.JivaReplicaIPsVK), ip, true)
}

//
func (p *VolumeMarkerBuilder) AddControllerStatuses(pod k8sApiV1.Pod) {
	status := string(pod.Status.Phase)

	// TODO
	// Deprecate
	// backward compatibility
	_ = p.AddMultiples(string(v1.ControllerStatusAPILbl), status, true)

	// new key representation
	_ = p.AddMultiples(string(v1.JivaControllerStatusVK), status, true)
}

// AddControllerContainerStatus is to fetch state of controller containers
func (p *VolumeMarkerBuilder) AddControllerContainerStatus(cp k8sApiV1.Pod) {

	p.AddContainerStatuses(cp, v1.ControllerContainerStatusVK)
}

// AddReplicaContainerStatus is to fetch state of replica containers
func (p *VolumeMarkerBuilder) AddReplicaContainerStatus(cp k8sApiV1.Pod) {

	p.AddContainerStatuses(cp, v1.ReplicaContainerStatusVK)
}

// AddContainerStatuses is to fetch the current state of containers
// inside the pod
func (p *VolumeMarkerBuilder) AddContainerStatuses(cp k8sApiV1.Pod, volumekey v1.VolumeKey) {
	for _, current := range cp.Status.ContainerStatuses {
		value := v1.NilVV
		if current.State.Waiting != nil {
			value = v1.ContainerWaitingVV
		}

		if current.State.Terminated != nil {
			value = v1.ContainerTerminatedVV
		}

		if current.State.Running != nil {
			if current.Ready {
				value = v1.ContainerRunningVV
			} else {
				value = v1.ContainerNotRunningVV
			}
		}
		_ = p.AddMultiples(string(volumekey), string(value), true)
	}
}

// IsVolumeRunning to compare the state of all containers in a pod
func (p *VolumeMarkerBuilder) IsVolumeRunning(pv *v1.Volume) bool {
	var cphase, rphase v1.VolumeValue
	cstate := pv.Annotations[string(v1.ControllerContainerStatusVK)]
	cresult := strings.Split(cstate, ",")

	for i := range cresult {
		if cresult[i] == string(v1.ContainerRunningVV) {
			cphase = v1.ContainerRunningVV
		} else {
			cphase = v1.ContainerNotRunningVV
		}

	}
	rstate := pv.Annotations[string(v1.ReplicaContainerStatusVK)]
	rresult := strings.Split(rstate, ",")

	for i := range rresult {
		if rresult[i] == string(v1.ContainerRunningVV) {
			rphase = v1.ContainerRunningVV
		} else {
			rphase = v1.ContainerNotRunningVV
		}
	}

	return cphase == v1.ContainerRunningVV && rphase == v1.ContainerRunningVV
}

//
func (p *VolumeMarkerBuilder) AddReplicaStatuses(pod k8sApiV1.Pod) {
	status := string(pod.Status.Phase)

	// TODO
	// Deprecate
	// backward compatibility
	_ = p.AddMultiples(string(v1.ReplicaStatusAPILbl), status, true)

	// new key representation
	_ = p.AddMultiples(string(v1.JivaReplicaStatusVK), status, true)
}

//
func (p *VolumeMarkerBuilder) AddReplicaCount(deploy k8sApisExtnsBeta1.Deployment) {
	count := fmt.Sprint(*deploy.Spec.Replicas)

	// TODO
	// Deprecate
	// backward compatibility
	_ = p.Add(string(v1.ReplicaCountAPILbl), count)

	// new key representation
	_ = p.Add(string(v1.JivaReplicasVK), count)
}

//
func (p *VolumeMarkerBuilder) AddVolumeCapacity(deploy k8sApisExtnsBeta1.Deployment) {
	con := deploy.Spec.Template.Spec.Containers[0]

	var capacity string
	for i, arg := range con.Args {
		if arg == "--size" {
			// since value of capacity is provided after --size
			capacity = con.Args[i+1]
			break
		}
	}

	// TODO
	// Deprecate
	// backward compatibility
	_ = p.Add(string(v1.VolumeSizeAPILbl), capacity)

	// new key representation
	_ = p.Add(string(v1.CapacityVK), capacity)
}

//
func (p *VolumeMarkerBuilder) AddIQN(volumeName string) {
	iqn := string(v1.JivaIqnFormatPrefix) + ":" + volumeName

	// TODO
	// Deprecate
	// backward compatibility
	_ = p.Add(string(v1.IQNAPILbl), iqn)

	// new key representation
	_ = p.Add(string(v1.JivaIQNVK), iqn)
}

//
func (p *VolumeMarkerBuilder) AddControllerClusterIP(svc k8sApiV1.Service) {
	ip := svc.Spec.ClusterIP

	// TODO
	// Deprecate
	// backward compatibility
	_ = p.Add(string(v1.ClusterIPsAPILbl), ip)

	// new key representation
	_ = p.Add(string(v1.JivaControllerClusterIPVK), ip)
}

//
func (p *VolumeMarkerBuilder) AddISCSITargetPortal(svc k8sApiV1.Service) {
	ip := strings.TrimSpace(svc.Spec.ClusterIP)
	ip = ip + ":" + string(v1.JivaISCSIPortDef)

	// TODO
	// Deprecate
	// backward compatibility
	_ = p.Add(string(v1.TargetPortalsAPILbl), ip)

	// new key representation
	_ = p.Add(string(v1.JivaTargetPortalVK), ip)
}

func (p *VolumeMarkerBuilder) AddVolumeType(value string) {
	_ = p.Add(string(v1.VolumeTypeVK), value)
}

//
func (p *VolumeMarkerBuilder) AddStoragePoolPolicy(value string) {
	_ = p.Add(string(v1.StoragePoolVK), value)
}

//
func (p *VolumeMarkerBuilder) AddMonitoringPolicy(value string) {
	_ = p.Add(string(v1.MonitorVK), value)
}

// Build returns the volume markers as a map of strings
func (p *VolumeMarkerBuilder) Build() map[string]string {
	markers := map[string]string{}
	for _, a := range p.Items {
		if a.IsMultiple {
			markers[a.Key] = a.GetValuesAsCommaSep()
		} else {
			markers[a.Key] = a.Value
		}
	}

	return markers
}

// AsAnnotations returns the volume markers as annotations
func (p *VolumeMarkerBuilder) AsAnnotations() map[string]string {
	return p.Build()
}

// AsLabels returns the volume markers as labels
func (p *VolumeMarkerBuilder) AsLabels() map[string]string {
	return p.Build()
}

type LabelK8sObject struct {
	// LabelKey is the label key that will be assigned
	// to the targetted K8s object
	LabelKey string

	// LabelValue is the label value that will be assigned
	// to the targetted K8s Object
	LabelValue string
}

func NewLabelK8sObject(key string, val string) (*LabelK8sObject, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("Key is missing in label")
	}

	if len(val) == 0 {
		return nil, fmt.Errorf("Value is missing in label")
	}

	return &LabelK8sObject{
		LabelKey:   key,
		LabelValue: val,
	}, nil
}

func (l *LabelK8sObject) generate() (string, string) {
	return l.LabelKey, l.LabelValue
}

type MonitoringSideCar struct {
	// TargetIP is the IP Address of the
	// service using which this sidecar will
	// pull metrics info
	TargetIP string

	// Image of this sidecar
	Image string `json:"image,omitempty"`

	// OImage is the wrapper over container image providing
	// utility methods
	OImage *OpenEBSImage

	// SideCar is the K8s type that represents a sidecar
	SideCar k8sApiV1.Container
}

// monSideCarTpl is the K8s type (used as a template) that gets
// built as a sidecar
var monSideCarTpl = k8sApiV1.Container{
	Name:  "maya-volume-exporter",
	Image: v1.DefaultMonitoringImage,
	Command: []string{
		"maya-volume-exporter",
	},
	Args: []string{
		"-c=http://__TARGET_IP__:9501",
	},
	Ports: []k8sApiV1.ContainerPort{
		k8sApiV1.ContainerPort{
			ContainerPort: 9500,
		},
	},
}

//
func NewMonitoringSideCar() *MonitoringSideCar {
	// create a new instance
	return &MonitoringSideCar{
		OImage: NewOpenEBSImage(v1.MonitorImageENVK),
	}
}

// Set sets the MonitoringSideCar instance with appropriate value(s)
//
// NOTE:
//  val can have following values:
//  1/ `true` or `1` or `yes` or `ok` or
//  2/ `image: some_repo/some_image:some_tag`
func (m *MonitoringSideCar) Set(val string) error {
	if util.CheckFalsy(val) {
		return fmt.Errorf("Monitoring is not enabled")
	}

	// truthy value indicates use of defaults in the sidecar
	// Otherwise specific values has been provided to generate the sidecar
	if !util.CheckTruthy(val) {
		// set with provided/specific values
		err := yaml.Unmarshal([]byte(val), m)
		if err != nil {
			return err
		}
	} else {
		// update sidecar's specific property(-ies)
		m.Image = m.OImage.GetImage(true)
	}

	// When deploying as sidecar, the targetIP should be 127.0.0.1
	m.TargetIP = "127.0.0.1"
	m.SideCar = monSideCarTpl

	return nil
}

// Get returns a K8s Container type from the yaml specification
func (m *MonitoringSideCar) Get() (k8sApiV1.Container, error) {

	m.SideCar.Args[0] = strings.Replace(m.SideCar.Args[0], "__TARGET_IP__", m.TargetIP, 1)

	if len(m.Image) != 0 {
		m.SideCar.Image = m.Image
	}

	return m.SideCar, nil
}

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

	// PVCOps provides a PVCInterface that exposes various CRUD operations
	// w.r.t PVC
	PVCOps() (k8sCoreV1.PersistentVolumeClaimInterface, error)
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

	// PVCOps provides all the CRUD & more operations associated
	// w.r.t a Persistent Volume Claim
	PVCOps2() (k8sCoreV1.PersistentVolumeClaimInterface, error)

	// StoragePoolOps provides all the CRUD & more operations associated
	// w.r.t a StoragePool
	//
	// NOTE:
	//  StoragePool is a K8s CRD resource
	StoragePoolOps() (oe_client_v1alpha1.StoragePoolInterface, error)

	// NamespaceOps provides a NamespaceInterface that exposes
	// various CRUD operations w.r.t Namespace
	NamespaceOps() (k8sCoreV1.NamespaceInterface, error)

	// DeploymentOps2 provides all the CRUD operations associated
	// w.r.t a Deployment
	DeploymentOps2() (k8sExtnsV1Beta1.DeploymentInterface, error)
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

	inOECS *versioned.Clientset

	outOECS *versioned.Clientset

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

// DeploymentOps is a utility function that provides a instance capable of
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

// DeploymentOps2 is a utility function that provides a instance capable of
// executing various k8s Deployment related operations.
func (k *k8sUtil) DeploymentOps2() (k8sExtnsV1Beta1.DeploymentInterface, error) {
	cs, err := k.getClientSet()
	if err != nil {
		return nil, err
	}

	// error out if still empty
	if len(k.volume.Namespace) == 0 {
		return nil, fmt.Errorf("Nil namespace")
	}

	return cs.ExtensionsV1beta1().Deployments(k.volume.Namespace), nil
}

//  NamespaceOps provides the NamespaceInterface object that exposes
// various CRUD operations
func (k *k8sUtil) NamespaceOps() (k8sCoreV1.NamespaceInterface, error) {
	cs, err := k.getClientSet()
	if err != nil {
		return nil, err
	}

	return cs.CoreV1().Namespaces(), nil
}

// PVCOps gets a PVCInterface that exposes various CRUD operations
// w.r.t PVC
func (k *k8sUtil) PVCOps() (k8sCoreV1.PersistentVolumeClaimInterface, error) {
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

	return cs.CoreV1().PersistentVolumeClaims(ns), nil
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

func (k *k8sUtil) PVCOps2() (k8sCoreV1.PersistentVolumeClaimInterface, error) {
	cs, err := k.getClientSet()
	if err != nil {
		return nil, err
	}

	// error out if still empty
	if len(k.volume.Namespace) == 0 {
		return nil, fmt.Errorf("Nil namespace")
	}

	return cs.CoreV1().PersistentVolumeClaims(k.volume.Namespace), nil
}

func (k *k8sUtil) StoragePoolOps() (oe_client_v1alpha1.StoragePoolInterface, error) {
	mcs, err := k.getOEClientSet()
	if err != nil {
		return nil, err
	}

	return mcs.OpenebsV1alpha1().StoragePools(), nil
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

// getOEClientSet is used to get a new http client capable
// of invoking OpenEBS CRD APIs.
func (k *k8sUtil) getOEClientSet() (*versioned.Clientset, error) {
	var cs *versioned.Clientset

	// Get if already available in current instance
	if k.inOECS != nil {
		return k.inOECS, nil
	}

	if k.outOECS != nil {
		return k.outOECS, nil
	}

	// Else get it fresh for this instance/http request
	inC, err := k.IsInClusterV2()
	if err != nil {
		return nil, err
	}

	// set based on in-cluster or out-of-cluster
	if inC {
		cs, err = k.getInClusterOECS()
		// set it for future retrievals in same http request
		k.inOECS = cs
	} else {
		cs, err = k.getOutClusterOECS()
		// set it for future retrievals in same http request
		k.outOECS = cs
	}

	if err != nil {
		return nil, err
	}

	return cs, nil
}

func getK8sConfig() (config *rest.Config, err error) {
	k8sMaster := v1.K8sMasterENV()
	kubeConfig := v1.KubeConfigENV()

	if len(k8sMaster) != 0 || len(kubeConfig) != 0 {
		// creates the config from k8sMaster or kubeConfig
		return clientcmd.BuildConfigFromFlags(k8sMaster, kubeConfig)
	}

	// creates the in-cluster config making use of the Pod's ENV & secrets
	return rest.InClusterConfig()
}

// getInClusterCS is used to initialize and return a new http client capable
// of invoking K8s APIs.
func (k *k8sUtil) getInClusterCS() (clientset *kubernetes.Clientset, err error) {
	config, err := getK8sConfig()
	if err != nil {
		return nil, err
	}

	// creates the in-cluster clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// getInClusterOECS is used to initialize and return a new http client capable
// of invoking OpenEBS CRD APIs.
func (k *k8sUtil) getInClusterOECS() (clientset *versioned.Clientset, err error) {
	config, err := getK8sConfig()
	if err != nil {
		return nil, err
	}

	// creates the in-cluster OE clientset
	clientset, err = versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// getOutClusterCS is used to initialize and return a new http client capable
// of invoking outside the cluster K8s APIs.
func (k *k8sUtil) getOutClusterCS() (*kubernetes.Clientset, error) {
	return nil, fmt.Errorf("out cluster clientset not supported in '%s'", k.Name())
}

// getOutClusterOECS is used to initialize and return a new http client capable
// of invoking outside the cluster K8s APIs.
func (k *k8sUtil) getOutClusterOECS() (*versioned.Clientset, error) {
	return nil, fmt.Errorf("out cluster OE clientset not supported in '%s'", k.Name())
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
