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
	"strings"
	"testing"

	"github.com/openebs/maya/pkg/util"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newCustomFuncsHolder is a utility function that returns a new instance of
// CustomFuncsHolder
func newCustomFuncsHolder(key, value, ctx string) *CustomFuncsHolder {
	if len(key) == 0 && len(value) == 0 {
		return &CustomFuncsHolder{}
	}

	pairs := map[string]string{
		key: value,
	}

	if ctx == "inputs" {
		return &CustomFuncsHolder{
			Inputs: pairs,
		}
	} else {
		return &CustomFuncsHolder{
			Stores: pairs,
		}
	}
}

func TestCustomFuncsHolderInputVal(t *testing.T) {
	tests := map[string]struct {
		key   string
		val   string
		isErr bool
	}{
		"valid input":                {key: "abc", val: "xyz", isErr: false},
		"empty key as input":         {key: "", val: "xyz", isErr: true},
		"empty value as input":       {key: "abc", val: "", isErr: true},
		"empty key & value as input": {key: "", val: "", isErr: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			h := newCustomFuncsHolder(test.key, test.val, "inputs")
			_, err := h.inputVal(test.key)

			if !test.isErr && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}
		})
	}
}

func TestCustomFuncsHolderMergeStores(t *testing.T) {
	tests := map[string]struct {
		key1   string
		key2   string
		val1   string
		val2   string
		optype string
	}{
		"new":                   {key1: "", val1: "", key2: "k", val2: "v", optype: "new"},
		"append":                {key1: "k1", val1: "v1", key2: "k2", val2: "v2", optype: "append"},
		"merge if empty":        {key1: "k1", val1: "", key2: "k1", val2: "v2", optype: "mergeifempty"},
		"no merge if not empty": {key1: "k1", val1: "v1", key2: "k1", val2: "v2", optype: "nomergeifnotempty"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			h := newCustomFuncsHolder(test.key1, test.val1, "stores")
			h.mergeStoresIfEmpty(map[string]string{
				test.key2: test.val2,
			})

			if test.optype == "new" && h.Stores[test.key2] != test.val2 {
				t.Fatalf("Expected: 'addition to stores' Actual: '%v'", h.Stores)
			}

			if test.optype == "append" && h.Stores[test.key1] != test.val1 {
				t.Fatalf("Expected: 'appending to stores' Actual: '%v'", h.Stores)
			}

			if test.optype == "append" && h.Stores[test.key2] != test.val2 {
				t.Fatalf("Expected: 'appending to stores' Actual: '%v'", h.Stores)
			}

			if test.optype == "mergeifempty" && h.Stores[test.key1] != test.val2 {
				t.Fatalf("Expected: 'merging of stores' Actual: '%v'", h.Stores)
			}

			if test.optype == "nomergeifnotempty" && h.Stores[test.key1] != test.val1 {
				t.Fatalf("Expected: 'no merging of stores' Actual: '%v'", h.Stores)
			}
		})
	}
}

func TestCustomFuncsHolderMergeInputsIfEmpty(t *testing.T) {
	tests := map[string]struct {
		key1   string
		key2   string
		val1   string
		val2   string
		optype string
	}{
		"new":            {key1: "", val1: "", key2: "k", val2: "v", optype: "new"},
		"append":         {key1: "k1", val1: "v1", key2: "k2", val2: "v2", optype: "append"},
		"merge if empty": {key1: "k1", val1: "", key2: "k1", val2: "v2", optype: "mergeifempty"},
		"nomerge":        {key1: "k1", val1: "v1", key2: "k1", val2: "v2", optype: "nomerge"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			h := newCustomFuncsHolder(test.key1, test.val1, "inputs")
			h.mergeInputsIfEmpty(map[string]string{
				test.key2: test.val2,
			})

			if test.optype == "new" && h.Inputs[test.key2] != test.val2 {
				t.Fatalf("Expected: 'new inputs' Actual: '%v'", h.Inputs)
			}

			if test.optype == "append" && h.Inputs[test.key1] != test.val1 {
				t.Fatalf("Expected: 'addition to inputs' Actual: '%v'", h.Inputs)
			}

			if test.optype == "append" && h.Inputs[test.key2] != test.val2 {
				t.Fatalf("Expected: 'addition to inputs' Actual: '%v'", h.Inputs)
			}

			if test.optype == "mergeifempty" && h.Inputs[test.key1] != test.val2 {
				t.Fatalf("Expected: 'merging of inputs' Actual: '%v'", h.Inputs)
			}

			if test.optype == "nomerge" && h.Inputs[test.key1] != test.val1 {
				t.Fatalf("Expected: 'no merging of inputs' Actual: '%v'", h.Inputs)
			}
		})
	}
}

