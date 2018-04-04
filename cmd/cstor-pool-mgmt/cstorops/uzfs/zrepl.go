package uzfs

import (
	"os/exec"
	"time"

	"github.com/golang/glog"
)

// CheckForZrepl is blocking call for checking status of zrepl in cstor-pool container.
func CheckForZrepl() {
	for {
		statuscmd := exec.Command("zpool", "status")
		_, err := statuscmd.CombinedOutput()
		if err != nil {
			time.Sleep(3 * time.Second)
			glog.Infof("Waiting for zrepl...")
			continue
		}
		break
	}
}
