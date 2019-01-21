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

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// fakeAlwaysTrue is a concrete implementation of daemonset Predicate
func fakeAlwaysTrue(d *daemonset) (string, bool) {
	return "fakeAlwaysTrue", true
}

// fakeAlwaysFalse is a concrete implementation of daemonset Predicate
func fakeAlwaysFalse(d *daemonset) (string, bool) {
	return "fakeAlwaysFalse", false
}

func TestPredicateFailedError(t *testing.T) {
	tests := map[string]struct {
		predicateMessage string
		expectedErr      string
	}{
		"always true":  {"fakeAlwaysTrue", "predicatefailed: fakeAlwaysTrue"},
		"always false": {"fakeAlwaysFalse", "predicatefailed: fakeAlwaysFalse"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			e := predicateFailedError(mock.predicateMessage)
			if e.Error() != mock.expectedErr {
				t.Fatalf("test '%s' failed: expected '%s': actual '%s'", name, mock.expectedErr, e.Error())
			}
		})
	}
}

func TestNew(t *testing.T) {
	d := New()
	if d == nil {
		t.Fatalf("test failed: expected not nil daemonset instance: actual nil")
	}
}

func TestBuilderBuild(t *testing.T) {
	d, _ := Builder().Build()
	if d == nil {
		t.Fatalf("test failed: expected not nil daemonset instance: actual nil")
	}
}

func TestBuilderValidation(t *testing.T) {
	tests := map[string]struct {
		checks  []Predicate
		isError bool
	}{
		"always true":  {[]Predicate{fakeAlwaysTrue}, false},
		"always false": {[]Predicate{fakeAlwaysFalse}, true},
		"true & false": {[]Predicate{fakeAlwaysTrue, fakeAlwaysFalse}, true},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Builder().AddChecks(mock.checks).Build()
			if mock.isError && err == nil {
				t.Fatalf("test '%s' failed: expected error: actual no error", name)
			}
			if !mock.isError && err != nil {
				t.Fatalf("test '%s' failed: expected no error: actual error '%+v'", name, err)
			}
			if !mock.isError && d == nil {
				t.Fatalf("test '%s' failed: expected not nil instance: actual nil instance", name)
			}
		})
	}
}

func TestBuilderAddChecks(t *testing.T) {
	tests := map[string]struct {
		checks        []Predicate
		expectedCount int
	}{
		"zero": {[]Predicate{}, 0},
		"one":  {[]Predicate{fakeAlwaysTrue}, 1},
		"two":  {[]Predicate{fakeAlwaysTrue, fakeAlwaysFalse}, 2},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := Builder().AddChecks(mock.checks)
			if len(b.checks) != mock.expectedCount {
				t.Fatalf("test '%s' failed: expected no of checks '%d': actual '%d'", name, mock.expectedCount, len(b.checks))
			}
		})
	}
}

func TestBuilderAddNodeSelector(t *testing.T) {
	tests := map[string]struct {
		key           string
		value         string
		expectedCount int
	}{
		"t1": {"kubernetes.io/hostname", "localhost", 1},
		"t2": {"my.io/storage", "ssd", 1},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := Builder().AddNodeSelector(mock.key, mock.value)
			if len(b.daemonset.daemon.Spec.Template.Spec.NodeSelector) != mock.expectedCount {
				t.Fatalf("test '%s' failed: expected nodeselector count '%d': actual '%d'", name, mock.expectedCount, len(b.daemonset.daemon.Spec.Template.Spec.NodeSelector))
			}
			if b.daemonset.daemon.Spec.Template.Spec.NodeSelector[mock.key] != mock.value {
				t.Fatalf("test '%s' failed: expected nodeselector value '%s': actual '%s'", name, mock.value, b.daemonset.daemon.Spec.Template.Spec.NodeSelector[mock.key])
			}
		})
	}
}

func TestNewAddNodeSelector(t *testing.T) {
	tests := map[string]struct {
		key           string
		value         string
		expectedCount int
	}{
		"t1": {"kubernetes.io/hostname", "localhost", 1},
		"t2": {"my.io/storage", "ssd", 1},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			daemon := New(AddNodeSelector(mock.key, mock.value))
			if len(daemon.Spec.Template.Spec.NodeSelector) != mock.expectedCount {
				t.Fatalf("test '%s' failed: expected nodeselector count '%d': actual '%d'", name, mock.expectedCount, len(daemon.Spec.Template.Spec.NodeSelector))
			}
			if daemon.Spec.Template.Spec.NodeSelector[mock.key] != mock.value {
				t.Fatalf("test '%s' failed: expected nodeselector value '%s': actual '%s'", name, mock.value, daemon.Spec.Template.Spec.NodeSelector[mock.key])
			}
		})
	}
}

