package command

import (
	"fmt"
	"log"
	"strings"

	"github.com/openebs/maya/pkg/client/jiva"
)

// SnapshotListCommand is a command implementation struct
type SnapshotListCommand struct {
	Meta
	Name string
}

// Help shows helpText for a particular CLI command
func (c *SnapshotListCommand) Help() string {
	helpText := `
	Usage: maya snapshot list -volname <vol>

	This command will list all the snapshots of a Volume.

	`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (c *SnapshotListCommand) Synopsis() string {
	return "Lists all the snapshots of a Volume"
}

// Run holds the flag values for CLI subcommands
func (c *SnapshotListCommand) Run(args []string) int {

	flags := c.Meta.FlagSet("volume snapshot", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }

	flags.StringVar(&c.Name, "volname", "", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	if err := ListSnapshot(c.Name); err != nil {
		log.Fatalf("Error running list-snapshot command: %v", err)
		return 1
	}
	return 0
}

// ListSnapshot is used to list snapshot
func ListSnapshot(name string) error {

	annotations, err := GetVolAnnotations(name)
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

	first := true
	//snapshots := []string{}
	for _, r := range replicas {
		if r.Mode != "RW" {
			continue
		}

		/*	if first {
					first = false
					chain, err := getChain(r.Address)
					if err != nil {
						return err
					}
					// Replica can just started and haven't prepare the head
					// file yet
					if len(chain) == 0 {
						break
					}
					snapshots = chain[1:]
					continue
				}

				chain, err := getChain(r.Address)
				if err != nil {
					return err
				}

				snapshots = Filter(snapshots, func(i string) bool {
					return Contains(chain, i)
				})
			}

			/*format := "%s\n"
			tw := tabwriter.NewWriter(os.Stdout, 0, 20, 1, ' ', 0)
			fmt.Fprintf(tw, format, "Snapshot_Name")
			for _, s := range snapshots {
				s = strings.TrimSuffix(strings.TrimPrefix(s, "volume-snap-"), ".img")
				fmt.Fprintf(tw, format, s)
			}
			tw.Flush()
		*/

		//for _, r := range replicas {
		if first {
			first = false
			snapdisk, err := getData(r.Address)
			if err != nil {
				return err
			}
			out := make([]string, len(snapdisk)+1)

			out[0] = "Name|Created At|Size"
			var i int

			for _, disk := range snapdisk {
				//	if !IsHeadDisk(disk.Name) {
				out[i+1] = fmt.Sprintf("%s|%s|%s",
					strings.TrimSuffix(strings.TrimPrefix(disk.Name, "volume-snap-"), ".img"),
					disk.Created,
					disk.Size)
				i = i + 1
				//	}
			}
			fmt.Println(formatList(out))
		}

	}

	return nil
}

// getChain is used to get chain
func getChain(address string) ([]string, error) {
	repClient, err := client.NewReplicaClient(address)
	if err != nil {
		return nil, err
	}

	r, err := repClient.GetReplica()
	if err != nil {
		return nil, err
	}

	return r.Chain, err
}

// getData is used to get data
func getData(address string) (map[string]client.DiskInfo, error) {
	repClient, err := client.NewReplicaClient(address)
	if err != nil {
		return nil, err
	}

	r, err := repClient.GetReplica()
	if err != nil {
		return nil, err
	}

	return r.Disks, err

}
