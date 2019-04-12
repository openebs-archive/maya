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

	"github.com/pkg/errors"

	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	upgrade "github.com/openebs/maya/pkg/upgrade/v1alpha1"
)

// Upgrade contains start options for openebs upgrade
type Upgrade struct {
	ConfigPath string
}

// String implements Stringer interface
func (u Upgrade) String() string {
	return stringer.Yaml("upgrade config", u)
}

// GoString implements GoStringer interface
func (u Upgrade) GoString() string {
	return u.String()
}

// Run runs various steps to upgrade unit of upgrades
// present in config.
func (u *Upgrade) Run() error {
	data, err := ioutil.ReadFile(u.ConfigPath)
	if err != nil {
		return errors.WithMessagef(err,
			"failed to run upgrade: failed to read config: %s", u)
	}

	cfg, err := upgrade.ConfigBuilderForRaw(data).
		AddCheckf(upgrade.IsCASTemplateName(), "missing castemplate name").
		AddCheckf(upgrade.IsResource(), "missing resource(s) for upgrade").
		AddCheckf(upgrade.IsValidResource(),
			"invalid resource: verify if namespace, kind and name were provided").
		AddCheckf(upgrade.IsSameKind(),
			"invalid resources: all resources should belong to same kind").
		Build()
	if err != nil {
		return errors.WithMessagef(err,
			"failed to run upgrade: %s", u)
	}

	el, err := ListEngineBuilderForConfig(cfg).
		Build()
	if err != nil {
		return errors.WithMessagef(err,
			"failed to run upgrade: %s", cfg)
	}

	err = el.Run()
	if err != nil {
		return errors.WithMessagef(err,
			"failed to run upgrade: %s", cfg)
	}

	return nil
}
