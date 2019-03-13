package validation

import (
	"regexp"

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
