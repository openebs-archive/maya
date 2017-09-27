package jiva

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/openebs/maya/command"
)

// SnapshotList to list the created snapshot for given volume
func SnapshotList(name string) error {
	annotations, err := command.GetVolumeSpec(name)
	if err != nil || annotations == nil {

		return err
	}
	controller, err := command.NewControllerClient(annotations.ControllerIP + ":9501")

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
			if err != err {
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
			fmt.Println(command.FormatList(out))
		}
	}
	return nil
}

// ListReplicas to get the details of all the existing replicas
// which contains address and mode of those replicas (RW/R/W) as well as
// resource information.
func (c *ControllerClient) ListReplicas(path string) ([]Replica, error) {
	var resp ReplicaCollection

	err := c.get(path+"/replicas", &resp)

	return resp.Data, err
}

// getChain contains the linked info related to replicas
func getChain(address string) ([]string, error) {
	repClient, err := command.NewReplicaClient(address)
	if err != nil {
		return nil, err
	}

	r, err := repClient.GetReplica()
	if err != nil {
		return nil, err
	}

	return r.Chain, err
}

// getData to get the linked Diskinfo related to replicas
func getData(address string) (map[string]command.DiskInfo, error) {
	repClient, err := command.NewReplicaClient(address)
	if err != nil {
		return nil, err
	}

	r, err := repClient.GetReplica()
	if err != nil {
		return nil, err
	}

	return r.Disks, err

}

// GetReplica will return the InfoReplica struct which contains info
// related to specific replica
func (c *ReplicaClient) GetReplica() (InfoReplica, error) {
	var replica InfoReplica

	err := c.get(c.Address+"/replicas/1", &replica)

	return replica, err
}

func (c *ReplicaClient) get(url string, obj interface{}) error {
	if !strings.HasPrefix(url, "http") {
		url = c.Address + url
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(obj)
}

func (c *ReplicaClient) post(path string, req, resp interface{}) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	bodyType := "application/json"
	url := path

	if !strings.HasPrefix(url, "http") {
		url = c.Address + path
	}

	httpResp, err := c.httpClient.Post(url, bodyType, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 300 {
		content, _ := ioutil.ReadAll(httpResp.Body)
		return fmt.Errorf("Bad response: %d %s: %s", httpResp.StatusCode, httpResp.Status, content)
	}

	if resp == nil {
		return nil
	}

	return json.NewDecoder(httpResp.Body).Decode(resp)
}
