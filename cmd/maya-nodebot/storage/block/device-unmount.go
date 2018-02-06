package block

import (
	"os/exec"
)

//UnMount is to unmount the specified disk - /mnt/<disk>
func UnMount(mnpt string, flag bool) error {
	if flag==true {
		mnpt = "/host" + mnpt
	}
	res, err := exec.Command("umount", mnpt).Output()
	if err != nil {
		return err
	}
	if len(res) == 0 {
		return nil
	}
	return err
}
