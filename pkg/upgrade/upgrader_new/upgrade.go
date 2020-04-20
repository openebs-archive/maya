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
	openebsclientset "github.com/openebs/api/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// UpgradeOptions ...
type UpgradeOptions func(*ResourcePatch, *Client) Upgrader

// Client ...
type Client struct {
	// kubeclientset is a standard kubernetes clientset
	KubeClientset kubernetes.Interface
	// openebsclientset is a openebs custom resource package generated for custom API group.
	OpenebsClientset openebsclientset.Interface
}

// Upgrade ...
type Upgrade struct {
	UpgradeMap map[string]UpgradeOptions
	*Client
}

func (u *Upgrade) initClient() error {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrap(err, "error building kubeconfig")
	}
	u.KubeClientset, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building kubernetes clientset")
	}
	u.OpenebsClientset, err = openebsclientset.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building openebs clientset")
	}
	return nil
}

// NewUpgrade ...
func NewUpgrade() *Upgrade {
	u := &Upgrade{
		UpgradeMap: map[string]UpgradeOptions{},
		Client:     &Client{},
	}
	u.initClient()
	u.RegisterAll()
	return u
}
