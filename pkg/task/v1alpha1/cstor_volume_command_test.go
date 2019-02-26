package v1alpha1

import (
	"fmt"
	"testing"
)

func TestCstorVolumeCommand(t *testing.T) {
	tests := map[string]struct {
		action            RunCommandAction
		isSupportedAction bool
	}{
		"test 101": {DeleteCommandAction, false},
		"test 102": {CreateCommandAction, false},
		"test 103": {ListCommandAction, false},
		"test 104": {GetCommandAction, false},
		"test 105": {PatchCommandAction, false},
		"test 106": {UpdateCommandAction, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := WithAction(Command(), mock.action)
			c := &cstorVolumeCommand{cmd}
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

func TestvalidateOptions(t *testing.T) {
	tests := map[string]struct {
		cstorVolCmd    *cstorVolumeCommand
		expectedOutput error
	}{
		"Empty volume name": {
			cstorVolCmd: &cstorVolumeCommand{
				RunCommand: &RunCommand{
					Data: RunCommandDataMap{"ip": RunCommandData("127.0.1"), "volname": RunCommandData(""), "capacity": RunCommandData("10G")},
				},
			},
			expectedOutput: fmt.Errorf("missing volume name"),
		},
		"Empty IP": {
			cstorVolCmd: &cstorVolumeCommand{
				RunCommand: &RunCommand{
					Data: RunCommandDataMap{"ip": RunCommandData(""), "volname": RunCommandData("vol1"), "capacity": RunCommandData("20G")},
				},
			},
			expectedOutput: fmt.Errorf("missing ip address"),
		},
		"Empty Capacity": {
			cstorVolCmd: &cstorVolumeCommand{
				RunCommand: &RunCommand{
					Data: RunCommandDataMap{"ip": RunCommandData("127.0.1"), "volname": RunCommandData("vol1"), "capacity": RunCommandData("")},
				},
			},
			expectedOutput: fmt.Errorf("missing volume capacity"),
		},
		"Populate all the values": {
			cstorVolCmd: &cstorVolumeCommand{
				RunCommand: &RunCommand{
					Data: RunCommandDataMap{"ip": RunCommandData("0.0.0.0"), "volname": RunCommandData("vol1"), "capacity": RunCommandData("5Zi")},
				},
			},
			expectedOutput: nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := test.cstorVolCmd.validateOptions()
			if err != nil && test.expectedOutput != nil {
				t.Errorf("Expected output was: %v \nbut got: %v", test.expectedOutput, err)
			}
		})
	}
}
