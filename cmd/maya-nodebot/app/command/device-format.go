package command

import (
	"errors"
	"fmt"

	"github.com/openebs/maya/cmd/maya-nodebot/storage/block"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

type CmdFormatOptions struct {
	disk  string
	ftype string
}

//NewSubCmdFormatAndMount formats the specified disk
func NewSubCmdFormat() *cobra.Command {
	options := CmdFormatOptions{}

	getCmd := &cobra.Command{
		Use:   "format",
		Short: "format disk",
		Long:  `the block devices on the storage area network can be formatted`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(), util.Fatal)
			res, err := block.Format(options.disk, options.ftype)
			if err != nil {
				fmt.Println("Could not format", options.disk, "with", options.ftype)
				util.CheckErr(err, util.Fatal)
			}
			fmt.Println(res)
		},
	}

	getCmd.Flags().StringVar(&options.disk, "disk", "",
		"disk name")
	getCmd.Flags().StringVar(&options.ftype, "type", "ext4",
		"formatting type")
	return getCmd
}

func (c *CmdFormatOptions) Validate() error {
	if c.disk == "" {
		return errors.New("--disk is missing. Please specify a disk")
	}
	return nil
}
