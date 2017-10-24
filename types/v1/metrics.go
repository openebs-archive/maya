package v1

// OpenEBSVolumeMetrics is used to store the collected metrics
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
	TotalWriteBlockCount string `json:"TotatWriteBlockCount"`

	UsedLogicalBlocks string `json:"UsedLogicalBlocks"`
	UsedBlocks        string `json:"UsedBlocks"`
	SectorSize        string `json:"SectorSize"`
}

type Resource struct {
	Id      string            `json:"id,omitempty"`
	Type    string            `json:"type,omitempty"`
	Links   map[string]string `json:"links"`
	Actions map[string]string `json:"actions"`
}
