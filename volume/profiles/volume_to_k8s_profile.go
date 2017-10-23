package profiles

import (
	"github.com/openebs/maya/types/v1"
	k8sClientV1Beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// VolumeToK8sDeployProfile is a structure with necessary
// properties & methods to convert an OpenEBS Volume kind
// to a K8s Deploy kind
type VolumeToK8sDeployProfile struct {
	// Transformer will transform the OpenEBS Volume Kind
	// to the desired K8s kind
	transformer *VolToK8sDeployTransformer
}

// NewVolumeToK8sDeployProfile provides an instance of VolumeK8sProfile
func NewVolumeToK8sDeployProfile(volType v1.VolumeType, vol *v1.Volume) *VolumeToK8sDeployProfile {
	return &VolumeToK8sDeployProfile{
		transformer: NewVolToK8sDeployTransformer(volType, vol, getVolToK8sDeployTransformers()),
	}
}

// Deployment will make use of transformers to convert an
// OpenEBS Kind to a K8s Deployment
func (p *VolumeToK8sDeployProfile) Deployment() (*k8sClientV1Beta1.Deployment, error) {
	return p.transformer.Transform()
}

func getVolToK8sDeployTransformers() []VolToK8sDeployTransType {
	// TODO Add the transformers !!
	return []VolToK8sDeployTransType{NodeTaintVolToK8sDeploy}
}
