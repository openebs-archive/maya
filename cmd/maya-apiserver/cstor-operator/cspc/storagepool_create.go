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

package cspc

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apiscsp "github.com/openebs/maya/pkg/cstor/newpool/v1alpha3"
	container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	pts "github.com/openebs/maya/pkg/kubernetes/podtemplatespec/v1alpha1"
	volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	"github.com/openebs/maya/pkg/version"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

// OpenEBSServiceAccount name of the openebs service accout with required
// permissions
const (
	OpenEBSServiceAccount = "openebs-maya-operator"
	// PoolMgmtContainerName is the name of cstor target container name
	PoolMgmtContainerName = "cstor-pool-mgmt"

	// PoolContainerName is the name of cstor target container name
	PoolContainerName = "cstor-pool"

	// PoolExporterContainerName is the name of cstor target container name
	PoolExporterContainerName = "maya-exporter"
)

var (
	// run container in privileged mode configuration that will be
	// applied to a container.
	privileged            = true
	defaultPoolMgmtMounts = []corev1.VolumeMount{
		corev1.VolumeMount{
			Name:      "device",
			MountPath: "/dev",
		},
		corev1.VolumeMount{
			Name:      "tmp",
			MountPath: "/tmp",
		},
		corev1.VolumeMount{
			Name:      "udev",
			MountPath: "/run/udev",
		},
	}
	// hostpathType represents the hostpath type
	hostpathTypeDirectory = corev1.HostPathDirectory

	// hostpathType represents the hostpath type
	hostpathTypeDirectoryOrCreate = corev1.HostPathDirectoryOrCreate
)

// CreateStoragePool creates the required resource to provision a cStor pool
func (pc *PoolConfig) CreateStoragePool() error {
	cspObj, err := pc.AlgorithmConfig.GetCSPSpec()
	if err != nil {
		return errors.Wrap(err, "failed to get CSP spec")
	}
	gotCSP, err := pc.createCSP(cspObj)

	if err != nil {
		return errors.Wrap(err, "failed to create CSP")
	}

	err = pc.createDeployForCSP(gotCSP)

	if err != nil {
		return errors.Wrapf(err, "failed to create deployment for CSP {%s}", gotCSP.Name)
	}

	return nil
}

func (pc *PoolConfig) createCSP(csp *apis.NewTestCStorPool) (*apis.NewTestCStorPool, error) {
	gotCSP, err := apiscsp.NewKubeClient().WithNamespace(pc.AlgorithmConfig.Namespace).Create(csp)
	return gotCSP, err
}

func (pc *PoolConfig) createPoolDeployment(deployObj *appsv1.Deployment) error {
	_, err := deploy.NewKubeClient().WithNamespace(pc.AlgorithmConfig.Namespace).Create(deployObj)
	return err
}

// GetPoolDeploySpec returns the pool deployment spec.
func (pc *PoolConfig) GetPoolDeploySpec(csp *apis.NewTestCStorPool) (*appsv1.Deployment, error) {
	deployObj, err := deploy.NewBuilder().
		WithName(csp.Name).
		WithNamespace(csp.Namespace).
		WithAnnotationsNew(getDeployAnnotations()).
		WithLabelsNew(getDeployLabels(csp)).
		WithNodeSelector(csp.Spec.NodeSelector).
		WithOwnerReferenceNew(getDeployOwnerReference(csp)).
		WithReplicas(getReplicaCount()).
		WithStrategyType(appsv1.RecreateDeploymentStrategyType).
		WithSelectorMatchLabelsNew(getDeployMatchLabels()).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(getPodLabels(csp)).
				WithAnnotationsNew(getPodAnnotations()).
				WithServiceAccountName(OpenEBSServiceAccount).
				// For CStor-Pool-Mgmt container
				WithContainerBuilders(
					container.NewBuilder().
						WithImage(getPoolMgmtImage()).
						WithName(PoolMgmtContainerName).
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithPrivilegedSecurityContext(&privileged).
						WithEnvsNew(getPoolMgmtEnv(csp)).
						// TODO : Resource and Limit
						WithVolumeMountsNew(getPoolMgmtMounts()),
					// For CStor-Pool container
					container.NewBuilder().
						WithImage(getPoolImage()).
						WithName(PoolContainerName).
						// TODO : Resource and Limit
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithPrivilegedSecurityContext(&privileged).
						WithPortsNew(getContainerPort(12000, 3232, 3233)).
						WithLivenessProbe(getPoolLivenessProbe()).
						WithEnvsNew(getPoolEnv(csp)).
						WithLifeCycle(getPoolLifeCycle()).
						WithVolumeMountsNew(getPoolMounts()),
					// For maya exporter
					container.NewBuilder().
						WithImage(getMayaExporterImage()).
						WithName(PoolExporterContainerName).
						// TODO : Resource and Limit
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithPrivilegedSecurityContext(&privileged).
						WithPortsNew(getContainerPort(9500)).
						WithCommandNew([]string{"maya-exporter"}).
						WithArgumentsNew([]string{"-e=pool"}).
						WithVolumeMountsNew(getPoolMounts()),
				).
				// TODO : Add toleration
				WithVolumeBuilders(
					volume.NewBuilder().
						WithName("device").
						WithHostPathAndType(
							"/dev",
							&hostpathTypeDirectory,
						),
					volume.NewBuilder().
						WithName("udev").
						WithHostPathAndType(
							"/run/udev",
							&hostpathTypeDirectory,
						),
					volume.NewBuilder().
						WithName("sparse").
						WithHostPathAndType(
							getSparseDirPath()+"shared-"+csp.Name,
							&hostpathTypeDirectoryOrCreate,
						),
					volume.NewBuilder().
						WithName("tmp").
						WithHostPathAndType(
							getSparseDirPath(),
							&hostpathTypeDirectoryOrCreate,
						),
				),
		).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build pool deployment object")
	}
	return deployObj, nil
}

