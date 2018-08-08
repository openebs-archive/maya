package command

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/spf13/cobra"
)

// TestNewCStorVolumeGrpc is to test cstor-volume-grpc command.
func TestNewCStorVolumeGrpc(t *testing.T) {
	cases := []struct {
		use string
	}{
		{"client"},
		{"server"},
	}
	cmd, err := NewCStorVolumeGrpc()
	if err != nil {
		t.Errorf("Unable to Instantiate cstor-volume-grpc")
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

func TestNewCmdServer(t *testing.T) {
	cases := []struct {
		use string
	}{
		{"start"},
	}
	cmd := NewCmdServer()
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
func TestNewCmdClient(t *testing.T) {
	cases := []struct {
		use string
	}{
		{"snapshot"},
	}
	cmd := NewCmdClient()
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
func TestNewCmdSnapshot(t *testing.T) {
	cases := []struct {
		use string
	}{
		{"create"},
		{"destroy"},
	}
	cmd := NewCmdSnapshot()
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

// TestNewCmdOptions is to test type of CLI command.
func TestNewCmdOptions(t *testing.T) {
	var expectedCmd *cobra.Command
	gotCmd := NewCmdOptions()
	if reflect.TypeOf(gotCmd) != reflect.TypeOf(expectedCmd) {
		t.Errorf("Expected: '%s' Actual: '%v'", reflect.TypeOf(gotCmd), reflect.TypeOf(expectedCmd))
	}
}
