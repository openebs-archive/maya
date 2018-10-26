package v1

import "encoding/json"

// VolumeMetrics is used to store the collected metrics
// all the stats exposed by jiva stored into OpenEBSVolumeMetrics fields
type VolumeMetrics struct {
	Resource        Resource
	RevisionCounter int64         `json:"RevisionCounter"`
	ReplicaCounter  int64         `json:"ReplicaCounter"`
	SCSIIOCount     map[int]int64 `json:"SCSIIOCount"`

	ReadIOPS            string `json:"ReadIOPS"`
	TotalReadTime       string `json:"TotalReadTime"`
	TotalReadBlockCount string `json:"TotalReadBlockCount"`

	WriteIOPS            string `json:"WriteIOPS"`
	TotalWriteTime       string `json:"TotalWriteTime"`
	TotalWriteBlockCount string `json:"TotalWriteBlockCount"`

	UsedLogicalBlocks string  `json:"UsedLogicalBlocks"`
	UsedBlocks        string  `json:"UsedBlocks"`
	SectorSize        string  `json:"SectorSize"`
	Size              string  `json:"Size"`
	UpTime            float64 `json:"UpTime"`
	Name              string  `json:"Name"`
}

// VolumeStats is used to store the collected metrics
// TODO: Make this generic, so that it can be used by mayactl
// and other components of maya.
type VolumeStats struct {
	Iqn                 string      `json:"iqn"`
	Reads               json.Number `json:"ReadIOPS"`
	TotalReadTime       json.Number `json:"TotalReadTime"`
	TotalReadBlockCount json.Number `json:"TotalReadBlockCount"`
	TotalReadBytes      json.Number `json:"TotalReadBytes"`

	Writes               json.Number `json:"WriteIOPS"`
	TotalWriteTime       json.Number `json:"TotalWriteTime"`
	TotalWriteBlockCount json.Number `json:"TotalWriteBlockCount"`
	TotalWriteBytes      json.Number `json:"TotalWriteBytes"`

	UsedLogicalBlocks json.Number `json:"UsedLogicalBlocks"`
	UsedBlocks        json.Number `json:"UsedBlocks"`
	SectorSize        json.Number `json:"SectorSize"`
	Size              json.Number `json:"Size"`
	UpTime            float64     `json:"UpTime"`
	Name              string      `json:"Name"`
}

type VolStatus struct {
	Resource        Resource
	ReplicaCounter  int64  `json:"replicacounter"`
	RevisionCounter string `json:"revisioncounter"`
}

type Resource struct {
	Id      string            `json:"id,omitempty"`
	Type    string            `json:"type,omitempty"`
	Links   map[string]string `json:"links"`
	Actions map[string]string `json:"actions"`
}

type Annotation struct {
	IQN    string `json:"Iqn"`
	Volume string `json:"Volume"`
	Portal string `json:"Portal"`
	Size   string `json:"Size"`
}

type StatsJSON struct {
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
