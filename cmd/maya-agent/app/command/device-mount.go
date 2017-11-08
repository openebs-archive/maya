package command

import (
	"github.com/openebs/maya/cmd/maya-agent/storage/block"
	"github.com/spf13/cobra"
)

//NewSubCmdMount mounts the specified disk
func NewSubCmdMount() *cobra.Command {
	var disk string
	getCmd := &cobra.Command{
		Use:   "mount",
		Short: "mount disk",
		Long:  `the block devices on the storage area network can be mount to /mnt/<disk>`,
		Run: func(cmd *cobra.Command, args []string) {

			block.Mount(disk)

		},
	}
	getCmd.Flags().StringVar(&disk, "disk", "sdc",
		"disk name")
	return getCmd
}
