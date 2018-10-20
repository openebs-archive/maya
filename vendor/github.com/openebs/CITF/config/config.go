/*
Copyright 2018 The OpenEBS Authors.
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

package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"os"
	"reflect"
	"strings"

	"github.com/openebs/CITF/common"
	"github.com/openebs/CITF/utils/log"
	yaml "gopkg.in/yaml.v2"
)

// Configuration is struct to hold the configurations of CITF
type Configuration struct {
	Environment    string `json:"environment,omitempty" yaml:"environment,omitempty"`
	Debug          bool   `json:"debug,omitempty" yaml:"debug,omitempty"`
	KubeMasterURL  string `json:"kubeMasterURL,omitempty" yaml:"kubeMasterURL,omitempty"`
	KubeConfigPath string `json:"kubeConfigPath,omitempty" yaml:"kubeConfigPath,omitempty"`
}

var (
	// Conf will contain configurations for CITF
	Conf        Configuration
	defaultConf Configuration
)

const (
	debugEnabledVal     = true
	debugDisabledVal    = false
	debugEnabledValStr  = "true"
	debugDisabledValStr = "false"
)

func init() {
	defaultConf = Configuration{
		Environment:    common.Minikube,
		Debug:          debugDisabledVal,
		KubeMasterURL:  "",
		KubeConfigPath: filepath.Join(os.Getenv("HOME"), ".kube", "config"),
	}

	// Set debug status to util packages
	SetDebugToUtilPackages(Debug())
}

// SetDebugToUtilPackages Enables/Disables debug in util packages
// this way we are decoupling util packages from others
func SetDebugToUtilPackages(debugEnabled bool) {
	log.DebugEnabled = debugEnabled
}

// LoadConf loads the configuration from the file which path is supplied
func LoadConf(confFilePath string) error {
	if len(confFilePath) == 0 {
		return nil
	}

	yamlBytes, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return fmt.Errorf("error reading file: %q. Error: %+v", confFilePath, err)
	}

	// Always pass pointer to the destination structure.
	// https://github.com/go-yaml/yaml/issues/224
	err = yaml.Unmarshal(yamlBytes, &Conf)
	if err != nil {
		return fmt.Errorf("error parsing file: %q. Error: %+v", confFilePath, err)
	}

	// Set debug status to util packages
	SetDebugToUtilPackages(Debug())
	return nil
}

// getConfValueByStringField returns value of the given field string in given Configuration
func getConfValueByStringField(conf Configuration, field string) string {
	r := reflect.ValueOf(conf)
	f := reflect.Indirect(r).FieldByName(field)
	return fmt.Sprintf("%v", f)
}

// GetDefaultValueByStringField returns the value of the given field string in Default Configuration
// fields should be in exact case as the field is present in struct Configuration
func GetDefaultValueByStringField(field string) string {
	return getConfValueByStringField(defaultConf, field)
}

// GetUserConfValueByStringField returns the value of the given field string in Default Configuration
// fields should be in exact case as the field is present in struct Configuration
func GetUserConfValueByStringField(field string) string {
	return getConfValueByStringField(Conf, field)
}

// GetConf returns the applicable configuration for the given field
func GetConf(field string) string {
	if value, ok := os.LookupEnv("CITF_CONF_" + strings.ToUpper(field)); ok {
		return value
	}
	if value := GetUserConfValueByStringField(field); len(value) != 0 {
		return value
	}
	return GetDefaultValueByStringField(field)
}

// Environment returns the environment which should be used in testing
func Environment() string {
	return GetConf("Environment")
}

// Debug returns the environment which should be used in testing
func Debug() bool {
	return strings.ToLower(GetConf("Debug")) == debugEnabledValStr
}

// Verbose is an alias of Debug which returns the environment which should be used in testing
var Verbose = Debug

// KubeMasterURL returns the URL of kube-master as per citf configurations
func KubeMasterURL() string {
	return GetConf("KubeMasterURL")
}

// KubeConfigPath returns the path of kube-config as per citf configurations
func KubeConfigPath() string {
	return GetConf("KubeConfigPath")
}
