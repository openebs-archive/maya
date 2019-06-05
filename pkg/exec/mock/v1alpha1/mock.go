// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"errors"

	types "github.com/openebs/maya/pkg/exec"
)

// runner is used to mock os/exec.
// Expected data is set against its
// instance.
type runner struct {
	// output is the expected return
	// value
	output []byte

	// isError flags this instance
	// to return error
	isError bool
}

// Builder ...
type Builder struct {
	runner *runner
}

// StdoutBuilder returns instance of Builder
func StdoutBuilder() *Builder {
	return &Builder{runner: &runner{}}
}

// WithOutput fills output field of runner struct
func (b *Builder) WithOutput(output string) *Builder {
	b.runner.output = []byte(output)
	return b
}

// Error set isError field of runner to true
func (b *Builder) Error() *Builder {
	b.runner.isError = true
	return b
}

// Build returns the instance of runner
func (b *Builder) Build() types.Runner {
	return b.runner
}

// RunCommandWithTimeoutContext mock the behaviour of actual RunCommandWithTimeoutContext.
func (r *runner) RunCommandWithTimeoutContext() ([]byte, error) {
	if r.isError {
		return nil, errors.New("dummy error")
	}
	return r.output, nil
}
