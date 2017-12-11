package command

import (
	"errors"
	"fmt"

	"github.com/openebs/maya/cmd/maya-nodebot/storage/iscsi"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

type CmdIscsiLogoutOptions struct {
	target string
}

//NewSubCmdIscsiLogout logs out of particular portal or all discovered portals
func NewSubCmdIscsiLogout() *cobra.Command {
	options := CmdIscsiLogoutOptions{}
	//var target string
	getCmd := &cobra.Command{
		Use:   "logout",
		Short: "logout of block devices with iscsi",
		Long:  `Single and multiple logout to set of block devices on the storage area network `,
		Run: func(cmd *cobra.Command, args []string) {

			util.CheckErr(options.Validate(), util.Fatal)
			res, err := iscsi.IscsiLogout(options.target)
			if err != nil {
				fmt.Println("Iscsi logout failure for portal", options.target)
				util.CheckErr(err, util.Fatal)
			}
			fmt.Println(res)

		},
	}
	getCmd.Flags().StringVar(&options.target, "portal", "",
		"target portal-ip to iscsi logout, 'all' to logout all")

	return getCmd
}

func (c *CmdIscsiLogoutOptions) Validate() error {
	if c.target == "" {
		return errors.New("--portal is missing. Please specify a portal-ip")
	}
	return nil
}
