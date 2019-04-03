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

package v1alpha1

import (
	"io/ioutil"

	"github.com/pkg/errors"

	log "github.com/golang/glog"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
	config "github.com/openebs/maya/pkg/upgrade/config/v1alpha1"
	engine "github.com/openebs/maya/pkg/upgrade/engine/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StartOptions contains start options for openebs upgrade
type StartOptions struct {
	ConfigPath string
}

// Run runs various steps to upgrade unit of upgrades
// present in config.
func (opt *StartOptions) Run() error {
	data, err := ioutil.ReadFile(opt.ConfigPath)
	if err != nil {
		log.Errorf("unable to read config from file : %v", err)
		return err
	}
	cfg, err := config.NewBuilder().
		WithYamlString(string(data)).
		AddCheckf(config.IsCASTemplateNamePresent(),
			"castemplate name not present").
		AddCheckf(config.IsResourcePresent(),
			"empty resource provided").
		AddCheckf(config.IsValidResource(),
			"resource should contains name namespace and kind").
		AddCheckf(config.IsSameKind(),
			"single job can not upgrade multiple kind of resource").
		Build()
	if err != nil {
		log.Errorf("upgrade config validation error : %v ", err)
		return err
	}
	castObj, err := cast.KubeClient().
		Get(cfg.CASTemplate, metav1.GetOptions{})
	if err != nil {
		return err
	}
	engines := []cast.Interface{}
	engineListErrors := []error{}
	engineRunErrors := []error{}
	for _, resource := range cfg.Resources {
		e, err := engine.New().
			WithCASTemplate(castObj).
			WithUnitOfUpgrade(&resource).
			WithRuntimeConfig(cfg.Data).
			Build()

		if err != nil {
			engineListErrors = append(engineListErrors, err)
			log.Errorf("error while getting engine for %s '%s' ", resource.Kind, resource.Name)
			continue
		}
		engines = append(engines, e)
	}

	for i, e := range engines {
		op, err := e.Run()
		if err != nil {
			log.Errorf("error while upgrading %s '%s' ",
				cfg.Resources[i].Kind, cfg.Resources[i].Name)
			engineRunErrors = append(engineRunErrors, err)
			continue
		}
		log.Infof("successfully upgraded %s '%s' ",
			cfg.Resources[i].Kind, cfg.Resources[i].Name)
		log.Infof("---------- %s '%s' upgrade result ----------\n%v",
			cfg.Resources[i].Kind, cfg.Resources[i].Name, string(op))
	}

	if len(engineListErrors) != 0 {
		return errors.Errorf("error while listing engines : %v", engineListErrors)
	}

	if len(engineRunErrors) != 0 {
		return errors.Errorf("error while running engines : %v", engineRunErrors)
	}

	return nil
}
