package jiva

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/golang/glog"
	client "github.com/openebs/maya/pkg/client/jiva"
)

// SnapshotList to list the created snapshot for given volume
func SnapshotList(name string, controllerIP string) (map[string]client.DiskInfo, error) {
	controller, err := client.NewControllerClient(controllerIP + ":9501")

	if err != nil {
		return nil, err
	}

	replicas, err := controller.ListReplicas(controller.Address)
	if err != nil {
		return nil, err
	}

	first := true
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
				return snapdisk, err
			}
			return snapdisk, nil

		}
	}
	return nil, nil
}

// getChain contains the linked info related to replicas
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

// getData to get the linked Diskinfo related to replicas
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

// CheckSnapshotExist check the existence of snapshot in chain of created snapshots
func CheckSnapshotExist(snapshot string, controllerIP string) error {
	controller, err := client.NewControllerClient(controllerIP + ":9501")

	glog.Infof("Validates existence of snapshot [%s] before create", snapshot)

	if err != nil {
		return err
	}

	replicas, err := controller.ListReplicas(controller.Address)
	if err != nil {
		return err
	}

	first := true
	for _, r := range replicas {
		if r.Mode != "RW" {
			continue
		}

		if first {
			first = false
			chain, _ := getChain(r.Address)
			_, index := getNameAndIndex(chain, snapshot)
			if index > 0 {
				return fmt.Errorf("snapshot [%s] already exists", snapshot)
			}
		}
		return err
	}
	return err
}

// getNameAndIndex get the name and index value based on the existence of
// snapshot. If snapshot is already exists the index value will be -1
// if not then any possitive number
func getNameAndIndex(chain []string, snapshot string) (string, int) {
	index := find(chain, snapshot)

	if index < 0 {
		snapshot = fmt.Sprintf("volume-snap-%s.img", snapshot)
		glog.Infof("Requested snapshot is: %v", snapshot)
		index = find(chain, snapshot)
	}

	if index < 0 {
		return "", index
	}
	return snapshot, index
}

func find(list []string, item string) int {
	for i, val := range list {
		if val == item {
			return i
		}
	}
	return -1
}
