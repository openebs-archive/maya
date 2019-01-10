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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestAsToleration(test *testing.T) {
	t := &toleration{
		corev1.Toleration{
			Key:      "app",
			Operator: corev1.TolerationOpExists,
			Value:    "jiva",
			Effect:   corev1.TaintEffectNoExecute,
		},
	}
	kt := t.asToleration()
	if kt.Key != t.Key || kt.Operator != t.Operator || kt.Value != t.Value || kt.Effect != t.Effect {
		test.Fatalf("test failed: expected '%+v' actual '%+v'", t, kt)
	}
}

func TestNoScheduleOnMaster(test *testing.T) {
	kt := NoScheduleOnMaster()
	if kt.Key != string(MasterNodeTolerationKey) || kt.Effect != corev1.TaintEffectNoSchedule {
		test.Fatalf("test failed: expected key '%s' effect '%s': actual '%+v'", MasterNodeTolerationKey, corev1.TaintEffectNoSchedule, kt)
	}
}

func TestBuilderNoSchedule(test *testing.T) {
	kt := Builder().NoSchedule().Build()
	if kt.Effect != corev1.TaintEffectNoSchedule {
		test.Fatalf("test failed: expected effect '%s': actual '%s'", corev1.TaintEffectNoSchedule, kt.Effect)
	}
}

func TestNewNoSchedule(test *testing.T) {
	kt := New(NoSchedule())
	if kt.Effect != corev1.TaintEffectNoSchedule {
		test.Fatalf("test failed: expected effect '%s': actual '%s'", corev1.TaintEffectNoSchedule, kt.Effect)
	}
}

func TestBuilderWithKey(t *testing.T) {
	tests := map[string]struct {
		key      string
		expected string
	}{
		"t1": {"openebs.io/key", "openebs.io/key"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			kt := Builder().WithKey(mock.key).Build()
			if kt.Key != mock.key {
				t.Fatalf("test failed: expected key '%s': actual '%s'", mock.key, kt.Key)
			}
		})
	}
}

func TestNewWithKey(t *testing.T) {
	tests := map[string]struct {
		key      string
		expected string
	}{
		"t1": {"openebs.io/key", "openebs.io/key"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			kt := New(WithKey(mock.key))
			if kt.Key != mock.key {
				t.Fatalf("test failed: expected key '%s': actual '%s'", mock.key, kt.Key)
			}
		})
	}
}

func TestBuilderWithEffect(t *testing.T) {
	tests := map[string]struct {
		effect corev1.TaintEffect
	}{
		"t1": {corev1.TaintEffectNoSchedule},
		"t2": {corev1.TaintEffectPreferNoSchedule},
		"t3": {corev1.TaintEffectNoExecute},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			kt := Builder().WithEffect(mock.effect).Build()
			if kt.Effect != mock.effect {
				t.Fatalf("test failed: expected effect '%s': actual '%s'", mock.effect, kt.Effect)
			}
		})
	}
}

func TestNewWithEffect(t *testing.T) {
	tests := map[string]struct {
		effect corev1.TaintEffect
	}{
		"t1": {corev1.TaintEffectNoSchedule},
		"t2": {corev1.TaintEffectPreferNoSchedule},
		"t3": {corev1.TaintEffectNoExecute},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			kt := New(WithEffect(mock.effect))
			if kt.Effect != mock.effect {
				t.Fatalf("test failed: expected effect '%s': actual '%s'", mock.effect, kt.Effect)
			}
		})
	}
}

func TestBuilderWithOperator(t *testing.T) {
	tests := map[string]struct {
		op corev1.TolerationOperator
	}{
		"t1": {corev1.TolerationOpExists},
		"t2": {corev1.TolerationOpEqual},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			kt := Builder().WithOperator(mock.op).Build()
			if kt.Operator != mock.op {
				t.Fatalf("test failed: expected operator '%s': actual '%s'", mock.op, kt.Operator)
			}
		})
	}
}

func TestNewWithOperator(t *testing.T) {
	tests := map[string]struct {
		op corev1.TolerationOperator
	}{
		"t1": {corev1.TolerationOpExists},
		"t2": {corev1.TolerationOpEqual},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			kt := New(WithOperator(mock.op))
			if kt.Operator != mock.op {
				t.Fatalf("test failed: expected operator '%s': actual '%s'", mock.op, kt.Operator)
			}
		})
	}
}

func TestBuilderWithValue(t *testing.T) {
	tests := map[string]struct {
		val string
	}{
		"t1": {"val1"},
		"t2": {"val2"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			kt := Builder().WithValue(mock.val).Build()
			if kt.Value != mock.val {
				t.Fatalf("test failed: expected value '%s': actual '%s'", mock.val, kt.Value)
			}
		})
	}
}

func TestNewWithValue(t *testing.T) {
	tests := map[string]struct {
		val string
	}{
		"t1": {"val1"},
		"t2": {"val2"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			kt := New(WithValue(mock.val))
			if kt.Value != mock.val {
				t.Fatalf("test failed: expected value '%s': actual '%s'", mock.val, kt.Value)
			}
		})
	}
}

func TestBuilderWithTolerationSeconds(t *testing.T) {
	tests := map[string]struct {
		seconds int64
	}{
		"t1": {300},
		"t2": {60},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			kt := Builder().WithTolerationSeconds(&mock.seconds).Build()
			if *kt.TolerationSeconds != mock.seconds {
				t.Fatalf("test failed: expected toleration seconds '%d': actual '%d'", mock.seconds, kt.TolerationSeconds)
			}
		})
	}
}

func TestNewWithTolerationSeconds(t *testing.T) {
	tests := map[string]struct {
		seconds int64
	}{
		"t1": {300},
		"t2": {60},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			kt := New(WithTolerationSeconds(&mock.seconds))
			if *kt.TolerationSeconds != mock.seconds {
				t.Fatalf("test failed: expected toleration seconds '%d': actual '%d'", mock.seconds, kt.TolerationSeconds)
			}
		})
	}
}
