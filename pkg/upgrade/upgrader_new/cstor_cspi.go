/*
Copyright 2019 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package upgrader

import (
	// apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"time"

	apis "github.com/openebs/api/pkg/apis/cstor/v1"
	"github.com/openebs/maya/pkg/upgrade/patch"
	"github.com/openebs/maya/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/klog"
)

// CSPIPatch is the patch required to upgrade cspi
type CSPIPatch struct {
	*ResourcePatch
	Namespace string
	Deploy    *patch.Deployment
	CSPI      *patch.CSPI
	*Client
}

// CSPIPatchOptions ...
type CSPIPatchOptions func(*CSPIPatch)

// WithCSPIResorcePatch ...
func WithCSPIResorcePatch(r *ResourcePatch) CSPIPatchOptions {
	return func(obj *CSPIPatch) {
		obj.ResourcePatch = r
	}
}

// WithCSPIClient ...
func WithCSPIClient(c *Client) CSPIPatchOptions {
	return func(obj *CSPIPatch) {
		obj.Client = c
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
	err = obj.CSPI.PreChecks(obj.From, obj.To)
	return err
}

// DeployUpgrade ...
func (obj *CSPIPatch) DeployUpgrade() error {
	err := obj.Deploy.Patch(obj.From, obj.To)
	if err != nil {
		return err
	}
	return nil
}

// CSPIUpgrade ...
func (obj *CSPIPatch) CSPIUpgrade() error {
	err := obj.CSPI.Patch(obj.From, obj.To)
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
	if err != nil {
		return err
	}
	err = obj.CSPIUpgrade()
	if err != nil {
		return err
	}
	err = obj.verifyCSPIVersionReconcile()
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
	obj.CSPI = patch.NewCSPI(
		patch.WithCSPIClient(obj.OpenebsClientset),
	)
	err = obj.CSPI.Get(obj.Name, obj.Namespace)
	if err != nil {
		return err
	}
	err = getCSPIDeployPatchData(obj)
	if err != nil {
		return err
	}
	err = getCSPIPatchData(obj)
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

func getCSPIPatchData(obj *CSPIPatch) error {
	newCSPI := obj.CSPI.Object.DeepCopy()
	err := transformCSPI(newCSPI, obj.ResourcePatch)
	if err != nil {
		return err
	}
	obj.CSPI.Data, err = util.GetPatchData(obj.CSPI.Object, newCSPI)
	return err
}

func transformCSPI(c *apis.CStorPoolInstance, res *ResourcePatch) error {
	c.Labels["openebs.io/version"] = res.To
	c.VersionDetails.Desired = res.To
	return nil
}

func (obj *CSPIPatch) verifyCSPIVersionReconcile() error {
	// get the latest cspi object
	err := obj.CSPI.Get(obj.Name, obj.Namespace)
	if err != nil {
		return err
	}
	// waiting for the current version to be equal to desired version
	for obj.CSPI.Object.VersionDetails.Status.Current != obj.To {
		klog.Infof("Verifying the reconciliation of version for %s", obj.CSPI.Object.Name)
		// Sleep equal to the default sync time
		time.Sleep(10 * time.Second)
		err = obj.CSPI.Get(obj.Name, obj.Namespace)
		if err != nil {
			return err
		}
		if obj.CSPI.Object.VersionDetails.Status.Message != "" {
			klog.Errorf("failed to reconcile: %s", obj.CSPI.Object.VersionDetails.Status.Reason)
		}
	}
	return nil
}
