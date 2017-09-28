package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/rancher/go-rancher/client"
)

type Status struct {
	Resource        client.Resource
	ReplicaCounter  int64  `json:"replicacounter"`
	RevisionCounter string `json:"revisioncounter"`
}

type VolumeStats struct {
	Resource        client.Resource
	RevisionCounter int64         `json:"RevisionCounter"`
	ReplicaCounter  int64         `json:"ReplicaCounter"`
	SCSIIOCount     map[int]int64 `json:"SCSIIOCount"`

	ReadIOPS            string `json:"ReadIOPS"`
	TotalReadTime       string `json:"TotalReadTime"`
	TotalReadBlockCount string `json:"TotalReadBlockCount"`

	WriteIOPS            string `json:"WriteIOPS"`
	TotalWriteTime       string `json:"TotalWriteTime"`
	TotalWriteBlockCount string `json:"TotatWriteBlockCount"`

	SectorSize        string `json:"SectorSize"`
	UsedBlocks        string `json:"UsedBlocks"`
	UsedLogicalBlocks string `json:"UsedLogicalBlocks"`
}

// VsmStatsCommand is a command implementation struct
type VsmStatsCommand struct {
	Meta
	Address     string
	Host        string
	Length      int
	Replica_ips string
	Json        string
}

// ReplicaClient is Client structure
type ReplicaClient struct {
	Address    string
	SyncAgent  string
	Host       string
	httpClient *http.Client
}

type ControllerClient struct {
	Address    string
	Host       string
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

	SectorSize  float64 `json:"SectorSize"`
	ActualUsed  float64 `json:"ActualUsed"`
	LogicalSize float64 `json:"LogicalSize"`
}

type Annotation struct {
	IQN    string `json:"Iqn"`
	Volume string `json:"Volume"`
	Portal string `json:"Portal"`
	Size   string `json:"Size"`
}

const (
	bytesToGB = 1073741824
	bytesToMB = 1048567
	mic_sec   = 1000000
	bytesToKB = 1024
	minwidth  = 0
	maxwidth  = 0
	padding   = 3
)

// Help shows helpText for a particular CLI command
func (c *VsmStatsCommand) Help() string {
	helpText := `
Usage: maya volume stats <vol> [-json]

This command displays Volume statistics including running status
and Read/Write.

Volume stats options:
    -json
      Output stats in json format

`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (c *VsmStatsCommand) Synopsis() string {
	return "Displays runtime statistics of a Volume"
}

// Run holds the flag values for CLI subcommands
func (c *VsmStatsCommand) Run(args []string) int {

	var (
		err, err1, err3 error
		err2, err4      int
		status          Status
		stats1, stats2  VolumeStats
		repStatus       string
	)
	statusArray := make([]string, 6)

	flags := c.Meta.FlagSet("volume stats", FlagSetClient)
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
		fmt.Println(err)
		return -1
	}

	if annotations.ControllerStatus != "Running" {
		fmt.Println("Volume not reachable")
		return -1
	}

	replicaCount := 0
	replicaStatus := strings.Split(annotations.ReplicaStatus, ",")
	for _, repStatus = range replicaStatus {
		if repStatus == "Pending" {
			statusArray[replicaCount] = "Unknown"
			statusArray[replicaCount+1] = "Unknown"
			statusArray[replicaCount+2] = "Unknown"
			replicaCount += 3
		}
	}

	replicas := strings.Split(annotations.Replicas, ",")
	for _, replica := range replicas {
		err, errCode1 := GetStatus(replica+":9502", &status)
		if err != nil {
			if errCode1 == 500 || strings.Contains(err.Error(), "EOF") {
				statusArray[replicaCount] = replica
				statusArray[replicaCount+1] = "Waiting"
				statusArray[replicaCount+2] = "Unknown"

			} else {
				statusArray[replicaCount] = replica
				statusArray[replicaCount+1] = "Offline"
				statusArray[replicaCount+2] = "Unknown"
			}
			replicaCount += 3
		} else {
			statusArray[replicaCount] = replica
			statusArray[replicaCount+1] = "Online"
			statusArray[replicaCount+2] = status.RevisionCounter
			replicaCount += 3
		}

	}
	//GetVolumeStats gets volume stats
	err1, err2 = GetVolumeStats(annotations.ClusterIP+":9501", &stats1)
	if err1 != nil {
		if (err2 == 500) || (err2 == 503) || err1 != nil {
			fmt.Println("Volume not Reachable\n", err1)
			return -1
		}
	} else {
		time.Sleep(1 * time.Second)
		err3, err4 = GetVolumeStats(annotations.ClusterIP+":9501", &stats2)
		if err3 != nil {
			if err4 == 500 || err4 == 503 || err3 != nil {
				fmt.Println("Volume not Reachable\n", err3)
				return -1
			}
		} else {

			//StatsOutput displays output
			StatsOutput(c, annotations, args, statusArray, stats1, stats2)

		}
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
		Host:       parts[0],
		Address:    address,
		SyncAgent:  syncAgent,
		httpClient: client,
	}, nil
}

