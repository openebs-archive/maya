package v1alpha2

import (
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
)

// GetPropertyValue will return value of given property for given pool
func GetPropertyValue(poolName, property string) (string, error) {
	ret, err := zfs.NewPoolGetProperty().
		WithScriptedMode(true).
		WithField("value").
		WithProperty(property).
		WithPool(poolName).
		Execute()
	return string(ret), err
}
