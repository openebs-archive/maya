package k8s

import (
	"encoding/json"
	//"fmt"
	//"time"

	"github.com/golang/glog"
	//TODO
	//"github.com/metral/memhog-operator/pkg/utils"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	// To authenticate against GKE clusters
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// #############################################################################

/*
 Note: The k8s.io/client-go lib's REST client.Get() and client.List()
 provide a means of accessing & working with a particular resource in the
 cluster.
 However, using client.Get() and client.List() on the REST client can
 become expensive if used multiple times.

 A user can benefit working with the cluster through its client by using
 a local cache store and event watches for better performance.
 This optimization is suggested and can be implemented with an:
 	- Informer: A local cache store & controller for state event handling on a
 	resource, that sync’s with the APIServer's state.
  See https://github.com/kubernetes/client-go/blob/v2.0.0/tools/cache/controller.go#L201-L221
 	- SharedInformer: A single, optimized local cache store & controller for
 	state event handling on multiple resources, syncing all stores &
 	controllers with the APIServer's state.
  See https://github.com/kubernetes/client-go/blob/v2.0.0/tools/cache/shared_informer.go#L31-L39
*/

// #############################################################################

// Configure & create an k8s API REST client for the StorageBackendAdaptor(SBA)
// resource in the k8s cluster.
func newSBAK8sClient(kubecfg *rest.Config, namespace string) (*rest.RESTClient, error) {
	// Update kubecfg to work with the SBA's API group, using the kubecfg
	// param as a baseline.
	addSBAToKubeConfig(kubecfg, CRDDomain, CRDVersionV1)

	// Add SBA's API group to the k8s api.Scheme to provide it with the
	// capability of doing conversions or a deep-copy on an SBA resource.
	addSBAToAPISchema(CRDDomain, CRDVersionV1)

	// Create the k8s API REST client
	client, err := rest.RESTClientFor(kubecfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Configure the attributes for the kubecfg used in the SBA API REST
// client.
func addSBAToKubeConfig(kubecfg *rest.Config, domain, version string) {
	groupversion := schema.GroupVersion{
		Group:   domain,
		Version: version,
	}

	// Set attributes in the kubecfg to reach and work with the
	// Submarine resource.
	kubecfg.GroupVersion = &groupversion
	kubecfg.APIPath = "/apis"
	kubecfg.ContentType = runtime.ContentTypeJSON
	kubecfg.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}
}

// Add the SBA types to the api.Scheme for when needing to do type
// conversions or a deep-copy of an SBA object.
func addSBAToAPISchema(domain, version string) {
	groupversion := schema.GroupVersion{
		Group:   domain,
		Version: version,
	}

	/*
		 Scheme defines methods for serializing and deserializing API objects, a type
		 registry for converting group, version, and kind information to and from Go
		 schemas, and mappings between Go schemas of different versions. A scheme is the
		 foundation for a versioned API and versioned configuration over time.

		 In a Scheme, a Type is a particular Go struct, a Version is a point-in-time
		 identifier for a particular representation of that Type (typically backwards
		 compatible), a Kind is the unique name for that Type within the Version, and a
		 Group identifies a set of Versions, Kinds, and Types that evolve over time. An
		 Unversioned Type is one that is not yet formally bound to a type and is promised
		 to be backwards compatible (effectively a "v1" of a Type that does not expect
		 to break in the future).

		 SchemeBuilder collects functions that add things to a scheme. It's to
		 allow code to compile without explicitly referencing generated types.
		 You should declare one in each package that will have generated deep-copy
		 or conversion functions.

		 Create a schemeBuilder that will ultimately add / register the
		 following types for groupversion into the api.Scheme used when performing
		 a deep-copy of an opject as to not mutate the original object.
			e.g. CopyObjToSubmarine()
	*/
	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			// AddKnownTypes registers all types passed in 'types' as being members of version 'version'.
			// All objects passed to types should be pointers to structs. The name that go reports for
			// the struct becomes the "kind" field when encoding. Version may not be empty - use the
			// APIVersionInternal constant if you have a type that does not have a formal version.
			scheme.AddKnownTypes(
				groupversion,
				&StorageBackendAdaptorSpec{},
				&StorageBackendAdaptorList{},
				&metav1.ListOptions{},
				&metav1.DeleteOptions{},
			)
			return nil
		})

	// AddToScheme applies all the stored functions to the scheme.A non-nil error
	// indicates that one function failed and the attempt was abandoned.
	schemeBuilder.AddToScheme(api.Scheme)
}

// Create a deep-copy of an SBASpec object
func CopyObjToSBA(obj interface{}) (*StorageBackendAdaptorSpec, error) {
	objCopy, err := api.Scheme.Copy(obj.(*StorageBackendAdaptorSpec))
	if err != nil {
		return nil, err
	}

	sba := objCopy.(*StorageBackendAdaptorSpec)
	if sba.Metadata.Annotations == nil {
		sba.Metadata.Annotations = make(map[string]string)
	}
	return sba, nil
}

// Attempt to deep copy an empty interface into an SubmarineList.
func CopyObjToSBAs(obj []interface{}) ([]StorageBackendAdaptorSpec, error) {
	sbas := []StorageBackendAdaptorSpec{}

	for _, o := range obj {
		sba, err := CopyObjToSBA(o)
		if err != nil {
			glog.Errorf("Failed to copy SBA object for sbaList: %v", err)
			return nil, err
		}
		sbas = append(sbas, *sba)
	}

	return sbas, nil
}

// #############################################################################

/*
 Note: The following code is boilerplate code needed to satisfy the
 Submarine as a resource in the cluster in terms of how it expects CRD's to
 be created, operate and used.
*/

// #############################################################################

// Required to satisfy Object interface
func (sba *StorageBackendAdaptorSpec) GetObjectKind() schema.ObjectKind {
	return &sba.TypeMeta
}

// Required to satisfy ObjectMetaAccessor interface
func (sba *StorageBackendAdaptorSpec) GetObjectMeta() metav1.Object {
	return &sba.Metadata
}

// Required to satisfy Object interface
func (sbas *StorageBackendAdaptorList) GetObjectKind() schema.ObjectKind {
	return &sbas.TypeMeta
}

// Required to satisfy ListMetaAccessor interface
func (sbas *StorageBackendAdaptorList) GetListMeta() metav1.List {
	return &sbas.Metadata
}

// #############################################################################

/*
 Note: The following code is used only to work around a known problem
 with third-party resources and ugorji. If/when these issues are resolved,
 the code below should no longer be required.
*/

// #############################################################################

type StorageBackendAdaptorListCopy StorageBackendAdaptorList
type StorageBackendAdaptorSpecCopy StorageBackendAdaptorSpec

func (sub *StorageBackendAdaptorSpec) UnmarshalJSON(data []byte) error {
	tmp := StorageBackendAdaptorSpecCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := StorageBackendAdaptorSpec(tmp)
	*sub = tmp2
	return nil
}

func (subs *StorageBackendAdaptorList) UnmarshalJSON(data []byte) error {
	tmp := StorageBackendAdaptorListCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := StorageBackendAdaptorList(tmp)
	*subs = tmp2
	return nil
}
