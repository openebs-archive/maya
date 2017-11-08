package block

import (
	"fmt"
	"os/exec"
)

//UnMount is to unmount the specified disk - /mnt/<disk>
func UnMount(disk string) {
	disk = "/mnt/" + disk
	res, err := exec.Command("umount", disk).Output()
	if err != nil {
		panic(err)
	}
	if len(res) == 0 {
		fmt.Println("Successfully unmounted : ", disk)
	}
}
