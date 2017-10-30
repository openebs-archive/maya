package jiva

import (
	"github.com/openebs/maya/command"
)

// SnapshotRevert will be responsible for reverting to a
// particular snapshot. If there is more then one snapshot has been
// created for a volume, then user can revert to any specific created
// snaphot for that particular volume.
func SnapshotRevert(snapshot string, controllerIP string) error {

	controller, err := command.NewControllerClient(controllerIP + ":9501")

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
