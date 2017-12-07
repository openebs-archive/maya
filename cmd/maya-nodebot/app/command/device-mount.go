package command

import (
	"errors"
	"fmt"

	"github.com/openebs/maya/cmd/maya-nodebot/storage/block"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

type CmdMountOptions struct {
	disk string
}

//NewSubCmdMount mounts the specified disk
func NewSubCmdMount() *cobra.Command {
	options := CmdMountOptions{}
	//var disk string
	getCmd := &cobra.Command{
		Use:   "mount",
		Short: "mount disk",
		Long:  `the block devices on the storage area network can be mount to /mnt/<disk>`,
		Run: func(cmd *cobra.Command, args []string) {

			util.CheckErr(options.Validate(), util.Fatal)
			mountpoint, err := block.Mount(options.disk)
			if err != nil {
				fmt.Println("Mounting failure for", options.disk)
				util.CheckErr(err, util.Fatal)
			}
			fmt.Println("Mounting successful at : ", mountpoint)
		},
	}
	getCmd.Flags().StringVar(&options.disk, "disk", "",
		"disk name")
	return getCmd
}

func (c *CmdMountOptions) Validate() error {
	if c.disk == "" {
		return errors.New("--disk is missing. Please specify a disk")
	}
	return nil
}
