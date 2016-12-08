package command

import (
	"os/exec"
	"testing"

	"github.com/mitchellh/cli"
)

func TestInstallMayaCommand_Implements(t *testing.T) {
	var _ cli.Command = &InstallMayaCommand{}
}

// We are using the OS `ls` command when required. Since it
// is generally available in all *nix environments. This is
// more or less sufficient to unit test.
// NOTE - Our target is to provide a good unit test coverage.
func TestInstallMayaCommand_Run(t *testing.T) {

	ui := new(cli.MockUi)

	cmd := &InstallMayaCommand{
		M:   Meta{Ui: ui},
		Cmd: exec.Command("ls", []string{"/home"}...),
	}

	if code := cmd.Run([]string{""}); code != 1 {
		t.Fatalf("expected exit 1, got: %d", code)
	}

}

func TestInstallMayaCommand_Negative(t *testing.T) {
	ui := new(cli.MockUi)

	cmd := &InstallMayaCommand{
		M:   Meta{Ui: ui},
		Cmd: exec.Command("ls", []string{".", "some", "bad", "args"}...),
	}

	// Fails on misuse
	if code := cmd.Run([]string{""}); code != 1 {
		t.Fatalf("expected exit code 1, got: %d", code)
	}

}
