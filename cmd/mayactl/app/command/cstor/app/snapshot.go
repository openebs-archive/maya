package cstor

import (
	"github.com/spf13/cobra"
	snapshot "github.com/openebs/maya/cmd/mayactl/app/command/cstor/snapshot"
)

func NewSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Prints version and other details relevant to maya",
		Long:  `Prints version and other details relevant to maya`,
	}
	cmd.AddCommand(
		snapshot.NewListCmd(),
		snapshot.NewCreateCmd(),
		snapshot.NewDestroyCmd(),
		snapshot.NewSendCmd(),
		snapshot.NewRecvCmd(),
		snapshot.NewRollbackCmd(),
	)
	return cmd
}
