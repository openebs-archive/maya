package block

import (
	"fmt"
	"os/exec"
)

//Format is to format the disk with specified type
func Format(disk, ftype string) (string, error) {
	diskDev := "/dev/" + disk
	fmt.Println("diskDev:", diskDev)
	fmt.Println("ftype:", ftype)
	res, err := exec.Command("mkfs", "-F", "-t", ftype, diskDev).Output()
	if err != nil {
		return "", err
	}
	return string(res), nil

}
