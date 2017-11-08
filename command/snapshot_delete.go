package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/openebs/maya/pkg/client/jiva"
	//	"github.com/rancher/go-rancher/client"
)

type SnapshotDeleteCommand struct {
	Meta
	Name  string
	Sname string
	//client *ControllerClient
}

/*func NewTask(controller string) *Task {
	return &Task{
		client: NewControllerClient(controller),
	}
}
*/

func (s *SnapshotDeleteCommand) Help() string {
	helpText := `
	Usage: maya snapshot delete -volname <vol>

	This command will delete all snapshots of a Volume.

	`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (s *SnapshotDeleteCommand) Synopsis() string {
	return "Deletes the snapshots of a Volume"
}
func (s *SnapshotDeleteCommand) Run(args []string) int {
	flags := s.Meta.FlagSet("snapshot", FlagSetClient)
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

	annotations, err := GetVolAnnotations(volume)
	if err != nil || annotations == nil {

		return err
	}
	controller, err := client.NewControllerClient(annotations.ControllerIP + ":9501")

	if err != nil {
		return err
	}

	replicas, err := controller.ListReplicas(controller.Address)
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

func (s *SnapshotDeleteCommand) isRebuilding(replicaInController *client.Replica) (bool, error) {
	repClient, err := client.NewReplicaClient(replicaInController.Address)
	if err != nil {
		return false, err
	}

	replica, err := repClient.GetReplica()
	if err != nil {
		return false, err
	}

	return replica.Rebuilding, nil
}

func (s *SnapshotDeleteCommand) markSnapshotAsRemoved(replicaInController *client.Replica, snapshot string) error {
	if replicaInController.Mode != "RW" {
		return fmt.Errorf("Can only mark snapshot as removed from replica in mode RW, got %s", replicaInController.Mode)
	}

	repClient, err := client.NewReplicaClient(replicaInController.Address)
	if err != nil {
		return err
	}

	if err := repClient.MarkDiskAsRemoved(snapshot); err != nil {
		return err
	}

	return nil
}
