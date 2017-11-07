package command

import (
	"github.com/openebs/maya/cmd/maya-agent/storage/iscsi"
	"github.com/spf13/cobra"
)

//NewSubCmdIscsiLogin logs in to particular portal or all discovered portals
func NewSubCmdIscsiLogin() *cobra.Command {
	var target string
	getCmd := &cobra.Command{
		Use:   "login",
		Short: "iscsi login to block devices",
		Long:  `Single and multiple login to set of block devices on the storage area network `,
		Run: func(cmd *cobra.Command, args []string) {

			iscsi.IscsiLogin(target)

		},
	}

	getCmd.Flags().StringVar(&target, "portal", "10.107.180.120",
		"target portal-ip to iscsi login, 'all' to login all")
	return getCmd
}
