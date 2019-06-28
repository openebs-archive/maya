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

package cvc

import (
	"os"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	pts "github.com/openebs/maya/pkg/kubernetes/podtemplatespec/v1alpha1"
	volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	tolerationSeconds = int64(30)

	deployreplicas int32 = 1

	privileged = true

	resyncInterval = "30"

	mountPropagation = corev1.MountPropagationBidirectional

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
)

func getDeployLabels(pvName string) map[string]string {

	return map[string]string{
		"app":                          "cstor-volume-manager",
		"openebs.io/target":            "cstor-target",
		"openebs.io/persistent-volume": pvName,
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
	}
}

func getDeployTemplateAnnotations() map[string]string {
	return map[string]string{
		"prometheus.io/path":   "/metrics",
		"prometheus.io/port":   "9500",
		"prometheus.io/scrape": "true",
	}
}

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

func getVolumeTargetImage() string {
	image, check := os.LookupEnv("OPENEBS_IO_CSTOR_TARGET_IMAGE")
	if check == false {
		image = "openebs/cstor-istgt:ci"
	}
	return image
}

func getVolumeMonitorImage() string {
	image, check := os.LookupEnv("OPENEBS_IO_VOLUME_MONITOR_IMAGE")
	if check == false {
		image = "openebs/m-exporter:ci"
	}
	return image
}

func getVolumeMgmtImage() string {
	image, check := os.LookupEnv("OPENEBS_IO_CSTOR_VOLUME_MGMT_IMAGE")
	if check == false {
		image = "openebs/cstor-volume-mgmt:ci"
	}
	return image
}

func getTargetDirPath(pvName string) string {
	dir, check := os.LookupEnv("OPENEBS_IO_CSTOR_TARGET_DIR")
	if check == false {
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

func createCStorTargetDeployment(
	vol *apis.CStorVolume,
) (*appsv1.Deployment, error) {

	deployObj, err := deploy.NewBuilder().
		WithName(vol.Name + "-target").
		WithLabelsNew(getDeployLabels(vol.Name)).
		WithAnnotationsNew(getDeployAnnotation()).
		WithReplicas(&deployreplicas).
		WithStrategyType(
			appsv1.RecreateDeploymentStrategyType,
		).
		WithSelectorMatchLabelsNew(getDeployMatchLabels(vol.Name)).
		WithTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(getDeployTemplateLabels(vol.Name)).
				WithAnnotationsNew(getDeployTemplateAnnotations()).
				WithServiceAccountName("openebs-maya-operator").
				//WithAffinity(getDeployTemplateAffinity()).
				// TODO
				//WithNodeSelectorNew().
				WithTolerationsNew(getDeployTemplateTolerations()...).
				WithContainerBuilders(
					container.NewBuilder().
						WithImage(getVolumeTargetImage()).
						WithName("cstor-istgt").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithPorts(getContainerPort(3260)).
						WithPrivilegedSecurityContext(&privileged).
						WithVolumeMounts(getTargetMgmtMounts()),
					container.NewBuilder().
						WithImage(getVolumeMonitorImage()).
						WithName("maya-volume-exporter").
						WithCommand([]string{"maya-exporter"}).
						WithArguments([]string{"-e=cstor"}).
						WithPorts(getContainerPort(9500)).
						WithVolumeMounts(getMonitorMounts()),
					container.NewBuilder().
						WithImage(getVolumeMgmtImage()).
						WithName("cstor-volume-mgmt").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithPorts(getContainerPort(80)).
						WithEnvs(getDeployTemplateEnvs(string(vol.UID))).
						WithPrivilegedSecurityContext(&privileged).
						WithVolumeMounts(getTargetMgmtMounts()),
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

	deploymentObj, err := deploy.NewKubeClient(deploy.WithNamespace("openebs")).Create(deployObj.Object)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to create deployment object")
	}

	return deploymentObj, nil
}
