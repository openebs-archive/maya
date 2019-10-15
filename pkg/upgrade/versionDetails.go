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

package upgrade

import (
	"strings"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	validCurrentVersions = []string{"1.0.0", "1.1.0", "1.2.0"}
	validDesiredVersion  = version.GetVersion()
)

// IsCurrentVersionValid ...
func IsCurrentVersionValid(vd apis.VersionDetails) bool {
	currentVersion := strings.Split(vd.Status.Current, "-")[0]
	return util.ContainsString(validCurrentVersions, currentVersion)
}

// IsDesiredVersionValid ...
func IsDesiredVersionValid(vd apis.VersionDetails) bool {
	desiredVersion := strings.Split(vd.Desired, "-")[0]
	return validDesiredVersion == desiredVersion
}

// Path ...
func Path(vd apis.VersionDetails) string {
	return strings.Split(vd.Status.Current, "-")[0] + "-" +
		strings.Split(vd.Desired, "-")[0]
}

// SetErrorStatus ...
func SetErrorStatus(vs apis.VersionStatus, msg string, err error) apis.VersionStatus {
	vs.Message = msg
	vs.Reason = err.Error()
	vs.LastUpdateTime = metav1.Now()
	return vs
}

// SetPendingStatus ...
func SetPendingStatus(vs apis.VersionStatus) apis.VersionStatus {
	vs.State = apis.ReconcileInProgress
	vs.LastUpdateTime = metav1.Now()
	return vs
}

// SetSuccessStatus ...
func SetSuccessStatus(vs apis.VersionStatus) apis.VersionStatus {
	vs.Current = validDesiredVersion
	vs.Message = ""
	vs.Reason = ""
	vs.State = apis.ReconcileComplete
	vs.LastUpdateTime = metav1.Now()
	return vs
}
