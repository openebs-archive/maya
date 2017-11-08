package iscsi

import (
	"fmt"
	"os/exec"
)

//IscsiLogin is to login to block devices in Storage area network
func IscsiLogin(target string) {
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
}
