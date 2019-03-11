/*
Copyright 2017 The OpenEBS Authors

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

// Package storagepool provides a specific implementation of CAS template engine
package storagepool

import (
	"fmt"
	"strings"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
)

// storagePoolEngine is capable of creating a storagepool via CAS template
//
// It implements following interfaces:
// - engine.CASCreator
//
// NOTE:
//  It overrides the Create method exposed by generic engine
type storagePoolEngine struct {
	engine        cast.Interface    // generic CAS template engine
	defaultConfig []v1alpha1.Config // default cas storagepool config found in CASTemplate
	openebsConfig []v1alpha1.Config // openebsConfig is the config that is provided
}

// NewStoragePoolEngine returns a new instance of storagePoolEngine
func NewStoragePoolEngine(
	castObj *v1alpha1.CASTemplate,
	openebsConfig string,
	key string,
	storagePoolValues map[string]interface{}) (e *storagePoolEngine, err error) {

	if len(strings.TrimSpace(key)) == 0 {
		err = fmt.Errorf("Failed to create cas template engine: nil storagepool key was provided")
		return
	}
	if len(storagePoolValues) == 0 {
		err = fmt.Errorf("Failed to create cas template engine: nil storagepool values was provided")
		return
	}
	openebsConf, err := cast.UnMarshallToConfig(openebsConfig)
	if err != nil {
		return
	}
	cEngine, err := cast.Engine(castObj, key, storagePoolValues)
	if err != nil {
		return
	}
	e = &storagePoolEngine{
		engine:        cEngine,
		defaultConfig: castObj.Spec.Defaults,
		openebsConfig: openebsConf,
	}
	return
}

// Create creates a storagepool
func (c *storagePoolEngine) Create() (op []byte, err error) {
	m, err := cast.ConfigToMap(cast.MergeConfig(c.openebsConfig, c.defaultConfig))
	if err != nil {
		return
	}
	// set customized config
	c.engine.SetConfig(m)
	// delegate to generic cas template engine
	return c.engine.Run()
}
