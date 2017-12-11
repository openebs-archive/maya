package block

import (
	"os/exec"
)

//UnMount is to unmount the specified disk - /mnt/<disk>
func UnMount(disk string) error {
	disk = "/mnt/" + disk
	res, err := exec.Command("umount", disk).Output()
	if err != nil {
		return err
	}
	if len(res) == 0 {
		return nil
	}
	return err
}
