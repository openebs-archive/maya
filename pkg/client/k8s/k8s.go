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
	"encoding/json"

	openebs "github.com/openebs/maya/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	api_oe_v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	api_oe_old "github.com/openebs/maya/types/v1"
	api_apps_v1beta1 "k8s.io/api/apps/v1beta1"
	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
	api_storage_v1 "k8s.io/api/storage/v1"

	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	typed_oe_v1alpha1 "github.com/openebs/maya/pkg/client/clientset/versioned/typed/openebs/v1alpha1"
	typed_apps_v1beta1 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	typed_core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	typed_ext_v1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	typed_storage_v1 "k8s.io/client-go/kubernetes/typed/storage/v1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sKind represents the Kinds understood by Kubernetes
type K8sKind string

const (
	// PodKK is a K8s Pod Kind
	PodKK K8sKind = "Pod"
	// DeploymentKK is a K8s Deployment Kind
	DeploymentKK K8sKind = "Deployment"
	// ConfigMapKK is a K8s ConfigMap Kind
	ConfigMapKK K8sKind = "ConfigMap"
	// ServiceKK is a K8s Service Kind
	ServiceKK K8sKind = "Service"
	// CRDKK is a K8s CustomResourceDefinition Kind
	CRDKK K8sKind = "CustomResourceDefinition"
	// StroagePoolCRKK is a K8s CR of kind StoragePool
	StroagePoolCRKK K8sKind = "StoragePool"
	// PersistentVolumeClaimKK is a K8s PersistentVolumeClaim Kind
	PersistentVolumeClaimKK K8sKind = "PersistentVolumeClaim"
)

//
type K8sAPIVersion string

const (
	ExtensionsV1Beta1KA K8sAPIVersion = "extensions/v1beta1"

	AppsV1B1KA K8sAPIVersion = "apps/v1beta1"

	CoreV1KA K8sAPIVersion = "v1"

	OEV1alpha1KA K8sAPIVersion = "openebs.io/v1alpha1"
)

// K8sClient provides the necessary utility to operate over
// various K8s Kind objects
type K8sClient struct {
	// ns refers to K8s namespace where the operation
	// will be performed
	ns string

	// cs refers to the Clientset capable of communicating
	// within the current K8s cluster
	cs *kubernetes.Clientset

	// oecs refers to the Clientset capable of communicating
	// within the current K8s cluster for OpenEBS objects
	oecs *openebs.Clientset

	// PVC refers to a K8s PersistentVolumeClaim object
	// NOTE: This property enables unit testing
	PVC *api_core_v1.PersistentVolumeClaim

	// Pod refers to a K8s Pod object
	// NOTE: This property enables unit testing
	Pod *api_core_v1.Pod

	// Service refers to a K8s Service object
	// NOTE: This property enables unit testing
	Service *api_core_v1.Service

	// ConfigMap refers to a K8s Service object
	// NOTE: This property enables unit testing
	ConfigMap *api_core_v1.ConfigMap

	// Deployment refers to a K8s Deployment object
	// NOTE: This property enables unit testing
	Deployment *api_extn_v1beta1.Deployment

	// StorageClass refers to a K8s StorageClass object
	// NOTE: This property is useful to mock
	// during unit testing
	StorageClass *api_storage_v1.StorageClass

	// StoragePool refers to a K8s StoragePool CRD object
	// NOTE: This property is useful to mock
	// during unit testing
	StoragePool *api_oe_v1alpha1.StoragePool

	// CASTemplate refers to a K8s CASTemplate custom resource
	// NOTE: This property is useful to mock
	// during unit testing
	CASTemplate *api_oe_v1alpha1.CASTemplate

	// various cert related to connecting to K8s API
	caCert     string
	caPath     string
	clientCert string
	clientKey  string
	insecure   bool
}

func NewK8sClient(ns string) (*K8sClient, error) {
	// get the appropriate clientset
	cs, err := getInClusterCS()
	if err != nil {
		return nil, err
	}

	// get the appropriate openebs clientset
	oecs, err := getInClusterOECS()
	if err != nil {
		return nil, err
	}

	return &K8sClient{
		ns:   ns,
		cs:   cs,
		oecs: oecs,
	}, nil
}