func TestBuilderAddInitContainers(t *testing.T) {
	tests := map[string]struct {
		containers    []corev1.Container
		expectedCount int
	}{
		"t1": {[]corev1.Container{corev1.Container{Name: "con1"}}, 1},
		"t2": {[]corev1.Container{corev1.Container{Name: "con1"}, corev1.Container{Name: "con2"}}, 2},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := Builder().AddInitContainers(mock.containers)
			if len(b.daemonset.daemon.Spec.Template.Spec.InitContainers) != mock.expectedCount {
				t.Fatalf("test '%s' failed: expected init containers count '%d': actual '%d'", name, mock.expectedCount, len(b.daemonset.daemon.Spec.Template.Spec.InitContainers))
			}
		})
	}
}

func TestNewAddInitContainer(t *testing.T) {
	tests := map[string]struct {
		container     corev1.Container
		expectedCount int
	}{
		"t1": {corev1.Container{Name: "con1"}, 1},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			daemon := New(AddInitContainer(mock.container))
			if len(daemon.Spec.Template.Spec.InitContainers) != mock.expectedCount {
				t.Fatalf("test '%s' failed: expected init containers count '%d': actual '%d'", name, mock.expectedCount, len(daemon.Spec.Template.Spec.InitContainers))
			}
		})
	}
}

func TestBuilderAddContainers(t *testing.T) {
	tests := map[string]struct {
		containers    []corev1.Container
		expectedCount int
	}{
		"t1": {[]corev1.Container{corev1.Container{Name: "con1"}}, 1},
		"t2": {[]corev1.Container{corev1.Container{Name: "con1"}, corev1.Container{Name: "con2"}}, 2},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := Builder().AddContainers(mock.containers)
			if len(b.daemonset.daemon.Spec.Template.Spec.Containers) != mock.expectedCount {
				t.Fatalf("test '%s' failed: expected containers count '%d': actual '%d'", name, mock.expectedCount, len(b.daemonset.daemon.Spec.Template.Spec.Containers))
			}
		})
	}
}

func TestNewAddContainer(t *testing.T) {
	tests := map[string]struct {
		container     corev1.Container
		expectedCount int
	}{
		"t1": {corev1.Container{Name: "con1"}, 1},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			daemon := New(AddContainer(mock.container))
			if len(daemon.Spec.Template.Spec.Containers) != mock.expectedCount {
				t.Fatalf("test '%s' failed: expected containers count '%d': actual '%d'", name, mock.expectedCount, len(daemon.Spec.Template.Spec.Containers))
			}
		})
	}
}

func TestBuilderAddTolerations(t *testing.T) {
	tests := map[string]struct {
		tolerations   []corev1.Toleration
		expectedCount int
	}{
		"t1": {[]corev1.Toleration{corev1.Toleration{Key: "one"}}, 1},
		"t2": {[]corev1.Toleration{corev1.Toleration{Key: "one"}, corev1.Toleration{Key: "two"}}, 2},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := Builder().AddTolerations(mock.tolerations)
			if len(b.daemonset.daemon.Spec.Template.Spec.Tolerations) != mock.expectedCount {
				t.Fatalf("test '%s' failed: expected tolerations count '%d': actual '%d'", name, mock.expectedCount, len(b.daemonset.daemon.Spec.Template.Spec.Tolerations))
			}
		})
	}
}

func TestNewAddToleration(t *testing.T) {
	tests := map[string]struct {
		toleration    corev1.Toleration
		expectedCount int
	}{
		"t1": {corev1.Toleration{Key: "one"}, 1},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			daemon := New(AddToleration(mock.toleration))
			if len(daemon.Spec.Template.Spec.Tolerations) != mock.expectedCount {
				t.Fatalf("test '%s' failed: expected tolerations count '%d': actual '%d'", name, mock.expectedCount, len(daemon.Spec.Template.Spec.Tolerations))
			}
		})
	}
}

func TestBuilderNoScheduleOnMaster(t *testing.T) {
	d, _ := Builder().NoScheduleOnMaster().Build()
	if len(d.Spec.Template.Spec.Tolerations) != 1 {
		t.Fatalf("test failed: expected toleration count '1': actual '%d'", len(d.Spec.Template.Spec.Tolerations))
	}
}

