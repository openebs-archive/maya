package util

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

// truthyValues maps a set of values which are considered as true
var truthyValues = map[string]bool{
	"1":    true,
	"YES":  true,
	"TRUE": true,
	"OK":   true,
}

// CheckTruthy checks for truthiness of the passed argument.
func CheckTruthy(truth string) bool {
	return truthyValues[strings.ToUpper(truth)]
}

// falsyValues maps a set of values which are considered as false
var falsyValues = map[string]bool{
	"0":     true,
	"NO":    true,
	"FALSE": true,
	"BLANK": true,
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
