package command

import (
	"encoding/json"
	"fmt"

	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rancher/go-rancher/client"
)

type Status struct {
	client.Resource
	ReplicaCounter  int64 `json:"replicacounter"`
	RevisionCounter int64 `json:"revisioncounter"`
}

type VolumeStats struct {
	client.Resource
	RevisionCounter int64         `json:"RevisionCounter"`
	ReplicaCounter  int64         `json:"ReplicaCounter"`
	SCSIIOCount     map[int]int64 `json:"SCSIIOCount"`

	ReadIOPS            string `json:"ReadIOPS"`
	TotalReadTime       string `json:"TotalReadTime"`
	TotalReadBlockCount string `json:"TotalReadBlockCount"`

	WriteIOPS            string `json:"WriteIOPS"`
	TotalWriteTime       string `json:"TotalWriteTime"`
	TotalWriteBlockCount string `json:"TotatWriteBlockCount"`
}

// VsmStatsCommand is a command implementation struct
type VsmStatsCommand struct {
	Meta
	address     string
	host        string
	length      int
	replica_ips string
	Json        string
}

// ReplicaClient is Client structure
type ReplicaClient struct {
	address    string
	syncAgent  string
	host       string
	httpClient *http.Client
}

type ControllerClient struct {
	address    string
	host       string
	httpClient *http.Client
}

