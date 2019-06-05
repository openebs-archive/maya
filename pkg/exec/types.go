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

package exec

// Runner interface implements various methods of running binaries which can be
// modified for unit testing.
type Runner interface {
	RunCommandWithTimeoutContext() ([]byte, error)
}

// BuilderInterface is used for building the object
// of runner
type BuilderInterface interface {
	Build() Runner
}
