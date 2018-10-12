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
	"fmt"

	menv "github.com/openebs/maya/pkg/env/v1alpha1"
)

// EnvLister abstracts listing of env structures based on version
type EnvLister func(v version) (l *envList, err error)

// EnvList returns a list of env instances based on version
func EnvList(v version) (l *envList, err error) {
	err = fmt.Errorf("invalid version '%+v': failed to list environment values", v)
	switch v {
	case version070:
		return envList070(), nil
	default:
		return
	}
	return
}

// envInstallConfig returns a list of environment variable info specific to
// install config
func envInstallConfig() (l *envList) {
	l = &envList{}
	l.Items = append(l.Items, &env{Key: InstallerConfigName, Value: "maya-install-config-default-0.7.0"})
	return
}

// envList070 returns a list of environment variable info specific to 0.7.0
// version
func envList070() (l *envList) {
	l = &envList{}
	l.Items = append(l.Items, &env{Key: DefaultCstorSparsePool, Value: "false"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateFeatureGateENVK, Value: "true"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToCreateJivaVolumeENVK, Value: "jiva-volume-create-default-0.7.0"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToReadJivaVolumeENVK, Value: "jiva-volume-read-default-0.7.0"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToDeleteJivaVolumeENVK, Value: "jiva-volume-delete-default-0.7.0"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToCreateCStorVolumeENVK, Value: "cstor-volume-create-default-0.7.0"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToReadCStorVolumeENVK, Value: "cstor-volume-read-default-0.7.0"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToDeleteCStorVolumeENVK, Value: "cstor-volume-delete-default-0.7.0"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToCreatePoolENVK, Value: "cstor-pool-create-default-0.7.0"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToDeletePoolENVK, Value: "cstor-pool-delete-default-0.7.0"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToListVolumeENVK, Value: "jiva-volume-list-default-0.6.0,jiva-volume-list-default-0.7.0,cstor-volume-list-default-0.7.0"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToCreateCStorSnapshotENVK, Value: "cstor-snapshot-create-default-0.7.0"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToDeleteCStorSnapshotENVK, Value: "cstor-snapshot-delete-default-0.7.0"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToCreateJivaSnapshotENVK, Value: "jiva-snapshot-create-default-0.7.0"})
	l.Items = append(l.Items, &env{Key: menv.CASTemplateToDeleteJivaSnapshotENVK, Value: "jiva-snapshot-delete-default-0.7.0"})
	return
}
