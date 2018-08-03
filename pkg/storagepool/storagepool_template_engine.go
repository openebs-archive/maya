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

// This file has storagepool specific implementation of cas template engine
package storagepool

import (
	"fmt"
	"strings"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/engine"
	"github.com/openebs/maya/pkg/util"
)

// casStoragePoolEngine is capable of creating a storagepool via CAS template
//
// It implements following interfaces:
// - engine.CASCreator
//
// NOTE:
//  It overrides the Create method exposed by generic CASEngine
type casStoragePoolEngine struct {
	// casEngine exposes generic CAS template operations
	casEngine *engine.CASEngine
	// defaultConfig is the default cas storagepool configurations found
	// in the CASTemplate
	defaultConfig []v1alpha1.Config
}

// NewCASStoragePoolEngine returns a new instance of casStoragePoolEngine based on
// the provided cas configs & runtime storagepool values
func NewCASStoragePoolEngine(
	casTemplate *v1alpha1.CASTemplate,
	runtimeKey string,
	runtimeStoragePoolValues map[string]interface{}) (storagePoolEngine *casStoragePoolEngine, err error) {

	if len(strings.TrimSpace(runtimeKey)) == 0 {
		err = fmt.Errorf("Failed to create cas template engine: nil runtime storagepool key was provided")
		return
	}

	if len(runtimeStoragePoolValues) == 0 {
		err = fmt.Errorf("Failed to create cas template engine: nil runtime storagepool values was provided")
		return
	}

	// make use of the generic CAS template engine
	cEngine, err := engine.NewCASEngine(casTemplate, runtimeKey, runtimeStoragePoolValues)
	if err != nil {
		return
	}

	storagePoolEngine = &casStoragePoolEngine{
		casEngine:     cEngine,
		defaultConfig: casTemplate.Spec.Defaults,
	}

	return
}

// addConfigToConfigTLP will add final cas storagepool configurations to ConfigTLP.
//
// NOTE:
//  This will enable templating a run task template as follows:
//
// {{ .Config.<ConfigName>.enabled }}
// {{ .Config.<ConfigName>.value }}
//
// NOTE:
//  Above parsing scheme is translated by running `go template` against the run
// task template
func (c *casStoragePoolEngine) addConfigToConfigTLP() error {
	var configName string
	allConfigsHierarchy := map[string]interface{}{}
	allConfigs := c.defaultConfig

	for _, config := range allConfigs {
		configName = strings.TrimSpace(config.Name)
		if len(configName) == 0 {
			return fmt.Errorf("Failed to merge config '%#v': missing config name", config)
		}

		configHierarchy := map[string]interface{}{
			configName: map[string]string{
				string(v1alpha1.EnabledPTP): config.Enabled,
				string(v1alpha1.ValuePTP):   config.Value,
			},
		}

		isMerged := util.MergeMapOfObjects(allConfigsHierarchy, configHierarchy)
		if !isMerged {
			return fmt.Errorf("Failed to merge config: unable to add config '%s' to config hierarchy", configName)
		}
	}

	// update merged config as the top level property
	c.casEngine.SetConfig(allConfigsHierarchy)
	return nil
}

// Create creates a storagepool
func (c *casStoragePoolEngine) Create() ([]byte, error) {
	// set customized CAS config as a top level property
	err := c.addConfigToConfigTLP()
	if err != nil {
		return nil, err
	}

	// delegate to generic cas template engine
	return c.casEngine.Run()
}
