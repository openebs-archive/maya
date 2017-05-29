package command

import (
	"bytes"
	"fmt"

	"github.com/mitchellh/cli"
)

// VersionCommand is a Command implementation prints the version.
type VersionCommand struct {
	Revision          string
	Version           string
	VersionPrerelease string
	Ui                cli.Ui
}

// Help shows helpText for a particular CLI command
func (c *VersionCommand) Help() string {
	return ""
}

// Run holds the flag values for CLI subcommands
func (c *VersionCommand) Run(_ []string) int {
	var versionString bytes.Buffer

	fmt.Fprintf(&versionString, "Maya v%s", c.Version)
	if c.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", c.VersionPrerelease)

		if c.Revision != "" {
			fmt.Fprintf(&versionString, " (%s)", c.Revision)
		}
	}

	c.Ui.Output(versionString.String())
	return 0
}

// Synopsis shows short information related to CLI command
func (c *VersionCommand) Synopsis() string {
	return "Prints the Maya version"
}
