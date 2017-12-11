package iscsi

import (
	"os/exec"
)

//IscsiDiscover is to discover block devices in Storage area network
func IscsiDiscover(target string) (string, error) {
	res, err := exec.Command("iscsiadm", "-m", "discovery", "-t", "sendtargets", "-p", target).Output()
	if err != nil {
		return "", err
	}
	return string(res), nil
	//fmt.Println("Device :", string(res))
}
