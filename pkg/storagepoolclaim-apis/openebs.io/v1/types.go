package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Rawstorageadaptor describes a Rawstorageadaptor.
type Storagepoolclaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata",omitempty`

	Spec StoragepoolclaimSpec `json:"spec"`
}

// RawstorageadaptorSpec is the spec for a RawstorageadaptorSpec
type StoragepoolclaimSpec struct {
	Name       string `json:"name"`
	Format     string `json:"format"`
	Mountpoint string `json:"mountpoint"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RawstorageadaptorList is a list of RawstorageadaptorList
type StoragepoolclaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Storagepoolclaim `json:"items"`
}
