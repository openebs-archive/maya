package cstor

import (
	"github.com/spf13/cobra"
	pool "github.com/openebs/maya/cmd/mayactl/app/command/cstor/pool"
)

func NewPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool",
		Short: "Prints version and other details relevant to maya",
		Long:  `Prints version and other details relevant to maya`,
	}
	cmd.AddCommand(
		pool.NewListCmd(),
		pool.NewCreateCmd(),
		pool.NewDestroyCmd(),
		pool.NewGetCmd(),
		pool.NewSetCmd(),
		pool.NewStatsCmd(),
		pool.NewIOStatsCmd(),
	)
	return cmd
}
