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

package v1alpha1

import (
	"strings"

	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	"github.com/pkg/errors"
	extn_v1_beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Predicate abstracts conditional logic w.r.t the deployment instance
//
// NOTE:
// predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*Deploy) bool

// Deploy is the wrapper over k8s deployment object
type Deploy struct {
	// kubernetes deployment instance
	object *extn_v1_beta1.Deployment
}

// Builder enables building an instance of
// deployment
type Builder struct {
	deployment *Deploy     // kubernetes deployment instance
	checks     []Predicate // predicate list for deploy
	errors     []error
}

// PredicateName type is wrapper over string.
// It is used to refer predicate and status msg.
type PredicateName string

const (
	// PredicateProgressDeadlineExceeded refer to
	// predicate IsProgressDeadlineExceeded.
	PredicateProgressDeadlineExceeded PredicateName = "ProgressDeadlineExceeded"
	// PredicateNotSpecSynced refer to predicate IsNotSpecSynced
	PredicateNotSpecSynced PredicateName = "NotSpecSynced"
	// PredicateOlderReplicaActive refer to predicate IsOlderReplicaActive
	PredicateOlderReplicaActive PredicateName = "OlderReplicaActive"
	// PredicateTerminationInProgress refer to predicate IsTerminationInProgress
	PredicateTerminationInProgress PredicateName = "TerminationInProgress"
	// PredicateUpdateInProgress refer to predicate IsUpdateInProgress.
	PredicateUpdateInProgress PredicateName = "UpdateInProgress"
)

// String implements the stringer interface
func (d *Deploy) String() string {
	return stringer.Yaml("deployment", d.object)
}

// GoString implements the goStringer interface
func (d *Deploy) GoString() string {
	return d.String()
}

// NewBuilder returns a new instance of builder meant for deployment
func NewBuilder() *Builder {
	return &Builder{
		deployment: &Deploy{
			object: &extn_v1_beta1.Deployment{},
		},
	}
}

// NewBuilderForAPIObject returns a new instance of builder
// for a given deployment object
func NewBuilderForAPIObject(deployment *extn_v1_beta1.Deployment) *Builder {
	b := NewBuilder()
	if deployment != nil {
		b.deployment.object = deployment
	} else {
		b.errors = append(b.errors,
			errors.New("nil deployment object given to get builder instance"))
	}
	return b
}

// Build returns a deployment instance
func (b *Builder) Build() (*Deploy, error) {
	err := b.validate()
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to build a deployment instance: %s",
			b.deployment.object)
	}
	return b.deployment, nil
}

func (b *Builder) validate() error {
	if len(b.errors) != 0 {
		return errors.Errorf("failed to validate: build errors were found: %v", b.errors)
	}
	return nil
}

// AddCheck adds the predicate as a condition to be validated
// against the deployment instance
func (b *Builder) AddCheck(p Predicate) *Builder {
	b.checks = append(b.checks, p)
	return b
}

// AddChecks adds the provided predicates as conditions to be
// validated against the deployment instance
func (b *Builder) AddChecks(p []Predicate) *Builder {
	for _, check := range p {
		check := check
		b.AddCheck(check)
	}
	return b
}

// IsRollout range over rolloutChecks map and check status of each predicate
// also it generates status message from rolloutStatuses using predicate key
func (d *Deploy) IsRollout() (PredicateName, bool) {
	for pk, p := range RolloutChecks {
		pk := pk
		p := p
		if p(d) {
			return pk, false
		}
	}
	return "", true
}

// FailedRollout returns rollout status message for fail condition
func (d *Deploy) FailedRollout(name PredicateName) *RolloutOutput {
	return &RolloutOutput{
		Message:     RolloutStatuses[name](d),
		IsRolledout: false,
	}
}

// SuccessRollout returns rollout status message for success condition
func (d *Deploy) SuccessRollout(name PredicateName) *RolloutOutput {
	return &RolloutOutput{
		Message:     "deployment successfully rolled out",
		IsRolledout: true,
	}
}

// RolloutStatus returns rollout message of deployment instance
func (d *Deploy) RolloutStatus() (op *RolloutOutput, err error) {
	pk, ok := d.IsRollout()
	if ok {
		return d.SuccessRollout(pk), nil
	}
	return d.FailedRollout(pk), nil
}

// RolloutStatusRaw returns rollout message of deployment instance
// in byte format
func (d *Deploy) RolloutStatusRaw() (op []byte, err error) {
	message, err := d.RolloutStatus()
	if err != nil {
		return nil, err
	}
	return NewRollout(
		withOutputObject(message)).
		Raw()
}

// IsProgressDeadlineExceeded is used to check update is timed out or not.
// If `Progressing` condition's reason is `ProgressDeadlineExceeded` then
// it is not rolled out.
func IsProgressDeadlineExceeded() Predicate {
	return func(d *Deploy) bool {
		return d.IsProgressDeadlineExceeded()
	}
}

