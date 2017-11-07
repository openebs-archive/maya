package command

import (
	"github.com/openebs/maya/cmd/maya-agent/storage/block"
	"github.com/spf13/cobra"
)

//NewSubCmdUnMount unmounts specified mounted disk
func NewSubCmdUnMount() *cobra.Command {
	var disk string
	getCmd := &cobra.Command{
		Use:   "unmount",
		Short: "unmount mounted disk",
		Long:  `specified mounted disk on the storage area network is unmounted`,
		Run: func(cmd *cobra.Command, args []string) {

			block.UnMount(disk)

		},
	}
	getCmd.Flags().StringVar(&disk, "disk", "sdc",
		"disk name")
	return getCmd
}
