package command

import (
	"os/exec"
	"testing"

	"github.com/mitchellh/cli"
)

func TestVsmListCommand_Implements(t *testing.T) {
	var _ cli.Command = &VsmListCommand{}
}

// We are using the OS `ls` command when required. Since it
// is generally available in all *nix environments. This is
// more or less sufficient to unit test `maya vsm-list`.
// NOTE - Our target is to provide a good unit test coverage.
func TestVsmListCommand_Run(t *testing.T) {

	ui := new(cli.MockUi)

	cmd := &VsmListCommand{
		M:   Meta{Ui: ui},
		Cmd: exec.Command("ls", []string{"/home"}...),
	}

	// Should return blank for no vsms
	if code := cmd.Run([]string{""}); code != 0 {
		t.Fatalf("expected exit 0, got: %d", code)
	}

}

func TestVsmListCommand_Negative(t *testing.T) {
	ui := new(cli.MockUi)

	cmd := &VsmListCommand{
		M:   Meta{Ui: ui},
		Cmd: exec.Command("ls", []string{".", "some", "bad", "args"}...),
	}

	// Fails on misuse
	if code := cmd.Run([]string{""}); code != 1 {
		t.Fatalf("expected exit code 1, got: %d", code)
	}

	// Internal Command will be executed
	// The internal command is not expected to be running
	// hence the expected return code is 1
	cmd.Cmd = nil
	if code := cmd.Run([]string{""}); code != 1 {
		t.Fatalf("expected exit code 1, got: %d", code)
	}
}
