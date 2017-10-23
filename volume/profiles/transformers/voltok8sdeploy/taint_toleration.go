package voltok8sdeploy

import (
	"github.com/openebs/maya/types/v1"
	k8sClientApiV1 "k8s.io/client-go/pkg/api/v1"
	k8sClientV1Beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// TaintTolerationTransformer is an implementation of
// 1. Transformer interface
// 2. K8sDeployTransformer interface
type TaintTolerationTransformer struct {
  // name is the name of this transformer
  name string

	// volume is the structure that represents an
	// OpenEBS volume
	volSpec v1.VolumeSpec

	// deploy is the structure that is created/updated after
	// transformation of node taint property
	deploy *k8sClientV1Beta1.Deployment
}

// NewTaintTolerationTransformer instantiates a new instance of
// NodeTaintTransformer
func NewTaintTolerationTransformer(volSpec v1.VolumeSpec, deploy *k8sClientV1Beta1.Deployment) *TaintTolerationTransformer {
  
	return &TaintTolerationTransformer{
	  name: "voltok8sdeploy/TaintToleration",
		volSpec: volSpec,
		deploy: deploy,
	}
}

// Version provides the version of this transformer
func (k *TaintTolerationTransformer) Version() (string, error) {
	return "1.0", nil
}

// Transform transforms the OpenEBS Volume into K8s Deployment
func (k *TaintTolerationTransformer) Transform() (*k8sClientV1Beta1.Deployment, error) {

  ver, err := k.Version()
  if err != nil {
    return nil, err
  }
  
  if strings.TrimSpace(k.volSpec.TaintToleration.Version) == "" {
    k.volSpec.TaintToleration.Version = profiles.TaintTolerationVolToK8sDeployVer
  }
  
  if ver != k.volSpec.TaintToleration.Version {
    // Skip the transformation minus errors, as this transformer is not
    // supposed to handle transformation
    return k.deploy, nil
  }

  if k.volSpec == nil {
    return nil, fmt.Errorf("Nil VolumeSpec provided to '%s: %s'", k.name, ver)
  }

  if k.deploy == nil {
    k.deploy = &k8sClientV1Beta1.Deployment{}
  }
  
 	// check if taint toleration needs to be set ?
	nTTs, reqd, err := k.isTaintTolerations()
	if err != nil {
		return nil, err
	}

	if reqd {		  
		err = addTaintTolerations(nTTs, k.deploy)
		if err != nil {
			return nil, err
		}
	}
  
  return k.deploy, nil
}

// isTaintTolerations indicates if taint toleration(s) is required
func (k *TaintTolerationTransformer) isTaintTolerations() ([]string, bool, error) {
	// Extract the taint toleration for controller
	nTTs, err := k.getTaintTolerations()
	if err != nil {
		return nil, false, err
	}

	if strings.TrimSpace(nTTs) == "" {
		return nil, false, nil
	}

	// nTTs is expected of below form
	// key=value:effect, key1=value1:effect1
	// __or__
	// key=value:effect
	return strings.Split(nTTs, ","), true, nil
}

// addTaintTolerations updates the Taint Toleration property
// against the provided K8s Deployment
func addTaintTolerations(taintTolerations []string, deploy *k8sClientV1Beta1.Deployment) error {

	// nTT is expected to be in key=value:effect format
	for _, tt := range taintTolerations {
		kveArr := strings.Split(tt, ":")
		if len(kveArr) != 2 {
			return fmt.Errorf("Invalid args '%s' provided for taint toleration", tt)
		}

		kv := kveArr[0]
		effect := strings.TrimSpace(kveArr[1])

		kvArr := strings.Split(kv, "=")
		if len(kvArr) != 2 {
			return fmt.Errorf("Invalid kv '%s' provided for taint toleration", kv)
		}
		k := strings.TrimSpace(kvArr[0])
		v := strings.TrimSpace(kvArr[1])

		// Setting to blank to validate later
		e := k8sApiV1.TaintEffect("")

		// Supports only these two effects
		if string(k8sClientApiV1.TaintEffectNoExecute) == effect {
			e = k8sClientApiV1.TaintEffectNoExecute
		} else if string(k8sClientApiV1.TaintEffectNoSchedule) == effect {
			e = k8sClientApiV1.TaintEffectNoSchedule
		}

		if string(e) == "" {
			return fmt.Errorf("Invalid effect '%s' provided for taint toleration", effect)
		}

		toleration := k8sApiV1.Toleration{
			Key:      k,
			Operator: k8sApiV1.TolerationOpEqual,
			Value:    v,
			Effect:   e,
		}

		tls := append(deploy.Spec.Template.Spec.Tolerations, toleration)
		deploy.Spec.Template.Spec.Tolerations = tls
	}

	return nil
}

// getTaintTolerations gets the taint tolerations if available
func (k *TaintTolerationTransformer) getTaintTolerations() (string, error) {
	val, err := taintTolerations(volSpec)
	if err != nil {
		return "", err
	}

	if val == "" {
		val, err = k.defaultTaintTolerations()
	}

	return val, err
}

// taintTolerations extracts the node taint tolerations
func (k *TaintTolerationTransformer) taintTolerations() (string, error) {
	val := k.volSpec.TaintToleration.KVEPairs

	if val != "" {
		return val, nil
	}

  val = strings.Split(v1.OSGetEnv(string(TaintTolerationEnv)), volSpec.Context),",")

	// else get from environment variable
	return , nil
}

// defaultTaintTolerations will fetch the default value for node
// taint tolerations
func (k *TaintTolerationTransformer) defaultTaintTolerations() (string, error) {
	// Controller node taint toleration property is optional. Hence returns blank
	// (i.e. not required) as default.
	return "", nil
}
