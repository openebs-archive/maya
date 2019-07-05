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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	extnv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// StoragePoolKindCSPC holds the value of CStorPoolCluster
	StoragePoolKindCSPC = "CStorPoolCluster"
	// APIVersion holds the value of OpenEBS version
	APIVersion = "openebs.io/v1alpha1"
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
	Object *extnv1beta1.Deployment
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
	return stringer.Yaml("deployment", d.Object)
}

// GoString implements the goStringer interface
func (d *Deploy) GoString() string {
	return d.String()
}

// NewBuilder returns a new instance of builder meant for deployment
func NewBuilder() *Builder {
	return &Builder{
		deployment: &Deploy{
			Object: &extnv1beta1.Deployment{},
		},
	}
}

func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: missing name"),
		)
		return b
	}
	b.deployment.Object.Name = name
	return b
}

func (b *Builder) WithNameSpace(ns string) *Builder {
	if len(ns) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: missing namespace"),
		)
		return b
	}
	b.deployment.Object.Namespace = ns
	return b
}

func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: missing labels"),
		)
		return b
	}
	b.deployment.Object.Labels = labels
	return b
}

func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: missing annotations"),
		)
		return b
	}
	b.deployment.Object.Annotations = annotations
	return b
}

func (b *Builder) WithOwnerReferences(csp *apis.NewTestCStorPool) *Builder {
	if csp == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: csp object is nil"),
		)
		return b
	}
	trueVal := true
	reference := metav1.OwnerReference{
		APIVersion:         APIVersion,
		Kind:               StoragePoolKindCSPC,
		UID:                csp.ObjectMeta.UID,
		Name:               csp.ObjectMeta.Name,
		BlockOwnerDeletion: &trueVal,
		Controller:         &trueVal,
	}
	b.deployment.Object.OwnerReferences = append(b.deployment.Object.OwnerReferences, reference)
	return b
}

func (b *Builder) WithReplicaCount(count *int32) *Builder {
	if count == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: replica count is nil"),
		)
		return b
	}
	b.deployment.Object.Spec.Replicas = count
	return b
}

func (b *Builder) WithDeploymentStrategy(strategy extnv1beta1.DeploymentStrategyType) *Builder {
	if strategy == "" {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: missing strategy type"),
		)
		return b
	}
	b.deployment.Object.Spec.Strategy.Type = strategy
	return b
}

func (b *Builder) WithSelector(selector *metav1.LabelSelector) *Builder {
	if selector == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: missing label selectors"),
		)
		return b
	}
	b.deployment.Object.Spec.Selector = selector
	return b
}

func (b *Builder) WithPodTemplateSpec(template *v1.PodTemplateSpec) *Builder {
	if template == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build deployment object: missing pod template spec"),
		)
		return b
	}
	b.deployment.Object.Spec.Template = *template
	return b
}

// NewBuilderForAPIObject returns a new instance of builder
// for a given deployment object
func NewBuilderForAPIObject(deployment *extnv1beta1.Deployment) *Builder {
	b := NewBuilder()
	if deployment != nil {
		b.deployment.Object = deployment
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
			b.deployment.Object)
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
		b.AddCheck(check)
	}
	return b
}

// IsRollout range over rolloutChecks map and check status of each predicate
// also it generates status message from rolloutStatuses using predicate key
func (d *Deploy) IsRollout() (PredicateName, bool) {
	for pk, p := range RolloutChecks {
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
func (d *Deploy) SuccessRollout() *RolloutOutput {
	return &RolloutOutput{
		Message:     "deployment successfully rolled out",
		IsRolledout: true,
	}
}

// RolloutStatus returns rollout message of deployment instance
func (d *Deploy) RolloutStatus() (op *RolloutOutput, err error) {
	pk, ok := d.IsRollout()
	if ok {
		return d.SuccessRollout(), nil
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
	for _, cond := range d.Object.Status.Conditions {
		if cond.Type == extnv1beta1.DeploymentProgressing &&
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
	return d.Object.Spec.Replicas != nil && d.Object.Status.UpdatedReplicas < *d.Object.Spec.Replicas
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
	return d.Object.Status.Replicas > d.Object.Status.UpdatedReplicas
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
	return d.Object.Status.AvailableReplicas < d.Object.Status.UpdatedReplicas
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
	return d.Object.Generation > d.Object.Status.ObservedGeneration
}
