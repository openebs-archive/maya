// CheckForIscsi is blocking call for checking status of istgt in cstor-istgt container.
package util

import (
	"time"

	"github.com/golang/glog"
)

const (
	IstgtConfPath        = "/usr/local/etc/istgt/istgt.conf"
	IstgtStatusCmd       = "STATUS"
	IstgtRefreshCmd      = "REFRESH"
	IstgtReplicaCmd      = "REPLICA"
	IstgtExecuteQuietCmd = "-q"
	ReplicaStatus        = "Replica status"
	WaitTimeForIscsi     = 3 * time.Second
)

func CheckForIscsi(UnixSockVar UnixSock) {
	for {
		_, err := UnixSockVar.SendCommand(IstgtStatusCmd)
		if err != nil {
			time.Sleep(WaitTimeForIscsi)
			glog.Warningf("Waiting for istgt... err : %v", err)
			continue
		}
		break
	}
}
