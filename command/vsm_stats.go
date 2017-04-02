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

type Stats struct {
	client.Resource
	ReplicaCounter  int64 `json:"replicacounter"`
	RevisionCounter int64 `json:"revisioncounter"`
}

type VsmStatsCommand struct {
	Meta
	address     string
	host        string
	length      int
	replica_ips string
}

type ReplicaClient struct {
	address    string
	syncAgent  string
	host       string
	httpClient *http.Client
}

func (c *VsmStatsCommand) Help() string {
	helpText := `
	Usage: maya vsm-stats [options] <replica-ip:port> 

  Display stats information about VSM.

  Stats Options:
`
	return strings.TrimSpace(helpText)
}

func (c *VsmStatsCommand) Synopsis() string {
	return "Display information about Vsm(s)"
}

func (c *VsmStatsCommand) Run(args []string) int {

	var stats Stats
	flags := c.Meta.FlagSet("vsm-stats", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	flags.StringVar(&c.replica_ips, "replica-ips", "", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}
	args = flags.Args()
	if len(args) < 1 {
		c.Ui.Error(c.Help())
		return 1
	}
	fmt.Printf("%2s %15s\n", "RevisionCounter", "ReplicaCounter")
	for counter := 0; counter <= 100; counter++ {
		time.Sleep(2 * time.Second)
		err := GetStatus(args[0], &stats)
		if err != nil {
			fmt.Println("\nERROR:", err)
			return -1
		}
		fmt.Printf("\r    %-17d %d", stats.RevisionCounter, stats.ReplicaCounter)
	}
	return 0
}

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

	timeout := time.Duration(30 * time.Second)
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

func GetStatus(address string, obj interface{}) error {
	replica, err := NewReplicaClient(address)
	if err != nil {
		return err
	}
	url := replica.address + "/stats"
	resp, err := replica.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(obj)
}
