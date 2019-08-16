/*
Copyright 2018 The OpenEBS Authors.

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

package poolinstancecontroller

import (
	"os"
	"reflect"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
)

const (
	// PoolPrefix is prefix for pool name
	PoolPrefix string = "cstor-"
)

// IsRightCStorPoolInstanceMgmt is to check if the pool request is for this pod.
func IsRightCStorPoolInstanceMgmt(cspi *apis.CStorPoolInstance) bool {
	return os.Getenv(string(common.OpenEBSIOCStorID)) == string(cspi.ObjectMeta.UID)
}

// IsDestroyed is to check if the call is for cStorPoolInstance destroy.
func IsDestroyed(cspi *apis.CStorPoolInstance) bool {
	return cspi.ObjectMeta.DeletionTimestamp != nil
}

// IsOnlyStatusChange is to check only status change of cStorPoolInstance object.
func IsOnlyStatusChange(ocspi, ncspi *apis.CStorPoolInstance) bool {
	if reflect.DeepEqual(ocspi.Spec, ncspi.Spec) &&
		!reflect.DeepEqual(ocspi.Status, ncspi.Status) {
		return true
	}
	return false
}

// IsStatusChange is to check only status change of cStorPoolInstance object.
func IsStatusChange(oldStatus, newStatus apis.CStorPoolStatus) bool {
	return !reflect.DeepEqual(oldStatus, newStatus)
}

// IsSyncEvent is to check if ResourceVersion of cStorPoolInstance object is not modifed.
func IsSyncEvent(ocspi, ncspi *apis.CStorPoolInstance) bool {
	return ncspi.ResourceVersion == ocspi.ResourceVersion
}

// IsEmptyStatus is to check if the status of cStorPoolInstance object is empty.
func IsEmptyStatus(cspi *apis.CStorPoolInstance) bool {
	return cspi.Status.Phase == apis.CStorPoolStatusEmpty
}

// IsPendingStatus is to check if the status of cStorPoolInstance object is pending.
func IsPendingStatus(cspi *apis.CStorPoolInstance) bool {
	return cspi.Status.Phase == apis.CStorPoolStatusPending
}

// IsErrorDuplicate is to check if the status of cStorPoolInstance object is error-duplicate.
func IsErrorDuplicate(cspi *apis.CStorPoolInstance) bool {
	return cspi.Status.Phase == apis.CStorPoolStatusErrorDuplicate
}

// IsDeletionFailedBefore is to make sure no other operation should happen if the
// status of cStorPoolInstance is deletion-failed.
func IsDeletionFailedBefore(cspi *apis.CStorPoolInstance) bool {
	return cspi.Status.Phase == apis.CStorPoolStatusDeletionFailed
}

// IsUIDSet check if UID is set or not
func IsUIDSet(cspi *apis.CStorPoolInstance) bool {
	return len(cspi.ObjectMeta.UID) != 0
}

// IsReconcileDisabled check if reconciliation is disabled for given object or not
func IsReconcileDisabled(cspi *apis.CStorPoolInstance) bool {
	return cspi.Annotations[string(apis.OpenEBSDisableReconcileKey)] == "true"
}

// IsHostNameChanged check if hostname for CSPI object is changed
func IsHostNameChanged(ocspi, ncspi *apis.CStorPoolInstance) bool {
	return ncspi.Spec.HostName != ocspi.Spec.HostName
}

// IsEmpty check if string is empty or not
func IsEmpty(s string) bool {
	return len(s) == 0
}

// ErrorWrapf wrap error
// If given err is nil then it will return new error
func ErrorWrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return errors.Errorf(format, args...)
	}

	return errors.Wrapf(err, format, args...)
}

// PoolName return pool name for given CSPI object
func PoolName(cspi *apis.CStorPoolInstance) string {
	return PoolPrefix + string(cspi.ObjectMeta.UID)
}
