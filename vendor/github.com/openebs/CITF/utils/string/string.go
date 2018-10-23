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

package string

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/openebs/CITF/utils/log"
	"gopkg.in/yaml.v2"
)

var logger log.Logger

// PrettyString returns the prettified string of the interface supplied. (If it can)
func PrettyString(in interface{}) string {
	jsonStr, err := json.MarshalIndent(in, "", "    ")
	logger.PrintlnDebugMessageIfError(err, "unable to marshal")
	if err != nil {
		return fmt.Sprintf("%+v", in)
	}

	return string(jsonStr)
}

// ReplaceHexCodesWithValue finds any string slice which is equivalent to hexadecimal value of
// a character then it replaces that with its value.
// If any error occurred while resolving it leaves that part as it is.
// e.g. "\\x20" in a string will be replaced with value of "\x20" i.e. space
func ReplaceHexCodesWithValue(s string) (string, error) {
	regexString := "\\\\[xX][0-9a-fA-F]{2}"
	pattern, _ := regexp.Compile(regexString) // This should be covered in unit test

	return pattern.ReplaceAllStringFunc(s, func(s string) string {
		bytes, err := hex.DecodeString(s[2:])
		logger.LogErrorf(err, "error occurred while resolving %q", s)
		if err != nil {
			return s
		}
		return string(bytes)
	}), nil
}

// ConvertMapI2MapS walks the given dynamic object recursively, and
// converts maps with interface{} key type to maps with string key type.
// This function comes handy if you want to marshal a dynamic object into
// JSON where maps with interface{} key type are not allowed.
//
// Recursion is implemented into values of the following types:
//   -map[interface{}]interface{}
//   -map[string]interface{}
//   -[]interface{}
//
// When converting map[interface{}]interface{} to map[string]interface{},
// fmt.Sprint() with default formatting is used to convert the key to a string key.
//
// Source: https://github.com/icza/dyno/blob/6009b3da28e195fd676c79e5bcbee68bcda793e3/dyno.go#L515
func ConvertMapI2MapS(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v2 := range x {
			switch k2 := k.(type) {
			case string: // Fast check if it's already a string
				m[k2] = ConvertMapI2MapS(v2)
			default:
				m[fmt.Sprint(k)] = ConvertMapI2MapS(v2)
			}
		}
		v = m

	case []interface{}:
		for i, v2 := range x {
			x[i] = ConvertMapI2MapS(v2)
		}

	case map[string]interface{}:
		for k, v2 := range x {
			x[k] = ConvertMapI2MapS(v2)
		}
	}

	return v
}

// ConvertYAMLtoJSON converts yaml bytes into json bytes
func ConvertYAMLtoJSON(yamlBytes []byte) ([]byte, error) {
	var body interface{}
	if err := yaml.Unmarshal(yamlBytes, &body); err != nil {
		return []byte{}, err
	}

	body = ConvertMapI2MapS(body)

	b, err := json.MarshalIndent(body, "", "    ")
	if err != nil {
		return []byte{}, err
	}

	return b, nil
}

// ConvertJSONtoYAML converts json bytes into yaml bytes
func ConvertJSONtoYAML(jsonBytes []byte) ([]byte, error) {
	var body interface{}
	if err := json.Unmarshal(jsonBytes, &body); err != nil {
		return []byte{}, err
	}

	body = ConvertMapI2MapS(body)

	b, err := yaml.Marshal(body)
	if err != nil {
		return []byte{}, err
	}

	return b, nil
}
