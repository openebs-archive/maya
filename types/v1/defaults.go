/*
Copyright 2017 The OpenEBS Authors.

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

package v1

import (
	"github.com/openebs/maya/pkg/util"
)

// These are a set of defaults
const (
	// DefaultVolumeType contains the default volume type
	DefaultVolumeType VolumeType = JivaVolumeType

	// DefaultOrchProvider contains the default orchestrator
	DefaultOrchProvider OrchProvider = K8sOrchProvider

	// DefaultNamespace contains the default namespace where
	// volume operations will be executed
	DefaultNamespace string = "default"

	// DefaultCapacity contains the default volume capacity
	DefaultCapacity string = "5G"

	// DefaultJivaControllerImage contains the default jiva controller
	// image
	DefaultJivaControllerImage string = "openebs/jiva:latest"

	// DefaultJivaReplicaImage contains the default jiva replica image
	DefaultJivaReplicaImage string = "openebs/jiva:latest"

	// DefaultStoragePool contains the name of default storage pool
	DefaultStoragePool string = "default"

	// DefaultHostPath contains the default host path value
	DefaultHostPath string = "/var/openebs"

	// DefaultMonitor contains the default value for volume
	// monitoring policy. Value of `false` indicates
	// volume monitoring is disabled by default.
	DefaultMonitor string = "false"

	// DefaultMonitorLabelKey contains the default value for Label key
	// used for volume monitoring polciy
	DefaultMonitorLabelKey string = "monitoring"

	// DefaultMonitorLabelValue contains the default value for Label value
	// used for volume monitoring polciy
	DefaultMonitorLabelValue string = "volume_exporter_prometheus"

	// DefaultMonitoringImage contains the default image for
	// volume monitoring
	DefaultMonitoringImage string = "openebs/m-exporter:latest"

	// DefaultNamespaceForListOps contains the default
	DefaultNamespaceForListOps string = "all-namespaces"
)

var (
	// DefaultJivaReplicas contains the default jiva replica count
	DefaultJivaReplicas *int32 = util.StrToInt32("2")

	// DefaultJivaControllers contains the default jiva controller
	// count
	DefaultJivaControllers *int32 = util.StrToInt32("1")
)
