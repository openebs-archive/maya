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
	"fmt"
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
// blank. This is required for cases where we would like to preserve the
// yaml/template format.
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

// splitKeyMap builds a list of map of key:value pairs where each map is set
// against a key. It accepts an array where each array item is converted into a
// map. Each map is set against a key known as primary key i.e. pkey. The
// resulting pkey:map pairs are set against the destination object. This
// function is same as 'keyMap' with user defined splitters as extra
// parameters here.
//
// NOTE:
//  Technically, it forms a map[string]interface{} structure where interface{}
// is a map[string]string.
//
// NOTE:
//  This is intended to be used as a template function
//
// NOTE:
//  This appends current value to original value if any.
//
// Example:
//  {{- "pkey=openebs--stor1=jiva--stor2=cstor" | splitList " " | splitKeyMap "-- =" "vals" .Target | noop -}}
//  {{- "co1=swarm--co2=k8s" | splitList " " | splitKeyMap "-- =" "vals" .Target | noop -}}
//  {{- "pkey=openebs--stor2=mstor" | splitList " " | splitKeyMap "-- =" "vals" .Target | noop -}}
//
// Above will result into following:
//  Target: map[string]interface{}{
//    "vals": map[string]interface{}{
//      "openebs": map[string]interface{}{
//        "stor1": "jiva",
//        "stor2": "cstor, mstor",
//      },
//      "pkey": map[string]interface{}{
//        "co1": "swarm",
//        "co2": "k8s",
//      },
//    },
//  }
//
// The assumption here is '.Target' is of type 'map[string]interface{}'
func splitKeyMap(splitters string, destinationFields string, destination map[string]interface{}, given []string) interface{} {
	var (
		primaryKey string
		key        string
		value      string
		destFields []string
		fields     []string
	)

	// defaultPKey is the default primary key if primary key (to build the
	// maps) is not specified
	defaultPKey := "pkey"
	// defaultPairItemsSplitter is the default delimiter to separate each keyvalue
	// pairs from a given string
	defaultPairItemsSplitter := ","
	// defaultPairSplitter is the default delimiter to split a pair i.e. split
	// a pair's key from its value
	defaultPairSplitter := "="
	// defaultAppendDelimiter is the default delimiter to append current value
	// to existing value at a particular path i.e. key path
	defaultAppendDelimiter := ", "
	// defaultValue is the default value to be set at a particular path i.e. key
	// path if value is empty for a keyvalue pair
	defaultValue := ""

	// destination fields is the path at which maps will be set
	if len(strings.TrimSpace(destinationFields)) != 0 {
		destFields = strings.Split(destinationFields, ".")
	}

	splitterItems := strings.Split(strings.TrimSpace(splitters), " ")
	// pairItemsSplitter to separate keyvalue pairs from a string
	pairItemsSplitter := ""
	// pairSplitter to separate key from its value of one pair
	pairSplitter := ""
	if len(splitterItems) == 2 {
		pairItemsSplitter = strings.TrimSpace(splitterItems[0])
		pairSplitter = strings.TrimSpace(splitterItems[1])
	}

	// default delimiter between pairs, if not set
	if len(pairItemsSplitter) == 0 {
		pairItemsSplitter = defaultPairItemsSplitter
	}

	// default delimiter of a pair, if not set
	if len(pairSplitter) == 0 {
		pairSplitter = defaultPairSplitter
	}

	// givenItem has a list of keyvalue pairs
	for _, givenItem := range given {
		// get all the kv pairs separated via pairItemsSplitter
		pairs := strings.Split(givenItem, pairItemsSplitter)

		// primary key is determined among the pairs
		pKeyPair := pickPrefix(defaultPKey+pairSplitter, pairs)

		// below logic is for setting the primary key value
		pKeyPairs := strings.Split(pKeyPair, pairSplitter)
		if len(pKeyPairs) == 2 {
			primaryKey = pKeyPairs[1]
		} else {
			// default to pkey as the primary key value
			primaryKey = defaultPKey
		}

		if len(primaryKey) == 0 {
			// default to pkey as the primary key value
			primaryKey = defaultPKey
		}

		for _, pair := range pairs {
			if pair == pKeyPair || len(pair) == 0 {
				// primary key value pair has already been considered & nothing
				// to be done for empty pair
				continue
			}

			// split the current pair by pairSplitter
			kvPairs := strings.Split(pair, pairSplitter)
			if len(kvPairs) == 0 {
				continue
			}

			// key value pair is determined here
			if len(kvPairs) == 2 {
				key = strings.TrimSpace(kvPairs[0])
				value = strings.TrimSpace(kvPairs[1])
			} else {
				key = strings.TrimSpace(kvPairs[0])
				value = defaultValue
			}

			if len(key) == 0 {
				// value can not be set at appropriate path if its key is empty
				// hence nothing to be done
				continue
			}

			// reset the path fields first
			fields = nil
			fields = append(destFields, primaryKey, key)

			// append to existing value if any
			origVal := strings.TrimSpace(util.GetNestedString(destination, fields...))
			if len(origVal) != 0 {
				value = strings.Join([]string{origVal, value}, defaultAppendDelimiter)
				// trim for cases where new value is empty
				value = strings.TrimSuffix(value, defaultAppendDelimiter)
			}

			// set the current given item value at destination object
			util.SetNestedField(destination, value, fields...)
		}
	}

	return destination
}

