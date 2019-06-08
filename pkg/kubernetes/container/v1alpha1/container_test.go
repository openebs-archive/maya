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
	"reflect"
	"testing"
)

// fakeAlwaysTrue is a concrete implementation of container Predicate
func fakeAlwaysTrue(d *container) (string, bool) {
	return "fakeAlwaysTrue", true
}

// fakeAlwaysFalse is a concrete implementation of container Predicate
func fakeAlwaysFalse(d *container) (string, bool) {
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
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			e := predicateFailedError(mock.predicateMessage)
			if e.Error() != mock.expectedErr {
				t.Fatalf("test '%s' failed: expected '%s': actual '%s'", name, mock.expectedErr, e.Error())
			}
		})
	}
}

func TestNewWithName(t *testing.T) {
	n := "con1"
	c := New(WithName(n))
	if c.Name != n {
		t.Fatalf("test failed: expected name '%s': actual '%s'", n, c.Name)
	}
}

func TestNewWithImage(t *testing.T) {
	i := "openebs.io/m-apiserver:1.0.0"
	c := New(WithImage(i))
	if c.Image != i {
		t.Fatalf("test failed: expected image '%s': actual '%s'", i, c.Image)
	}
}

func TestNewWithCommand(t *testing.T) {
	cmd := []string{"kubectl", "get", "po"}
	c := New(WithCommand(cmd))
	if !reflect.DeepEqual(c.Command, cmd) {
		t.Fatalf("test failed: expected command '%q': actual '%q'", cmd, c.Command)
	}
}

func TestNewWithArguments(t *testing.T) {
	args := []string{"-o", "yaml"}
	c := New(WithArguments(args))
	if !reflect.DeepEqual(c.Args, args) {
		t.Fatalf("test failed: expected arguments '%q': actual '%q'", args, c.Args)
	}
}

func TestBuilderBuild(t *testing.T) {
	_, err := NewBuilder().Build()
	if err != nil {
		t.Fatalf("test failed: expected no err: actual '%+v'", err)
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
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			_, err := NewBuilder().AddChecks(mock.checks).Build()
			if mock.isError && err == nil {
				t.Fatalf("test '%s' failed: expected error: actual no error", name)
			}
			if !mock.isError && err != nil {
				t.Fatalf("test '%s' failed: expected no error: actual error '%+v'", name, err)
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
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().AddChecks(mock.checks)
			if len(b.checks) != mock.expectedCount {
				t.Fatalf("test '%s' failed: expected no of checks '%d': actual '%d'", name, mock.expectedCount, len(b.checks))
			}
		})
	}
}

func TestBuilderWithName(t *testing.T) {
	tests := map[string]struct {
		name         string
		expectedName string
	}{
		"t1": {"nginx", "nginx"},
		"t2": {"maya", "maya"},
		"t3": {"ndm", "ndm"},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			c, _ := NewBuilder().WithName(mock.name).Build()
			if c.Name != mock.expectedName {
				t.Fatalf("test '%s' failed: expected name '%s': actual '%s'", name, mock.expectedName, c.Name)
			}
		})
	}
}

func TestBuilderWithImage(t *testing.T) {
	tests := map[string]struct {
		image         string
		expectedImage string
	}{
		"t1": {"nginx:1.0.0", "nginx:1.0.0"},
		"t2": {"openebs.io/maya:1.0", "openebs.io/maya:1.0"},
		"t3": {"openebs.io/ndm:1.0", "openebs.io/ndm:1.0"},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			c, _ := NewBuilder().WithImage(mock.image).Build()
			if c.Image != mock.expectedImage {
				t.Fatalf("test '%s' failed: expected image '%s': actual '%s'", name, mock.expectedImage, c.Image)
			}
		})
	}
}

func TestBuilderWithCommand(t *testing.T) {
	tests := map[string]struct {
		cmd         []string
		expectedCmd []string
	}{
		"t1": {[]string{"kubectl", "get", "po"}, []string{"kubectl", "get", "po"}},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			c, _ := NewBuilder().WithCommand(mock.cmd).Build()
			if !reflect.DeepEqual(c.Command, mock.expectedCmd) {
				t.Fatalf("test '%s' failed: expected command '%q': actual '%q'", name, mock.expectedCmd, c.Command)
			}
		})
	}
}

func TestBuilderWithArguments(t *testing.T) {
	tests := map[string]struct {
		args         []string
		expectedArgs []string
	}{
		"t1": {[]string{"-jsonpath", "metadata.name"}, []string{"-jsonpath", "metadata.name"}},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			c, _ := NewBuilder().WithArguments(mock.args).Build()
			if !reflect.DeepEqual(c.Args, mock.expectedArgs) {
				t.Fatalf("test '%s' failed: expected arguments '%q': actual '%q'", name, mock.expectedArgs, c.Args)
			}
		})
	}
}
