package cstor

import (
	"github.com/spf13/cobra"
	volume "github.com/openebs/maya/cmd/mayactl/app/command/cstor/volume"
)

func NewVolumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume",
		Short: "Prints version and other details relevant to maya",
		Long:  `Prints version and other details relevant to maya`,
	}
	cmd.AddCommand(
		volume.NewListCmd(),
		volume.NewCreateCmd(),
		volume.NewDestroyCmd(),
		volume.NewGetCmd(),
		volume.NewSetCmd(),
	)
	return cmd
}