// IsProgressDeadlineExceeded is used to check update is timed out or not.
// If `Progressing` condition's reason is `ProgressDeadlineExceeded` then
// it is not rolled out.
func (d *Deploy) IsProgressDeadlineExceeded() bool {
	for _, cond := range d.object.Status.Conditions {
		cond := cond
		if cond.Type == extn_v1_beta1.DeploymentProgressing &&
			cond.Reason == "ProgressDeadlineExceeded" {
			return true
		}
	}
	return false
}

// IsOlderReplicaActive check if older replica's are still active or not
// if Status.UpdatedReplicas < *Spec.Replicas then some of the replicas
// are updated and some of them are not.
func IsOlderReplicaActive() Predicate {
	return func(d *Deploy) bool {
		return d.IsOlderReplicaActive()
	}
}

// IsOlderReplicaActive check if older replica's are still active or not
// if Status.UpdatedReplicas < *Spec.Replicas then some of the replicas
// are updated and some of them are not.
func (d *Deploy) IsOlderReplicaActive() bool {
	return d.object.Spec.Replicas != nil && d.object.Status.UpdatedReplicas < *d.object.Spec.Replicas
}

// IsTerminationInProgress checks for older replicas are waiting to
// terminate or not. If Status.Replicas > Status.UpdatedReplicas then
// some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to
// come into running state then terminate.
func IsTerminationInProgress() Predicate {
	return func(d *Deploy) bool {
		return d.IsTerminationInProgress()
	}
}

// IsTerminationInProgress checks for older replicas are waiting to
// terminate or not. If Status.Replicas > Status.UpdatedReplicas then
// some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to
// come into running state then terminate.
func (d *Deploy) IsTerminationInProgress() bool {
	return d.object.Status.Replicas > d.object.Status.UpdatedReplicas
}

// IsUpdateInProgress Checks if all the replicas are updated or not. If
// Status.AvailableReplicas < Status.UpdatedReplicas then all the older
// replicas are not there but there are less number of availableReplicas
func IsUpdateInProgress() Predicate {
	return func(d *Deploy) bool {
		return d.IsUpdateInProgress()
	}
}

// IsUpdateInProgress Checks if all the replicas are updated or not. If
// Status.AvailableReplicas < Status.UpdatedReplicas then all the older
// replicas are not there but there are less number of availableReplicas
func (d *Deploy) IsUpdateInProgress() bool {
	return d.object.Status.AvailableReplicas < d.object.Status.UpdatedReplicas
}

// IsNotSyncSpec compare generation in status and spec and check if deployment
// spec is synced or not. If Generation <= Status.ObservedGeneration then
// deployment spec is not updated yet.
func IsNotSyncSpec() Predicate {
	return func(d *Deploy) bool {
		return d.IsNotSyncSpec()
	}
}

// IsNotSyncSpec compare generation in status and spec and check if deployment
// spec is synced or not. If Generation <= Status.ObservedGeneration then
// deployment spec is not updated yet.
func (d *Deploy) IsNotSyncSpec() bool {
	return d.object.Generation > d.object.Status.ObservedGeneration
}

// DeployList is the list of deployments
type DeployList struct {
	items []*Deploy
}

// ListBuilder enables building an instance of
// deploymentList
type ListBuilder struct {
	list *DeployList
	//output []map[string]interface{}
	output  outputMap
	filters filterList
	errors  []error
}

// filterList is a list of filters
type filterList []Filter

// outputMap is a map of desired output function
// and their reference key
type outputMap map[*Output]string

// Filter abstracts filtering logic w.r.t the
// deployment list instance
type Filter func(*Deploy) bool

// Output represents the desired output to
// be given against a deployment instance
type Output func(*Deploy) interface{}

// NewListBuilder returns an instance of
// list builder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{
		list:   &DeployList{},
		output: outputMap{},
	}
}

// ListBuilderForRuntimeObject returns a list builder instance
// for deployment list
func ListBuilderForRuntimeObject(obj runtime.Object) *ListBuilder {
	var (
		dl *extn_v1_beta1.DeploymentList
		ok bool
	)
	lb := NewListBuilder()
	if obj == nil {
		lb.errors = append(lb.errors, errors.New("failed to build instance: nil runtime.Object given"))
		return lb
	}
	// Convert the runtime.Object to its desired type i.e.
	// DeploymentList here
	if dl, ok = obj.(*extn_v1_beta1.DeploymentList); !ok {
		lb.errors = append(lb.errors, errors.New(
			"failed to build instance: unable to typecast given object to deployment list"))
		return lb
	}
	// Iterate over deployment list objects and
	// insert it into the ListBuilder instance
	for _, d := range dl.Items {
		d := d
		deploy := &Deploy{}
		deploy.object = &d
		lb.list.items = append(lb.list.items, deploy)
	}
	return lb
}

