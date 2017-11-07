package command

import (
	"github.com/openebs/maya/cmd/maya-agent/storage/block"
	"github.com/openebs/maya/cmd/maya-agent/types/v1"
	"github.com/spf13/cobra"
)

// NewSubCmdListBlockDevice is to list block device is created
func NewSubCmdListBlockDevice() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "list",
		Short: "List block devices",
		Long: `the set of block devices on the local machine/minion
		can be listed`,
		Run: func(cmd *cobra.Command, args []string) {
			//resJsonDecoded is the decoded value of block disk
			var resJsonDecoded v1.BlockDeviceInfo
			err := block.ListBlockExec(&resJsonDecoded)
			if err != nil {
				panic(err)
			}
			//to print after formatting to end user
			block.FormatOutputForUser(&resJsonDecoded)

		},
	}

	return getCmd
}
