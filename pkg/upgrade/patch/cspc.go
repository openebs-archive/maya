package patch

import (
	"strings"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cspc "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

// CSPC ...
type CSPC struct {
	Object *apis.CStorPoolCluster
	Data   []byte
}

// NewCSPC ...
func NewCSPC() *CSPC {
	return &CSPC{}
}

// PreChecks ...
func (c *CSPC) PreChecks(from, to string) error {
	if c.Object == nil {
		return errors.Errorf("nil cspc object")
	}
	version := strings.Split(c.Object.VersionDetails.Status.Current, "-")[0]
	if version != from && version != to {
		return errors.Errorf(
			"cspc version %s is neither %s nor %s",
			version,
			from,
			to,
		)
	}
	return nil
}

// Patch ...
func (c *CSPC) Patch(from, to string) error {
	klog.Info("patching cspc ", c.Object.Name)
	client := cspc.NewKubeClient(cspc.WithKubeConfigPath("/var/run/kubernetes/admin.kubeconfig"))
	version := c.Object.VersionDetails.Desired
	if version == to {
		klog.Infof("cspc already in %s version", to)
		return nil
	}
	if version == from {
		patch := c.Data
		_, err := client.WithNamespace(c.Object.Namespace).Patch(
			c.Object.Name,
			types.MergePatchType,
			[]byte(patch),
		)
		if err != nil {
			return errors.Wrapf(
				err,
				"failed to patch cspc %s",
				c.Object.Name,
			)
		}
		klog.Infof("cspc %s patched", c.Object.Name)
	}
	return nil
}

// Get ...
func (c *CSPC) Get(name, namespace string) error {
	cspc, err := cspc.NewKubeClient(cspc.WithKubeConfigPath("/var/run/kubernetes/admin.kubeconfig")).WithNamespace(namespace).
		Get(name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get cspc %s in %s namespace", name, namespace)
	}
	c.Object = cspc
	return nil
}
