package cstor

import (
	"github.com/openebs/maya/cmd/cstor-volume-grpc/app/command"
)

func Snapshot(volName, snapName, targetIP string) (interface{}, error) {
	return command.CreateSnapshot(volName, snapName, targetIP)
}

func SnapshotDelete(volName, snapName, targetIP string) (interface{}, error) {
	return command.DestroySnapshot(volName, snapName, targetIP)
}
