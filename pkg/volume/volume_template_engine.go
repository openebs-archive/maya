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

// Package volume contains specific implementation of cas template engine
package volume

import (
	"fmt"
	"strings"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
)

// volumeEngine is capable of executing a CAS volume related operation via
// CAS template
type volumeEngine struct {
	// engine exposes generic CAS template operations
	engine cast.Interface
	// defaultConfig is the default cas volume configurations found
	// in the CASTemplate
	defaultConfig []v1alpha1.Config
	// casConfigSC is the cas volume config found in the StorageClass
	casConfigSC []v1alpha1.Config
	// casConfigPVC is the cas volume config found in the PersistentVolumeClaim
	casConfigPVC []v1alpha1.Config
}

// NewVolumeEngine returns a new instance of casVolumeEngine based on
// the provided cas configs & runtime volume values
//
// NOTE:
//  runtime volume values set at **runtime** by openebs storage provisioner
// (a kubernetes dynamic storage provisioner)
func NewVolumeEngine(
	casConfigPVC string,
	casConfigSC string,
	castObj *v1alpha1.CASTemplate,
	key string,
	volumeValues map[string]interface{}) (e *volumeEngine, err error) {

	if len(strings.TrimSpace(key)) == 0 {
		err = fmt.Errorf("failed to create cas template engine: nil volume key was provided")
		return
	}
	if len(volumeValues) == 0 {
		err = fmt.Errorf("failed to create cas template engine: nil volume values was provided")
		return
	}
	// CAS config from  PersistentVolumeClaim
	casConfPVC, err := cast.UnMarshallToConfig(casConfigPVC)
	if err != nil {
		return
	}
	// CAS config from StorageClass
	casConfSC, err := cast.UnMarshallToConfig(casConfigSC)
	if err != nil {
		return
	}
	// make use of the generic CAS template engine
	cEngine, err := cast.Engine(castObj, key, volumeValues)
	if err != nil {
		return
	}
	e = &volumeEngine{
		engine:        cEngine,
		defaultConfig: castObj.Spec.Defaults,
		casConfigSC:   casConfSC,
		casConfigPVC:  casConfPVC,
	}
	return
}

// prepareFinalConfig returns the merge of CAS configs from
// PersistentVolumeClaim, StorageClass & CAS Template's default config
//
// NOTE:
//  The priority of config merge is as follows:
//  PersistentVolumeClaim >> StorageClass >> CAS Template Default Config
func (c *volumeEngine) prepareFinalConfig() (final []v1alpha1.Config) {
	// merge unique config elements from SC with config from PVC
	mc := cast.MergeConfig(c.casConfigPVC, c.casConfigSC)
	// merge above resulting config with default config from CASTemplate
	return cast.MergeConfig(mc, c.defaultConfig)
}

// Run executes a CAS volume related operation
func (c *volumeEngine) Run() (op []byte, err error) {
	m, err := cast.ConfigToMap(c.prepareFinalConfig())
	if err != nil {
		return
	}
	// set customized config
	c.engine.SetConfig(m)
	// delegate to generic cas template engine
	return c.engine.Run()
}