// keyMap builds a list of map of key:value pairs where each map is set
// against a key. It accepts an array where each array item is converted into a
// map. Each map is set against a key known as primary key i.e. pkey. The
// resulting pkey:map pairs are set against the destination object.
//
// NOTE:
//  Technically, it forms a map[string]interface{} structure where interface{}
// is a map[string]string.
//
// NOTE:
//  This is intended to be used as a template function
//
// NOTE:
//  This appends current value to original value if any.
//
// Example:
//  {{- "pkey=openebs,stor1=jiva,stor2=cstor" | splitList " " | keyMap "vals" .Target | noop -}}
//  {{- "co1=swarm,co2=k8s" | splitList " " | keyMap "vals" .Target | noop -}}
//  {{- "pkey=openebs,stor2=mstor" | splitList " " | keyMap "vals" .Target | noop -}}
//
// Above will result into following:
//  Target: map[string]interface{}{
//    "vals": map[string]interface{}{
//      "openebs": map[string]interface{}{
//        "stor1": "jiva",
//        "stor2": "cstor, mstor",
//      },
//      "pkey": map[string]interface{}{
//        "co1": "swarm",
//        "co2": "k8s",
//      },
//    },
//  }
//
// The assumption here is '.Target' is of type 'map[string]interface{}'
func keyMap(destinationFields string, destination map[string]interface{}, given []string) interface{} {
	return splitKeyMap(", =", destinationFields, destination, given)
}