func TestMayaYamlV2AsMapOfObjects(t *testing.T) {
	tests := map[string]struct {
		yaml    string
		inputk1 string
		inputv1 string
		inputk2 string
		inputv2 string
		inputk3 string
		inputv3 string
		isErr   bool
	}{
		"pass inputs & verify as well": {
			inputk1: "type",
			inputv1: "testing",
			inputk2: "operator",
			inputv2: "testops",
			inputk3: "operator-version",
			inputv3: "v2",
			isErr:   false,
			yaml: `
apiVersion: v1
kind: Service
metadata:
  name: ABC
  labels:
    openebs.io/type: {{inputs "type"}}
    openebs.io/operator: {{inputs "operator"}}
    operator.openebs.io/version: {{inputs "operator-version"}}
`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			y := &MayaYamlV2{
				Yaml: test.yaml,
				CustomFuncsHolder: CustomFuncsHolder{
					Inputs: map[string]string{
						test.inputk1: test.inputv1,
						test.inputk2: test.inputv2,
						test.inputk3: test.inputv3,
					},
				},
			}

			obj, err := y.asMapOfObjects()

			if !test.isErr && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}

			val1 := util.GetNestedString(obj, []string{"metadata", "labels", "openebs.io/type"}...)
			val2 := util.GetNestedString(obj, []string{"metadata", "labels", "openebs.io/operator"}...)
			val3 := util.GetNestedString(obj, []string{"metadata", "labels", "operator.openebs.io/version"}...)

			if !test.isErr && test.inputv1 != val1 {
				t.Fatalf("Expected: '%s' Actual: '%s'", test.inputv1, val1)
			}

			if !test.isErr && test.inputv2 != val2 {
				t.Fatalf("Expected: '%s' Actual: '%s'", test.inputv2, val2)
			}

			if !test.isErr && test.inputv3 != val3 {
				t.Fatalf("Expected: '%s' Actual: '%s'", test.inputv3, val3)
			}
		})
	}
}

func TestTemplateMetaAsTemplateMeta(t *testing.T) {

	yml := `
apiVersion: {{inputs "version"}}
kind: {{inputs "kind"}}
namespace: {{inputs "namespace"}}
action: {{inputs "action"}}
`

	tests := map[string]struct {
		yaml      string
		kindKey   string
		kindVal   string
		verKey    string
		verVal    string
		nsKey     string
		nsVal     string
		actionKey string
		actionVal string
		version   string
		isErr     bool
	}{
		"meta for deployment": {
			verKey:    "version",
			verVal:    "extensions/v1beta1",
			kindKey:   "kind",
			kindVal:   "Deployment",
			nsKey:     "namespace",
			nsVal:     "default",
			actionKey: "action",
			actionVal: "get",
			isErr:     false,
			yaml:      yml,
		},
		"meta for service": {
			verKey:    "version",
			verVal:    "v1",
			kindKey:   "kind",
			kindVal:   "Service",
			nsKey:     "namespace",
			nsVal:     "system",
			actionKey: "action",
			actionVal: "put",
			isErr:     false,
			yaml:      yml,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewTemplateMeta(test.yaml, map[string]string{
				test.verKey:    test.verVal,
				test.kindKey:   test.kindVal,
				test.nsKey:     test.nsVal,
				test.actionKey: test.actionVal,
			})

			m, err := m.asTemplateMeta()

			if !test.isErr && err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%s'", err)
			}

			if !test.isErr && test.verVal != m.APIVersion {
				t.Fatalf("Expected: '%s' Actual: '%s'", test.verVal, m.APIVersion)
			}

			if !test.isErr && test.kindVal != m.Kind {
				t.Fatalf("Expected: '%s' Actual: '%s'", test.kindVal, m.Kind)
			}

			if !test.isErr && test.nsVal != m.Namespace {
				t.Fatalf("Expected: '%s' Actual: '%s'", test.nsVal, m.Namespace)
			}

			if !test.isErr && MayaRunAction(test.actionVal) != m.Action {
				t.Fatalf("Expected: '%s' Actual: '%s'", test.actionVal, m.Action)
			}
		})
	}
}

