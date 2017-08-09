package command

import (
	"fmt"
	"os"
	"strings"
)

type SnapshotDeleteCommand struct {
	Meta
	Name  string
	Sname string
	//client *ControllerClient
}

func (s *SnapshotDeleteCommand) Help() string {
	helpText := `
Usage: maya vsm-snapshot delete -name <vsm-name> 
         
  Command to delete the snapshot.
`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (s *SnapshotDeleteCommand) Synopsis() string {
	return "delete the snapshots"
}

func (s *SnapshotDeleteCommand) Run(args []string) int {
	flags := s.Meta.FlagSet("vsm-snapshot", FlagSetClient)
	flags.Usage = func() { s.Ui.Output(s.Help()) }

	flags.StringVar(&s.Name, "volname", "", "")
	flags.StringVar(&s.Sname, "snapname", "", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	if err := s.DeleteSnapshot(s.Name, s.Sname); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to delete %s: %v\n", s.Sname, err)
		return 1
	}

	return 0
}

func (s *SnapshotDeleteCommand) DeleteSnapshot(volume string, snapshot string) error {
	var err error
	//var path string
	annotations, err := GetVolAnnotations(volume)
	if err != nil || annotations == nil {

		return err
	}
	controller, err := NewControllerClient(annotations.ControllerIP + ":9501")

	if err != nil {
		return err
	}

	replicas, err := controller.ListReplicas(controller.address)
	if err != nil {
		return err
	}

	for _, r := range replicas {
		if ok, err := s.isRebuilding(&r); err != nil {
			return err
		} else if ok {
			return fmt.Errorf("Can not remove a snapshot because %s is rebuilding", r.Address)
		}
	}

	for _, replica := range replicas {
		if err = s.markSnapshotAsRemoved(&replica, snapshot); err != nil {
			return err
		}
	}

	return nil
}

func (s *SnapshotDeleteCommand) isRebuilding(replicaInController *Replica) (bool, error) {
	repClient, err := NewReplicaClient(replicaInController.Address)
	if err != nil {
		return false, err
	}

	replica, err := repClient.GetReplica()
	if err != nil {
		return false, err
	}

	return replica.Rebuilding, nil
}

func (s *SnapshotDeleteCommand) markSnapshotAsRemoved(replicaInController *Replica, snapshot string) error {
	if replicaInController.Mode != "RW" {
		return fmt.Errorf("Can only mark snapshot as removed from replica in mode RW, got %s", replicaInController.Mode)
	}

	repClient, err := NewReplicaClient(replicaInController.Address)
	if err != nil {
		return err
	}

	if err := repClient.MarkDiskAsRemoved(snapshot); err != nil {
		return err
	}

	return nil
}

func (c *ReplicaClient) MarkDiskAsRemoved(disk string) error {

	_, err := c.GetReplica()
	if err != nil {
		return err
	}
	//url := "/replicas/1?action=markdiskasremoved"
	url := "/replicas/1?action=removedisk"

	return c.post(url, &MarkDiskAsRemovedInput{
		Name: disk,
	}, nil)
}