func TestNewNoScheduleOnMaster(t *testing.T) {
	d := New(NoScheduleOnMaster())
	if len(d.Spec.Template.Spec.Tolerations) != 1 {
		t.Fatalf("test failed: expected toleration count '1': actual '%d'", len(d.Spec.Template.Spec.Tolerations))
	}
}

func TestBuilderWithSpec(t *testing.T) {
	tests := map[string]struct {
		daemon        *appsv1.DaemonSet
		expectedName  string
		expectedImage string
	}{
		"t1": {&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dmn",
			},
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Image: "m-apiserver",
							},
						},
					},
				},
			},
		}, "dmn", "m-apiserver"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			d, _ := Builder().WithSpec(mock.daemon).Build()
			if d.Name != mock.expectedName {
				t.Fatalf("test failed: expected name '%s' actual '%s'", mock.expectedName, d.Name)
			}
			if d.Spec.Template.Spec.Containers[0].Image != mock.expectedImage {
				t.Fatalf("test '%s' failed: expected image '%s' actual '%s'", name, mock.expectedImage, d.Spec.Template.Spec.Containers[0].Image)
			}
		})
	}
}

func TestNewWithSpec(t *testing.T) {
	tests := map[string]struct {
		daemon        *appsv1.DaemonSet
		expectedName  string
		expectedImage string
	}{
		"t1": {&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dmn",
			},
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Image: "m-apiserver",
							},
						},
					},
				},
			},
		}, "dmn", "m-apiserver"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			d := New(WithSpec(mock.daemon))
			if d.Name != mock.expectedName {
				t.Fatalf("test failed: expected name '%s' actual '%s'", mock.expectedName, d.Name)
			}
			if d.Spec.Template.Spec.Containers[0].Image != mock.expectedImage {
				t.Fatalf("test '%s' failed: expected image '%s' actual '%s'", name, mock.expectedImage, d.Spec.Template.Spec.Containers[0].Image)
			}
		})
	}
}

func TestBuilderWithTemplate(t *testing.T) {
	tests := map[string]struct {
		daemon        string
		data          interface{}
		expectedName  string
		expectedImage string
	}{
		"t1": {`
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: {{.name}}
  namespace: kube-system
  labels:
    k8s-app: fluentd-logging
spec:
  selector:
    matchLabels:
      name: {{.name}}
  template:
    metadata:
      labels:
        name: {{.name}}
    spec:
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: fluentd-elasticsearch
        image: {{.image}}
        resources:
          limits:
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 200Mi
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
      terminationGracePeriodSeconds: 30
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
`,
			map[string]string{
				"name":  "fluentd-elasticsearch",
				"image": "k8s.gcr.io/fluentd-elasticsearch:1.20",
			},
			"fluentd-elasticsearch", "k8s.gcr.io/fluentd-elasticsearch:1.20"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			d, _ := Builder().WithTemplate(mock.daemon, mock.data).Build()
			if d.Name != mock.expectedName {
				t.Fatalf("test failed: expected name '%s' actual '%s'", mock.expectedName, d.Name)
			}
			if d.Spec.Template.Spec.Containers[0].Image != mock.expectedImage {
				t.Fatalf("test '%s' failed: expected image '%s' actual '%s'", name, mock.expectedImage, d.Spec.Template.Spec.Containers[0].Image)
			}
		})
	}
}

func TestNewWithTemplate(t *testing.T) {
	tests := map[string]struct {
		daemon        string
		data          interface{}
		expectedName  string
		expectedImage string
	}{
		"t1": {`
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: {{.name}}
  namespace: kube-system
  labels:
    k8s-app: fluentd-logging
spec:
  selector:
    matchLabels:
      name: {{.name}}
  template:
    metadata:
      labels:
        name: {{.name}}
    spec:
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: fluentd-elasticsearch
        image: {{.image}}
        resources:
          limits:
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 200Mi
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
      terminationGracePeriodSeconds: 30
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
`,
			map[string]string{
				"name":  "fluentd-elasticsearch",
				"image": "k8s.gcr.io/fluentd-elasticsearch:1.20",
			},
			"fluentd-elasticsearch", "k8s.gcr.io/fluentd-elasticsearch:1.20"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			d := New(WithTemplate(mock.daemon, mock.data))
			if d.Name != mock.expectedName {
				t.Fatalf("test failed: expected name '%s' actual '%s'", mock.expectedName, d.Name)
			}
			if d.Spec.Template.Spec.Containers[0].Image != mock.expectedImage {
				t.Fatalf("test '%s' failed: expected image '%s' actual '%s'", name, mock.expectedImage, d.Spec.Template.Spec.Containers[0].Image)
			}
		})
	}
}
