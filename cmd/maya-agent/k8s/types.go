package k8s

import (
	magentyptesv1 "github.com/openebs/maya/cmd/maya-agent/types/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// #############################################################################

// Note: The following code is boilerplate code needed to satisfy the
// StorageBackendAdaptor as a resource in the cluster in terms of how it
// expects CRD's to be created, operate and used.

// #############################################################################


// StorageBackendAdaptorSpec is a specification for a storage backend adaptor resource
type StorageBackendAdaptorSpec struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ObjectMeta `json:"metadata"`

	Spec magentyptesv1.StorageBackendAdaptor `json:"spec"`
}

// StorageBackendAdaptorList struct is a list of storage backend adaptor 
type StorageBackendAdaptorList struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ListMeta `json:"metadata"`

	Items []StorageBackendAdaptorSpec `json:"items"`
}
