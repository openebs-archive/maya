package cstor
import (
        "github.com/spf13/cobra"
)
func NewCstorCmd() *cobra.Command {
        cmd := &cobra.Command{
                Use:   "cstor",
                Short: "Prints version and other details relevant to maya",
                Long:  `Prints version and other details relevant to maya`,
        }
        cmd.AddCommand(
                NewPoolCmd(),
                NewVolumeCmd(),
                NewSnapshotCmd(),
        )
        return cmd
}
