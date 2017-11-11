package snapshot

import (
	"errors"
	"fmt"

	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

/*type CmdSnaphotCreateOptions struct {
	volName  string
	snapName string
}*/

// NewCmdSnapshotRevert reverts a snapshot of OpenEBS Volume
func NewCmdSnapshotRevert() *cobra.Command {
	options := CmdSnaphotCreateOptions{}

	cmd := &cobra.Command{
		Use:   "revert",
		Short: "Reverts to specific snapshot of a Volume",
		Long:  "Reverts to specific snapshot of a Volume",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			util.CheckErr(options.RunSnapshotRevert(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "", options.volName,
		"unique volume name.")
	cmd.MarkPersistentFlagRequired("volname")
	cmd.MarkPersistentFlagRequired("snapname")

	cmd.Flags().StringVarP(&options.snapName, "snapname", "", options.snapName,
		"unique snapshot name")

	return cmd
}

// RunSnapshotRevert does tasks related to mayaserver.
func (c *CmdSnaphotCreateOptions) RunSnapshotRevert(cmd *cobra.Command) error {
	fmt.Println("Executing volume snapshot revert ...")

	resp := mapiserver.RevertSnapshot(c.volName, c.snapName)
	if resp != nil {
		return errors.New(fmt.Sprintf("Error: %v", resp))
	}

	fmt.Println("Reverting to snapshot [%s] of volume [%s]", c.snapName, c.volName)
	return nil
}
