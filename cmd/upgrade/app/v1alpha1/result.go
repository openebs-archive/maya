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
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	upgrade "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	objectmeta "github.com/openebs/maya/pkg/kubernetes/objectmeta/v1alpha1"
	ownerreference "github.com/openebs/maya/pkg/kubernetes/ownerreference/v1alpha1"
	typemeta "github.com/openebs/maya/pkg/kubernetes/typemeta/v1alpha1"
	upgraderesult "github.com/openebs/maya/pkg/upgrade/result/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	jobNameLabelKey       = "openebs.io/upgradejob"
	itemNameLabelKey      = "openebs.io/upgradeitemname"
	itemNamespaceLabelKey = "openebs.io/upgradeitemnamespace"
	itemKindLabelKey      = "openebs.io/upgradeitemkind"
)

// UpgradeResult is a wrapper over upgrade.UpgradeResult struct
type UpgradeResult struct {
	object *upgrade.UpgradeResult
}

// UpgradeResultBuilder helps to build UpgradeResult instance
type UpgradeResultBuilder struct {
	InstanceName      string
	InstanceNamespace string
	InstanceUID       types.UID
	UpgradeConfig     *upgrade.UpgradeConfig
	ResourceDetails   *upgrade.ResourceDetails
	Tasks             []upgrade.UpgradeResultTask
	UpgradeResult     *UpgradeResult
	errors            []error
}

// String implements GoStringer interface
func (urb *UpgradeResultBuilder) String() string {
	return stringer.Yaml("upgraderesult builder", urb)
}

// GoString implements GoStringer interface
func (urb *UpgradeResultBuilder) GoString() string {
	return urb.String()
}

// NewUpgradeResultBuilder returns a new UpgradeResult instance
func NewUpgradeResultBuilder() *UpgradeResultBuilder {
	return &UpgradeResultBuilder{
		UpgradeResult: &UpgradeResult{},
		errors:        []error{},
	}
}

// WithInstanceName adds Name in UpgradeResult instance
func (urb *UpgradeResultBuilder) WithInstanceName(name string) *UpgradeResultBuilder {
	urb.InstanceName = name
	return urb
}

// WithInstanceNamespace adds Namespace in UpgradeResult instance
func (urb *UpgradeResultBuilder) WithInstanceNamespace(namespace string) *UpgradeResultBuilder {
	urb.InstanceNamespace = namespace
	return urb
}

// WithInstanceUID adds UID in UpgradeResult instance
func (urb *UpgradeResultBuilder) WithInstanceUID(uid types.UID) *UpgradeResultBuilder {
	urb.InstanceUID = uid
	return urb
}

// WithUpgradeConfig ...adds UpgradeConfig in UpgradeResult instance
func (urb *UpgradeResultBuilder) WithUpgradeConfig(config *upgrade.UpgradeConfig) *UpgradeResultBuilder {
	urb.UpgradeConfig = config
	return urb
}

// WithResourceDetails adds ResourceDetails in UpgradeResult instance
func (urb *UpgradeResultBuilder) WithResourceDetails(resource *upgrade.ResourceDetails) *UpgradeResultBuilder {
	urb.ResourceDetails = resource
	return urb
}

// WithTasks adds Tasks in UpgradeResult instance
func (urb *UpgradeResultBuilder) WithTasks(tasks []upgrade.UpgradeResultTask) *UpgradeResultBuilder {
	urb.Tasks = tasks
	return urb
}

// validate validates UpgradeResultBuilder instance
func (urb *UpgradeResultBuilder) validate() error {
	if len(urb.errors) != 0 {
		return errors.Errorf("failed to validate: build errors were found: %v", urb.errors)
	}
	validationErrs := []error{}
	if urb.InstanceName == "" {
		validationErrs = append(validationErrs, errors.New("missing instance name"))
	}
	if urb.InstanceNamespace == "" {
		validationErrs = append(validationErrs, errors.New("missing instance namespace"))
	}
	if urb.InstanceUID == "" {
		validationErrs = append(validationErrs, errors.New("missing instance uid"))
	}
	if urb.UpgradeConfig == nil {
		validationErrs = append(validationErrs, errors.New("missing upgrade config"))
	}
	if urb.ResourceDetails == nil {
		validationErrs = append(validationErrs, errors.New("missing resource details"))
	}
	if len(urb.Tasks) == 0 {
		validationErrs = append(validationErrs, errors.New("missing tasks"))
	}
	if len(validationErrs) != 0 {
		urb.errors = append(urb.errors, validationErrs...)
		return errors.Errorf("validation error(s) found: %v", validationErrs)
	}
	return nil
}

