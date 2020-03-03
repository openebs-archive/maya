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
	"strings"

	apis "github.com/openebs/api/pkg/apis/cstor/v1"
	clientset "github.com/openebs/api/pkg/client/clientset/versioned"
	//apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

// CSPC ...
type CSPC struct {
	Object *apis.CStorPoolCluster
	Data   []byte
	Client *clientset.Clientset
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
	version := c.Object.VersionDetails.Desired
	if version == to {
		klog.Infof("cspc already in %s version", to)
		return nil
	}
	if version == from {
		patch := c.Data
		_, err := c.Client.CstorV1().CStorPoolClusters(c.Object.Namespace).Patch(
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
	cspcObj, err := c.Client.CstorV1().CStorPoolClusters(namespace).
		Get(name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get cspc %s in %s namespace", name, namespace)
	}
	c.Object = cspcObj
	return nil
}
