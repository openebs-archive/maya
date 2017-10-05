package command

import (
	"fmt"
	"os"
	"strings"
)

type SnapshotRevertCommand struct {
	Meta
	Name  string
	Sname string
	//client *ControllerClient
}

func (s *SnapshotRevertCommand) Help() string {
	helpText := `
Usage: maya snapshot revert -volname <vol> -snapname <snap>
							           
This command will revert to specific snapshot of a Volume.

`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (s *SnapshotRevertCommand) Synopsis() string {
	return "Reverts to specific snapshot of a Volume"
}
func (s *SnapshotRevertCommand) Run(args []string) int {
	flags := s.Meta.FlagSet("vsm-snapshot", FlagSetClient)
	flags.Usage = func() { s.Ui.Output(s.Help()) }

	flags.StringVar(&s.Name, "volname", "", "")
	flags.StringVar(&s.Sname, "snapname", "", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}
	var c *ControllerClient
	if err := c.RevertSnapshot(s.Name, s.Sname); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to revert snapshot %s: %v\n", s.Sname, err)
		return 1
	}
	fmt.Println("Snapshot reverted:", s.Sname)
	return 0
}

func (c *ControllerClient) RevertSnapshot(volname string, snapshot string) error {

	annotations, err := GetVolAnnotations(volname)
	if err != nil || annotations == nil {

		return err
	}

	if annotations.ControllerStatus != "Running" {
		fmt.Println("Volume not reachable")
		return err
	}
	controller, err := NewControllerClient(annotations.ControllerIP + ":9501")

	if err != nil {
		return err
	}

	//var c *ControllerClient
	volume, err := GetVolume(controller.Address)
	if err != nil {
		return err
	}

	url := controller.Address + "/volumes/" + volume.Id + "?action=revert"

	return c.post(url, RevertInput{
		Name: snapshot,
	}, nil)
}
