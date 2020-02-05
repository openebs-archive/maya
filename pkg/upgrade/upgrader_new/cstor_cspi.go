package upgrader

import (
	"github.com/openebs/maya/pkg/upgrade/patch"
	"github.com/openebs/maya/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
)

// CSPIPatch is the patch required to upgrade cspi
type CSPIPatch struct {
	*ResourcePatch
	Namespace string
	Deploy    *patch.Deployment
}

// CSPIPatchOptions ...
type CSPIPatchOptions func(*CSPIPatch)

// WithCSPIResorcePatch ...
func WithCSPIResorcePatch(r *ResourcePatch) CSPIPatchOptions {
	return func(obj *CSPIPatch) {
		obj.ResourcePatch = r
	}
}

// WithCSPIDeploy ...
func WithCSPIDeploy(t *patch.Deployment) CSPIPatchOptions {
	return func(obj *CSPIPatch) {
		obj.Deploy = t
	}
}

// NewCSPIPatch ...
func NewCSPIPatch(opts ...CSPIPatchOptions) *CSPIPatch {
	obj := &CSPIPatch{}
	for _, o := range opts {
		o(obj)
	}
	return obj
}

// PreUpgrade ...
func (obj *CSPIPatch) PreUpgrade() error {
	err := obj.Deploy.PreChecks(obj.From, obj.To)
	if err != nil {
		return err
	}
	return nil
}

// DeployUpgrade ...
func (obj *CSPIPatch) DeployUpgrade() error {
	err := obj.Deploy.Patch(obj.From, obj.To)
	if err != nil {
		return err
	}
	return nil
}

// Upgrade execute the steps to upgrade cspi
func (obj *CSPIPatch) Upgrade() error {
	err := obj.Init()
	if err != nil {
		return err
	}
	err = obj.PreUpgrade()
	if err != nil {
		return err
	}
	err = obj.DeployUpgrade()
	return err
}

// Init initializes all the fields of the CSPIPatch
func (obj *CSPIPatch) Init() error {
	obj.Deploy = patch.NewDeployment()
	obj.Namespace = obj.OpenebsNamespace
	label := "openebs.io/cstor-pool-instance=" + obj.Name
	err := obj.Deploy.Get(label, obj.Namespace)
	if err != nil {
		return err
	}
	err = getCSPIDeployPatchData(obj)
	return err
}

func getCSPIDeployPatchData(obobj *CSPIPatch) error {
	newDeploy := obobj.Deploy.Object.DeepCopy()
	err := transformCSPIDeploy(newDeploy, obobj.ResourcePatch)
	if err != nil {
		return err
	}
	obobj.Deploy.Data, err = util.GetPatchData(obobj.Deploy.Object, newDeploy)
	return err
}

func transformCSPIDeploy(d *appsv1.Deployment, res *ResourcePatch) error {
	// update deployment images
	tag := res.To
	if res.ImageTag != "" {
		tag = res.ImageTag
	}
	cons := len(d.Spec.Template.Spec.Containers)
	for i := 0; i < cons; i++ {
		url, err := getImageURL(
			d.Spec.Template.Spec.Containers[i].Image,
			res.BaseURL,
		)
		if err != nil {
			return err
		}
		d.Spec.Template.Spec.Containers[i].Image = url + ":" + tag
	}
	d.Labels["openebs.io/version"] = res.To
	d.Spec.Template.Labels["openebs.io/version"] = res.To
	return nil
}
