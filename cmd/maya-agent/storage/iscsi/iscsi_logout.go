package iscsi

import (
	"fmt"
	"os/exec"
)

//IscsiLogout logs out of logged in block devices
func IscsiLogout(target string) {
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
}