func getReplicaCount() *int32 {
	var count int32 = 1
	return &count
}

func getDeployOwnerReference(csp *apis.NewTestCStorPool) []metav1.OwnerReference {
	OwnerReference := []metav1.OwnerReference{
		*metav1.NewControllerRef(csp, apis.SchemeGroupVersion.WithKind("NewTestCStorPool")),
	}
	return OwnerReference
}

// TODO: Use builder for labels and annotations
func getDeployLabels(csp *apis.NewTestCStorPool) map[string]string {
	return map[string]string{
		string(apis.CStorPoolClusterCPK): csp.Annotations[string(apis.CStorPoolClusterCPK)],
		"app":                            "cstor-pool",
		"openebs.io/cstor-pool":          csp.Name,
		"openebs.io/version":             version.GetVersion(),
	}
}

func getDeployAnnotations() map[string]string {
	return map[string]string{
		"openebs.io/monitoring": "pool_exporter_prometheus",
	}
}

func getPodLabels(csp *apis.NewTestCStorPool) map[string]string {
	return getDeployLabels(csp)
}

func getPodAnnotations() map[string]string {
	return map[string]string{
		"openebs.io/monitoring": "pool_exporter_prometheus",
		"prometheus.io/path":    "/metrics",
		"prometheus.io/port":    "9500",
		"prometheus.io/scrape":  "true",
	}
}

func getDeployMatchLabels() map[string]string {
	return map[string]string{
		"app": "cstor-pool",
	}
}

// getVolumeTargetImage returns Volume target image
// retrieves the value of the environment variable named
// by the key.
func getPoolMgmtImage() string {
	image, present := os.LookupEnv("OPENEBS_IO_CSTOR_POOL_MGMT_IMAGE")
	if !present {
		image = "openebs/cstor-pool-mgmt:ci"
	}
	return image
}

// getVolumeTargetImage returns Volume target image
// retrieves the value of the environment variable named
// by the key.
func getPoolImage() string {
	image, present := os.LookupEnv("OPENEBS_IO_CSTOR_POOL_IMAGE")
	if !present {
		image = "openebs/cstor-pool:ci"
	}
	return image
}

// getVolumeTargetImage returns Volume target image
// retrieves the value of the environment variable named
// by the key.
func getMayaExporterImage() string {
	image, present := os.LookupEnv("OPENEBS_IO_CSTOR_POOL_EXPORTER_IMAGE")
	if !present {
		image = "openebs/m-exporter:ci"
	}
	return image
}

func getContainerPort(port ...int32) []corev1.ContainerPort {
	var containerPorts []corev1.ContainerPort
	for _, p := range port {
		containerPorts = append(containerPorts, corev1.ContainerPort{ContainerPort: p, Protocol: "TCP"})
	}
	return containerPorts
}

func getPoolMgmtMounts() []corev1.VolumeMount {
	return append(
		defaultPoolMgmtMounts,
		corev1.VolumeMount{
			Name:      "sparse",
			MountPath: getSparseDirPath(),
		},
	)
}

func getSparseDirPath() string {
	dir, present := os.LookupEnv("OPENEBS_IO_CSTOR_POOL_SPARSE_DIR")
	if !present {
		dir = "/var/openebs/sparse"
	}
	return dir
}

func getPoolMgmtEnv(csp *apis.NewTestCStorPool) []corev1.EnvVar {
	var env []corev1.EnvVar
	return append(
		env,
		corev1.EnvVar{
			Name:  "OPENEBS_IO_CSTOR_ID",
			Value: string(csp.GetUID()),
		},
		corev1.EnvVar{
			Name: "RESYNC_INTERVAL",
			// TODO : Add tunable
			Value: "30",
		},
		corev1.EnvVar{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		corev1.EnvVar{
			Name: "NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	)
}

func getPoolLivenessProbe() *corev1.Probe {
	probe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"/bin/sh", "-c", "zfs set io.openebs:livenesstimestap='$(date)' cstor-$OPENEBS_IO_CSTOR_ID"},
			},
		},
		FailureThreshold:    3,
		InitialDelaySeconds: 300,
		PeriodSeconds:       10,
		TimeoutSeconds:      300,
	}
	return probe
}

func getPoolMounts() []corev1.VolumeMount {
	return getPoolMgmtMounts()
}

func getPoolEnv(csp *apis.NewTestCStorPool) []corev1.EnvVar {
	var env []corev1.EnvVar
	return append(
		env,
		corev1.EnvVar{
			Name:  "OPENEBS_IO_CSTOR_ID",
			Value: string(csp.GetUID()),
		},
	)
}

func getPoolLifeCycle() *corev1.Lifecycle {
	lc := &corev1.Lifecycle{
		PostStart: &corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"/bin/sh", "-c", "sleep 2"},
			},
		},
	}
	return lc
}
