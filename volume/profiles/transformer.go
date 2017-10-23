package profiles

import (
	k8sClientV1Beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// Transformer provides common methods for transforming any
// OpenEBS' Kind to any target's Kind
type Transformer interface {
	// Version provides the version of this transformer
	Version() (string, error)
}

// K8sDeployTransformer provides method to transform any
// OpenEBS Kind to corresponding K8s Deployment
type K8sDeployTransformer interface {
	// Transform transforms any OpenEBS Kind to K8s Deployment
	Transform() (*k8sClientV1Beta1.Deployment, error)
}
