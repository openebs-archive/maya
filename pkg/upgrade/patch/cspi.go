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
package patch

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cspi "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

// CSPI ...
type CSPI struct {
	Object *apis.CStorPoolInstance
	Data   []byte
}

// NewCSPI ...
func NewCSPI() *CSPI {
	return &CSPI{}
}

// PreChecks ...
func (c *CSPI) PreChecks(from, to string) error {
	if c.Object == nil {
		return errors.Errorf("nil cspi object")
	}
	version := c.Object.Labels[string(apis.OpenEBSVersionKey)]
	if version != from && version != to {
		return errors.Errorf(
			"cspi version %s is neither %s nor %s",
			version,
			from,
			to,
		)
	}
	return nil
}

// Patch ...
func (c *CSPI) Patch(from, to string) error {
	klog.Info("patching cspi ", c.Object.Name)
	client := cspi.NewKubeClient(cspi.WithKubeConfigPath("/var/run/kubernetes/admin.kubeconfig"))
	version := c.Object.Labels[string(apis.OpenEBSVersionKey)]
	if version == to {
		klog.Infof("cspi already in %s version", to)
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
				"failed to patch cspi %s",
				c.Object.Name,
			)
		}
		klog.Infof("cspi %s patched", c.Object.Name)
	}
	return nil
}

// Get ...
func (c *CSPI) Get(name, namespace string) error {
	cspi, err := cspi.NewKubeClient(cspi.WithKubeConfigPath("/var/run/kubernetes/admin.kubeconfig")).WithNamespace(namespace).
		Get(name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get cspi %s in %s namespace", name, namespace)
	}
	c.Object = cspi
	return nil
}
