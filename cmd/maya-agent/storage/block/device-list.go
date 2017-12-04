package block

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/openebs/maya/cmd/maya-agent/types/v1"
)

//ListBlockExec is for running os cmds for block disk and json parsing
func ListBlockExec(resJsonDecoded *v1.BlockDeviceInfo) error {
	//list block devices in json format
	ListBlockCommand := v1.OsCommand{"lsblk", "-J"}
	res, err := exec.Command(ListBlockCommand.Command, ListBlockCommand.Flag).Output()
	if err != nil {
		panic(err)
	}

	//decode the json output
	return json.Unmarshal(res, &resJsonDecoded)
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
