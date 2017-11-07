package command

import (
	"github.com/spf13/cobra"
)

// NewCmdBlockDevice and its nested children are created
func NewCmdIscsi() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iscsi",
		Short: "Operations using iscsi",
		Long: `The block devices on the local machine/minion can be 
		operated using maya-agent`,
	}
	//New sub command to list block device is added
	cmd.AddCommand(
		NewSubCmdIscsiDiscover(),
		NewSubCmdIscsiLogin(),
		NewSubCmdIscsiLogout(),
	)

	return cmd
}
