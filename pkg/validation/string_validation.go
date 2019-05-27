// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
