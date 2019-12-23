/*
Copyright 2019 The OpenEBS Authors.

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

package v1alpha1

import (
	"io/ioutil"
	"path"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	upgrade "github.com/openebs/maya/pkg/upgrade/v1alpha1"
)

// Upgrade contains configurations to perform upgrade
type Upgrade struct {
	// ConfigPath represents the configuration that
	// is provided to upgrade as its input
	ConfigPath string

	// Config represents the config instance
	// built from ConfigPath
	Config *apis.UpgradeConfig
}

// String implements Stringer interface
func (u Upgrade) String() string {
	return stringer.Yaml("upgrade", u)
}

// GoString implements GoStringer interface
func (u Upgrade) GoString() string {
	return u.String()
}

// NewUpgradeForConfigPath takes config file path and add
// config in upgrade instance
func NewUpgradeForConfigPath(filePath string) (*Upgrade, error) {
	data, err := ioutil.ReadFile(path.Clean(filePath))
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to initialize upgrade instance: failed to read config: %s", filePath)
	}

	cfg, err := upgrade.ConfigBuilderForRaw(data).
		AddCheckf(upgrade.IsCASTemplateName(), "missing castemplate name").
		AddCheckf(upgrade.IsResource(), "missing resource(s)").
		AddCheckf(upgrade.IsValidResource(),
			"invalid resource: verify if namespace, kind and name were provided").
		AddCheckf(upgrade.IsSameKind(),
			"invalid resources: all resources should belong to same kind").
		Build()
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to instantiate upgrade instance: config path {%s}", filePath)
	}
	return &Upgrade{
		ConfigPath: filePath,
		Config:     cfg,
	}, nil
}

// Run runs various steps to upgrade unit of upgrades
// present in config.
func (u *Upgrade) Run() error {
	e, err := ExecutorBuilderForConfig(u.Config).
		Build()
	if err != nil {
		return errors.Wrapf(err,
			"failed to run upgrade: %s", u)
	}

	err = e.Execute()
	if err != nil {
		return errors.Wrapf(err,
			"failed to run upgrade: %s", u)
	}
	return nil
}
