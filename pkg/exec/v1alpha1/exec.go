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
	"os/exec"
	"time"

	"context"

	types "github.com/openebs/maya/pkg/exec"
	"github.com/pkg/errors"
)

// runner implements the Runner interface
type runner struct {
	timeout time.Duration
	command string
	args    []string
}

// Builder is used for filling the values of runner struct
type Builder struct {
	runner *runner
}

// StdoutBuilder returns instance of builder
func StdoutBuilder() *Builder {
	return &Builder{runner: &runner{}}
}

// WithTimeout fill timeout field of runner struct
func (b *Builder) WithTimeout(timeout time.Duration) *Builder {
	b.runner.timeout = timeout
	return b
}

// WithCommand fill command field of runner struct
func (b *Builder) WithCommand(cmd string) *Builder {
	b.runner.command = cmd
	return b
}

// WithArgs fill args field of runner struct
func (b *Builder) WithArgs(args ...string) *Builder {
	b.runner.args = args
	return b
}

// Build returns the instance of runner
func (b *Builder) Build() types.Runner {
	return b.runner
}

// RunCommandWithTimeoutContext executes command provides and returns stdout
// error. If command does not returns within given timout interval command will
// be killed and return "Context time exceeded"
func (r runner) RunCommandWithTimeoutContext() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	// #nosec
	out, err := exec.CommandContext(ctx, r.command, r.args...).CombinedOutput()
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, errors.Wrapf(ctx.Err(), "Failed to run command: %v %v", r.command, r.args)
		default:
			return nil, err
		}
	}
	return out, nil
}