// scOps is a utility function that provides a instance capable of
// executing various K8s StorageClass related operations
func (k *K8sClient) storageV1SCOps() typed_storage_v1.StorageClassInterface {
	return k.cs.StorageV1().StorageClasses()
}

// GetStorageV1SC fetches the K8s StorageClass specs based on
// the provided name
func (k *K8sClient) GetStorageV1SC(name string, opts mach_apis_meta_v1.GetOptions) (*api_storage_v1.StorageClass, error) {
	if k.StorageClass != nil {
		return k.StorageClass, nil
	}

	scops := k.storageV1SCOps()
	return scops.Get(name, opts)
}

// oeV1alpha1SPOps is a utility function that provides a instance capable of
// executing various OpenEBS StoragePool related operations
func (k *K8sClient) oeV1alpha1SPOps() typed_oe_v1alpha1.StoragePoolInterface {
	return k.oecs.OpenebsV1alpha1().StoragePools()
}

// GetOEV1alpha1SP fetches the OpenEBS StoragePool specs based on
// the provided name
func (k *K8sClient) GetOEV1alpha1SP(name string) (*api_oe_v1alpha1.StoragePool, error) {
	if k.StoragePool != nil {
		return k.StoragePool, nil
	}

	spOps := k.oeV1alpha1SPOps()
	return spOps.Get(name, mach_apis_meta_v1.GetOptions{})
}

// oeV1alpha1CASTOps is a utility function that provides a instance capable of
// executing various OpenEBS CASTemplate related operations
func (k *K8sClient) oeV1alpha1CASTOps() typed_oe_v1alpha1.CASTemplateInterface {
	return k.oecs.OpenebsV1alpha1().CASTemplates()
}

// GetOEV1alpha1CAST fetches the OpenEBS CASTemplate specs based on
// the provided name
func (k *K8sClient) GetOEV1alpha1CAST(name string, opts mach_apis_meta_v1.GetOptions) (*api_oe_v1alpha1.CASTemplate, error) {
	if k.CASTemplate != nil {
		return k.CASTemplate, nil
	}

	castOps := k.oeV1alpha1CASTOps()
	return castOps.Get(name, opts)
}

// cmOps is a utility function that provides a instance capable of
// executing various K8s ConfigMap related operations.
func (k *K8sClient) cmOps() typed_core_v1.ConfigMapInterface {
	return k.cs.CoreV1().ConfigMaps(k.ns)
}

// GetConfigMap fetches the K8s ConfigMap with the provided name
func (k *K8sClient) GetConfigMap(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.ConfigMap, error) {
	if k.ConfigMap != nil {
		return k.ConfigMap, nil
	}

	cops := k.cmOps()
	return cops.Get(name, opts)
}

// coreV1PVCOps is a utility function that provides a instance capable of
// executing various K8s PVC related operations.
func (k *K8sClient) coreV1PVCOps() typed_core_v1.PersistentVolumeClaimInterface {
	return k.cs.CoreV1().PersistentVolumeClaims(k.ns)
}

// GetPVC fetches the K8s PVC with the provided name
func (k *K8sClient) GetPVC(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.PersistentVolumeClaim, error) {
	if k.PVC != nil {
		return k.PVC, nil
	}

	pops := k.coreV1PVCOps()
	return pops.Get(name, opts)
}

// GetCoreV1PVCAsRaw fetches the K8s PVC with the provided name
func (k *K8sClient) GetCoreV1PVCAsRaw(name string) (result []byte, err error) {
	result, err = k.cs.CoreV1().RESTClient().
		Get().
		Namespace(k.ns).
		Resource("persistentvolumeclaims").
		Name(name).
		VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
		DoRaw()

	return
}

