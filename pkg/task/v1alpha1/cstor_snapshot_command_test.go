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
	"testing"
)

func TestCstorSnapshotCommand(t *testing.T) {
	tests := map[string]struct {
		action            RunCommandAction
		isSupportedAction bool
	}{
		"test 101": {DeleteCommandAction, true},
		"test 102": {CreateCommandAction, true},
		"test 103": {ListCommandAction, false},
		"test 104": {GetCommandAction, false},
		"test 105": {PatchCommandAction, false},
		"test 106": {UpdateCommandAction, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := WithAction(Command(), mock.action)
			c := &cstorSnapshotCommand{cmd}
			result := c.Run()

			if !mock.isSupportedAction && result.Error() != ErrorNotSupportedAction {
				t.Fatalf("Test '%s' failed: expected 'ErrorNotSupportedAction': actual '%s': result '%s'", name, result.Error(), result)
			}

			if mock.isSupportedAction && result.Error() == ErrorNotSupportedAction {
				t.Fatalf("Test '%s' failed: expected 'supported action': actual 'ErrorNotSupportedAction': result '%s'", name, result)
			}
		})
	}
}
