package command

import (
	"errors"
	"fmt"

	"github.com/openebs/maya/cmd/maya-nodebot/storage/block"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

type CmdUnmountOptions struct {
	disk string
}

//NewSubCmdUnMount unmounts specified mounted disk
func NewSubCmdUnMount() *cobra.Command {
	options := CmdUnmountOptions{}
	//var disk string
	getCmd := &cobra.Command{
		Use:   "unmount",
		Short: "unmount mounted disk",
		Long:  `specified mounted disk on the storage area network is unmounted`,
		Run: func(cmd *cobra.Command, args []string) {

			util.CheckErr(options.Validate(), util.Fatal)
			err := block.UnMount(options.disk)
			if err != nil {
				fmt.Println("Unmounting failure for", options.disk)
				util.CheckErr(err, util.Fatal)
			}
			fmt.Println("Unmounting successful for : ", options.disk)

		},
	}
	getCmd.Flags().StringVar(&options.disk, "disk", "",
		"disk name")
	return getCmd
}

func (c *CmdUnmountOptions) Validate() error {
	if c.disk == "" {
		return errors.New("--disk is missing. Please specify a disk")
	}
	return nil
}
