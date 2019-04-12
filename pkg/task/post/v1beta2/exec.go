/*
Copyright 2019 The OpenEBS Authors

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

package v1beta2

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/ghodss/yaml"
	runtask "github.com/openebs/maya/pkg/apis/openebs.io/runtask/v1beta2"
	deployList "github.com/openebs/maya/pkg/task/post/operations/v1beta2"
	"github.com/openebs/maya/pkg/template"
	flag "github.com/spf13/pflag"
)

const (
	deploymentList         = "deploymentlist"
	asTemplatedBytesOutput = "''"
)

// Post represents the executor struct for runtask's post operations
type Post struct {
	// postTask holds the task's post information
	PostTask       *runtask.Post
	values         map[string]interface{}
	metaID         string
	metaAPIVersion string
	metaKind       string
	metaAction     string
	forFlag        ForFlag
}

// ForFlag specifies the flags supported by for field
// of runtask post field
type ForFlag struct {
	kind       string
	objectPath string
	jsonPath   string
}

// Builder enables building an instance of
// post
type Builder struct {
	Executor *Post
	errors   []error
}

// NewBuilder returns a new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		Executor: &Post{
			PostTask: &runtask.Post{},
			values:   make(map[string]interface{}),
		},
	}
}

// WithTemplate returns a builder instance with post and
// template values
func (b *Builder) WithTemplate(context, yamlTemplate string,
	values map[string]interface{}) *Builder {
	p := &runtask.Post{}
	if len(yamlTemplate) == 0 {
		// nothing needs to be done
		b.errors = append(b.errors, errors.Errorf("empty post yaml given: %+v", yamlTemplate))
		return b
	}
	// transform the yaml with provided templateValues
	t, err := template.AsTemplatedBytes(context, yamlTemplate, values)
	if err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	// TODO: Need to have a better approach to handle this case.
	//
	// Check if string of templated bytes is single quotes ('')
	// if yes, then do nothing and return (this is required
	// to handle the case where only go-templating is being used in runtask)
	if string(t) == asTemplatedBytesOutput {
		return b
	}
	// unmarshal the yaml bytes into Post struct
	err = yaml.Unmarshal(t, p)
	if err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	b.Executor.PostTask = p
	b.Executor.values = values
	return b
}

// WithMetaID returns a builder instance after
// setting runtask meta id in builder
func (b *Builder) WithMetaID(id string) *Builder {
	b.Executor.metaID = id
	return b
}

// WithMetaAPIVersion returns a builder instance after
// setting runtask meta APIVersion in builder
func (b *Builder) WithMetaAPIVersion(apiVersion string) *Builder {
	b.Executor.metaAPIVersion = apiVersion
	return b
}

// WithMetaAction returns a builder instance after
// setting runtask meta action in builder
func (b *Builder) WithMetaAction(action string) *Builder {
	b.Executor.metaAction = action
	return b
}

// WithMetaKind returns a builder instance after
// setting runtask meta kind in builder
func (b *Builder) WithMetaKind(kind string) *Builder {
	b.Executor.metaKind = kind
	return b
}

// Build returns the final instance of post
func (b *Builder) Build() (*Post, error) {
	if len(b.errors) != 0 {
		return nil, errors.Errorf("%v", b.errors)
	}
	return b.Executor, nil
}

// Execute will execute all the runtask post operations
func (p *Post) Execute() error {
	pTask := p.PostTask
	for _, operation := range pTask.Operations {
		operation := operation
		err := p.executeOp(&operation)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Post) executeOp(o *runtask.Operation) (err error) {
	// register and parse the for flags if defined
	err = registerAndParseForFlags(o.For)
	if err != nil {
		return err
	}
	// var result interface{}
	// set value of kind flag of for if its not set
	err = p.setPostKindIfEmpty()
	if err != nil {
		return err
	}
	// set the dataPath to default dataPath i.e.
	// RuntimeObject (top-level property for saving current
	// runtask data)
	dataPath := p.setDataPathIfEmpty()
	run := strings.ToLower(o.Run)
	switch strings.ToLower(p.forFlag.kind) {
	case deploymentList:
		dlist, err := deployList.
			BuilderForPostValues(run, o.WithFilter, o.WithOutput).
			WithDataPath(dataPath).
			WithTemplateValues(p.values).
			Build()
		if err != nil {
			return err
		}
		_, err = dlist.ExecuteOp()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported kind for runtask post operation, %s", p.forFlag.kind)
	}
	// TODO: Save this result at path specified by 'as' field of runtask post
	as := o.As
	fmt.Println(as)
	return nil
}

func (p *Post) setPostKindIfEmpty() error {
	// Check if the kind flag has been set, if not
	// then use the kind set in meta of runtask
	// Note: For list action, kind would be kind + action
	// i.e. action = list and kind = deployment then kind flag
	// will be set as deploymentlist
	if p.forFlag.kind == "" {
		if p.metaAction == "" || p.metaKind == "" {
			return errors.Errorf("Empty value for kind or action found, kind: %s, action: %s", p.metaKind, p.metaAction)
		}
		if p.metaAction == "list" {
			p.forFlag.kind = p.metaKind + p.metaAction
		} else {
			p.forFlag.kind = p.metaKind
		}
	}
	return nil
}

func (p *Post) setDataPathIfEmpty() (dataPath string) {
	// Check if the dataPath is set or not, if not
	// set the dataPath as runtime.Object top-level property
	// i.e. RuntimeObject or json result top-level property i.e.
	// JsonResult
	if p.forFlag.jsonPath != "" {
		dataPath = p.forFlag.jsonPath
	} else if p.forFlag.objectPath != "" {
		dataPath = p.forFlag.objectPath
	} else {
		// If none of the above flags are set then
		// use "RuntimeObject" as the top-level property
		// to get data
		dataPath = string(runtask.CurrentRuntimeObjectTLP)
	}
	return dataPath
}

func registerAndParseForFlags(For []string) error {
	ff := ForFlag{}
	// forFs is the flagset having all the flags defined for "For" field
	// of post operations
	forFs := flag.NewFlagSet("for", flag.ExitOnError)
	forFs.StringVar(&ff.kind, "kind", "", "represents the kind of resource")
	forFs.StringVar(&ff.objectPath, "objectPath", "", "represents the runtime object path for fetching the details")
	forFs.StringVar(&ff.jsonPath, "jsonPath", "", "represents the jsonResult path for fetching the details")
	if err := forFs.Parse(For); err != nil {
		return errors.Errorf("failed to parse post forFlags: given forFlags: %s, error: %v", For, err)
	}
	return nil
}
