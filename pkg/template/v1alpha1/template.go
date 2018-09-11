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
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	stdtemplate "text/template"
)

// template encapsulates standard template instance
type template struct {
	tpl *stdtemplate.Template
}

// templateMiddleware abstracts updating template instance
type templateMiddleware func(given *template) (updated *template)

// StandardTemplate updates the template instance with go standard templating
// features
func StandardTemplate(context string) templateMiddleware {
	return func(given *template) (updated *template) {
		if given == nil {
			return given
		}
		if len(context) == 0 {
			context = "standardtpl"
		}
		if given.tpl == nil {
			given.tpl = stdtemplate.New(context)
		}
		return given
	}
}

// SprigTemplate updates the template instance with sprig based templating features
func SprigTemplate(context string) templateMiddleware {
	return func(given *template) (updated *template) {
		if len(context) == 0 {
			context = "sprigtpl"
		}
		given = StandardTemplate(context)(given)
		if given == nil || given.tpl == nil {
			return given
		}
		given.tpl.Funcs(sprig.TxtFuncMap())
		return given
	}
}

// Templater abstracts executing a template and returning the templated
// result
type Templater func(context, given string, values map[string]interface{}) (updated string, err error)

// TextTemplate executes the provided template document against the provided
// template values
//
// NOTE:
//  This is an implementation of Templater
func TextTemplate(context, given string, values map[string]interface{}) (updated string, err error) {
	t := SprigTemplate(context)(&template{})
	if t == nil || t.tpl == nil {
		err = fmt.Errorf("failed to initialize templating: failed to text template '%s' '%s'", context, given)
		return
	}

	p, err := t.tpl.Parse(string(given))
	if err != nil {
		err = errors.Wrapf(err, "failed to parse '%s' text template '%s'", context, given)
		return
	}

	// buf is an io.Writer implementation as required by the template
	var buf bytes.Buffer

	// go template the parsed yaml against the provided template values & write
	// the result into the buffer
	err = p.Execute(&buf, values)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute '%s' text template '%s' against '%+v'", context, given, values)
		return
	}

	return buf.String(), nil
}
