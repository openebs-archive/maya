package v1

// StorageBackendAdaptor contains the generic options that apply to all the
// StorageBackendAdaptors (K8s CRD yaml)
type StorageBackendAdaptor struct {
	NodeFilter        string `json:"nodeFilter"`
	MaxAllocationSize string `json:"maxAllocationSize"`
}

// HostDirSBA contains the generic options that apply to all the
// HostDirStorageBackendAdaptors (K8s CRD yaml)
type HostDirSBA struct {
	StorageBackendAdaptor
	HostDir string `json:"hostDir"`
}
