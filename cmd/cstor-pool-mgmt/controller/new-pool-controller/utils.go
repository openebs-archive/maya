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

package poolcontroller

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

// IsRightCStorPoolMgmt is to check if the pool request is for this pod.
func IsRightCStorPoolMgmt(csp *apis.NewTestCStorPool) bool {
	return os.Getenv(string(common.OpenEBSIOCStorID)) == string(csp.ObjectMeta.UID)
}

// IsDestroyed is to check if the call is for cStorPool destroy.
func IsDestroyed(csp *apis.NewTestCStorPool) bool {
	return csp.ObjectMeta.DeletionTimestamp != nil
}

// IsOnlyStatusChange is to check only status change of cStorPool object.
func IsOnlyStatusChange(ocsp, ncsp *apis.NewTestCStorPool) bool {
	if reflect.DeepEqual(ocsp.Spec, ncsp.Spec) &&
		!reflect.DeepEqual(ocsp.Status, ncsp.Status) {
		return true
	}
	return false
}

// IsStatusChange is to check only status change of cStorPool object.
func IsStatusChange(oldStatus, newStatus apis.CStorPoolStatus) bool {
	return !reflect.DeepEqual(oldStatus, newStatus)
}

// IsSyncEvent is to check if ResourceVersion of cStorPool object is not modifed.
func IsSyncEvent(ocsp, ncsp *apis.NewTestCStorPool) bool {
	return ncsp.ResourceVersion == ocsp.ResourceVersion
}

// IsEmptyStatus is to check if the status of cStorPool object is empty.
func IsEmptyStatus(csp *apis.NewTestCStorPool) bool {
	return csp.Status.Phase == apis.CStorPoolStatusEmpty
}

// IsPendingStatus is to check if the status of cStorPool object is pending.
func IsPendingStatus(csp *apis.NewTestCStorPool) bool {
	return csp.Status.Phase == apis.CStorPoolStatusPending
}

// IsErrorDuplicate is to check if the status of cStorPool object is error-duplicate.
func IsErrorDuplicate(csp *apis.NewTestCStorPool) bool {
	return csp.Status.Phase == apis.CStorPoolStatusErrorDuplicate
}

// IsDeletionFailedBefore is to make sure no other operation should happen if the
// status of cStorPool is deletion-failed.
func IsDeletionFailedBefore(csp *apis.NewTestCStorPool) bool {
	return csp.Status.Phase == apis.CStorPoolStatusDeletionFailed
}

// IsUIDSet check if UID is set or not
func IsUIDSet(csp *apis.NewTestCStorPool) bool {
	return len(csp.ObjectMeta.UID) != 0
}

// IsReconcileDisabled check if reconciliation is disabled for given object or not
func IsReconcileDisabled(csp *apis.NewTestCStorPool) bool {
	return csp.Annotations[string(apis.OpenEBSDisableReconcileKey)] == "true"
}

// IsHostNameChanged check if hostname for CSP object is changed
func IsHostNameChanged(ocsp, ncsp *apis.NewTestCStorPool) bool {
	return ncsp.Spec.HostName != ocsp.Spec.HostName
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

// PoolName return pool name for given CSP object
func PoolName(csp *apis.NewTestCStorPool) string {
	return PoolPrefix + string(csp.ObjectMeta.UID)
}