// TestMayaBytes tests if a yaml marshalls to bytes.
// The tests are written in a table format to provide various
// scenarios of test data.
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

// MockMayaAnyK8s will be helpful in mocking the methods
// of MayaAnyK8s
type MockMayaAnyK8s struct {
	kind       string
	apiVersion string
	owner      string
	suffixName string
	yaml       string
	isError    bool
}

func (m MockMayaAnyK8s) NewMayaAnyK8s() *MayaAnyK8s {
	return &MayaAnyK8s{
		Kind:       m.kind,
		APIVersion: m.apiVersion,
		MayaYaml: MayaYaml{
			Yaml: m.yaml,
			MayaPlaceholders: MayaPlaceholders{
				Owner: m.owner,
			},
		},
	}
}

func (m MockMayaAnyK8s) TestObjectMeta(obj mach_apis_meta_v1.ObjectMeta, err error, t *testing.T) {
	if !m.isError && err != nil {
		t.Fatalf("Expected: 'no error' Actual: '%s'", err)
	}

	if !m.isError && !strings.HasPrefix(obj.Name, m.owner) {
		t.Fatalf("Expected Name: '%s%s' Actual: '%s'", m.owner, m.suffixName, obj.Name)
	}

	if !m.isError && !strings.HasSuffix(obj.Name, m.suffixName) {
		t.Fatalf("Expected Name: '%s%s' Actual: '%s'", m.owner, m.suffixName, obj.Name)
	}
}

func (m MockMayaAnyK8s) TestTypeMeta(meta mach_apis_meta_v1.TypeMeta, err error, t *testing.T) {
	if !m.isError && err != nil {
		t.Fatalf("Expected: 'no error' Actual: '%s'", err)
	}

	if !m.isError && meta.APIVersion != m.apiVersion {
		t.Fatalf("Expected APIVersion: '%s' Actual: '%s'", m.apiVersion, meta.APIVersion)
	}

	if !m.isError && meta.Kind != m.kind {
		t.Fatalf("Expected Kind: '%s' Actual: '%s'", m.kind, meta.Kind)
	}
}

// TestMayaAnyK8sGenerateService tests if a yaml
// marshalls to a K8s Service object. The tests are
// written in a table format to provide various
// scenarios of test data.
func TestMayaAnyK8sGenerateService(t *testing.T) {
	tests := map[string]struct {
		kind    string
		yaml    string
		isError bool
	}{
		"blank service":   {kind: "Service", yaml: "", isError: true},
		"hello service":   {kind: "Service", yaml: "Hello!!", isError: true},
		"invalid service": {kind: "blah", yaml: "Junk!!", isError: true},
		"valid service": {
			kind:    "Service",
			isError: false,
			yaml: `
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
`},
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

// TestMayaAnyK8sGenerateServiceTemplated tests if a templated yaml
// marshalls to a K8s Service object
func TestMayaAnyK8sGenerateServiceTemplated(t *testing.T) {
	tests := map[string]MockMayaAnyK8s{
		"templated service": {
			kind:       "Service",
			apiVersion: "v1",
			owner:      "pv-123-abc",
			suffixName: "-svc",
			isError:    false,
			yaml: `
apiVersion: {{.APIVersion}}
kind: {{.Kind}}
metadata:
  name: {{.Owner}}-svc
spec:
  ports:
  - name: api
    port: 5656
    protocol: TCP
    targetPort: 5656
  selector:
    name: maya-apiserver
  sessionAffinity: None
`},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			ma := mock.NewMayaAnyK8s()
			s, err := ma.GenerateService()

			if mock.isError && s != nil {
				t.Fatalf("Expected: 'nil service' Actual: '%v'", s)
			}
			mock.TestObjectMeta(s.ObjectMeta, err, t)
			mock.TestTypeMeta(s.TypeMeta, err, t)
		})
	}
}

