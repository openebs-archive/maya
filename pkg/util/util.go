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

package util

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

var (
	// ErrMAPIADDRNotSet is the new error to display this error if MAPI_ADDR is not set.
	ErrMAPIADDRNotSet = errors.New("MAPI_ADDR environment variable not set")
	// ErrInternalServerError is the new error to raise if an error occurs while rendering the service
	ErrInternalServerError = errors.New("Internal Server Error")
	// ErrServerUnavailable is the new error to raise if the server is not available
	ErrServerUnavailable = errors.New("Server Unavailable")
	// ErrServerNotReachable is the new error to raise if the server is not reachable
	ErrServerNotReachable = errors.New("Server Not Reachable")
	// ErrPageNotFound is the new error to raise if the page is not found
	ErrPageNotFound = errors.New("Page Not Found")
)

// truthyValues maps a set of values which are considered as true
var truthyValues = map[string]bool{
	"1":       true,
	"YES":     true,
	"TRUE":    true,
	"OK":      true,
	"ENABLED": true,
	"ON":      true,
}

// CheckTruthy checks for truthiness of the passed argument.
func CheckTruthy(truth string) bool {
	return truthyValues[strings.ToUpper(truth)]
}

// falsyValues maps a set of values which are considered as false
var falsyValues = map[string]bool{
	"0":        true,
	"NO":       true,
	"FALSE":    true,
	"BLANK":    true,
	"DISABLED": true,
	"OFF":      true,
}

// CheckFalsy checks for non-truthiness of the passed argument.
func CheckFalsy(falsy string) bool {
	if len(falsy) == 0 {
		falsy = "blank"
	}
	return falsyValues[strings.ToUpper(falsy)]
}

// CheckErr to handle command errors
func CheckErr(err error, handleErr func(string)) {
	if err == nil {
		return
	}
	handleErr(err.Error())
}

// Fatal prints the message (if provided) and then exits. If V(2) or greater,
// glog.Fatal is invoked for extended information.
func Fatal(msg string) {
	if glog.V(2) {
		glog.FatalDepth(2, msg)
	}
	if len(msg) > 0 {
		// add newline if needed
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Fprint(os.Stderr, msg)
	}
	os.Exit(1)
}

// StringToInt32 converts a string type to corresponding
// *int32 type
func StringToInt32(val string) (*int32, error) {
	if len(val) == 0 {
		return nil, fmt.Errorf("Nil value to convert")
	}

	n, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return nil, err
	}
	n32 := int32(n)
	return &n32, nil
}

// StrToInt32 converts a string type to corresponding
// *int32 type
//
// NOTE:
//  This swallows the error if any
func StrToInt32(val string) *int32 {
	n32, _ := StringToInt32(val)
	return n32
}

// ContainsString returns true if the provided element is present in the
// provided array
func ContainsString(stringarr []string, element string) bool {
	for _, elem := range stringarr {
		if elem == element {
			return true
		}
	}
	return false
}

// ContainsKey returns true if the provided key is present in the provided map
func ContainsKey(mapOfObjs map[string]interface{}, key string) bool {
	for k := range mapOfObjs {
		if k == key {
			return true
		}
	}
	return false
}

// ContainKeys returns true if all the provided keys are present in the
// provided map
func ContainKeys(mapOfObjs map[string]interface{}, keys []string) bool {
	if len(keys) == 0 || len(mapOfObjs) == 0 {
		return false
	}

	allKeys := []string{}
	for k := range mapOfObjs {
		allKeys = append(allKeys, k)
	}

	for _, expectedKey := range keys {
		if !ContainsString(allKeys, expectedKey) {
			return false
		}
	}

	return true
}

// MergeMaps merges maps and returns the resulting map.
// map priority increases with order i.e. MergeMaps(m1,m2)
// will result in a map with overriding values from m2
func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
