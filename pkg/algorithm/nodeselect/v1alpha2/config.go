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

package v1alpha2

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
)

const (
	// HostName holds the hostname key for kubernetes node.
	HostName = "kubernetes.io/hostname"
)

// Config embeds CSPC object and namespace where openebs is installed.
type Config struct {
	// CSPC is the CStorPoolCluster object.
	CSPC *apis.CStorPoolCluster
	// Namespace is the namespace where openebs is installed.
	Namespace string
	// VisitedNodes is a map which contains the node names which has already been
	// processed for pool provisioning
	VisitedNodes map[string]bool
}

// Builder embeds the Config object.
type Builder struct {
	ConfigObj *Config
	errs      []error
}

// NewBuilder returns an empty instance of Builder object
func NewBuilder() *Builder {
	return &Builder{
		ConfigObj: &Config{
			CSPC:         &apis.CStorPoolCluster{},
			Namespace:    "",
			VisitedNodes: make(map[string]bool),
		},
	}
}

// WithNameSpace sets the Namespace field of config object with provided value.
func (b *Builder) WithNameSpace(ns string) *Builder {
	if len(ns) == 0 {
		b.errs = append(b.errs, errors.New("failed to build algorithm config object: missing namespace"))
		return b
	}
	b.ConfigObj.Namespace = ns
	return b
}

// WithCSPC sets the CSPC field of the config object with the provided value.
func (b *Builder) WithCSPC(cspc *apis.CStorPoolCluster) *Builder {
	if cspc == nil {
		b.errs = append(b.errs, errors.New("failed to build algorithm config object: nil cspc object"))
		return b
	}
	b.ConfigObj.CSPC = cspc
	return b
}

// Build returns the Config  instance
func (b *Builder) Build() (*Config, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.ConfigObj, nil
}
