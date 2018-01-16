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

package maya

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/ghodss/yaml"
)

// CustomFuncsHolder contains properties that are
// used to build custom text/template functions
type CustomFuncsHolder struct {
	// Inputs contains the k:v pairs required to
	// set the template's placeholders
	Inputs map[string]string `json:"inputs"`
	// Stores contains the k:v pairs out of resulting
	// actions on the template's embedded object
	Stores map[string]string
}

func customFuncVal(pairs map[string]string, context, key string) (string, error) {
	if len(pairs) == 0 {
		return "", fmt.Errorf("No %s found", context)
	}

	if len(key) == 0 {
		return "", fmt.Errorf("Missing %s key", context)
	}

	val := pairs[key]
	if len(val) == 0 {
		return "", fmt.Errorf("Nil value for %s key '%s'", context, key)
	}

	return val, nil
}

func (f *CustomFuncsHolder) inputVal(key string) (string, error) {
	return customFuncVal(f.Inputs, "inputs", key)
}

func (f *CustomFuncsHolder) storeVal(key string) (string, error) {
	return customFuncVal(f.Stores, "stores", key)
}

func (f *CustomFuncsHolder) setStore(key, value string) {
	f.Stores[key] = value
}

func (f *CustomFuncsHolder) setInput(key, value string) {
	f.Inputs[key] = value
}

func (f *CustomFuncsHolder) setStoreIfEmpty(key, value string) {
	if len(f.Stores[key]) == 0 {
		f.Stores[key] = value
	} else {
		// TODO log
	}
}

func (f *CustomFuncsHolder) setInputIfEmpty(key, value string) {
	if len(f.Inputs[key]) == 0 {
		f.Inputs[key] = value
	} else {
		// TODO log
	}
}

// mergeInputsIfEmpty is an append minus override
//
// NOTE: Immutability is fundamental to having systems
// that work as per instructions and later to ensure
// better debuggability
func (f *CustomFuncsHolder) mergeInputsIfEmpty(inputs map[string]string) {
	if len(f.Inputs) == 0 {
		f.Inputs = inputs
		return
	}

	for k, v := range inputs {
		f.setInputIfEmpty(k, v)
	}
}

// mergeStoresIfEmpty is an append minus override
//
// NOTE: Immutability is fundamental to having systems
// that work as per instructions and later to ensure
// better debuggability
func (f *CustomFuncsHolder) mergeStoresIfEmpty(stores map[string]string) {
	if len(f.Stores) == 0 {
		f.Stores = stores
		return
	}

	for k, v := range stores {
		f.setStoreIfEmpty(k, v)
	}
}

// MayaYamlV2 represents a yaml definition
//
// This yaml is expected to be marshalled into
// corresponding go struct. YAMLs will be used to specify
// DevOps intents.
type MayaYamlV2 struct {
	// Yaml represents a templated yaml in string format
	Yaml string

	// CustomFuncsHolder contains the values menat to replace
	// the placeholders set in the YAML. It also exposes functions
	// as text/template's custom functions. We need custom functions
	// to retain the user friendliness of the YAML.
	CustomFuncsHolder

	// YmlInBytes represents the templated yaml in
	// byte slice format. This is typically the result of applying
	// text/template on the YAML that has ocassional placeholders.
	YmlInBytes []byte
}

// getYaml provides the yaml in string format
func (m *MayaYamlV2) getYaml() (string, error) {
	if len(m.Yaml) == 0 {
		return "", fmt.Errorf("Yaml is not set")
	}

	return m.Yaml, nil
}

// asTemplatedBytes returns a byte slice format of its yaml
// representation. This []byte is derived after applying the
// text/template.
func (m *MayaYamlV2) asTemplatedBytes() ([]byte, error) {
	yml, err := m.getYaml()
	if err != nil {
		return nil, err
	}

	tpl := template.New("mayayaml")
	// Any maya yaml exposes below custom functions. These are
	// used to set placeholders at runtime.
	tpl.Funcs(template.FuncMap{
		"inputs": m.inputVal,
		"stores": m.storeVal,
	})

	tpl, err = tpl.Parse(yml)
	if err != nil {
		return nil, err
	}

	// this has implementation of io.Writer
	// that is required by the template
	var buf bytes.Buffer

	// execute the parsed yaml against this instance
	// & write the result into the buffer
	err = tpl.Execute(&buf, m)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// asMapOfObjects returns a map of any objects
// that corresponds to this instance's yaml
//
// NOTE: This is required in the cases where developer does not
// know the targetted go struct to unmarshall this yaml into.
// However, this developer will be interested in a particular
// piece from the entire yaml. Developer will use this
// method before trying to get the desired piece.
func (m *MayaYamlV2) asMapOfObjects() (map[string]interface{}, error) {
	// templated & then unmarshall-ed version of this yaml
	b, err := m.asTemplatedBytes()
	if err != nil {
		return nil, err
	}

	// Any given YAML can be unmarshalled into a map of any objects
	var obj map[string]interface{}
	err = yaml.Unmarshal(b, &obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
