/*
Copyright 2018 The OpenEBS Authors

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

package v1alpha1

import (
	"bytes"
	"github.com/pkg/errors"
	"text/template"
)

// TextTemplater abstracts executing a template and returning the templated
// result
type TextTemplater func(context, given string, values map[string]interface{}) (updated string, err error)

// TemplateIt executes the provided template document against the provided
// template values
//
// NOTE:
//  This is an implementation of TextTemplater
func TemplateIt(context, given string, values map[string]interface{}) (updated string, err error) {
	t := template.New(context)
	t, err = t.Parse(string(given))
	if err != nil {
		err = errors.Wrapf(err, "failed to parse '%s' text template '%s'", context, given)
		return
	}

	// buf is an io.Writer implementation as required by the template
	var buf bytes.Buffer

	// go template the parsed yaml against the provided template values & write
	// the result into the buffer
	err = t.Execute(&buf, values)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute '%s' text template '%s' against '%+v'", context, given, values)
		return
	}

	return buf.String(), nil
}
