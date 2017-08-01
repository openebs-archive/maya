package command

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
)

type SnapshotListCommand struct {
	Meta
	Name string
}

func (c *SnapshotListCommand) Help() string {
	helpText := `
	Usage: maya vsm-snapshot list 
	    
	Command to list the snapshot.
`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (c *SnapshotListCommand) Synopsis() string {
	return "List the snapshots"
}

func (c *SnapshotListCommand) Run(args []string) int {

	flags := c.Meta.FlagSet("vsm-snapshot", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }

	flags.StringVar(&c.Name, "name", "", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	if err := ListSnapshot(c.Name); err != nil {
		log.Fatalf("Error running list-snapshot command: %v", err)
		return 1
	}
	return 0
}
func ListSnapshot(name string) error {
	//var annotations Annotations
	annotations, err := GetVolAnnotations(name)
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

	first := true
	snapshots := []string{}
	for _, r := range replicas {
		if r.Mode != "RW" {
			continue
		}

		if first {
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

	format := "%s\n"
	tw := tabwriter.NewWriter(os.Stdout, 0, 20, 1, ' ', 0)
	fmt.Fprintf(tw, format, "Snapshot_Name")
	for _, s := range snapshots {
		s = strings.TrimSuffix(strings.TrimPrefix(s, "volume-snap-"), ".img")
		fmt.Fprintf(tw, format, s)
	}
	tw.Flush()

	return nil
}

func (c *ControllerClient) ListReplicas(path string) ([]Replica, error) {
	var resp ReplicaCollection

	err := c.get(path+"/replicas", &resp)

	return resp.Data, err
}

func getChain(address string) ([]string, error) {
	repClient, err := NewReplicaClient(address)
	if err != nil {
		return nil, err
	}

	r, err := repClient.GetReplica()
	if err != nil {
		return nil, err
	}

	return r.Chain, err
}

func (c *ReplicaClient) GetReplica() (InfoReplica, error) {
	var replica InfoReplica

	err := c.get(c.address+"/replicas/1", &replica)
	return replica, err
}

func (c *ReplicaClient) get(url string, obj interface{}) error {
	if !strings.HasPrefix(url, "http") {
		url = c.address + url
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(obj)
}
