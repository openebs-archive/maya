package block

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/openebs/maya/cmd/maya-agent/types/v1"
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
		NewSubCmdIscsiDiscover(),
		NewSubCmdIscsiLogin(),
		NewSubCmdIscsiLogout(),
		NewSubCmdFormatAndMount(),
		NewSubCmdUnMount(),
	//	NewSubCmdCreatePartiton(), //will be defined later
	//	NewSubCmdDeletePartiton(),
	//	NewSubCmdShowPartiton(),
	)

	return cmd
}

// NewSubCmdListBlockDevice is to list block device is created
func NewSubCmdListBlockDevice() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "list",
		Short: "List block devices",
		Long: `the set of block devices on the local machine/minion
		can be listed`,
		Run: func(cmd *cobra.Command, args []string) {
			//resJsonDecoded is the decoded value of block disk
			var resJsonDecoded v1.BlockDeviceInfo
			err := ListBlockExec(&resJsonDecoded)
			if err != nil {
				panic(err)
			}
			//to print after formatting to end user
			FormatOutputForUser(&resJsonDecoded)

		},
	}

	return getCmd
}

//ListBlockExec is for running os cmds for block disk and json parsing
func ListBlockExec(resJsonDecoded *v1.BlockDeviceInfo) error {
	//list block devices in json format
	ListBlockCommand := v1.OsCommand{"lsblk", "-J"}
	res, err := exec.Command(ListBlockCommand.Command, ListBlockCommand.Flag).Output()
	if err != nil {
		panic(err)
	}

	//decode the json output
	err = json.Unmarshal(res, &resJsonDecoded)
	if err != nil {
		return err
	}
	return nil
}

//FormatOutputForUser is to print disk details to end user only with necessary fields
func FormatOutputForUser(resJsonDecoded *v1.BlockDeviceInfo) {
	fmt.Printf("%v  %9v  %4v  %4v\n", "Name", "Size", "Type", "Mountpoint")
	for _, v := range resJsonDecoded.Blockdevices {
		if v.Type == "disk" && (v.Mountpoint == "" || v.Mountpoint == "/" ||
			strings.HasPrefix(v.Mountpoint, "/mnt/")) {
			if v.Mountpoint == "" {
				v.Mountpoint = "null"
			}
			//Print parent details
			fmt.Printf("%v  %9v  %5v  %5v\n", v.Name, v.Size, v.Type, v.Mountpoint)
			if v.Children != nil {
				for _, u := range v.Children {
					if u.Type == "part" {
						if u.Mountpoint == "" {
							u.Mountpoint = "null"
						}
						//Print children details
						fmt.Printf("|_%v  %6v  %5v  %5v\n", u.Name, u.Size, u.Type, u.Mountpoint)

					}
				}
			}
		}
	}
}
