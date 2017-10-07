package command

import (
	"os/exec"
	"testing"

	"github.com/mitchellh/cli"
)

func TestVsmUpdateCommand_Implements(t *testing.T) {
	var _ cli.Command = &VsmUpdateCommand{}
}

// We are using the OS `ls` command when required. Since it
// is generally available in all *nix environments. This is
// more or less sufficient to unit test.
// NOTE - Our target is to provide a good unit test coverage.
func TestVsmUpdateCommand_Run(t *testing.T) {

	ui := new(cli.MockUi)

	cmd := &VsmUpdateCommand{
		M:   Meta{Ui: ui},
		Cmd: exec.Command("ls", []string{"/home"}...),
	}

	if code := cmd.Run([]string{""}); code != 0 {
		t.Fatalf("expected exit 0, got: %d", code)
	}

}

func TestVsmUpdateCommand_Negative(t *testing.T) {
	ui := new(cli.MockUi)

	cmd := &VsmUpdateCommand{
		M:   Meta{Ui: ui},
		Cmd: exec.Command("ls", []string{".", "some", "bad", "args"}...),
	}

	// Fails on misuse
	if code := cmd.Run([]string{""}); code != 2 {
		t.Fatalf("expected exit code 2, got: %d", code)
	}

	// Execute internal command with no args
	cmd.Cmd = nil
	if code := cmd.Run([]string{""}); code != 1 {
		t.Fatalf("expected exit code 1, got: %d", code)
	}

	// Execute internal command with bad arguments
	cmd.Cmd = nil
	if code := cmd.Run([]string{"some", "bad", "args"}); code != 1 {
		t.Fatalf("expected exit code 1, got: %d", code)
	}

}
