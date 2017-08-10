package command

import (
	"strings"

	"github.com/mitchellh/cli"
)

// SnapshotCommand is a Command implementation that just shows help for
// the subcommands nested below it.
type SnapshotCommand struct {
}

func (c *SnapshotCommand) Run(args []string) int {
	return cli.RunResultHelp
}

func (c *SnapshotCommand) Help() string {
	helpText := `
Usage: maya vsm-snapshot <subcommand> [options] [args]

  This command has subcommands for creating a snapshot of Vsm
  and list them. 

  snapshot operations.

  Create a snapshot:

      $ maya vsm-snapshot create -volname <vol> -snapname <snapshot>

  list a snapshot:

      $ maya vsm-snapshot list -name <vsm-name>
  
  Remove a snapshot:

     $ maya vsm-snapshot rm -volname <vol> -snapname <snapshot>

  Revert a snapshot:

     $ maya vsm-snapshot revert -volname <vol> -snapname <snapshot>


`
	return strings.TrimSpace(helpText)
}
func (c *SnapshotCommand) Synopsis() string {
	return "Create a snapshot of a Volume"
}
