package command

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

// TestNewCStorPoolMgmt is to test cstor-pool-mgmt command.
func TestNewCStorPoolMgmt(t *testing.T) {
	cases := []struct {
		use string
	}{
		{"start"},
	}
	cmd, err := NewCStorPoolMgmt()
	if err != nil {
		t.Errorf("Unable to Instantiate cstor-pool-mgmt")
	}
	cmds := cmd.Commands()
	if len(cmds) != len(cases) {
		t.Errorf("ExpectedCommands: %d ActualCommands: '%d'", len(cases), len(cmds))
	}
	for i, c := range cases {
		if c.use != cmds[i].Use {
			t.Errorf("ExpectedCommand: '%s' ActualCommand: '%s'", c.use, cmds[i].Use)
		}
	}
}

// TestRun is to test running cstor-pool-mgmt without sub-commands.
func TestRun(t *testing.T) {
	var cmd *cobra.Command
	err := Run(cmd)
	if err != nil {
		t.Errorf("Expected: '%s' Actual: '%s'", "nil", err)
	}
}

// TestNewCmdOptions is to test type of CLI command.
func TestNewCmdOptions(t *testing.T) {
	var expectedCmd *cobra.Command
	gotCmd := NewCmdOptions()
	if reflect.TypeOf(gotCmd) != reflect.TypeOf(expectedCmd) {
		t.Errorf("Expected: '%s' Actual: '%v'", reflect.TypeOf(gotCmd), reflect.TypeOf(expectedCmd))
	}
}
