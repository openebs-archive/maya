package command

import (
	"errors"
	"fmt"

	"github.com/openebs/maya/cmd/maya-nodebot/storage/iscsi"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

type CmdIscsiLoginOptions struct {
	target string
}

//NewSubCmdIscsiLogin logs in to particular portal or all discovered portals
func NewSubCmdIscsiLogin() *cobra.Command {
	options := CmdIscsiLoginOptions{}
	//var target string
	getCmd := &cobra.Command{
		Use:   "login",
		Short: "iscsi login to block devices",
		Long:  `Single and multiple login to set of block devices on the storage area network `,
		Run: func(cmd *cobra.Command, args []string) {

			util.CheckErr(options.Validate(), util.Fatal)
			res, err := iscsi.IscsiLogin(options.target)
			if err != nil {
				fmt.Println("Iscsi login failure for portal", options.target)
				util.CheckErr(err, util.Fatal)
			}
			fmt.Println(res)

		},
	}

	getCmd.Flags().StringVar(&options.target, "portal", "",
		"target portal-ip to iscsi login, 'all' to login all discovery")
	return getCmd
}

func (c *CmdIscsiLoginOptions) Validate() error {
	if c.target == "" {
		return errors.New("--portal is missing. Please specify a portal-ip")
	}
	return nil
}
