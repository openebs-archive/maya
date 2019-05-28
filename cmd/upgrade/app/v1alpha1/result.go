/*
Copyright 2019 The OpenEBS Authors.

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
	upgrade "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	objectmeta "github.com/openebs/maya/pkg/kubernetes/objectmeta/v1alpha1"
	typemeta "github.com/openebs/maya/pkg/kubernetes/typemeta/v1alpha1"
	upgraderesult "github.com/openebs/maya/pkg/upgrade/result/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// labelJobName label contains name of job in which upgrade
	// process is running
	labelJobName = "upgradejob.openebs.io/name"
	// labelItemName label contains name of unit of upgrade
	labelItemName = "upgradeitem.openebs.io/name"
	// labelItemNamespace label contains namespace of unit of upgrade
	labelItemNamespace = "upgradeitem.openebs.io/namespace"
	// labelItemKind label contains kind of unit of upgrade
	labelItemKind = "upgradeitem.openebs.io/kind"
)

// UpgradeResult is a wrapper over upgrade.UpgradeResult struct
type UpgradeResult struct {
	object *upgrade.UpgradeResult
}

// UpgradeResultGetOrCreateBuilder helps to get or create UpgradeResult instance
type UpgradeResultGetOrCreateBuilder struct {
	*errors.ErrorList
	SelfNamespace   string
	Owner           *metav1.OwnerReference      // owner reference for upgrade result cr
	UpgradeConfig   *upgrade.UpgradeConfig      // runtime config for upgrade
	ResourceDetails *upgrade.ResourceDetails    // unit of upgrade details
	Tasks           []upgrade.UpgradeResultTask // list of runtasks used to upgrade a resource
	UpgradeResult   *UpgradeResult
}

// String implements GoStringer interface
func (urb *UpgradeResultGetOrCreateBuilder) String() string {
	return stringer.Yaml("upgrade result get or create builder", urb)
}

// GoString implements GoStringer interface
func (urb *UpgradeResultGetOrCreateBuilder) GoString() string {
	return urb.String()
}

// NewUpgradeResultGetOrCreateBuilder returns a new UpgradeResult instance
func NewUpgradeResultGetOrCreateBuilder() *UpgradeResultGetOrCreateBuilder {
	return &UpgradeResultGetOrCreateBuilder{
		UpgradeResult: &UpgradeResult{},
		ErrorList:     &errors.ErrorList{},
	}
}

// WithSelfNamespace adds Namespace in UpgradeResult instance
func (urb *UpgradeResultGetOrCreateBuilder) WithSelfNamespace(
	namespace string) *UpgradeResultGetOrCreateBuilder {
	urb.SelfNamespace = namespace
	return urb
}

// WithOwner adds OwnerReference in UpgradeResult instance
func (urb *UpgradeResultGetOrCreateBuilder) WithOwner(
	owner *metav1.OwnerReference) *UpgradeResultGetOrCreateBuilder {
	urb.Owner = owner
	return urb
}

// WithUpgradeConfig adds UpgradeConfig in UpgradeResult instance
func (urb *UpgradeResultGetOrCreateBuilder) WithUpgradeConfig(
	config *upgrade.UpgradeConfig) *UpgradeResultGetOrCreateBuilder {
	urb.UpgradeConfig = config
	return urb
}

// WithResourceDetails adds ResourceDetails in UpgradeResult instance
func (urb *UpgradeResultGetOrCreateBuilder) WithResourceDetails(
	resource *upgrade.ResourceDetails) *UpgradeResultGetOrCreateBuilder {
	urb.ResourceDetails = resource
	return urb
}

// WithTasks adds Tasks in UpgradeResult instance
func (urb *UpgradeResultGetOrCreateBuilder) WithTasks(
	tasks []upgrade.UpgradeResultTask) *UpgradeResultGetOrCreateBuilder {
	urb.Tasks = tasks
	return urb
}

// validate validates UpgradeResultGetOrCreateBuilder instance
func (urb *UpgradeResultGetOrCreateBuilder) validate() error {
	if len(urb.ErrorList.Errors) != 0 {
		return urb.ErrorList.WithStack("failed to validate upgrade result get or create")
	}
	validationErrs := &errors.ErrorList{}

	if urb.SelfNamespace == "" {
		validationErrs.Errors = append(validationErrs.Errors,
			errors.New("missing self namespace"))
	}
	if urb.Owner == nil {
		validationErrs.Errors = append(validationErrs.Errors,
			errors.New("missing self owner"))
	}
	if urb.UpgradeConfig == nil {
		validationErrs.Errors = append(validationErrs.Errors,
			errors.New("missing upgrade config"))
	}
	if urb.ResourceDetails == nil {
		validationErrs.Errors = append(validationErrs.Errors,
			errors.New("missing resource details"))
	}
	if len(urb.Tasks) == 0 {
		validationErrs.Errors = append(validationErrs.Errors,
			errors.New("missing tasks"))
	}
	if len(validationErrs.Errors) != 0 {
		urb.Errors = append(urb.Errors, validationErrs.Errors...)
		return validationErrs.WithStack("failed to validate get or create upgrade result")
	}
	return nil
}

// GetOrCreate builds a new instance of UpgradeResult with the
// helps of UpgradeResultGetOrCreateBuilder. Upgrade result cr
// is required to maintain resiliency in upgrade.
func (urb *UpgradeResultGetOrCreateBuilder) GetOrCreate() (
	res *upgrade.UpgradeResult, err error) {
	err = urb.validate()
	if err != nil {
		return nil,
			errors.Wrapf(err, "failed to get or create upgrade result: %s", urb)
	}
	l := labelJobName + "=" + urb.Owner.Name +
		"," + labelItemName + "=" + urb.ResourceDetails.Name +
		"," + labelItemNamespace + "=" + urb.ResourceDetails.Namespace +
		"," + labelItemKind + "=" + urb.ResourceDetails.Kind
	opts := metav1.ListOptions{
		LabelSelector: l,
	}
	urList, err := upgraderesult.NewKubeClient().
		WithNamespace(urb.SelfNamespace).
		List(opts)
	if err != nil {
		return nil,
			errors.Wrapf(err, "failed to get or create upgrade result: %s", urb)
	}
	switch urCount := len(urList.Items); urCount {
	case 0:
		ur, err := urb.buildUpgradeResult()
		if err != nil {
			return nil,
				errors.Wrapf(err, "failed to get or create upgrade result: %s", urb)
		}

		urb.UpgradeResult.object, err = upgraderesult.NewKubeClient().
			WithNamespace(urb.SelfNamespace).
			Create(ur)
		if err != nil {
			return nil,
				errors.Wrapf(err, "failed to get or create upgrade result: failed to create: %s", urb)
		}
		return urb.UpgradeResult.object, nil
	case 1:
		return &urList.Items[0], nil
	default:
		return nil,
			errors.Errorf(
				"failed to get or create upgrade result builder: more than one upgrade result instances were found for resource {%v}: upgrade result instances {%v}",
				urb.ResourceDetails, urList)
	}
}

// buildUpgradeResult returns UpgradeResult Object
func (urb *UpgradeResultGetOrCreateBuilder) buildUpgradeResult() (
	*upgrade.UpgradeResult, error) {
	tm, err := urb.getTypeMeta()
	if err != nil {
		return nil, err
	}

	om, err := urb.getObjectMeta()
	if err != nil {
		return nil, err
	}

	return upgraderesult.NewBuilder().
		WithTypeMeta(*tm).
		WithObjectMeta(*om).
		WithResultConfig(*urb.ResourceDetails, urb.UpgradeConfig.Data...).
		WithTasks(urb.Tasks...).
		Build()
}

// getTypeMeta returns metav1.TypeMeta for upgrade result cr
func (urb *UpgradeResultGetOrCreateBuilder) getTypeMeta() (
	tm *metav1.TypeMeta, err error) {
	return typemeta.NewBuilder().
		WithKind("UpgradeResult").
		WithAPIVersion("openebs.io/v1alpha1").
		Build()
}

// getObjectMeta returns metav1.ObjectMeta for upgrade result cr.
func (urb *UpgradeResultGetOrCreateBuilder) getObjectMeta() (
	tm *metav1.ObjectMeta, err error) {
	labels := map[string]string{
		labelJobName:       urb.Owner.Name,
		labelItemName:      urb.ResourceDetails.Name,
		labelItemNamespace: urb.ResourceDetails.Namespace,
		labelItemKind:      urb.ResourceDetails.Kind,
	}
	return objectmeta.NewBuilder().
		WithGenerateName(urb.Owner.Name + "-").
		WithNamespace(urb.SelfNamespace).
		WithLabels(labels).
		WithOwnerReferences(*urb.Owner).
		Build()
}
