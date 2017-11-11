package snapshot

import (
	"errors"
	"fmt"

	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

// NewCmdSnapshotCreate creates a snapshot of OpenEBS Volume
func NewCmdSnapshotList() *cobra.Command {
	options := CmdSnaphotCreateOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all the snapshots of a Volume",
		//Long:  SnapshotCreateCommandHelpText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.ValidateList(cmd), util.Fatal)
			util.CheckErr(options.RunSnapshotList(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "n", options.volName,
		"unique volume name.")
	cmd.MarkPersistentFlagRequired("volname")
	return cmd
}

// Validate validates the flag values
func (c *CmdSnaphotCreateOptions) ValidateList(cmd *cobra.Command) error {
	if c.volName == "" {
		return errors.New("--volname is missing. Please specify an unique name")
	}
	return nil
}

// RunSnapshotCreate does tasks related to mayaserver.
func (c *CmdSnaphotCreateOptions) RunSnapshotList(cmd *cobra.Command) error {
	fmt.Println("Executing volume snapshot list...")

	resp := mapiserver.ListSnapshot(c.volName)
	if resp != nil {
		return errors.New(fmt.Sprintf("Error: %v", resp))
	}

	fmt.Printf("Volume snapshots are:%v\n", resp)

	return nil
}
