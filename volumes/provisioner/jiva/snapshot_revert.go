package jiva

import (
	"fmt"

	"github.com/openebs/maya/command"
)

// SnapshotRevert will be responsible for reverting to a
// particular snapshot. If there is more then one snapshot has been
// created for a volume, then user can revert to any specific created
// snaphot for that particular volume.
func SnapshotRevert(volname string, snapshot string) error {

	annotations, err := command.GetVolumeSpec(volname)
	if err != nil || annotations == nil {

		return err
	}

	if annotations.ControllerStatus != "Running" {
		fmt.Println("Volume not reachable")
		return err
	}
	controller, err := command.NewControllerClient(annotations.ControllerIP + ":9501")

	if err != nil {
		return err
	}

	//var c *ControllerClient
	volume, err := command.GetVolume(controller.Address)
	if err != nil {
		return err
	}

	url := controller.Address + "/volumes/" + volume.Id + "?action=revert"
	var c ControllerClient
	return c.post(url, RevertInput{
		Name: snapshot,
	}, nil)
}
