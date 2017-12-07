/*
Copyright 2017 The OpenEBS Authors

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

package maya

import (
	"testing"
)

func TestMayaBytes(t *testing.T) {
	tests := []struct {
		yaml  string
		isErr bool
	}{
		{"", true},
		{"Hello!!", false},
	}

	for _, test := range tests {
		my := &MayaYaml{
			Yaml: test.yaml,
		}

		_, err := my.Bytes()

		if !test.isErr && err != nil {
			t.Fatalf("Expected: 'no error' Actual: '%s'", err)
		}
	}
}

func TestMayaConfigMapLoad(t *testing.T) {
	tests := map[string]struct {
		yaml    string
		isError bool
	}{
		"blank yaml": {yaml: "", isError: true},
		"hello yaml": {yaml: "Hello!!", isError: true},
		"valid struct yaml": {yaml: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-hello
data:
  hey: hello
`, isError: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mc := NewMayaConfigMap(test.yaml)
			err := mc.Load()

			if !test.isError && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}

			if test.isError && mc.ConfigMap != nil {
				t.Fatalf("Expected: 'nil maya config map' Actual: '%v'", mc.ConfigMap)
			}
		})
	}
}

func TestMayaConfigMapLoadAll(t *testing.T) {
	tests := map[string]struct {
		yaml          string
		isK8sEmbedded bool
	}{
		"no embedded k8s yaml": {yaml: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-hello
`, isK8sEmbedded: false},
		"embedded k8s yaml": {yaml: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-hello
data:
  namespace: default
  apiVersion: v1
  kind: ConfigMap
  yaml: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: embedded-cm
`, isK8sEmbedded: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mc := NewMayaConfigMap(test.yaml)
			err := mc.LoadAll()

			if err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}

			if test.isK8sEmbedded && mc.EK8sObject == nil {
				t.Fatalf("Expected: 'embedded k8s object' Actual: '%v'", mc.EK8sObject)
			}

			if test.isK8sEmbedded && len(mc.EK8sObject.Kind) == 0 {
				t.Fatalf("Expected: 'embedded k8s kind' Actual: '%s'", mc.EK8sObject.Kind)
			}
		})
	}
}

func TestMayaContainerLoad(t *testing.T) {
	tests := []struct {
		yaml  string
		isErr bool
	}{
		{"", true},
		{"Hello!!", true},
		{`
name: maya-apiserver
image: openebs/m-apiserver:test
ports:
- containerPort: 5656
`, false},
	}

	for _, test := range tests {
		mc := NewMayaContainer(test.yaml)

		err := mc.Load()

		if !test.isErr && err != nil {
			t.Fatalf("Expected: 'no error' Actual: '%s'", err)
		}

		if test.isErr && len(mc.Container.Name) != 0 {
			t.Fatalf("Expected: 'nil maya container' Actual: '%v'", mc.Container)
		}
	}
}

func TestMayaDeploymentLoad(t *testing.T) {
	tests := []struct {
		yaml  string
		isErr bool
	}{
		{"", true},
		{"Hello!!", true},
		{`
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: maya-apiserver
  namespace: default
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: maya-apiserver
    spec:
      serviceAccountName: openebs-maya-operator
      containers:
      - name: maya-apiserver
        imagePullPolicy: Always
        image: openebs/m-apiserver:test
        ports:
        - containerPort: 5656
`, false},
	}

	for _, test := range tests {
		md := NewMayaDeployment(test.yaml)

		err := md.Load()

		if !test.isErr && err != nil {
			t.Fatalf("Expected: 'no error' Actual: '%s'", err)
		}

		if test.isErr && md.Deployment != nil {
			t.Fatalf("Expected: 'nil maya deployment' Actual: '%v'", md)
		}
	}
}

func TestMayaDeploymentAddContainer(t *testing.T) {

	deploy := `
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: maya-apiserver
  namespace: default
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: maya-apiserver
    spec:
      serviceAccountName: openebs-maya-operator
      containers:
      - name: maya-apiserver
        imagePullPolicy: Always
        image: openebs/m-apiserver:test
        ports:
        - containerPort: 5656
`

	md := NewMayaDeployment(deploy)
	err := md.Load()
	if err != nil {
		t.Fatalf("Error in test logic: '%v'", err)
	}

	tests := []struct {
		containerYaml string
		isError       bool
	}{
		{containerYaml: "", isError: true},
		{containerYaml: "Hello!!", isError: true},
		{containerYaml: `
name: maya-apiserver-2
imagePullPolicy: Always
image: openebs/m-apiserver:test
ports:
- containerPort: 5657
`, isError: false},
	}

	for _, test := range tests {
		err := md.AddContainer(test.containerYaml)

		if !test.isError && err != nil {
			t.Fatalf("Expected: 'no error' Actual: '%s'", err)
		}

		if !test.isError && len(md.Deployment.Spec.Template.Spec.Containers) != 2 {
			t.Fatalf("Expected: '2 containers in maya deployment' Actual: '%d'", len(md.Deployment.Spec.Template.Spec.Containers))
		}
	}
}
