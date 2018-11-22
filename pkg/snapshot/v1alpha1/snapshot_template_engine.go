/*
Copyright 2018 The OpenEBS Authors

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

package snapshot

import (
	"errors"
	"strings"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/engine"
)

// snapshotEngine is capable of creating a CAS snapshot via CAS template
//
// It implements following interfaces:
// - engine.CASCreator
//
// NOTE:
//  It overrides the Create method exposed by generic CASEngine
type snapshotEngine struct {
	// engine exposes generic CAS template operations
	engine engine.Interface
	// defaultConfig is the default cas volume configurations found
	// in the CASTemplate
	defaultConfig []v1alpha1.Config
	// casConfigSC is the cas volume config found in the StorageClass
	casConfigSC []v1alpha1.Config
	// casConfigSnap is the cas volume config found in the PersistentVolumeClaim
	casConfigSnap []v1alpha1.Config
}

// prepareFinalConfig returns the merge of CAS configs from
// VolumeSnapshot, StorageClass & CAS Template's default config
//
// NOTE:
//  The priority of config merge is as follows:
//  VolumeSnapshot >> StorageClass >> CAS Template Default Config
func (c *snapshotEngine) prepareFinalConfig() (final []v1alpha1.Config) {
	// merge unique config elements from SC with config from PVC
	mc := engine.MergeConfig(c.casConfigSnap, c.casConfigSC)
	// merge above resulting config with default config from CASTemplate
	return engine.MergeConfig(mc, c.defaultConfig)
}

// SnapshotEngine returns a new instance of snapshotEngine based on
// the provided cas configs & runtime snapshot values
//
// NOTE:
//  runtime snapshot values set at **runtime** by openebs storage provisioner
// (a kubernetes dynamic storage provisioner)
func SnapshotEngine(
	casConfigSC string,
	casConfigSnap string,
	casTemplate *v1alpha1.CASTemplate,
	key string,
	snapshotValues map[string]interface{}) (snapEngine *snapshotEngine, err error) {

	if len(strings.TrimSpace(key)) == 0 {
		err = errors.New("failed to create cas template engine: nil snapshot key was provided")
		return
	}
	if len(snapshotValues) == 0 {
		err = errors.New("failed to create cas template engine: nil snapshot values was provided")
		return
	}
	// CAS config from StorageClass
	casConfSC, err := engine.UnMarshallToConfig(casConfigSC)
	if err != nil {
		return
	}

	// CAS config from VolumeSnapshot
	casConfSnap, err := engine.UnMarshallToConfig(casConfigSnap)
	if err != nil {
		return
	}

	// make use of the generic CAS template engine
	cEngine, err := engine.New(casTemplate, key, snapshotValues)
	if err != nil {
		return
	}

	snapEngine = &snapshotEngine{
		engine:        cEngine,
		casConfigSC:   casConfSC,
		casConfigSnap: casConfSnap,
		defaultConfig: casTemplate.Spec.Defaults,
	}
	return
}

// Run executes a CAS volume related operation
func (c *snapshotEngine) Run() (op []byte, err error) {
	m, err := engine.ConfigToMap(c.prepareFinalConfig())
	if err != nil {
		return
	}
	// set customized config
	c.engine.SetConfig(m)
	// delegate to generic cas template engine
	return c.engine.Run()
}
