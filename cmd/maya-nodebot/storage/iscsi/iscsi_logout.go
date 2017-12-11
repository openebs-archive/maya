package iscsi

import (
	"os/exec"
)

//IscsiLogout logs out of logged in block devices
func IscsiLogout(target string) (string, error) {
	var res []byte
	var err error
	if target == "all" {
		res, err = exec.Command("iscsiadm", "-m", "node", "-u").Output()
	} else {
		res, err = exec.Command("iscsiadm", "-m", "node", "-p", target, "-u").Output()
	}
	if err != nil {
		return "", err
	}
	return string(res), nil
	//fmt.Println("Device(s) :", string(res))
}
