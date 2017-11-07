package block

import (
	"fmt"
	"os/exec"
)

//Mount is to mount the specified disk to /mnt/<disk>
func Mount(disk string) {
	mountpoint := "/mnt/" + disk
	diskDev := "/dev/" + disk
	//p flag is to return no error if the directory is available already
	res, err := exec.Command("mkdir", "-p", mountpoint).Output()
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
}
