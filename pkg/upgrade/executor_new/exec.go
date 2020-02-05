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
