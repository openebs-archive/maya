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

// This file has volume specific implementation of cas template engine
package volume

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/engine"
	"github.com/openebs/maya/pkg/util"
)

func unMarshallToConfig(config string) (configs []v1alpha1.Config, err error) {
	err = yaml.Unmarshal([]byte(config), &configs)
	return
}

// mergeConfig will merge the unique configuration elements of lowPriorityConfig
// into highPriorityConfig and return the result
func mergeConfig(highPriorityConfig, lowPriorityConfig []v1alpha1.Config) (final []v1alpha1.Config) {
	var prioritized []string

	for _, pc := range highPriorityConfig {
		final = append(final, pc)
		prioritized = append(prioritized, strings.TrimSpace(pc.Name))
	}

	for _, lc := range lowPriorityConfig {
		if !util.ContainsString(prioritized, strings.TrimSpace(lc.Name)) {
			final = append(final, lc)
		}
	}

	return
}

// casVolumeEngine is capable of creating a CAS volume via CAS template
//
// It implements following interfaces:
// - engine.CASCreator
//
// NOTE:
//  It overrides the Create method exposed by generic CASEngine
type casVolumeEngine struct {
	// casEngine exposes generic CAS template operations
	casEngine *engine.CASEngine
	// defaultConfig is the default cas volume configurations found
	// in the CASTemplate
	defaultConfig []v1alpha1.Config
	// casConfigSC is the cas volume config found in the StorageClass
	casConfigSC []v1alpha1.Config
	// casConfigPVC is the cas volume config found in the PersistentVolumeClaim
	casConfigPVC []v1alpha1.Config
}

// NewCASVolumeEngine returns a new instance of casVolumeEngine based on
// the provided cas configs & runtime volume values
//
// NOTE:
//  runtime volume values set at **runtime** by openebs storage provisioner
// (a kubernetes dynamic storage provisioner)
func NewCASVolumeEngine(
	casConfigPVC string,
	casConfigSC string,
	casTemplate *v1alpha1.CASTemplate,
	runtimeKey string,
	runtimeVolumeValues map[string]interface{}) (volumeEngine *casVolumeEngine, err error) {

	if len(strings.TrimSpace(runtimeKey)) == 0 {
		err = fmt.Errorf("failed to create cas template engine: nil runtime volume key was provided")
		return
	}

	if len(runtimeVolumeValues) == 0 {
		err = fmt.Errorf("failed to create cas template engine: nil runtime volume values was provided")
		return
	}

	// CAS config from  PersistentVolumeClaim
	casConfPVC, err := unMarshallToConfig(casConfigPVC)
	if err != nil {
		return
	}

	// CAS config from StorageClass
	casConfSC, err := unMarshallToConfig(casConfigSC)
	if err != nil {
		return
	}

	// make use of the generic CAS template engine
	cEngine, err := engine.NewCASEngine(casTemplate, runtimeKey, runtimeVolumeValues)
	if err != nil {
		return
	}

	volumeEngine = &casVolumeEngine{
		casEngine:     cEngine,
		defaultConfig: casTemplate.Spec.Defaults,
		casConfigSC:   casConfSC,
		casConfigPVC:  casConfPVC,
	}

	return
}

// prepareFinalConfig returns the final config which is a result of merge
// of CAS configs from PersistentVolumeClaim, StorageClass & CAS Template's
// default config.
//
// NOTE:
//  The priority of config merge is as follows:
//  PersistentVolumeClaim >> StorageClass >> CAS Template Default Config
func (c *casVolumeEngine) prepareFinalConfig() (final []v1alpha1.Config) {
	// merge unique config elements from SC with config from PVC
	mc := mergeConfig(c.casConfigPVC, c.casConfigSC)

	// merge above resulting config with default config from CASTemplate
	final = mergeConfig(mc, c.defaultConfig)

	return
}

// Create creates a CAS volume
func (c *casVolumeEngine) Create() ([]byte, error) {
	// set customized CAS config as a top level property
	err := c.casEngine.AddConfigToConfigTLP(c.prepareFinalConfig())
	if err != nil {
		return nil, err
	}

	// delegate to generic cas template engine
	return c.casEngine.Run()
}
