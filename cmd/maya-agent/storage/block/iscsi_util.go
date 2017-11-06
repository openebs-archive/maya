package block

import (
	"fmt"
	"os/exec"

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
			res, err := exec.Command("iscsiadm", "-m", "discovery", "-t", "sendtargets", "-p", target).Output()
			if err != nil {
				panic(err)
			}
			fmt.Println("Device :", string(res))

		},
	}
	getCmd.Flags().StringVar(&target, "portal", "127.0.0.1", "target portal-ip to iscsi discover")
	return getCmd
}

//NewSubCmdIscsiLogin logs in to particular portal or all discovered portals
func NewSubCmdIscsiLogin() *cobra.Command {
	var target string
	getCmd := &cobra.Command{
		Use:   "login",
		Short: "iscsi login to block devices",
		Long:  `Single and multiple login to set of block devices on the storage area network `,
		Run: func(cmd *cobra.Command, args []string) {

			var res []byte
			var err error
			if target == "all" {
				res, err = exec.Command("iscsiadm", "-m", "node", "-l").Output()
			} else {
				res, err = exec.Command("iscsiadm", "-m", "node", "-p", target, "-l").Output()
			}
			if err != nil {
				panic(err)
			}
			fmt.Println("Device(s) :", string(res))

		},
	}

	getCmd.Flags().StringVar(&target, "portal", "10.107.180.120",
		"target portal-ip to iscsi login, 'all' to login all")
	return getCmd
}

//NewSubCmdIscsiLogout logs out of particular portal or all discovered portals
func NewSubCmdIscsiLogout() *cobra.Command {
	var target string
	getCmd := &cobra.Command{
		Use:   "logout",
		Short: "logout of block devices with iscsi",
		Long:  `Single and multiple logout to set of block devices on the storage area network `,
		Run: func(cmd *cobra.Command, args []string) {

			var res []byte
			var err error
			if target == "all" {
				res, err = exec.Command("iscsiadm", "-m", "node", "-u").Output()
			} else {
				res, err = exec.Command("iscsiadm", "-m", "node", "-p", target, "-u").Output()
			}
			if err != nil {
				panic(err)
			}
			fmt.Println("Device(s) :", string(res))
		},
	}
	getCmd.Flags().StringVar(&target, "portal", "127.0.0.1",
		"target portal-ip to iscsi logout, 'all' to logout all")

	return getCmd
}

//NewSubCmdFormatAndMount formats and mounts the specified disk
func NewSubCmdFormatAndMount() *cobra.Command {
	var disk string
	getCmd := &cobra.Command{
		Use:   "format-mount",
		Short: "format and mount disk",
		Long:  `the block devices on the storage area network can be formatted and mount`,
		Run: func(cmd *cobra.Command, args []string) {

			diskDev := "/dev/" + disk
			fmt.Println("diskDev:", diskDev)
			res, err := exec.Command("mkfs.ext4", "-F", diskDev).Output()
			if err != nil {
				panic(err)
			}
			fmt.Println("res:", string(res))

			mountpoint := "/mnt/" + disk
			res, err = exec.Command("mkdir", "-p", mountpoint).Output()
			if err != nil {
				panic(err)
			}

			res, err = exec.Command("mount", diskDev, mountpoint).Output()
			if err != nil {
				panic(err)
			}
			if len(res) == 0 {
				fmt.Println("Successfully mounted on: ", mountpoint)
			} else {
				fmt.Println("Mounting failure")
			}

		},
	}
	getCmd.Flags().StringVar(&disk, "disk", "sdc",
		"disk name")
	return getCmd
}

//NewSubCmdUnMount unmounts specified mounted disk
func NewSubCmdUnMount() *cobra.Command {
	var disk string
	getCmd := &cobra.Command{
		Use:   "unmount",
		Short: "unmount mounted disk",
		Long:  `specified mounted disk on the storage area network is unmounted`,
		Run: func(cmd *cobra.Command, args []string) {
			disk = "/mnt/" + disk
			res, err := exec.Command("umount", disk).Output()
			if err != nil {
				panic(err)
			}
			if len(res) == 0 {
				fmt.Println("Successfully unmounted : ", disk)
			}
		},
	}
	getCmd.Flags().StringVar(&disk, "disk", "sdc",
		"disk name")
	return getCmd
}
