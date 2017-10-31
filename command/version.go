package command

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/openebs/maya/orchprovider"
	"github.com/openebs/maya/pkg/client/mapiserver"
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
	helpText := `
	Usage: maya version

	This command provides versioning and other details relevant to maya.

	`
	return strings.TrimSpace(helpText)
}

// Run holds the flag values for CLI subcommands
func (c *VersionCommand) Run(_ []string) int {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "Maya %s", c.Version)
	if c.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "%s", c.VersionPrerelease)

		if c.Revision != "" {
			fmt.Fprintf(&versionString, " (%s)", c.Revision)
		}
	}

	c.Ui.Output(versionString.String())

	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("OS/Arch:", runtime.GOOS, "/", runtime.GOARCH)

	fmt.Println("m-apiserver url: ", mapiserver.GetURL())
	fmt.Println("m-apiserver status: ", mapiserver.GetConnectionStatus())

	fmt.Println("Provider: ", orchprovider.DetectOrchProviderFromEnv())

	return 0
}

// Synopsis shows short information related to CLI command
func (c *VersionCommand) Synopsis() string {
	return "Prints version and other details relevant to maya"
}
