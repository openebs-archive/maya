package command

import (
	"github.com/openebs/maya/cmd/maya-agent/storage/iscsi"
	"github.com/spf13/cobra"
)

//NewSubCmdIscsiDiscover discovers iscsi block devices
func NewSubCmdIscsiDiscover() *cobra.Command {
	var target string
	getCmd := &cobra.Command{
		Use:   "discover",
		Short: "discover block device with iscsi",
		Long:  `the  block device is discovered on the storage area network with specified target`,
		Run: func(cmd *cobra.Command, args []string) {

			iscsi.IscsiDiscover(target)

		},
	}
	getCmd.Flags().StringVar(&target, "portal", "127.0.0.1", "target portal-ip to iscsi discover")
	return getCmd
}
