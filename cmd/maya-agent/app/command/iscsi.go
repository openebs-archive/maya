package command

import (
	"github.com/spf13/cobra"
)

// NewCmdIscsi creates NewCmdBlockDevice and its nested children
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