// nestedKeyMap builds a nested map from the given string(s). Each string item is
// split as per the provided set of delimiters and is transformed into a
// hierarchical path that is used to set the value. Value here implies the last
// resulting split item once all the splits are perfomed. The resulting map is
// then set into the provided destination object.
//
// NOTE:
//  This is intended to be used as a template function
//
// NOTE:
//  This appends current value to original value if any.
//
// Example:
//  {{- "default/mypod@app=jiva openebs/mypod@app=cstor" | splitList " " | nestedKeyMap "@ =" .Target | noop -}}
//  {{- "default/mypod@backend=true" | splitList " " | nestedKeyMap "@ =" .Target | noop -}}
//  {{- "litmus/mypod@backend=true" | splitList " " | nestedKeyMap "/ @ =" .Target | noop -}}
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
func nestedKeyMap(delimiters string, destination map[string]interface{}, given []string) interface{} {
	var (
		nestedKeys  []string
		nestedValue string
	)
	// get all the splitters which are separated by a space i.e. " "
	splitters := strings.Split(delimiters, " ")

	for _, givenItem := range given {
		nestedKeys = nil
		nestedValue = givenItem

		for _, splitItem := range splitters {
			// split into two substrings only the current given item by the
			// current split item
			kv := strings.SplitN(nestedValue, splitItem, 2)
			if len(kv) != 2 {
				continue
			}

			k := strings.TrimSpace(kv[0])
			if len(k) != 0 {
				nestedKeys = append(nestedKeys, k)
			}
			nestedValue = strings.TrimSpace(kv[1])
		}

		if len(nestedKeys) == 0 {
			// there is no specific path to set the value
			// hence avoid setting this value
			continue
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

// addTo stores the provided value at specific hierarchy as mentioned in the
// fields inside the values object.
//
// NOTE:
//  This hierarchy along with the provided value is added or appended
// (as comma separated) in the values object.
//
// NOTE:
//  fields is represented as a single string with each field separated by dot
// i.e. '.'
//
// Example:
// {{- "Hi" | addTo "TaskResult.msg" .Values | noop -}}
// {{- "Hello" | addTo "TaskResult.msg" .Values | noop -}}
// {{- .Values.TaskResult.msg -}}
//
// Above will result in printing 'Hi,Hello'
// Assumption here is .Values is of type map[string]interface{}
func addTo(fields string, values map[string]interface{}, value string) string {
	newVal := strings.TrimSpace(value)
	// no need to do anything if provided value is empty
	if len(newVal) == 0 {
		// return what was provided
		return value
	}

	fieldsArr := strings.Split(fields, ".")
	oldValue := util.GetNestedString(values, fieldsArr...)

	// append to the old value if any
	if len(oldValue) != 0 {
		newVal = strings.Join([]string{oldValue, newVal}, ", ")
	}
	util.SetNestedField(values, newVal, fieldsArr...)

	// return what was provided
	return value
}

// saveAs stores the provided value at specific hierarchy as mentioned in the
// fields inside the values object.
//
// NOTE:
//  This hierarchy along with the provided value is added or updated
// (i.e. overriden) in the values object.
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

// notFoundErr throws NotFoundError if given object is empty
//
// NOTE:
//  This function is intended to be used as a go template function
//
// Example:
// {{- "" | notFoundErr "empty object" | saveIf "errMsg" .Values | noop -}}
//
// Above returns NotFoundError during template execution. However this
// does not result in a runtime error.
//
// {{- "I am not empty" | notFoundErr "empty object" | saveIf "errMsg" .Values | noop -}}
//
// Above does not return any error during template execution.
//
// Assumption here is .Values is of type 'map[string]interface{}'
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
// NOTE:
//  This is intended to be used as a go template function
//
// Example:
// {{- "" | empty | verifyErr "empty value provided" | saveIf "errMsg" .Values | noop -}}
//
// Above returns VerifyError during template execution. However this
// does not result in a runtime error.
//
// {{- "I am not empty" | empty | verifyErr "empty value provided" | saveIf "errMsg" .Values | noop -}}
//
// Above does not return any error during template execution.
//
// Assumption here is .Values is of type 'map[string]interface{}'
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

// ToYaml takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
//
// NOTE: Borrowed from a similar function in helm
//  https://github.com/kubernetes/helm/blob/master/pkg/chartutil/files.go
func ToYaml(v interface{}) (yamlstr string) {
	data, err := yaml.Marshal(v)
	if err != nil {
		// error is handled
		yamlstr = fmt.Sprintf("error: %s", err.Error())
		return
	}

	yamlstr = string(data)
	return
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
		"toYaml":       ToYaml,
		"fromYaml":     fromYaml,
		"jsonpath":     jsonPath,
		"saveAs":       saveAs,
		"saveIf":       saveIf,
		"addTo":        addTo,
		"noop":         noop,
		"notFoundErr":  notFoundErr,
		"verifyErr":    verifyErr,
		"isLen":        isLen,
		"nestedKeyMap": nestedKeyMap,
		"keyMap":       keyMap,
		"splitKeyMap":  splitKeyMap,
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