type StatsArr struct {
	IQN    string `json:"Iqn"`
	Volume string `json:"Volume"`
	Portal string `json:"Portal"`
	Size   string `json:"Size"`

	ReadIOPS  int64 `json:"ReadIOPS"`
	WriteIOPS int64 `json:"WriteIOPS"`

	ReadThroughput  float64 `json:"ReadThroughput"`
	WriteThroughput float64 `json:"WriteThroughput"`

	ReadLatency  float64 `json:"ReadLatency"`
	WriteLatency float64 `json:"WriteLatency"`

	AvgReadBlockSize  int64 `json:"AvgReadBlockSize"`
	AvgWriteBlockSize int64 `json:"AvgWriteBlockSize"`
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
		err, err1, err2 error
		status          Status
		stats1          VolumeStats
		stats2          VolumeStats
		statusArray     []string
		ReadLatency     int64
		WriteLatency    int64

		AvgReadBlockCountPS  int64
		AvgWriteBlockCountPS int64
	)

	flags := c.Meta.FlagSet("vsm-stats", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	flags.StringVar(&c.Json, "json", "", "")

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
	if annotations.ControllerStatus != "Running" {
		fmt.Println("Volume not reachable")
		return -1
	}
	replicas := strings.Split(annotations.Replicas, ",")
	for _, replica := range replicas {
		err, errCode1 := GetStatus(replica+":9502", &status)
		if err != nil {
			if errCode1 == 500 || strings.Contains(err.Error(), "EOF") {
				statusArray = append(statusArray, fmt.Sprintf("%-15s %-12s%-10s", replica, "Waiting", "Unknown"))

			} else {
				statusArray = append(statusArray, fmt.Sprintf("%-15s %-12s%-10s", replica, "Offline", "Unknown"))
			}
		} else {
			statusArray = append(statusArray, fmt.Sprintf("%-15s %-10s  %d", replica, "Online", status.RevisionCounter))
		}
	}

	//Get VolumeStats
	err1, _ = GetVolumeStats(annotations.ClusterIP+":9501", &stats1)
	time.Sleep(1 * time.Second)
	err2, _ = GetVolumeStats(annotations.ClusterIP+":9501", &stats2)

	if (err1 != nil) || (err2 != nil) {
		fmt.Println("Volume not reachable")
	}
	ReadIOPSi, _ := strconv.ParseInt(stats1.ReadIOPS, 10, 64)
	ReadIOPSf, _ := strconv.ParseInt(stats2.ReadIOPS, 10, 64)
	ReadIOPSPS := ReadIOPSf - ReadIOPSi

	ReadTimePSi, _ := strconv.ParseInt(stats1.TotalReadTime, 10, 64)
	ReadTimePSf, _ := strconv.ParseInt(stats2.TotalReadTime, 10, 64)
	ReadTimePS := ReadTimePSf - ReadTimePSi

	ReadBlockCountPSi, _ := strconv.ParseInt(stats1.TotalReadBlockCount, 10, 64)
	ReadBlockCountPSf, _ := strconv.ParseInt(stats2.TotalReadBlockCount, 10, 64)
	ReadBlockCountPS := ReadBlockCountPSf - ReadBlockCountPSi

	RThroughput := ReadBlockCountPS
	if ReadIOPSPS != 0 {
		ReadLatency = ReadTimePS / ReadIOPSPS
		AvgReadBlockCountPS = ReadBlockCountPS / ReadIOPSPS
	} else {
		ReadLatency = 0
		AvgReadBlockCountPS = 0
	}

	WriteIOPSi, _ := strconv.ParseInt(stats1.WriteIOPS, 10, 64)
	WriteIOPSf, _ := strconv.ParseInt(stats2.WriteIOPS, 10, 64)
	WriteIOPSPS := WriteIOPSf - WriteIOPSi

	WriteTimePSi, _ := strconv.ParseInt(stats1.TotalWriteTime, 10, 64)
	WriteTimePSf, _ := strconv.ParseInt(stats2.TotalWriteTime, 10, 64)
	WriteTimePS := WriteTimePSf - WriteTimePSi

	WriteBlockCountPSi, _ := strconv.ParseInt(stats1.TotalWriteBlockCount, 10, 64)
	WriteBlockCountPSf, _ := strconv.ParseInt(stats2.TotalWriteBlockCount, 10, 64)
	WriteBlockCountPS := WriteBlockCountPSf - WriteBlockCountPSi

	WThroughput := WriteBlockCountPS
	if WriteIOPSPS != 0 {
		WriteLatency = WriteTimePS / WriteIOPSPS
		AvgWriteBlockCountPS = WriteBlockCountPS / WriteIOPSPS
	} else {
		WriteLatency = 0
		AvgWriteBlockCountPS = 0
	}

	fmt.Println("------------------------------------")
	// json formatting and showing default output
	if c.Json == "json" {

		stat1 := StatsArr{

			IQN:    annotations.Iqn,
			Volume: args[0],
			Portal: annotations.TargetPortal,
			Size:   annotations.VolSize,

			ReadIOPS:  ReadIOPSPS,
			WriteIOPS: WriteIOPSPS,

			ReadThroughput:  float64(RThroughput) / 104857,
			WriteThroughput: float64(WThroughput) / 104857,

			ReadLatency:  float64(ReadLatency) / 1000000,
			WriteLatency: float64(WriteLatency) / 1000000,

			AvgReadBlockSize:  AvgReadBlockCountPS / 1024,
			AvgWriteBlockSize: AvgWriteBlockCountPS / 1024,
		}

		data, err := json.Marshal(stat1)

		if err != nil {

			panic(err)
		}

		os.Stdout.Write(data)

		fmt.Println("\n------------------------------------")

	} else {
		fmt.Printf("%7s: %-48s\n", "IQN", annotations.Iqn)
		fmt.Printf("%7s: %-16s\n", "Volume", args[0])
		fmt.Printf("%7s: %-15s\n", "Portal", annotations.TargetPortal)
		fmt.Printf("%7s: %-6s\n\n", "Size", annotations.VolSize)
		fmt.Printf("%s         %s      %s\n", "Replica", "Status", "DataUpdateIndex")

		for i, _ := range statusArray {
			fmt.Printf("%s\n", statusArray[i])
		}
		fmt.Println("------------------------------------")

		// Printing in tabular form
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight|tabwriter.Debug)
		fmt.Fprintf(w, "r/s\tw/s\tr(MB/s)\tw(MB/s)\trLat(ms)\twLat(ms)\trBlk(KB)\twBlk(KB)\t\n")
		fmt.Fprintf(w, "%d\t%d\t%.3f\t%.3f\t%.3f\t%.3f\t%d\t%d\t\n", ReadIOPSPS, WriteIOPSPS, float64(RThroughput)/1048576, float64(WThroughput)/1048576, float64(ReadLatency)/1000000, float64(WriteLatency)/1000000, AvgReadBlockCountPS/1024, AvgWriteBlockCountPS/1024)
		w.Flush()

	}
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

// NewControllerClient create the new replica client
func NewControllerClient(address string) (*ControllerClient, error) {
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

	timeout := time.Duration(2 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	return &ControllerClient{
		host:       parts[0],
		address:    address,
		httpClient: client,
	}, nil
}

// GetStatus will return json response and statusCode
func GetVolumeStats(address string, obj interface{}) (error, int) {
	controller, err := NewControllerClient(address)
	if err != nil {
		return err, -1
	}
	url := controller.address + "/stats"
	resp, err := controller.httpClient.Get(url)
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
