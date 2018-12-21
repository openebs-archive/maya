package command

import (
	"strconv"
	"testing"

	"github.com/spf13/cobra"
)

// TestNewCStorVolumeMgmt is to test cstor-volume-mgmt command.
func TestNewCStorVolumeMgmt(t *testing.T) {
	cases := []struct {
		use string
	}{
		{"start"},
	}
	cmd, err := NewCStorVolumeMgmt()
	if err != nil {
		t.Errorf("Unable to Instantiate cstor-volume-mgmt")
	}
	cmds := cmd.Commands()
	if len(cmds) != len(cases) {
		t.Errorf("ExpectedCommands: %d ActualCommands: '%d'", len(cases), len(cmds))
	}
	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if c.use != cmds[i].Use {
				t.Errorf("ExpectedCommand: '%s' ActualCommand: '%s'", c.use, cmds[i].Use)
			}
		})
	}
}

// TestRun is to test running cstor-volume-mgmt without sub-commands.
func TestRun(t *testing.T) {
	var cmd *cobra.Command
	err := Run(cmd)
	if err != nil {
		t.Errorf("Expected: '%s' Actual: '%s'", "nil", err)
	}
}
