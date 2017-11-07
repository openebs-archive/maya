package command

import (
	"github.com/spf13/cobra"
)

// NewCmdBlockDevice and its nested children are created
func NewCmdBlockDevice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device",
		Short: "Operations on block devices",
		Long: `The block devices on the local machine/minion can be 
		operated using maya-agent`,
	}
	//New sub command to list block device is added
	cmd.AddCommand(
		NewSubCmdListBlockDevice(),
		NewSubCmdFormat(),
		NewSubCmdMount(),
		NewSubCmdUnMount(),

	//	will be defined later
	//	NewSubCmdCreatePartiton(),
	//	NewSubCmdDeletePartiton(),
	//	NewSubCmdShowPartiton(),
	)

	return cmd
}
