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

  This command has subcommands related to Volume.

  Volume operations.

  Create a Volume:

      $ maya volume create -volname <vol> -size <size>

  list a Volume:

      $ maya volume list
  
  Delete a Volume:

     $ maya volume delete -volname <vol>

  Stats of Volume:

     $ maya volume stats <volname>
`
	return strings.TrimSpace(helpText)
}
func (c *VolumeCommand) Synopsis() string {
	return "Creates a OpenEBS Volume"
}