// TestMayaAnyK8sGenerateDeployment tests if a yaml
// marshalls to a K8s Deployment object. The tests are
// written in a table format to provide various
// scenarios of test data.
func TestMayaAnyK8sGenerateDeployment(t *testing.T) {
	tests := map[string]struct {
		kind    string
		yaml    string
		isError bool
	}{
		"blank deployment":   {kind: "Deployment", yaml: "", isError: true},
		"hello deployment":   {kind: "Deployment", yaml: "Hello!!", isError: true},
		"invalid deployment": {kind: "blah", yaml: "Junk!!", isError: true},
		"valid deployment": {kind: "Deployment", yaml: `
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

// TestMayaAnyK8sGenerateDeploymentTemplated tests if a
// templated yaml marshalls to a K8s Deployment object
func TestMayaAnyK8sGenerateDeploymentTemplated(t *testing.T) {
	tests := map[string]MockMayaAnyK8s{
		"templated deployment": {
			kind:       "Deployment",
			apiVersion: "apps/v1beta1",
			owner:      "pv-123",
			suffixName: "-dep",
			isError:    false,
			yaml: `
apiVersion: {{.APIVersion}}
kind: {{.Kind}}
metadata:
  name: {{.Owner}}-dep
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
`},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			ma := mock.NewMayaAnyK8s()
			d, err := ma.GenerateDeployment()

			if mock.isError && d != nil {
				t.Fatalf("Expected: 'nil deployment' Actual: '%v'", d)
			}
			mock.TestObjectMeta(d.ObjectMeta, err, t)
			mock.TestTypeMeta(d.TypeMeta, err, t)
		})
	}
}

// TestMayaAnyK8sGenerateCofigMap tests if a yaml marshalls
// to ConfigMap. The tests are written in a table format to
// provide various scenarios of test data.
func TestMayaAnyK8sGenerateCofigMap(t *testing.T) {
	tests := map[string]MockMayaAnyK8s{
		"blank yaml is invalid configmap": {
			kind:    "ConfigMap",
			isError: true,
			yaml:    "",
		},
		"hello yaml is invalid configmap": {
			kind:    "ConfigMap",
			isError: true,
			yaml:    "Hello",
		},
		"valid configmap": {
			apiVersion: "v1",
			kind:       "ConfigMap",
			owner:      "my",
			suffixName: "-cm",
			isError:    false,
			yaml: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-cm
data:
  x: y
  a: b
  c: d
`},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			ma := mock.NewMayaAnyK8s()
			cm, err := ma.GenerateConfigMap()

			if mock.isError && cm != nil {
				t.Fatalf("Expected: 'nil configmap' Actual: '%v'", cm)
			}

			if !mock.isError {
				mock.TestObjectMeta(cm.ObjectMeta, err, t)
				mock.TestTypeMeta(cm.TypeMeta, err, t)
			}
		})
	}
}

// TestMayaAnyK8sGenerateCofigMapTemplated tests if a
// templated yaml marshalls to a K8s ConfigMap object
func TestMayaAnyK8sGenerateCofigMapTemplated(t *testing.T) {
	tests := map[string]MockMayaAnyK8s{
		"a templated as well as valid configmap": {
			kind:       "ConfigMap",
			apiVersion: "v1",
			owner:      "pv-123",
			suffixName: "-cm",
			isError:    false,
			yaml: `
apiVersion: {{.APIVersion}}
kind: {{.Kind}}
metadata:
  name: {{.Owner}}-cm
data:
  x: y
  a: b
  c: d
`},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			ma := mock.NewMayaAnyK8s()
			cm, err := ma.GenerateConfigMap()

			if mock.isError && cm != nil {
				t.Fatalf("Expected: 'nil deployment' Actual: '%v'", cm)
			}
			mock.TestObjectMeta(cm.ObjectMeta, err, t)
			mock.TestTypeMeta(cm.TypeMeta, err, t)
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
