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

package retry

import (
	"fmt"
	"time"
)

// Action ...
type Action func(attempt uint) error

// Model ...
type Model struct {
	retry    uint
	waitTime time.Duration
}

// Times ...
func Times(retry uint) *Model {
	model := Model{}
	return model.Times(retry)
}

// Times ...
func (model *Model) Times(retry uint) *Model {
	model.retry = retry
	return model
}

// Wait ...
func Wait(waitTime time.Duration) *Model {
	model := Model{}
	return model.Wait(waitTime)
}

// Wait ...
func (model *Model) Wait(waitTime time.Duration) *Model {
	model.waitTime = waitTime
	return model
}

// Try ...
func (model Model) Try(action Action) error {
	if action == nil {
		return fmt.Errorf("no action specified")
	}

	var err error
	for attempt := uint(0); (0 == attempt || nil != err) && attempt <= model.retry; attempt++ {
		if attempt > 0 && model.waitTime > 0 {
			time.Sleep(model.waitTime)
		}

		err = action(attempt)
	}

	return err
}
