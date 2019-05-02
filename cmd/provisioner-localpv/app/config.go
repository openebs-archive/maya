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

package app

import (
	"fmt"
	"path/filepath"
	"strings"

	mconfig "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
)

const (
	KeyPVBasePath     = "BasePath"
	KeyPVRelativePath = "RelativePath"
	KeyPVAbsolutePath = "AbsolutePath"
)

const (
	// BetaStorageClassAnnotation represents the beta/previous StorageClass annotation.
	// It's currently still used and will be held for backwards compatibility
	BetaStorageClassAnnotation = "volume.beta.kubernetes.io/storage-class"
)

//CASConfigParser creates a new CASConfigPVC struct by
// parsing and merging the configuration provided in the PVC
// annotation - cas.openebs.io/config with the
// default configuration of the provisioner.
func (p *Provisioner) CASConfigParser(pvName string, pvc *v1.PersistentVolumeClaim) (*CASConfigPVC, error) {

	pvConfig := p.defaultConfig

	//TODO Fetch the SC and its configuration
	scName := GetStorageClassName(pvc)

	// extract the cas volume config from pvc
	pvcCASConfigStr := pvc.ObjectMeta.Annotations[string(mconfig.CASConfigKey)]
	if len(strings.TrimSpace(pvcCASConfigStr)) != 0 {
		pvcCASConfig, err := cast.UnMarshallToConfig(pvcCASConfigStr)
		if err == nil {
			pvConfig = cast.MergeConfig(pvcCASConfig, pvConfig)
		}
	}

	pvConfigMap, err := cast.ConfigToMap(pvConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read configuration for pvc %v", pvc.ObjectMeta.Name)
	}

	c := &CASConfigPVC{
		pvName:  pvName,
		pvcName: pvc.ObjectMeta.Name,
		scName:  scName,
		config:  pvConfigMap,
	}
	return c, nil
}

//GetPath returns a valid PV path based on the configuration
// or an error. The Path is constructed using the following rules:
// If AbsolutePath is specified return it. (Future)
// If PVPath is specified, suffix it with BasePath and return it. (Future)
// If neither of above are specified, suffix the PVName to BasePath
//  and return it
// Also before returning the path, validate that path is safe
//  and matches the filters specified in StorageClass.
func (c *CASConfigPVC) GetPath() (string, error) {
	//This feature need to be supported with some more
	// security checks are in place, so that rouge pods
	// don't get access to node directories.
	//absolutePath := c.getConfigValue(KeyPVAbsolutePath)
	//if len(strings.TrimSpace(absolutePath)) != 0 {
	//	return c.validatePath(absolutePath)
	//}

	basePath := c.getConfigValue(KeyPVBasePath)
	if len(strings.TrimSpace(basePath)) == 0 {
		return "", fmt.Errorf("configuration error, no base path was specified")
	}

	//This feature need to be supported after the
	// security checks are in place.
	//pvRelPath := c.getConfigValue(KeyPVRelativePath)
	//if len(strings.TrimSpace(pvRelPath)) == 0 {
	//	pvRelPath = c.pvName
	//}

	pvRelPath := c.pvName
	path := filepath.Join(basePath, pvRelPath)

	return c.validatePath(path)
}

//getConfigValue is a utility function to extract the value
// of the `key` from the ConfigMap object - which is
// map[string]interface{map[string][string]}
func (c *CASConfigPVC) getConfigValue(key string) string {
	if configObj, ok := util.GetNestedField(c.config, key).(map[string]string); ok {
		if val, p := configObj[string(mconfig.ValuePTP)]; p {
			return val
		}
	}
	return ""
}

//validatePath checks for the sanity of the PV path
func (c *CASConfigPVC) validatePath(path string) (string, error) {
	//Validate that path is well formed.
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	//Validate that root directories are not used as PV paths
	//Valid path should have a basepath (which is not /) and
	// a subpath for the PV.
	path = strings.TrimSuffix(path, "/")
	parentDir, volumeDir := c.extractSubPath(path)
	//parentDir = strings.TrimSuffix(parentDir, "/")
	//volumeDir = strings.TrimSuffix(volumeDir, "/")
	if parentDir == "" || volumeDir == "" {
		// it covers the `/` case
		return "", fmt.Errorf("invalid path %v for cleanup: cannot find parent dir or volume dir", path)
	}

	//TODO: Validate against blacklist or whitelist of paths
	return path, nil
}

//extractSubPath is utility function to split directory from path
func (c *CASConfigPVC) extractSubPath(path string) (string, string) {
	parentDir, volumeDir := filepath.Split(path)
	parentDir = strings.TrimSuffix(parentDir, "/")
	volumeDir = strings.TrimSuffix(volumeDir, "/")
	return parentDir, volumeDir
}

// GetStorageClassName extracts the StorageClass name from PVC
func GetStorageClassName(pvc *v1.PersistentVolumeClaim) *string {
	// Use beta annotation first
	if class, found := pvc.Annotations[BetaStorageClassAnnotation]; found {
		return &class
	}
	return pvc.Spec.StorageClassName
}
