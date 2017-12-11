package command

import (
	"errors"
	"fmt"

	"github.com/openebs/maya/cmd/maya-nodebot/storage/iscsi"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

type CmdIscsiDiscoverOptions struct {
	target string
}

//NewSubCmdIscsiDiscover discovers iscsi block devices
func NewSubCmdIscsiDiscover() *cobra.Command {
	options := CmdIscsiDiscoverOptions{}

	getCmd := &cobra.Command{
		Use:   "discover",
		Short: "discover block device with iscsi",
		Long:  `the  block device is discovered on the storage area network with specified target`,
		Run: func(cmd *cobra.Command, args []string) {

			util.CheckErr(options.Validate(), util.Fatal)
			res, err := iscsi.IscsiDiscover(options.target)
			if err != nil {
				fmt.Println("Iscsi discovery failure for portal", options.target)
				util.CheckErr(err, util.Fatal)
			}
			fmt.Println(res)

		},
	}
	getCmd.Flags().StringVar(&options.target, "portal", "", "target portal-ip to iscsi discover")
	return getCmd
}

func (c *CmdIscsiDiscoverOptions) Validate() error {
	if c.target == "" {
		return errors.New("--portal is missing. Please specify a portal-ip")
	}
	return nil
}
