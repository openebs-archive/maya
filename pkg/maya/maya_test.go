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

func TestMayaContainerReload(t *testing.T) {
	tests := []struct {
		yaml  string
		isErr bool
	}{
		{"", true},
		{"Hello!!", true},
		{`
name: maya-apiserver
imagePullPolicy: Always
image: openebs/m-apiserver:test
ports:
- containerPort: 5656
`, false},
	}

	for _, test := range tests {
		mc := NewMayaContainer("")

		err := mc.Reload(test.yaml)

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

	tests := []struct {
		containerYaml string
		isErr         bool
	}{
		{"", true},
		{"Hello!!", true},
		{`
name: maya-apiserver-2
image: openebs/m-apiserver:test
`, false},
	}

	for _, test := range tests {
		md := NewMayaDeployment(deploy).SetMayaContainer()
		md.Load()
		err := md.AddContainer(test.containerYaml)

		if !test.isErr && err != nil {
			t.Fatalf("Expected: 'no error' Actual: '%s'", err)
		}

		if test.isErr && len(md.MayaContainer.Container.Name) != 0 {
			t.Fatalf("Expected: 'nil maya deployment container' Actual: '%v'", md.MayaContainer.Container)
		}

		if !test.isErr && len(md.Deployment.Spec.Template.Spec.Containers) != 2 {
			t.Fatalf("Expected: '2 containers in maya deployment' Actual: '%d'", len(md.Deployment.Spec.Template.Spec.Containers))
		}
	}
}
