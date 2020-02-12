package upgrader

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cspi "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	"github.com/openebs/maya/pkg/upgrade/patch"
	"github.com/openebs/maya/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CSPCPatch is the patch required to upgrade CSPC
type CSPCPatch struct {
	*ResourcePatch
	Namespace string
	CSPC      *patch.CSPC
}

// CSPCPatchOptions ...
type CSPCPatchOptions func(*CSPCPatch)

// WithCSPCResorcePatch ...
func WithCSPCResorcePatch(r *ResourcePatch) CSPCPatchOptions {
	return func(obj *CSPCPatch) {
		obj.ResourcePatch = r
	}
}

// NewCSPCPatch ...
func NewCSPCPatch(opts ...CSPCPatchOptions) *CSPCPatch {
	obj := &CSPCPatch{}
	for _, o := range opts {
		o(obj)
	}
	return obj
}

// PreUpgrade ...
func (obj *CSPCPatch) PreUpgrade() error {
	err := obj.CSPC.PreChecks(obj.From, obj.To)
	return err
}

// Init initializes all the fields of the CSPCPatch
func (obj *CSPCPatch) Init() error {
	obj.Namespace = obj.OpenebsNamespace
	obj.CSPC = patch.NewCSPC()
	err := obj.CSPC.Get(obj.Name, obj.Namespace)
	if err != nil {
		return err
	}
	err = getCSPCPatchData(obj)
	return err
}

func getCSPCPatchData(obj *CSPCPatch) error {
	newCSPC := obj.CSPC.Object.DeepCopy()
	err := transformCSPC(newCSPC, obj.ResourcePatch)
	if err != nil {
		return err
	}
	obj.CSPC.Data, err = util.GetPatchData(obj.CSPC.Object, newCSPC)
	return err
}

func transformCSPC(c *apis.CStorPoolCluster, res *ResourcePatch) error {
	c.VersionDetails.Desired = res.To
	return nil
}

// CSPCUpgrade ...
func (obj *CSPCPatch) CSPCUpgrade() error {
	err := obj.CSPC.Patch(obj.From, obj.To)
	if err != nil {
		return err
	}
	return nil
}

// Upgrade execute the steps to upgrade CSPC
func (obj *CSPCPatch) Upgrade() error {
	err := obj.Init()
	if err != nil {
		return err
	}
	err = obj.PreUpgrade()
	if err != nil {
		return err
	}
	res := *obj.ResourcePatch
	cspiList, err := cspi.NewKubeClient(cspi.WithKubeConfigPath("/var/run/kubernetes/admin.kubeconfig")).List(
		metav1.ListOptions{
			LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + obj.Name,
		},
	)
	if err != nil {
		return err
	}
	for _, cspiObj := range cspiList.Items {
		res.Name = cspiObj.Name
		dependant := NewCSPIPatch(
			WithCSPIResorcePatch(&res),
		)
		err = dependant.Upgrade()
		if err != nil {
			return err
		}
	}
	err = obj.CSPCUpgrade()
	return err
}
