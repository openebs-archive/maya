package profiles

import (
	"github.com/openebs/maya/types/v1"
	k8sClientV1Beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// VolToK8sDeployTransformer is an implementation of
// 1. Transformer interface
// 2. K8sDeployTransformer interface
type VolToK8sDeployTransformer struct {
	// volType is the type of OpenEBS Volume Kind
	volType v1.VolumeType

	// volume is the structure that represents an
	// OpenEBS volume
	volume *v1.Volume

	// deploy is the structure that is generated after
	// transformation of OpenEBS volume
	deploy *k8sClientV1Beta1.Deployment

	// transTypes is a list (i.e. chain) of transformer
	// types to be executed to transform an OpenEBS Volume
	// to corresponding K8s Deployment
	transTypes []VolToK8sDeployTransType
}

// NewVolToK8sDeployTransformer instantiates a new instance of
// VolToK8sDeployTransformer
func NewVolToK8sDeployTransformer(volType v1.VolumeType, volume *v1.Volume, transTypes []VolToK8sDeployTransType) *VolToK8sDeployTransformer {
	return &VolToK8sDeployTransformer{
		volType:    volType,
		volume:     volume,
		transTypes: transTypes,
	}
}

// Version provides the version of this transformer
func (k *VolToK8sDeployTransformer) Version() (string, error) {
	return "1.0", nil
}

// IsVolumeTypeSupported indicates if this transformer can transform
// the OpenEBS Volume Type.
func (k *VolToK8sDeployTransformer) IsVolumeTypeSupported() (bool, error) {

	if k.volType == v1.JivaVolume {
		return true, nil
	} else {
		return false, nil
	}
}

// Transform transforms the OpenEBS Kind into K8s Deployment kind
func (k *VolToK8sDeployTransformer) Transform() (*k8sClientV1Beta1.Deployment, error) {

	for _, tt := range k.transTypes {
		t, err := GetVolToK8sDeployTrans(k.volume, k.deploy, tt)
		if err != nil {
			return nil, err
		}
		d, err := t.Transform()
		if err != nil {
			return nil, err
		}
		k.deploy = d
	}
	return k.deploy, nil
}
