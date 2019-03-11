package validation

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// ValidateString checks whether the string matches with provided regular
// expression or not
func ValidateString(str, expr string) (bool, error) {
	reg, err := regexp.Compile(expr)
	if err != nil {
		return false, errors.Wrapf(err, "failed to process regular expresion")
	}

	return reg.MatchString(str), nil
}

// Checkforstring returns true if searching string present in given slices
func Checkforstring(stringArray []string, searchStr string) bool {
	for _, str := range stringArray {
		if strings.Contains(str, searchStr) {
			return true
		}
	}
	return false
}
