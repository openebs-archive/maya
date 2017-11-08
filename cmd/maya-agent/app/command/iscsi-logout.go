package command

import (
	"github.com/openebs/maya/cmd/maya-agent/storage/iscsi"
	"github.com/spf13/cobra"
)

//NewSubCmdIscsiLogout logs out of particular portal or all discovered portals
func NewSubCmdIscsiLogout() *cobra.Command {
	var target string
	getCmd := &cobra.Command{
		Use:   "logout",
		Short: "logout of block devices with iscsi",
		Long:  `Single and multiple logout to set of block devices on the storage area network `,
		Run: func(cmd *cobra.Command, args []string) {

			iscsi.IscsiLogout(target)

		},
	}
	getCmd.Flags().StringVar(&target, "portal", "127.0.0.1",
		"target portal-ip to iscsi logout, 'all' to logout all")

	return getCmd
}
