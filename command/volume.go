package command

import (
	"strings"

	"github.com/mitchellh/cli"
)

// VolumeCommand is a Command implementation that just shows help for
// the subcommands nested below it.
type VolumeCommand struct {
}

func (c *VolumeCommand) Run(args []string) int {
	return cli.RunResultHelp
}

func (c *VolumeCommand) Help() string {
	helpText := `
Usage: maya volume <subcommand> [options] [args]

This command provides operations related to a Volume.

    Create a Volume:
    $ maya volume create -volname <vol> -size <size>

    List Volumes:
    $ maya volume list
  
    Delete a Volume:
    $ maya volume delete -volname <vol>

    Statistics of a Volume:
    $ maya volume stats <vol>

`
	return strings.TrimSpace(helpText)
}
func (c *VolumeCommand) Synopsis() string {
	return "Provides operations related to a Volume"
}