// List returns a list of deployments after doing
// all the filtering and validations
func (lb *ListBuilder) List() (*DeployList, error) {
	if len(lb.errors) != 0 {
		return nil, errors.Errorf("%v", lb.errors)
	}
	if lb.filters == nil || len(lb.filters) == 0 {
		return lb.list, nil
	}
	filtered := &DeployList{}
	for _, d := range lb.list.items {
		d := d
		if lb.filters.all(d) {
			filtered.items = append(filtered.items, d)
		}
	}
	return filtered, nil
}

// all returns true if all the predicates
// succeed against the provided deployment
// instance
func (f filterList) all(d *Deploy) bool {
	for _, filter := range f {
		filter := filter
		if !filter(d) {
			return false
		}
	}
	return true
}

// AddFilter adds the filter to be applied against the
// deployment instance
func (lb *ListBuilder) AddFilter(f Filter) *ListBuilder {
	lb.filters = append(lb.filters, f)
	return lb
}

// AddFilters adds the provided filters to be applied against
// the deployment instance
func (lb *ListBuilder) AddFilters(filters ...Filter) *ListBuilder {
	for _, filter := range filters {
		filter := filter
		lb.AddFilter(filter)
	}
	return lb
}

// HasLabel returns HasLabel filter for
// the given label
func HasLabel(label string) Filter {
	return func(d *Deploy) bool {
		return d.HasLabel(label)
	}
}

// HasLabel checks if the given label is
// present or not for a particular deployment
// object
func (d *Deploy) HasLabel(label string) bool {
	labels := d.object.GetLabels()
	if _, exist := labels[label]; exist {
		return true
	}
	return false
}

// HasLabels returns IsLabel filter for
// the given label
func HasLabels(labels ...string) Filter {
	return func(d *Deploy) bool {
		return d.HasLabels(labels...)
	}
}

// HasLabels checks if the given labels are
// present or not for a particular deployment
// object
func (d *Deploy) HasLabels(labels ...string) bool {
	const (
		keyIndex   = 0
		valueIndex = 1
	)
	gotLabels := d.object.GetLabels()
	for _, label := range labels {
		label := label
		var labelDetails []string
		// get the label key and value by splitting
		// it based on delimiter '=' or ':'
		if strings.Contains(label, "=") {
			labelDetails = strings.Split(label, "=")
		} else {
			labelDetails = strings.Split(label, ":")
		}
		if lValue, exist := gotLabels[labelDetails[keyIndex]]; exist {
			if lValue != labelDetails[valueIndex] {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// Name returns an output instance
// for getting name of a deployment
func Name() Output {
	return func(d *Deploy) interface{} {
		return d.Name()
	}
}

// Name returns the name of the given deployment
func (d *Deploy) Name() interface{} {
	return d.object.GetName()
}

// Namespace returns an output instance
// for getting namespace of a deployment
func Namespace() Output {
	return func(d *Deploy) interface{} {
		return d.Namespace()
	}
}

// Namespace returns the namespace of the given deployment
func (d *Deploy) Namespace() interface{} {
	return d.object.GetNamespace()
}

// Labels returns an output instance
// for getting labels of a deployment
func Labels() Output {
	return func(d *Deploy) interface{} {
		return d.Labels()
	}
}

// Labels returns the labels of the given deployment
func (d *Deploy) Labels() interface{} {
	return d.object.GetLabels()
}

// WithOutput returns a listBuilder instance having
// the desired output key added
func (lb *ListBuilder) WithOutput(o Output, referenceKey string) *ListBuilder {
	if o == nil || referenceKey == "" {
		lb.errors = append(lb.errors, errors.Errorf(
			"nil reference key given for output %v", o))
		return lb
	}
	lb.output[&o] = referenceKey
	return lb
}

// TupleList enables building a tuple list instance
// against a deployment list
type TupleList []map[string]interface{}

// GetTupleList returns a tuple list based on the desired
// outputs provided
func (lb *ListBuilder) GetTupleList() (TupleList, error) {
	tList := TupleList{}
	dList, err := lb.List()
	if err != nil {
		return nil, errors.Errorf(
			"failed to get tuple list: error: %v", err)
	}
	for _, d := range dList.items {
		d := d
		deployDetails := d.getDesiredDetails(lb.output)
		tList = append(tList, deployDetails)
	}
	return tList, nil
}

func (d *Deploy) getDesiredDetails(desiredDetails outputMap) map[string]interface{} {
	deployDetails := make(map[string]interface{})
	for out, refKey := range desiredDetails {
		out := out
		refKey := refKey
		i := (*out)(d)
		deployDetails[refKey] = i
	}
	return deployDetails
}
