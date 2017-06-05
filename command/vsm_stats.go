package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/go-rancher/client"
)

// Stats implementation stats struct
type Stats struct {
	client.Resource
	ReplicaCounter  int64 `json:"replicacounter"`
	RevisionCounter int64 `json:"revisioncounter"`
}

// VsmStatsCommand is a command implementation struct
type VsmStatsCommand struct {
	Meta
	address     string
	host        string
	length      int
	replica_ips string
}

// ReplicaClient is Client structure
type ReplicaClient struct {
	address    string
	syncAgent  string
	host       string
	httpClient *http.Client
}

// Help shows helpText for a particular CLI command
func (c *VsmStatsCommand) Help() string {
	helpText := `
	Usage: maya vsm-stats <vsm-name> 

  Display VSM Stats.

`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (c *VsmStatsCommand) Synopsis() string {
	return "Display VSM Stats"
}

// Run holds the flag values for CLI subcommands
func (c *VsmStatsCommand) Run(args []string) int {

	var (
		err        error
		stats      Stats
		statsArray []string
	)

	flags := c.Meta.FlagSet("vsm-stats", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }

	if err := flags.Parse(args); err != nil {
		return 1
	}
	args = flags.Args()
	if len(args) < 1 {
		c.Ui.Error(c.Help())
		return 1
	}

	annotations, err := GetVolAnnotations(args[0])
	if err != nil || annotations == nil {
		return -1
	}

	for _, replica := range annotations.Replicas {
		err, errCode := GetStatus(replica+":9502", &stats)
		if err != nil {
			if errCode == 500 || strings.Contains(err.Error(), "EOF") {
				statsArray = append(statsArray, fmt.Sprintf("%-15s %-12s%-10s", replica, "Waiting", "Unknown"))

			} else {
				statsArray = append(statsArray, fmt.Sprintf("%-15s %-12s%-10s", replica, "Offline", "Unknown"))
			}
		} else {
			statsArray = append(statsArray, fmt.Sprintf("%-15s %-10s  %d", replica, "Online", stats.RevisionCounter))
		}
	}
	fmt.Println("------------------------------------")
	fmt.Printf("%7s: %-48s\n", "IQN", annotations.Iqn)
	fmt.Printf("%7s: %-16s\n", "Volume", args[0])
	fmt.Printf("%7s: %-15s\n", "Portal", annotations.VolAddr)
	fmt.Printf("%7s: %-6s\n\n", "Size", annotations.VolSize)
	fmt.Printf("%s         %s      %s\n", "Replica", "Status", "DataUpdateIndex")
	for i, _ := range statsArray {
		fmt.Printf("%s\n", statsArray[i])
	}
	fmt.Println("------------------------------------")
	return 0
}

// NewReplicaClient create the new replica client
func NewReplicaClient(address string) (*ReplicaClient, error) {
	if strings.HasPrefix(address, "tcp://") {
		address = address[6:]
	}

	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}

	if !strings.HasSuffix(address, "/v1") {
		address += "/v1"
	}

	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(u.Host, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("Invalid address %s, must have a port in it", address)
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	syncAgent := strings.Replace(address, fmt.Sprintf(":%d", port), fmt.Sprintf(":%d", port+2), -1)

	timeout := time.Duration(2 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	return &ReplicaClient{
		host:       parts[0],
		address:    address,
		syncAgent:  syncAgent,
		httpClient: client,
	}, nil
}

// GetStatus will return json response and statusCode
func GetStatus(address string, obj interface{}) (error, int) {
	replica, err := NewReplicaClient(address)
	if err != nil {
		return err, -1
	}
	url := replica.address + "/stats"
	resp, err := replica.httpClient.Get(url)
	if resp != nil {
		if resp.StatusCode == 500 {
			return err, 500
		} else if resp.StatusCode == 503 {
			return err, 503
		}
	} else {
		return err, -1
	}
	if err != nil {
		return err, -1
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(obj), 0
}
