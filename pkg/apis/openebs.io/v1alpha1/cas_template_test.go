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
	"strings"
	"testing"

	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCASConfigString(t *testing.T) {
	tests := map[string]struct {
		config              Config
		expectedStringParts []string
	}{
		"cas config": {
			Config{
				Name:    "Replicas",
				Value:   "2",
				Enabled: "false",
				List:    []string{"openebs.io/cpu-node"},
			},
			[]string{"name: Replicas", `value: "2"`, `enabled: "false"`, `"openebs.io/cpu-node"`},
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			ymlstr := mock.config.String()
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}

func TestCASConfigJSONIndent(t *testing.T) {
	tests := map[string]struct {
		context             string
		config              Config
		expectedStringParts []string
	}{
		"cas config": {
			"my cas config",
			Config{
				Name:    "Replicas",
				Value:   "2",
				Enabled: "false",
			},
			[]string{`"name": "Replicas"`, `"value": "2"`, `"enabled": "false"`},
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			jsonstr := stringer.JSONIndent(mock.context, mock.config)
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(jsonstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, jsonstr)
				}
			}
		})
	}
}

func TestCASTemplateString(t *testing.T) {
	tests := map[string]struct {
		cast                *CASTemplate
		expectedStringParts []string
	}{
		"castemplate": {
			&CASTemplate{
				ObjectMeta: metav1.ObjectMeta{Name: "my-cast", Namespace: "default"},
				Spec: CASTemplateSpec{
					TaskNamespace: "openebs",
					RunTasks: RunTasks{
						Tasks: []string{"rt1", "rt2"},
					},
					OutputTask: "rt3",
					Fallback:   "cast2",
				},
			},
			[]string{"taskNamespace: openebs", "- rt1", "- rt2", "output: rt3", "fallback: cast2"},
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			ymlstr := mock.cast.String()
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}

func TestCASTemplateJSONIndent(t *testing.T) {
	tests := map[string]struct {
		context             string
		cast                *CASTemplate
		expectedStringParts []string
	}{
		"castemplate": {
			"my castemplate",
			&CASTemplate{
				ObjectMeta: metav1.ObjectMeta{Name: "my-cast", Namespace: "default"},
				Spec: CASTemplateSpec{
					TaskNamespace: "openebs",
					RunTasks: RunTasks{
						Tasks: []string{"rt1", "rt2"},
					},
					OutputTask: "rt3",
					Fallback:   "cast2",
				},
			},
			[]string{`"taskNamespace": "openebs"`, "rt1", "rt2", `"output": "rt3"`, `"fallback": "cast2"`},
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			jsonstr := stringer.JSONIndent(mock.context, mock.cast)
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(jsonstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, jsonstr)
				}
			}
		})
	}
}

func TestRuntaskString(t *testing.T) {
	tests := map[string]struct {
		runtask             *RunTask
		expectedStringParts []string
	}{
		"runtask": {
			&RunTask{
				ObjectMeta: metav1.ObjectMeta{Name: "myrt", Namespace: "open"},
				Spec: RunTaskSpec{
					Meta:    "meta details",
					Task:    "task details",
					PostRun: "post run details",
				},
			},
			[]string{"meta: meta details", "task: task details", "post: post run details"},
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			ymlstr := mock.runtask.String()
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}

func TestRunTaskJSONIndent(t *testing.T) {
	tests := map[string]struct {
		context             string
		runtask             *RunTask
		expectedStringParts []string
	}{
		"runtask": {
			"my runtask",
			&RunTask{
				ObjectMeta: metav1.ObjectMeta{Name: "myrt", Namespace: "open"},
				Spec: RunTaskSpec{
					Meta:    "meta details",
					Task:    "task details",
					PostRun: "post run details",
				},
			},
			[]string{`"meta": "meta details"`, `"task": "task details"`, `"post": "post run details"`},
		},
	}
	for name, mock := range tests {
		mock := mock // pin it
		name := name // pin it
		t.Run(name, func(t *testing.T) {
			jsonstr := stringer.JSONIndent(mock.context, mock.runtask)
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(jsonstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, jsonstr)
				}
			}
		})
	}
}
