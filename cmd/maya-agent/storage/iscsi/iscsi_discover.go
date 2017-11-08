package iscsi

import (
	"fmt"
	"os/exec"
)

//IscsiDiscover is to discover block devices in Storage area network
func IscsiDiscover(target string) {
	res, err := exec.Command("iscsiadm", "-m", "discovery", "-t", "sendtargets", "-p", target).Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("Device :", string(res))
}
