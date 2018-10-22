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
	engine *engine.CASEngine
	// defaultConfig is the default cas snapshot configurations found
	// in the CASTemplate
	defaultConfig []v1alpha1.Config
}

// SnapshotEngine returns a new instance of snapshotEngine based on
// the provided cas configs & runtime snapshot values
//
// NOTE:
//  runtime snapshot values set at **runtime** by openebs storage provisioner
// (a kubernetes dynamic storage provisioner)
func SnapshotEngine(
	casTemplate *v1alpha1.CASTemplate,
	runtimeKey string,
	runtimeSnapshotValues map[string]interface{}) (snapEngine *snapshotEngine, err error) {

	if len(strings.TrimSpace(runtimeKey)) == 0 {
		err = errors.New("failed to create cas template engine: nil runtime snapshot key was provided")
		return
	}

	if len(runtimeSnapshotValues) == 0 {
		err = errors.New("failed to create cas template engine: nil runtime snapshot values was provided")
		return
	}

	// make use of the generic CAS template engine
	cEngine, err := engine.NewCASEngine(casTemplate, runtimeKey, runtimeSnapshotValues)
	if err != nil {
		return
	}

	snapEngine = &snapshotEngine{
		engine:        cEngine,
		defaultConfig: casTemplate.Spec.Defaults,
	}

	return
}

// Create creates a CAS snapshot
func (c *snapshotEngine) Create() ([]byte, error) {
	// set customized CAS config as a top level property
	err := c.engine.AddConfigToConfigTLP(c.defaultConfig)
	if err != nil {
		return nil, err
	}

	// delegate to generic cas template engine
	return c.engine.Run()
}
