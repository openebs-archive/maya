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

package cstorvolumeclaim

import (
	"os"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	pts "github.com/openebs/maya/pkg/kubernetes/podtemplatespec/v1alpha1"
	volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	"github.com/openebs/maya/pkg/version"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (

	// tolerationSeconds represents the period of time the toleration
	// tolerates the taint
	tolerationSeconds = int64(30)
	// deployreplicas is replica count for target deployment
	deployreplicas int32 = 1

	// run container in privileged mode configuration that will be
	// applied to a container.
	privileged = true

	resyncInterval = "30"
	// MountPropagationBidirectional means that the volume in a container will
	// receive new mounts from the host or other containers, and its own mounts
	// will be propagated from the container to the host or other containers.
	mountPropagation = corev1.MountPropagationBidirectional
	// hostpathType represents the hostpath type
	hostpathType = corev1.HostPathDirectoryOrCreate

	defaultMounts = []corev1.VolumeMount{
		corev1.VolumeMount{
			Name:      "sockfile",
			MountPath: "/var/run",
		},
		corev1.VolumeMount{
			Name:      "conf",
			MountPath: "/usr/local/etc/istgt",
		},
	}
	// OpenEBSServiceAccount name of the openebs service accout with required
	// permissions
	OpenEBSServiceAccount = "openebs-maya-operator"
	// TargetContainerName is the name of cstor target container name
	TargetContainerName = "cstor-istgt"
	// MonitorContainerName is the name of monitor container name
	MonitorContainerName = "maya-volume-exporter"
	// MgmtContainerName is the container name of cstor volume mgmt side car
	MgmtContainerName = "cstor-volume-mgmt"
)

func getDeployLabels(pvName string) map[string]string {
	return map[string]string{
		"app":                            "cstor-volume-manager",
		"openebs.io/target":              "cstor-target",
		"openebs.io/storage-engine-type": "cstor",
		"openebs.io/cas-type":            "cstor",
		"openebs.io/persistent-volume":   pvName,
		"openebs.io/version":             version.GetVersion(),
	}
}

func getDeployAnnotation() map[string]string {
	return map[string]string{
		"openebs.io/volume-monitor": "true",
		"openebs.io/volume-type":    "cstor",
	}
}

func getDeployMatchLabels(pvName string) map[string]string {
	return map[string]string{
		"app":                          "cstor-volume-manager",
		"openebs.io/target":            "cstor-target",
		"openebs.io/persistent-volume": pvName,
	}
}

func getDeployTemplateLabels(pvName string) map[string]string {
	return map[string]string{
		"monitoring":                   "volume_exporter_prometheus",
		"app":                          "cstor-volume-manager",
		"openebs.io/target":            "cstor-target",
		"openebs.io/persistent-volume": pvName,
		"openebs.io/version":           version.GetVersion(),
	}
}

func getDeployTemplateAnnotations() map[string]string {
	return map[string]string{
		"prometheus.io/path":   "/metrics",
		"prometheus.io/port":   "9500",
		"prometheus.io/scrape": "true",
	}
}

func getDeployOwnerReference(volume *apis.CStorVolume) []metav1.OwnerReference {
	OwnerReference := []metav1.OwnerReference{
		*metav1.NewControllerRef(volume, apis.SchemeGroupVersion.WithKind("CStorVolume")),
	}
	return OwnerReference
}

// getDeployTemplateAffinity returns affinities
// for target deployement
func getDeployTemplateAffinity() *corev1.Affinity {
	return &corev1.Affinity{
		PodAffinity: &corev1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							metav1.LabelSelectorRequirement{
								Key:      "statefulset.kubernetes.io/pod-name",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{},
							},
						},
					},
				},
			},
		},
	}
}

// getDeployTemplateTolerations returns the array of toleration
// for target deployement
func getDeployTemplateTolerations() []corev1.Toleration {
	return []corev1.Toleration{
		corev1.Toleration{
			Effect:            corev1.TaintEffectNoExecute,
			Key:               "node.alpha.kubernetes.io/notReady",
			Operator:          corev1.TolerationOpExists,
			TolerationSeconds: &tolerationSeconds,
		},
		corev1.Toleration{
			Effect:            corev1.TaintEffectNoExecute,
			Key:               "node.alpha.kubernetes.io/unreachable",
			Operator:          corev1.TolerationOpExists,
			TolerationSeconds: &tolerationSeconds,
		},
		corev1.Toleration{
			Effect:            corev1.TaintEffectNoExecute,
			Key:               "node.kubernetes.io/not-ready",
			Operator:          corev1.TolerationOpExists,
			TolerationSeconds: &tolerationSeconds,
		},
		corev1.Toleration{
			Effect:            corev1.TaintEffectNoExecute,
			Key:               "node.kubernetes.io/unreachable",
			Operator:          corev1.TolerationOpExists,
			TolerationSeconds: &tolerationSeconds,
		},
	}
}

func getMonitorMounts() []corev1.VolumeMount {
	return defaultMounts
}

func getTargetMgmtMounts() []corev1.VolumeMount {
	return append(
		defaultMounts,
		corev1.VolumeMount{
			Name:             "tmp",
			MountPath:        "/tmp",
			MountPropagation: &mountPropagation,
		},
	)
}

