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
package executor

import (
	upgrader "github.com/openebs/maya/pkg/upgrade/upgrader_new"
)

// UpgradeOptions ...
type UpgradeOptions func(*upgrader.ResourcePatch) upgrader.Upgrader

// Upgrade ...
type Upgrade struct {
	UpgradeMap map[string]UpgradeOptions
}

// NewUpgrade ...
func NewUpgrade() *Upgrade {
	u := &Upgrade{
		UpgradeMap: map[string]UpgradeOptions{},
	}
	u.RegisterAll()
	return u
}

// Exec ...
func Exec(fromVersion, toVersion, kind, name,
	openebsNamespace, urlprefix, imagetag string) error {
	rp := upgrader.NewResourcePatch(
		upgrader.FromVersion(fromVersion),
		upgrader.ToVersion(toVersion),
		upgrader.WithName(name),
		upgrader.WithOpenebsNamespace(openebsNamespace),
		upgrader.WithBaseURL(urlprefix),
		upgrader.WithImageTag(imagetag),
	)
	u := NewUpgrade()
	err := u.UpgradeMap[kind](rp).Upgrade()
	if err != nil {
		return err
	}
	return nil
}
