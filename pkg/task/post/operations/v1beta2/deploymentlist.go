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

	"k8s.io/apimachinery/pkg/runtime"

	deployment_extnv1beta1 "github.com/openebs/maya/pkg/kubernetes/deployment/extnv1beta1/v1alpha1"
	"github.com/openebs/maya/pkg/task/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
)

const (
	getTupleList = "gettuplelist"
)

// DeploymentList represents the details for fetching
// the desired values from a deployment list object
type DeploymentList struct {
	Run          string
	DataPath     string
	Values       map[string]interface{}
	WithFilterFs *filter
	WithOutputFs *output
}

// Builder enables building an instance of
// DeploymentList
type Builder struct {
	DeploymentList *DeploymentList
	errors         []error
}

type output struct {
	name      bool
	namespace bool
}

type filter struct {
	isLabel isLabels
}

type isLabels struct {
	labels []string
}

func (labelArr *isLabels) String() string {
	return fmt.Sprint(labelArr.labels)
}

func (labelArr *isLabels) Set(value string) error {
	if value == "" {
		return nil
	}
	labelArr.labels = append(labelArr.labels, value)
	return nil
}

func (labelArr *isLabels) Type() string {
	return "isLabels"
}

// NewBuilder returns a new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		DeploymentList: &DeploymentList{
			Values:       make(map[string]interface{}),
			WithFilterFs: &filter{},
			WithOutputFs: &output{},
		},
	}
}

// BuilderForPostValues returns a builder instance
// after filling the required post values
func BuilderForPostValues(run string,
	withFilter []string, withOutput []string) *Builder {
	b := NewBuilder()
	f := &filter{}
	o := &output{}

	// withFilter is the flagset having all the flags defined for "withFilter" field
	// of post operations
	withFilterFs := flag.NewFlagSet("withFilter", flag.ExitOnError)
	withFilterFs.Var(&f.isLabel, "isLabel", "checks for given labels against a resource")

	// withOutput is the flagset having all the flags defined for "withOutput" field
	// of post operations
	withOutputFs := flag.NewFlagSet("withOutput", flag.ExitOnError)
	withOutputFs.BoolVar(&o.name, "name", false, "set if output should contain resource name")
	withOutputFs.BoolVar(&o.namespace, "namespace", false, "set if output should contain resource namespace")

	if err := withFilterFs.Parse(withFilter); err != nil {
		b.errors = append(b.errors, errors.Errorf("error parsing withFilter flags, error: %v", err))
	}
	if err := withOutputFs.Parse(withOutput); err != nil {
		b.errors = append(b.errors, errors.Errorf("error parsing withOutput flags, error: %v", err))
	}
	b.DeploymentList.WithFilterFs = f
	b.DeploymentList.WithOutputFs = o
	b.DeploymentList.Run = run
	return b
}

// WithDataPath returns a builder instance with
// objectPath/jsonPath of the saved result
func (b *Builder) WithDataPath(path string) *Builder {
	// Trim unnecessary spaces or dots
	path = strings.Trim(strings.TrimSpace(path), ".")
	b.DeploymentList.DataPath = path
	return b
}

// WithTemplateValues returns a builder instance with
// templateValues map
func (b *Builder) WithTemplateValues(values map[string]interface{}) *Builder {
	b.DeploymentList.Values = values
	return b
}

// Build returns the final instance of post
func (b *Builder) Build() (*DeploymentList, error) {
	if len(b.errors) != 0 {
		return nil, errors.Errorf("%v", b.errors)
	}
	return b.DeploymentList, nil
}

// ExecuteOp executes the post operation on a
// deploymentList instance
func (d *DeploymentList) ExecuteOp() (result interface{}, err error) {
	switch d.Run {
	case getTupleList:
		result, err = d.getTupleList()
	default:
		return result, errors.Errorf(
			"unsupported runtask post operation, `%s` for deploymentList", d.Run)
	}
	if err != nil {
		return result, err
	}
	return result, nil
}

func (d *DeploymentList) getTupleList() (tList []map[string]interface{}, err error) {
	var (
		dListObj runtime.Object
		ok       bool
		lb       *deployment_extnv1beta1.ListBuilder
	)
	fields := strings.Split(d.DataPath, ".")
	// TODO: Need to have a function to get runtime.Object
	// directly instead of interface{}
	data := util.GetNestedField(d.Values, fields...)
	if dListObj, ok = data.(runtime.Object); !ok {
		return nil, errors.New("failed to get tuple list: list given is not runtime.Object type")
	}
	// Form the build instance
	lb = deployment_extnv1beta1.
		ListBuilderForRuntimeObject(dListObj)
	// Call the required filters and output build functions
	// based on provided flags
	lb = d.addFuncForFlags(lb)
	// Call the target function now i.e. tupleList for getting the list
	// of tuples
	tList, err = lb.GetTupleList()
	if err != nil {
		return nil, err
	}
	return tList, nil
}

func (d *DeploymentList) addFuncForFlags(
	lb *deployment_extnv1beta1.ListBuilder) *deployment_extnv1beta1.ListBuilder {
	// Check for the withFilter flags which has been set and
	// accordingly call the corresponding build functions
	//
	// Check if the isLabel flag is set or not
	if len(d.WithFilterFs.isLabel.labels) != 0 {
		lb = lb.AddFilter(deployment_extnv1beta1.HasLabels(d.WithFilterFs.isLabel.labels...))
	}
	// Check for the withOutput flags that has been set and
	// accordingly call the corresponding build functions
	//
	// Check if the name flag is set or not
	if d.WithOutputFs.name {
		lb = lb.WithOutput(deployment_extnv1beta1.Name(), "name")
	}
	//Check if the namespace flag is set
	if d.WithOutputFs.namespace {
		lb = lb.WithOutput(deployment_extnv1beta1.Namespace(), "namespace")
	}
	return lb
}

// TODO: Need to have these functions at a common place
//
// saveAs stores the provided value at specific hierarchy as mentioned in the
// fields inside the values object.
//
// NOTE:
//  This hierarchy along with the provided value is added or updated
// (i.e. overridden) in the values object.
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
func saveAs(fields string, destination map[string]interface{}, given interface{}) interface{} {
	fieldsArr := strings.Split(fields, ".")
	// save the run task command result in specific way
	r, ok := given.(v1alpha1.RunCommandResult)
	if ok {
		resultpath := append(fieldsArr, "result")
		util.SetNestedField(destination, r.Result(), resultpath...)
		errpath := append(fieldsArr, "error")
		util.SetNestedField(destination, r.Error(), errpath...)
		debugpath := append(fieldsArr, "debug")
		util.SetNestedField(destination, r.Debug(), debugpath...)
		return given
	}
	util.SetNestedField(destination, given, fieldsArr...)
	return given
}
