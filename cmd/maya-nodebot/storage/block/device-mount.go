package block

import (
	"os/exec"
)

//Mount is to mount the specified disk to /mnt/<disk>
func Mount(disk, mountpoint string, flag bool) (string, error) {
	if flag==true {
		disk = "/host" + disk
		mountpoint = "/host" + mountpoint
	}
	//p flag is to return no error if the directory is available already
	res, err := exec.Command("mkdir", "-p", mountpoint).Output()
	if err != nil {
		return "", err
	}

	res, err = exec.Command("mount", disk, mountpoint).Output()
	if err != nil {
		return "", err
	}
	if len(res) == 0 {
		return mountpoint, nil
	} else {
		return "", err
	}
}
