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
	tests := map[string]struct {
		yaml  string
		isErr bool
	}{
		"blank yaml": {"", true},
		"hello yaml": {"Hello!!", false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			my := &MayaYaml{
				Yaml: test.yaml,
			}

			_, err := my.Bytes()

			if !test.isErr && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}
		})
	}
}

func TestMayaAnyK8sGenerateService(t *testing.T) {
	tests := map[string]struct {
		kind    string
		yaml    string
		isError bool
	}{
		"blank service":   {kind: "service", yaml: "", isError: true},
		"hello service":   {kind: "service", yaml: "Hello!!", isError: true},
		"invalid service": {kind: "blah", yaml: "Junk!!", isError: true},
		"valid service": {kind: "service", yaml: `
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  ports:
  - name: api
    port: 5656
    protocol: TCP
    targetPort: 5656
  selector:
    name: maya-apiserver
  sessionAffinity: None
`, isError: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ma := &MayaAnyK8s{
				Kind: test.kind,
				MayaYaml: MayaYaml{
					Yaml: test.yaml,
				},
			}
			s, err := ma.GenerateService()

			if !test.isError && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}

			if test.isError && s != nil {
				t.Fatalf("Expected: 'nil service' Actual: '%v'", s)
			}
		})
	}
}

func TestMayaAnyK8sGenerateDeployment(t *testing.T) {
	tests := map[string]struct {
		kind    string
		yaml    string
		isError bool
	}{
		"blank deployment":   {kind: "deployment", yaml: "", isError: true},
		"hello deployment":   {kind: "deployment", yaml: "Hello!!", isError: true},
		"invalid deployment": {kind: "blah", yaml: "Junk!!", isError: true},
		"valid deployment": {kind: "deployment", yaml: `
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
`, isError: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ma := &MayaAnyK8s{
				Kind: test.kind,
				MayaYaml: MayaYaml{
					Yaml: test.yaml,
				},
			}
			d, err := ma.GenerateDeployment()

			if !test.isError && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}

			if test.isError && d != nil {
				t.Fatalf("Expected: 'nil deployment' Actual: '%v'", d)
			}
		})
	}
}

func TestMayaConfigMapLoad(t *testing.T) {
	tests := map[string]struct {
		yaml    string
		isError bool
	}{
		"blank config map": {yaml: "", isError: true},
		"hello config map": {yaml: "Hello!!", isError: true},
		"valid config map": {yaml: `
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

func TestMayaConfigMapLoadEmbeddedK8s(t *testing.T) {
	tests := map[string]struct {
		yaml          string
		isK8sEmbedded bool
	}{
		"config map without embedded k8s": {yaml: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-hello
`, isK8sEmbedded: false},
		"config map with embedded k8s": {yaml: `
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
			err := mc.LoadEmbeddedK8s()

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

func TestMayaServiceLoad(t *testing.T) {
	tests := map[string]struct {
		yaml    string
		isError bool
	}{
		"blank service": {yaml: "", isError: true},
		"hello service": {yaml: "Hello!!", isError: true},
		"valid service": {yaml: `
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  ports:
  - name: api
    port: 5656
    protocol: TCP
    targetPort: 5656
  selector:
    name: maya-apiserver
  sessionAffinity: None
`, isError: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			s := NewMayaService(test.yaml)
			err := s.Load()

			if !test.isError && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}

			if test.isError && s.Service != nil {
				t.Fatalf("Expected: 'nil service' Actual: '%v'", s.Service)
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
	tests := map[string]struct {
		yaml  string
		isErr bool
	}{
		"blank yaml": {"", true},
		"hello yaml": {"Hello!!", true},
		"valid yaml": {`
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

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			md := NewMayaDeployment(test.yaml)

			err := md.Load()

			if !test.isErr && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}

			if test.isErr && md.Deployment != nil {
				t.Fatalf("Expected: 'nil maya deployment' Actual: '%v'", md)
			}
		})
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

	tests := map[string]struct {
		containerYaml string
		isError       bool
	}{
		"blank yaml": {containerYaml: "", isError: true},
		"hello yaml": {containerYaml: "Hello!!", isError: true},
		"valid struct yaml": {containerYaml: `
name: maya-apiserver-2
imagePullPolicy: Always
image: openebs/m-apiserver:test
ports:
- containerPort: 5657
`, isError: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := md.AddContainer(test.containerYaml)

			if !test.isError && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}

			if !test.isError && len(md.Deployment.Spec.Template.Spec.Containers) != 2 {
				t.Fatalf("Expected: '2 containers in maya deployment' Actual: '%d'", len(md.Deployment.Spec.Template.Spec.Containers))
			}
		})
	}
}