// GetStatus will return json response and statusCode
func GetStatus(address string, obj interface{}) (error, int) {
	replica, err := NewReplicaClient(address)
	if err != nil {
		return err, -1
	}
	url := replica.Address + "/stats"
	resp, err := replica.httpClient.Get(url)
	if resp != nil {
		if resp.StatusCode == 500 {
			return errors.New("Internal Server Error"), 500
		} else if resp.StatusCode == 503 {
			return errors.New("Service Unavailable"), 503
		}
	} else {
		return errors.New("Server Not Reachable"), -1
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
		Host:       parts[0],
		Address:    address,
		httpClient: client,
	}, nil
}

// GetStatus will return json response and statusCode
func GetVolumeStats(address string, obj interface{}) (error, int) {
	controller, err := NewControllerClient(address)
	if err != nil {
		return err, -1
	}
	url := controller.Address + "/stats"
	resp, err := controller.httpClient.Get(url)
	if resp != nil {
		if resp.StatusCode == 500 {
			return errors.New("Internal Server Error"), 500
		} else if resp.StatusCode == 503 {
			return errors.New("Service Unavailable"), 503
		}
	} else {
		return errors.New("Server Not Reachable"), -1
	}
	if err != nil {
		return err, -1
	}
	defer resp.Body.Close()
	rc := json.NewDecoder(resp.Body).Decode(obj)
	return rc, 0
}