// getDeployTemplateEnvs return the common env required for
// cstorvolume target deployment
func getDeployTemplateEnvs(cstorid string) []corev1.EnvVar {
	return []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "OPENEBS_IO_CSTOR_VOLUME_ID",
			Value: cstorid,
		},
		corev1.EnvVar{
			Name:  "RESYNC_INTERVAL",
			Value: resyncInterval,
		},
		corev1.EnvVar{
			Name: "NODE_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "spec.nodeName",
				},
			},
		},
		corev1.EnvVar{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
	}
}

// getVolumeTargetImage returns Volume target image
// retrieves the value of the environment variable named
// by the key.
func getVolumeTargetImage() string {
	image, present := os.LookupEnv("OPENEBS_IO_CSTOR_TARGET_IMAGE")
	if !present {
		image = "openebs/cstor-istgt:ci"
	}
	return image
}

// getVolumeMonitorImage returns monitor image
// retrieves the value of the environment variable named
// by the key.
func getVolumeMonitorImage() string {
	image, present := os.LookupEnv("OPENEBS_IO_VOLUME_MONITOR_IMAGE")
	if !present {
		image = "openebs/m-exporter:ci"
	}
	return image
}

// getVolumeMgmtImage returns volume mgmt side image
// retrieves the value of the environment variable named
// by the key.
func getVolumeMgmtImage() string {
	image, present := os.LookupEnv("OPENEBS_IO_CSTOR_VOLUME_MGMT_IMAGE")
	if !present {
		image = "openebs/cstor-volume-mgmt:ci"
	}
	return image
}

// getTargetDirPath returns cstor target volume directory for a
// given volume, retrieves the value of the environment variable named
// by the key.
func getTargetDirPath(pvName string) string {
	dir, present := os.LookupEnv("OPENEBS_IO_CSTOR_TARGET_DIR")
	if !present {
		dir = "/var/openebs"
	}
	return dir + "/shared-" + pvName + "-target"
}

func getContainerPort(port int32) []corev1.ContainerPort {
	return []corev1.ContainerPort{
		corev1.ContainerPort{
			ContainerPort: port,
		},
	}
}

// getOrCreateCStorTargetDeployment get or create the cstor target deployment
// for a given cstorvolume.
func getOrCreateCStorTargetDeployment(
	vol *apis.CStorVolume,
) (*appsv1.Deployment, error) {

	deployObj, err := deploy.NewKubeClient(deploy.WithNamespace("openebs")).
		Get(vol.Name + "-target")

	if err != nil && !k8serror.IsNotFound(err) {
		return nil, errors.Wrapf(
			err,
			"failed to get cstorvolume target {%v}",
			deployObj.Name,
		)
	}

	if k8serror.IsNotFound(err) {
		deployObj, err = deploy.NewBuilder().
			WithName(vol.Name + "-target").
			WithLabelsNew(getDeployLabels(vol.Name)).
			WithAnnotationsNew(getDeployAnnotation()).
			WithOwnerReferenceNew(getDeployOwnerReference(vol)).
			WithReplicas(&deployreplicas).
			WithStrategyType(
				appsv1.RecreateDeploymentStrategyType,
			).
			WithSelectorMatchLabelsNew(getDeployMatchLabels(vol.Name)).
			WithPodTemplateSpecBuilder(
				pts.NewBuilder().
					WithLabelsNew(getDeployTemplateLabels(vol.Name)).
					WithAnnotationsNew(getDeployTemplateAnnotations()).
					WithServiceAccountName(OpenEBSServiceAccount).
					//WithAffinity(getDeployTemplateAffinity()).
					// TODO use of selector and affinity
					//WithNodeSelectorNew().
					WithTolerationsNew(getDeployTemplateTolerations()...).
					WithContainerBuilders(
						container.NewBuilder().
							WithImage(getVolumeTargetImage()).
							WithName(TargetContainerName).
							WithImagePullPolicy(corev1.PullIfNotPresent).
							WithPortsNew(getContainerPort(3260)).
							WithPrivilegedSecurityContext(&privileged).
							WithVolumeMountsNew(getTargetMgmtMounts()),
						container.NewBuilder().
							WithImage(getVolumeMonitorImage()).
							WithName(MonitorContainerName).
							WithCommandNew([]string{"maya-exporter"}).
							WithArgumentsNew([]string{"-e=cstor"}).
							WithPortsNew(getContainerPort(9500)).
							WithVolumeMountsNew(getMonitorMounts()),
						container.NewBuilder().
							WithImage(getVolumeMgmtImage()).
							WithName(MgmtContainerName).
							WithImagePullPolicy(corev1.PullIfNotPresent).
							WithPortsNew(getContainerPort(80)).
							WithEnvsNew(getDeployTemplateEnvs(string(vol.UID))).
							WithPrivilegedSecurityContext(&privileged).
							WithVolumeMountsNew(getTargetMgmtMounts()),
					).
					WithVolumeBuilders(
						volume.NewBuilder().
							WithName("sockfile").
							WithEmptyDir(&corev1.EmptyDirVolumeSource{}),
						volume.NewBuilder().
							WithName("conf").
							WithEmptyDir(&corev1.EmptyDirVolumeSource{}),
						volume.NewBuilder().
							WithName("tmp").
							WithHostPathAndType(
								getTargetDirPath(vol.Name),
								&hostpathType,
							),
					),
			).
			Build()

		if err != nil {
			return nil, errors.Wrapf(err, "failed to build deployment object")
		}

		deployObj, err = deploy.NewKubeClient(deploy.WithNamespace("openebs")).Create(deployObj)

		if err != nil {
			return nil, errors.Wrapf(err, "failed to create deployment object")
		}
	}
	return deployObj, nil
}