// Build builds a new instance of UpgradeResult with the
// helps of UpgradeResultBuilder
func (urb *UpgradeResultBuilder) Build() (res *upgrade.UpgradeResult, err error) {
	err = urb.validate()
	if err != nil {
		return nil,
			errors.Wrapf(err, "failed to build UpgradeResultCR: %s", urb)
	}
	l := jobNameLabelKey + "=" + urb.InstanceName +
		"," + itemNameLabelKey + "=" + urb.ResourceDetails.Name +
		"," + itemNamespaceLabelKey + "=" + urb.ResourceDetails.Namespace +
		"," + itemKindLabelKey + "=" + urb.ResourceDetails.Kind
	opts := metav1.ListOptions{
		LabelSelector: l,
	}
	urList, err := upgraderesult.KubeClient(
		upgraderesult.WithNamespace(urb.InstanceNamespace)).
		List(opts)
	if err != nil {
		return nil,
			errors.Wrapf(err, "failed to build UpgradeResultCR: %s", urb)
	}
	if len(urList.Items) == 1 {
		return &urList.Items[0], nil
	} else if len(urList.Items) == 0 {
		ur, err := urb.getUpgradeResultObj()
		if err != nil {
			return nil,
				errors.Wrapf(err, "failed to build UpgradeResultCR: %s", urb)
		}

		urb.UpgradeResult.object, err = upgraderesult.KubeClient(
			upgraderesult.WithNamespace(urb.InstanceNamespace)).
			Create(ur)
		if err != nil {
			return nil,
				errors.Wrapf(err, "failed to build UpgradeResultCR: %s", urb)
		}
		return urb.UpgradeResult.object, nil
	}
	return nil,
		errors.Errorf(`failed to build UpgradeResultCR:
		multiple upgrade result cr found for resource: %v
		upgrade result crs: %v`, urb.ResourceDetails, urList)
}

// getUpgradeResultObj returns UpgradeResult Object for given resource
func (urb *UpgradeResultBuilder) getUpgradeResultObj() (*upgrade.UpgradeResult, error) {
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
func (urb *UpgradeResultBuilder) getTypeMeta() (tm *metav1.TypeMeta, err error) {
	return typemeta.NewBuilder().
		WithKind("UpgradeResult").
		WithAPIVersion("openebs.io/v1alpha1").
		Build()
}

// getTypeMeta returns metav1.ObjectMeta for upgrade result cr
func (urb *UpgradeResultBuilder) getObjectMeta() (tm *metav1.ObjectMeta, err error) {
	rand.Seed(time.Now().UnixNano())
	letters := "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 5)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	name := urb.UpgradeConfig.CASTemplate + "-" + string(b)

	oRef, err := urb.getOwnerReference()
	if err != nil {
		return nil, err
	}
	labels := map[string]string{
		jobNameLabelKey:       urb.InstanceName,
		itemNameLabelKey:      urb.ResourceDetails.Name,
		itemNamespaceLabelKey: urb.ResourceDetails.Namespace,
		itemKindLabelKey:      urb.ResourceDetails.Kind,
	}
	return objectmeta.NewBuilder().
		WithName(name).
		WithNamespace(urb.InstanceNamespace).
		WithLabels(labels).
		WithOwnerReferences(*oRef).
		Build()
}

// getTypeMeta returns metav1.OwnerReference for upgrade result cr
func (urb *UpgradeResultBuilder) getOwnerReference() (oRef *metav1.OwnerReference, err error) {
	ctrlOpt := true
	blockOwnerDeletionOption := true
	return ownerreference.NewBuilder().
		WithName(urb.InstanceName).
		WithKind("Job").
		WithAPIVersion("batch/v1").
		WithUID(urb.InstanceUID).
		WithControllerOption(&ctrlOpt).
		WithBlockOwnerDeletionOption(&blockOwnerDeletionOption).
		Build()
}