// StatsOutput will return error code if any otherwise return zero
func StatsOutput(c *VsmStatsCommand, annotations *Annotations, args []string, statusArray []string, stats1 VolumeStats, stats2 VolumeStats) error {

	var (
		err          error
		ReadLatency  int64
		WriteLatency int64

		AvgReadBlockCountPS  int64
		AvgWriteBlockCountPS int64
	)

	// 10 and 64 represents decimal and bits respectively
	iReadIOPS, _ := strconv.ParseInt(stats1.ReadIOPS, 10, 64) // Initial
	fReadIOPS, _ := strconv.ParseInt(stats2.ReadIOPS, 10, 64) // Final
	readIOPS := fReadIOPS - iReadIOPS

	iReadTimePS, _ := strconv.ParseInt(stats1.TotalReadTime, 10, 64)
	fReadTimePS, _ := strconv.ParseInt(stats2.TotalReadTime, 10, 64)
	readTimePS := fReadTimePS - iReadTimePS

	iReadBlockCountPS, _ := strconv.ParseInt(stats1.TotalReadBlockCount, 10, 64)
	fReadBlockCountPS, _ := strconv.ParseInt(stats2.TotalReadBlockCount, 10, 64)
	readBlockCountPS := fReadBlockCountPS - iReadBlockCountPS

	rThroughput := readBlockCountPS
	if readIOPS != 0 {
		ReadLatency = readTimePS / readIOPS
		AvgReadBlockCountPS = readBlockCountPS / readIOPS
	} else {
		ReadLatency = 0
		AvgReadBlockCountPS = 0
	}

	iWriteIOPS, _ := strconv.ParseInt(stats1.WriteIOPS, 10, 64)
	fWriteIOPS, _ := strconv.ParseInt(stats2.WriteIOPS, 10, 64)
	writeIOPS := fWriteIOPS - iWriteIOPS

	iWriteTimePS, _ := strconv.ParseInt(stats1.TotalWriteTime, 10, 64)
	fWriteTimePS, _ := strconv.ParseInt(stats2.TotalWriteTime, 10, 64)
	writeTimePS := fWriteTimePS - iWriteTimePS

	iWriteBlockCountPS, _ := strconv.ParseInt(stats1.TotalWriteBlockCount, 10, 64)
	fWriteBlockCountPS, _ := strconv.ParseInt(stats2.TotalWriteBlockCount, 10, 64)
	writeBlockCountPS := fWriteBlockCountPS - iWriteBlockCountPS

	wThroughput := writeBlockCountPS
	if writeIOPS != 0 {
		WriteLatency = writeTimePS / writeIOPS
		AvgWriteBlockCountPS = writeBlockCountPS / writeIOPS
	} else {
		WriteLatency = 0
		AvgWriteBlockCountPS = 0
	}

	sectorSize, _ := strconv.ParseFloat(stats2.SectorSize, 64) // Sector Size
	sectorSize = sectorSize

	logicalSize, _ := strconv.ParseFloat(stats2.UsedBlocks, 64) // Logical Size
	logicalSize = logicalSize * sectorSize

	actualUsed, _ := strconv.ParseFloat(stats2.UsedLogicalBlocks, 64) // Actual Used
	actualUsed = actualUsed * sectorSize

	annotation := Annotation{
		IQN:    annotations.Iqn,
		Volume: args[0],
		Portal: annotations.TargetPortal,
		Size:   annotations.VolSize,
	}

	if c.Json == "json" {

		stat1 := StatsArr{

			IQN:    annotations.Iqn,
			Volume: args[0],
			Portal: annotations.TargetPortal,
			Size:   annotations.VolSize,

			ReadIOPS:  readIOPS,
			WriteIOPS: writeIOPS,

			ReadThroughput:  float64(rThroughput) / bytesToMB, // bytes to MB
			WriteThroughput: float64(wThroughput) / bytesToMB,

			ReadLatency:  float64(ReadLatency) / mic_sec, // Microsecond
			WriteLatency: float64(WriteLatency) / mic_sec,

			AvgReadBlockSize:  AvgReadBlockCountPS / bytesToKB, // Bytes to KB
			AvgWriteBlockSize: AvgWriteBlockCountPS / bytesToKB,

			SectorSize:  sectorSize,
			ActualUsed:  actualUsed / bytesToGB,
			LogicalSize: logicalSize / bytesToGB,
		}

		data, err := json.MarshalIndent(stat1, "", "\t")

		if err != nil {
			fmt.Println("Can't Marshal the data ", err)
		}

		os.Stdout.Write(data)

	} else {

		// Printing using template
		tmpl, err1 := template.New("test").Parse("IQN     : {{.IQN}}\nVolume  : {{.Volume}}\nPortal  : {{.Portal}}\nSize    : {{.Size}}")
		err = err1
		if err != nil {
			fmt.Println("Can't Parse the template ", err)
		}
		err = tmpl.Execute(os.Stdout, annotation)
		if err != nil {
			fmt.Println("Can't execute the template ", err)
		}

		// Printing in tabular form
		q := tabwriter.NewWriter(os.Stdout, minwidth, maxwidth, padding, ' ', tabwriter.AlignRight|tabwriter.Debug)
		fmt.Fprintf(q, "\n\nReplica\tStatus\tDataUpdateIndex\t\n")
		fmt.Fprintf(q, "\t\t\t\n")
		for i := 0; i < 4; i += 3 {

			fmt.Fprintf(q, "%s\t%s\t%s\t\n", statusArray[i], statusArray[i+1], statusArray[i+2])
		}

		q.Flush()

		w := tabwriter.NewWriter(os.Stdout, minwidth, maxwidth, padding, ' ', tabwriter.AlignRight|tabwriter.Debug)
		fmt.Println("\n----------- Performance Stats -----------\n")
		fmt.Fprintf(w, "r/s\tw/s\tr(MB/s)\tw(MB/s)\trLat(ms)\twLat(ms)\t\n")
		fmt.Fprintf(w, "%d\t%d\t%.3f\t%.3f\t%.3f\t%.3f\t\n", readIOPS, writeIOPS, float64(rThroughput)/bytesToMB, float64(wThroughput)/bytesToMB, float64(ReadLatency)/mic_sec, float64(WriteLatency)/mic_sec)
		w.Flush()

		x := tabwriter.NewWriter(os.Stdout, minwidth, maxwidth, padding, ' ', tabwriter.AlignRight|tabwriter.Debug)
		fmt.Println("\n------------ Capacity Stats -------------\n")
		fmt.Fprintf(x, "Logical(GB)\tUsed(GB)\t\n")
		fmt.Fprintf(x, "%.3f\t%.3f\t\n", logicalSize/bytesToGB, actualUsed/bytesToGB)
		x.Flush()
	}
	return err
}
