/*
Copyright 2017 The OpenEBS Authors
Copyright 2016 The Kubernetes Authors
Copyright (C) 2013 Masterminds

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

package template

import (
	"bytes"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/util"
)

// empty returns true if the given value has the zero value for its type.
//
// This function is taken as-is from https://github.com/Masterminds/sprig
func empty(given interface{}) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return true
	}

	// Basically adapted from text/template.isTrue
	switch g.Kind() {
	default:
		return g.IsNil()
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return g.Len() == 0
	case reflect.Bool:
		return g.Bool() == false
	case reflect.Complex64, reflect.Complex128:
		return g.Complex() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return g.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return g.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return g.Float() == 0
	case reflect.Struct:
		return false
	}
	return true
}

// NotFoundError represents an error due to a missing object
type NotFoundError struct {
	err string
}

func (e *NotFoundError) Error() string {
	return e.err
}

// VerifyError represents an error due to a failure in verification
type VerifyError struct {
	err string
}

func (e *VerifyError) Error() string {
	return e.err
}

// isLen returns true if the expected value matches the given object's
// length
//
// This function is intended to be used as a go template function.
//
// Example:
// {{- "abc def" | splitList " " | isLen 2 | not | verifyErr "count is not two" | noop -}}
func isLen(expected int, given interface{}) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return false
	}

	// Basically adapted from text/template.isTrue
	switch g.Kind() {
	default:
		return false
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return g.Len() == expected
	}
	return false
}

// jsonPath returns the value at a given jsonpath of a json doc. This resulting
// value is returned as a string.
//
// This function is intended to be used as a go template function.
//
// Example:
// {{- jsonpath .JsonDoc "{.items[*].kind}" | saveAs "TaskResult.kinds" .Values | noop -}}
//
//  Above runs jsonpath template function against .JsonDoc value with the
// provided json path and then saves the result into specific path in Values
// object i.e. at .Values.TaskResult.kinds
//
//  The assumptions here are:
// - '.Values' is of type 'map[string]interface{}'
// - '.JsonDoc' is of type '[]byte'
func jsonPath(json []byte, path string) string {
	jq := NewJsonQuery("templated-jsonpath", json, path)
	output, err := jq.Query()
	if err != nil {
		return err.Error()
	}
	return output
}

// noop as its name suggests does nothing
//
// NOTE:
//  It is intended to be used as a template function where the output of another
// template function is piped to noop to be consumed and this in-turn returns a
// blank.
//
// e.g.
//
// {{- saveAs "path1.path2" .Values "abc" | noop -}}
// {{ val: .Values.path1.path2 }}
//
// Above template with some template values available in .Values will get
// templated as:
//
// val: abc
//
// Alternative approach without use of noop as a template function will be:
//
// {{- $noop := saveAs "path1.path2" .Values "abc" -}}
// {{ val: .Values.path1.path2 }}
//
// Above template with some template values available in .Values will get
// templated like above as:
//
// val: abc
func noop(given ...interface{}) (op string) {
	return
}

// pickSuffix picks and returns an item from an array of strings. It picks the
// first item whose suffix matches the provided match.
//
// NOTE:
//  This is intended to be used as a template function
//
// Example:
//  {{- "jiva-rep, cstor" | splitList ", " | pickSuffix "-rep" | noop -}}
func pickSuffix(match string, given []string) (matched string) {
	for _, givenItem := range given {
		if strings.HasSuffix(givenItem, match) {
			matched = givenItem
			return
		}
	}

	return
}

// pickPrefix picks and returns an item from an array of strings. It picks the
// first item whose prefix matches the provided match.
//
// NOTE:
//  This is intended to be used as a template function
//
// Example:
//  {{- "pvc-jiva-rep, cstor" | splitList ", " | pickPrefix "pvc-" | noop -}}
func pickPrefix(match string, given []string) (matched string) {
	for _, givenItem := range given {
		if strings.HasPrefix(givenItem, match) {
			matched = givenItem
			return
		}
	}

	return
}

// pickContains picks and returns an item from an array of strings. It picks the
// first item which contains the provided match.
//
// NOTE:
//  This is intended to be used as a template function
//
// Example:
//  {{- "pvc-jiva-rep, cstor" | splitList ", " | pickContains "-jiva-" | noop -}}
func pickContains(match string, given []string) (matched string) {
	for _, givenItem := range given {
		if strings.Contains(givenItem, match) {
			matched = givenItem
			return
		}
	}

	return
}

// asKeyMap builds a list of map of key:value pairs where each map is set
// against a key. It accepts an array where each array item is converted into a
// map. Each map is set against a key known as primary key i.e. pkey. The
// resulting pkey:map pairs are set against the destination object.
//
// NOTE:
//  This is intended to be used as a template function
//
// NOTE:
//  This appends current value to original value if any.
//
// Example:
//  {{- "pkey=openebs stor1=jiva stor2=cstor" | splitList " " | asKeyMap "vals" .Target | noop -}}
//  {{- "co1=swarm co2=k8s" | splitList " " | asKeyMap "vals" .Target | noop -}}
//  {{- "pkey=openebs stor2=mstor" | splitList " " | asKeyMap "vals" .Target | noop -}}
//
// Above will result into following:
//  Target: map[string]interface{}{
//    "vals": map[string]interface{}{
//      "openebs": map[string]interface{}{
//        "stor1": "jiva",
//        "stor2": "cstor mstor",
//      },
//      "pkey": map[string]interface{}{
//        "co1": "swarm",
//        "co2": "k8s",
//      },
//    },
//  }
//
// The assumption here is '.Target' is of type 'map[string]interface{}'
func asKeyMap(destinationFields string, destination map[string]interface{}, given []string) interface{} {
	var (
		primaryKey string
		key        string
		value      string
		destFields []string
		fields     []string
	)

	// destFields is the path at which maps will be set
	if len(strings.TrimSpace(destinationFields)) != 0 {
		destFields = strings.Split(destinationFields, ".")
	}

	for _, givenItem := range given {
		// get all the kv pairs
		pairs := strings.Split(givenItem, " ")

		// primary key is determined for each given item
		pKeyPair := pickPrefix("pkey=", pairs)
		pKeyPairs := strings.Split(pKeyPair, "=")
		if len(pKeyPairs) == 2 {
			primaryKey = pKeyPairs[1]
		}

		if len(primaryKey) == 0 {
			// default to pkey
			primaryKey = "pkey"
		}

		for _, pair := range pairs {
			if pair == pKeyPair || len(pair) == 0 {
				// primary key value pair has already been considered & nothing
				// to be done for empty pair
				continue
			}

			// split the current pair by "="
			kvPairs := strings.Split(pair, "=")
			if len(kvPairs) == 0 {
				continue
			}

			// key value pair is determined here
			if len(kvPairs) == 2 {
				key = kvPairs[0]
				value = kvPairs[1]
			} else {
				key = kvPairs[0]
				value = ""
			}

			if len(key) == 0 {
				// nothing needs to be done
				continue
			}

			// reset the fields first
			fields = nil
			fields = append(destFields, primaryKey, key)

			// append to existing value if any
			origVal := strings.TrimSpace(util.GetNestedString(destination, fields...))
			if len(origVal) != 0 {
				value = strings.Join([]string{origVal, value}, ", ")
				value = strings.TrimSuffix(value, ", ")
			}

			// set the current given item value at destination object
			util.SetNestedField(destination, value, fields...)
		}
	}

	return destination
}

// asNestedMap builds a nested map from the given string(s). These strings are
// split as per the provided delimiters fields. The resulting map is set into the
// provided destination object.
//
// NOTE:
//  This is intended to be used as a template function
//
// NOTE:
//  This appends current value to original value if any.
//
// Example:
//  {{- "default/mypod@app=jiva openebs/mypod@app=cstor" | splitList " " | asNestedMap "@ =" .Target | noop -}}
//  {{- "default/mypod@backend=true" | splitList " " | asNestedMap "@ =" .Target | noop -}}
//  {{- "litmus/mypod@backend=true" | splitList " " | asNestedMap "/ @ =" .Target | noop -}}
//
// Above will result into following:
//  Target: map[string]interface{}{
//    "default/mypod": map[string]interface{}{
//      "app": "jiva",
//      "backend": true,
//    },
//    "openebs/mypod": map[string]interface{}{
//      "app": "cstor",
//    },
//    "litmus": map[string]interface{}{
//      "mypod": map[string]interface{}{
//        "backend": true,
//      },
//    },
//  }
//
// Above assumes that .Target is defined as a map[string]interface{} before
// executing the go template
func asNestedMap(delimiters string, destination map[string]interface{}, given []string) interface{} {
	var (
		nestedKeys  []string
		nestedValue string
	)
	// get all the splitters
	splitters := strings.Split(delimiters, " ")

	for _, givenItem := range given {
		nestedKeys = nil
		nestedValue = givenItem

		for _, splitItem := range splitters {
			// split the current given item by the current split item
			kv := strings.Split(nestedValue, splitItem)
			if len(kv) != 2 {
				continue
			}
			nestedKeys = append(nestedKeys, kv[0])
			nestedValue = kv[1]
		}

		// append to existing value if any
		origVal := strings.TrimSpace(util.GetNestedString(destination, nestedKeys...))
		if len(origVal) != 0 {
			nestedValue = strings.Join([]string{origVal, nestedValue}, ", ")
			nestedValue = strings.TrimSuffix(nestedValue, ", ")
		}

		// set the current given item value at destination object
		util.SetNestedField(destination, nestedValue, nestedKeys...)
	}

	return destination
}

// saveAs stores the provided value at specific hierarchy as mentioned in the
// fields inside the values object.
//
// NOTE:
//  This hierarchy along with the provided value is formed/updated in the
// values object.
//
// NOTE:
//  fields is represented as a single string with each field separated by dot
// i.e. '.'
//
// Example:
// {{- "Hi" | saveAs "TaskResult.msg" .Values | noop -}}
// {{- .Values.TaskResult.msg -}}
//
// Above will result in printing 'Hi'
// Assumption here is .Values is of type map[string]interface{}
func saveAs(fields string, values map[string]interface{}, value interface{}) interface{} {
	fieldsArr := strings.Split(fields, ".")
	util.SetNestedField(values, value, fieldsArr...)
	return value
}

// saveIf stores the provided value at specific hierarchy as mentioned in the
// fields. However, the provided value is not set if the hierarchy i.e. the path
// determined by the fields has a non-empty value set previously.
//
// NOTE:
//  This hierarchy along with the provided value is formed/updated in the
// values object.
//
// NOTE:
//  fields is represented as a single string with each field separated by dot
// i.e. '.'
//
// Example:
// {{- "Hi" | saveIf "TaskResult.msg" .Values | noop -}}
// {{- "Hi There" | saveIf "TaskResult.msg" .Values | noop -}}
// {{- .Values.TaskResult.msg -}}
//
// Above will print 'Hi'
//
// Example:
// {{- "Hi" | saveIf "TaskResult.msg" .Values | noop -}}
// {{- "Hi There" | saveAs "TaskResult.msg" .Values | noop -}}
// {{- .Values.TaskResult.msg -}}
//
// Above will print 'Hi There'
//
// Assumption here is .Values is of type 'map[string]interface{}'
func saveIf(fields string, values map[string]interface{}, value interface{}) interface{} {
	fieldsArr := strings.Split(fields, ".")
	oldValue := util.GetNestedField(values, fieldsArr...)
	// will not override the old value
	if !empty(oldValue) {
		return oldValue
	}

	util.SetNestedField(values, value, fieldsArr...)
	return value
}

// notFoundErr returns NotFoundError if given object is empty
//
// Example:
// {{- "" | notFoundErr "empty object" | toString | saveAs "TaskResult.notFoundErrMsg" .Values | noop -}}
// {{- .Values.TaskResult.notFoundErrMsg -}}
//
// Above will print 'empty object'
//
// {{- "" | empty | notFoundErr "replica pod(s) not found" | saveIf "createlistrep.notFoundErr" .Values | noop -}}
//
// Above stores *template.NotFoundError at .Values.createlistrep.notFoundErr
//
// In both the cases assumption is '.Values' is of type 'map[string]interface{}'
func notFoundErr(errMessage string, given interface{}) (err error) {
	if !empty(given) {
		// no error if not empty
		return
	}

	if len(errMessage) == 0 {
		errMessage = "item is not found"
	}

	err = &NotFoundError{
		err: errMessage,
	}
	return
}

// verifyErr returns VerifyError if given verification flag failed i.e. is true
//
// Example:
// {{- "" | empty | verifyErr "name is missing" | toString | saveAs "TaskResult.verifyErrMsg" .Values | noop -}}
// {{- .Values.TaskResult.verifyErrMsg -}}
//
// Above prints 'name is missing'
//
// {{- "" | empty | verifyErr "replica pod(s) not found" | saveIf "createlistrep.verifyErr" .Values | noop -}}
//
// Above stores *template.VerifyError at .Values.createlistrep.verifyErr
//
// In both the cases assumption is '.Values' is of type 'map[string]interface{}'
func verifyErr(errMessage string, hasVerificationFailed bool) (err error) {
	if !hasVerificationFailed {
		// no error if verification did pass successfully
		return
	}

	if len(errMessage) == 0 {
		errMessage = "verification failed"
	}

	err = &VerifyError{
		err: errMessage,
	}
	return
}

// toYaml takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
//
// NOTE: Borrowed from a similar function in helm
//  https://github.com/kubernetes/helm/blob/master/pkg/chartutil/files.go
func toYaml(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}

// fromYaml converts a YAML document into a map[string]interface{}.
//
// This is not a general-purpose YAML parser, and will not parse all valid
// YAML documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
//
// NOTE: Borrowed from helm
// https://github.com/kubernetes/helm/blob/master/pkg/chartutil/files.go
func fromYaml(str string) map[string]interface{} {
	m := map[string]interface{}{}

	if err := yaml.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

// funcMap returns the set of template functions supported in this library
func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()

	// Add some extra templating functions
	extra := template.FuncMap{
		"pickContains": pickContains,
		"pickSuffix":   pickSuffix,
		"pickPrefix":   pickPrefix,
		"toYaml":       toYaml,
		"fromYaml":     fromYaml,
		"jsonpath":     jsonPath,
		"saveAs":       saveAs,
		"saveIf":       saveIf,
		"noop":         noop,
		"notFoundErr":  notFoundErr,
		"verifyErr":    verifyErr,
		"isLen":        isLen,
		"asNestedMap":  asNestedMap,
		"asKeyMap":     asKeyMap,
	}

	for k, v := range extra {
		f[k] = v
	}

	return f
}

// AsTemplatedBytes returns a byte slice
// based on the provided yaml & values
func AsTemplatedBytes(context string, yml string, values map[string]interface{}) ([]byte, error) {
	tpl := template.New(context + "YamlTpl")

	// Any maya yaml exposes below templating functions
	tpl.Funcs(funcMap())

	tpl, err := tpl.Parse(yml)
	if err != nil {
		return nil, err
	}

	// buf is an io.Writer implementation
	// as required by the template
	var buf bytes.Buffer

	// execute the parsed yaml against this instance
	// & write the result into the buffer
	err = tpl.Execute(&buf, values)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// AsMapOfObjects returns a map of objects based on the provided yaml & values
func AsMapOfObjects(yml string, values map[string]interface{}) (map[string]interface{}, error) {
	// templated & then unmarshall-ed version of this yaml
	b, err := AsTemplatedBytes("MapOfObjects", yml, values)
	if err != nil {
		return nil, err
	}

	// Any given YAML can be unmarshalled into a map of arbitrary objects
	var obj map[string]interface{}
	err = yaml.Unmarshal(b, &obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// AsMapOfStrings returns a map of strings based on the provided yaml & values
func AsMapOfStrings(context string, yml string, values map[string]interface{}) (map[string]string, error) {
	// templated & then unmarshall-ed version of this yaml
	b, err := AsTemplatedBytes(context+"MapOfStrings", yml, values)
	if err != nil {
		return nil, err
	}

	// Any given YAML can be unmarshalled into a map of strings
	var obj map[string]string
	err = yaml.Unmarshal(b, &obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
