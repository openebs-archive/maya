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

This command provides operations related to snapshot of a Volume.

    Create snapshot:
    $ maya snapshot create -volname <vol> -snapname <snap>

    List snapshots:
    $ maya snapshot list -volname <vol>
  
    Revert to snapshot:
    $ maya snapshot revert -volname <vol> -snapname <snap>

`
	return strings.TrimSpace(helpText)
}
func (c *SnapshotCommand) Synopsis() string {
	return "Provides operations related to snapshot of a Volume"
}