// GetExtnV1B1DeploymentAsRaw fetches the K8s Deployment with the provided name
func (k *K8sClient) GetExtnV1B1DeploymentAsRaw(name string) (result []byte, err error) {
	result, err = k.cs.ExtensionsV1beta1().RESTClient().
		Get().
		Namespace(k.ns).
		Resource("deployments").
		Name(name).
		VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
		DoRaw()

	return
}

// GetAppsV1B1DeploymentAsRaw fetches the K8s Deployment with the provided name
func (k *K8sClient) GetAppsV1B1DeploymentAsRaw(name string) (result []byte, err error) {
	result, err = k.cs.AppsV1beta1().RESTClient().
		Get().
		Namespace(k.ns).
		Resource("deployments").
		Name(name).
		VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
		DoRaw()

	return
}

// GetOEV1alpha1SPAsRaw fetches the OpenEBS SP with the provided name
func (k *K8sClient) GetOEV1alpha1SPAsRaw(name string) (result []byte, err error) {
	sp, err := k.GetOEV1alpha1SP(name)
	if err != nil {
		return
	}

	return json.Marshal(sp)

	// TODO
	//  A better way needs to be determined to get or use raw bytes of a resource.
	// These lines will be removed or refactor-ed once we conclude on this better
	// approach.
	//
	//result, err = k.oecs.OpenebsV1alpha1().RESTClient().
	//	Get().
	//	Namespace(k.ns).
	//	Resource("storagepools").
	//	Name(name).
	//	VersionedParams(&mach_apis_meta_v1.GetOptions{}, scheme.ParameterCodec).
	//	DoRaw()

	//return
}

// podOps is a utility function that provides a instance capable of
// executing various K8s pod related operations.
func (k *K8sClient) podOps() typed_core_v1.PodInterface {
	return k.cs.CoreV1().Pods(k.ns)
}

// GetPod fetches the K8s Pod with the provided name
func (k *K8sClient) GetPod(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.Pod, error) {
	if k.Pod != nil {
		return k.Pod, nil
	}

	pops := k.podOps()
	return pops.Get(name, opts)
}

// GetPods fetches the K8s Pods
func (k *K8sClient) GetPods() ([]api_core_v1.Pod, error) {
	podLists, err := k.cs.Core().Pods(k.ns).List(mach_apis_meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return podLists.Items, nil
}

// ListCoreV1PodAsRaw fetches a list of K8s Pods with the provided options
func (k *K8sClient) ListCoreV1PodAsRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	result, err = k.cs.CoreV1().RESTClient().Get().
		Namespace(k.ns).
		Resource("pods").
		VersionedParams(&opts, scheme.ParameterCodec).
		DoRaw()

	return
}

// ListCoreV1ServiceAsRaw fetches a list of K8s Services as per the provided options
func (k *K8sClient) ListCoreV1ServiceAsRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	result, err = k.cs.CoreV1().RESTClient().Get().
		Namespace(k.ns).
		Resource("services").
		VersionedParams(&opts, scheme.ParameterCodec).
		DoRaw()

	return
}

// ListExtnV1B1DeploymentAsRaw fetches a list of K8s Deployments as per the
// provided options
func (k *K8sClient) ListExtnV1B1DeploymentAsRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	result, err = k.cs.ExtensionsV1beta1().RESTClient().Get().
		Namespace(k.ns).
		Resource("deployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		DoRaw()

	return
}

