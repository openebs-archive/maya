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
Usage: maya snapshot <subcommand> [options] [args]

  This command has subcommands related to snapshot of Volume.

  snapshot operations.

  Create snapshot:

      $ maya snapshot create -volname <vol> -snapname <snapshot>

  list snapshot:

      $ maya snapshot list -volname <volume-name>
  
  Revert to snapshot:

     $ maya snapshot revert -volname <vol> -snapname <snapshot>


`
	return strings.TrimSpace(helpText)
}
func (c *SnapshotCommand) Synopsis() string {
	return "Creates snapshot of a Volume"
}
