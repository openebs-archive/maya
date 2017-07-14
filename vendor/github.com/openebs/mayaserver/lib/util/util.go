package util

import (
	"strings"
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