// ListAppsV1B1DeploymentAsRaw fetches a list of K8s Deployments as per the
// provided options
func (k *K8sClient) ListAppsV1B1DeploymentAsRaw(opts mach_apis_meta_v1.ListOptions) (result []byte, err error) {
	result, err = k.cs.AppsV1beta1().RESTClient().Get().
		Namespace(k.ns).
		Resource("deployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		DoRaw()

	return
}

// serviceOps is a utility function that provides a instance capable of
// executing various k8s service related operations.
func (k *K8sClient) serviceOps() typed_core_v1.ServiceInterface {
	return k.cs.CoreV1().Services(k.ns)
}

// GetService fetches the K8s Service with the provided name
func (k *K8sClient) GetService(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.Service, error) {
	if k.Service != nil {
		return k.Service, nil
	}

	sops := k.serviceOps()
	return sops.Get(name, opts)
}

// coreV1ServiceOps is a utility function that provides a instance capable of
// executing various k8s service related operations.
func (k *K8sClient) coreV1ServiceOps() typed_core_v1.ServiceInterface {
	return k.cs.CoreV1().Services(k.ns)
}

// CreateCoreV1Service creates a K8s Service
func (k *K8sClient) CreateCoreV1Service(svc *api_core_v1.Service) (*api_core_v1.Service, error) {
	sops := k.coreV1ServiceOps()
	return sops.Create(svc)
}

// DeleteCoreV1Service deletes a K8s Service
func (k *K8sClient) DeleteCoreV1Service(name string) error {
	sops := k.coreV1ServiceOps()
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return sops.Delete(name, &mach_apis_meta_v1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	})
}

// TODO deprecate
//
// deploymentOps is a utility function that provides a instance capable of
// executing various k8s Deployment related operations.
func (k *K8sClient) deploymentOps() typed_ext_v1beta1.DeploymentInterface {
	return k.cs.ExtensionsV1beta1().Deployments(k.ns)
}

// extnV1B1DeploymentOps is a utility function that provides a instance capable of
// executing various k8s Deployment related operations.
func (k *K8sClient) extnV1B1DeploymentOps() typed_ext_v1beta1.DeploymentInterface {
	return k.cs.ExtensionsV1beta1().Deployments(k.ns)
}

// GetDeployment fetches the K8s Deployment with the provided name
func (k *K8sClient) GetDeployment(name string, opts mach_apis_meta_v1.GetOptions) (*api_extn_v1beta1.Deployment, error) {
	if k.Deployment != nil {
		return k.Deployment, nil
	}

	dops := k.deploymentOps()
	return dops.Get(name, opts)
}

// CreateExtnV1B1Deployment creates a K8s Deployment
func (k *K8sClient) CreateExtnV1B1Deployment(d *api_extn_v1beta1.Deployment) (*api_extn_v1beta1.Deployment, error) {
	dops := k.extnV1B1DeploymentOps()
	return dops.Create(d)
}

// CreateExtnV1B1DeploymentAsRaw creates a K8s Deployment
func (k *K8sClient) CreateExtnV1B1DeploymentAsRaw(d *api_extn_v1beta1.Deployment) (result []byte, err error) {
	deploy, err := k.CreateExtnV1B1Deployment(d)
	if err != nil {
		return
	}

	return json.Marshal(deploy)

	// TODO
	//  A better way needs to be determined to get or use raw bytes of a resource.
	// These lines will be removed or refactor-ed once we conclude on this better
	// approach.
	//
	//result, err = k.cs.ExtensionsV1beta1().RESTClient().
	//	Put().
	//	Namespace(k.ns).
	//	Resource("deployments").
	//	Body(d).
	//	DoRaw()

	//return
}

// CreateAppsV1B1DeploymentAsRaw creates a K8s Deployment
func (k *K8sClient) CreateAppsV1B1DeploymentAsRaw(d *api_apps_v1beta1.Deployment) (result []byte, err error) {
	deploy, err := k.CreateAppsV1B1Deployment(d)
	if err != nil {
		return
	}

	return json.Marshal(deploy)

	// TODO
	//  A better way needs to be determined to get or use raw bytes of a resource.
	// These lines will be removed or refactor-ed once we conclude on this better
	// approach.
	//
	//result, err = k.cs.AppsV1beta1().RESTClient().
	//	Put().
	//	Namespace(k.ns).
	//	Resource("deployments").
	//	Body(d).
	//	DoRaw()

	//return
}

// CreateCoreV1ServiceAsRaw creates a K8s Service
func (k *K8sClient) CreateCoreV1ServiceAsRaw(s *api_core_v1.Service) (result []byte, err error) {
	svc, err := k.CreateCoreV1Service(s)
	if err != nil {
		return
	}

	return json.Marshal(svc)

	// TODO
	//  A better way needs to be determined to get or use raw bytes of a resource.
	// These lines will be removed or refactor-ed once we conclude on this better
	// approach.
	//
	//result, err = k.cs.CoreV1().RESTClient().
	//	Put().
	//	Namespace(k.ns).
	//	Resource("services").
	//	Body(s).
	//	DoRaw()

	//return
}

