package command

import (
	"github.com/openebs/maya/cmd/maya-agent/storage/block"
	"github.com/spf13/cobra"
)

//NewSubCmdFormatAndMount formats the specified disk
func NewSubCmdFormat() *cobra.Command {
	var disk string
	var ftype string
	getCmd := &cobra.Command{
		Use:   "format",
		Short: "format disk",
		Long:  `the block devices on the storage area network can be formatted`,
		Run: func(cmd *cobra.Command, args []string) {

			block.Format(disk, ftype)

		},
	}
	getCmd.Flags().StringVar(&disk, "disk", "sdc",
		"disk name")
	getCmd.Flags().StringVar(&ftype, "type", "ext4",
		"formatting type")
	return getCmd
}
