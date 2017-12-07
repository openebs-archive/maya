package command

import (
"errors"
	"github.com/openebs/maya/cmd/maya-nodebot/mcrd"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

type CmdStartOptions struct {
	kubeconfig string
}

//NewSubCmdIscsiLogin logs in to particular portal or all discovered portals
func NewCmdStart() *cobra.Command {
	options := CmdStartOptions{}
	//var target string
	getCmd := &cobra.Command{
		Use:   "start",
		Short: "crd",
		Long:  ` crd `,
		Run: func(cmd *cobra.Command, args []string) {

			util.CheckErr(options.Validate(), util.Fatal)
			mcrd.Checkcrd(options.kubeconfig)

			/*
				res, err := iscsi.IscsiLogin(options.target)
				if err != nil {
					fmt.Println("Iscsi login failure for portal", options.target)
					util.CheckErr(err, util.Fatal)
				}
				fmt.Println(res)
			*/

		},
	}

	getCmd.Flags().StringVar(&options.kubeconfig, "kubeconfig", "$HOME/.kube/config",
		`kubeconfig needs to be specified if out of cluster`)
	return getCmd
}

func (c *CmdStartOptions) Validate() error {
	if c.kubeconfig == "" {
		return errors.New("--kubeconfig is missing. Please specify kubeconfig")
	}
	return nil
}