// PatchExtnV1B1Deployment patches the K8s Deployment with the provided patches
func (k *K8sClient) PatchExtnV1B1Deployment(name string, patchType types.PatchType, patches []byte) (*api_extn_v1beta1.Deployment, error) {
	dops := k.extnV1B1DeploymentOps()
	return dops.Patch(name, patchType, patches)
}

// PatchExtnV1B1Deployment patches the K8s Deployment with the provided patches
func (k *K8sClient) PatchExtnV1B1DeploymentAsRaw(name string, patchType types.PatchType, patches []byte) (result []byte, err error) {
	result, err = k.cs.ExtensionsV1beta1().RESTClient().Patch(patchType).
		Namespace(k.ns).
		Resource("deployments").
		Name(name).
		Body(patches).
		DoRaw()

	return
}

// DeleteExtnV1B1Deployment deletes the K8s Deployment with the provided name
func (k *K8sClient) DeleteExtnV1B1Deployment(name string) error {
	dops := k.extnV1B1DeploymentOps()
	// ensure all the dependants are deleted
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return dops.Delete(name, &mach_apis_meta_v1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	})
}

// appsV1B1DeploymentOps is a utility function that provides a instance capable of
// executing various k8s Deployment related operations.
func (k *K8sClient) appsV1B1DeploymentOps() typed_apps_v1beta1.DeploymentInterface {
	return k.cs.AppsV1beta1().Deployments(k.ns)
}

// GetAppsV1B1Deployment fetches the K8s Deployment with the provided name
func (k *K8sClient) GetAppsV1B1Deployment(name string, opts mach_apis_meta_v1.GetOptions) (*api_apps_v1beta1.Deployment, error) {
	dops := k.appsV1B1DeploymentOps()
	return dops.Get(name, opts)
}

// CreateAppsV1B1Deployment creates the K8s Deployment with the provided name
func (k *K8sClient) CreateAppsV1B1Deployment(d *api_apps_v1beta1.Deployment) (*api_apps_v1beta1.Deployment, error) {
	dops := k.appsV1B1DeploymentOps()
	return dops.Create(d)
}

// DeleteAppsV1B1Deployment deletes the K8s Deployment with the provided name
func (k *K8sClient) DeleteAppsV1B1Deployment(name string) error {
	dops := k.appsV1B1DeploymentOps()
	// ensure all the dependants are deleted
	deletePropagation := mach_apis_meta_v1.DeletePropagationForeground
	return dops.Delete(name, &mach_apis_meta_v1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	})
}

func getK8sConfig() (config *rest.Config, err error) {
	k8sMaster := api_oe_old.K8sMasterENV()
	kubeConfig := api_oe_old.KubeConfigENV()

	if len(k8sMaster) != 0 || len(kubeConfig) != 0 {
		// creates the config from k8sMaster or kubeConfig
		return clientcmd.BuildConfigFromFlags(k8sMaster, kubeConfig)
	}

	// creates the in-cluster config making use of the Pod's ENV & secrets
	return rest.InClusterConfig()
}

// getInClusterCS is used to initialize and return a new http client capable
// of invoking K8s APIs within the cluster
func getInClusterCS() (clientset *kubernetes.Clientset, err error) {
	config, err := getK8sConfig()
	if err != nil {
		return nil, err
	}

	// creates the in-cluster kubernetes clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// getInClusterOECS is used to initialize and return a new http client capable
// of invoking OpenEBS CRD APIs within the cluster
func getInClusterOECS() (clientset *openebs.Clientset, err error) {
	config, err := getK8sConfig()
	if err != nil {
		return nil, err
	}

	// creates the in-cluster openebs clientset
	clientset, err = openebs.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
