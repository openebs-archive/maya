package jiva

import (
	"net/http"
	"strings"
)

const (
	metadataSuffix     = ".meta"
	imgSuffix          = ".img"
	volumeMetaData     = "volume.meta"
	defaultSectorSize  = 4096
	headPrefix         = "volume-head-"
	headSuffix         = ".img"
	headName           = headPrefix + "%03d" + headSuffix
	diskPrefix         = "volume-snap-"
	diskSuffix         = ".img"
	diskName           = diskPrefix + "%s" + diskSuffix
	maximumChainLength = 250
)

// Volumes contains parameters to describe a volume
type Volumes struct {
	Resource
	Name         string `json:"name"`
	ReplicaCount int    `json:"replicaCount"`
	Endpoint     string `json:"endpoint"`
}

// VolumeCollection is a collection of volume
type VolumeCollection struct {
	Collection
	Data []Volumes `json:"data"`
}

// Replica contains information about a replica
type Replica struct {
	Resource
	Address string `json:"address"`
	Mode    string `json:"mode"`
}

// InfoReplica describes stats of a Replica
type InfoReplica struct {
	Resource
	Dirty           bool                `json:"dirty"`
	Rebuilding      bool                `json:"rebuilding"`
	Head            string              `json:"head"`
	Parent          string              `json:"parent"`
	Size            string              `json:"size"`
	SectorSize      int64               `json:"sectorSize"`
	State           string              `json:"state"`
	Chain           []string            `json:"chain"`
	Disks           map[string]DiskInfo `json:"disks"`
	RemainSnapshots int                 `json:"remainsnapshots"`
	RevisionCounter int64               `json:"revisioncounter"`
}

// DiskInfo contains information about a disk
type DiskInfo struct {
	Name        string   `json:"name"`
	Parent      string   `json:"parent"`
	Children    []string `json:"children"`
	Removed     bool     `json:"removed"`
	UserCreated bool     `json:"usercreated"`
	Created     string   `json:"created"`
	Size        string   `json:"size"`
}

// ReplicaCollection is a collection of a replica
type ReplicaCollection struct {
	Collection
	Data []Replica `json:"data"`
}

// MarkDiskAsRemovedInput contains disk name to be removed
type MarkDiskAsRemovedInput struct {
	Resource
	Name string `json:"name"`
}

// ReplicaClient is Client structure
type ReplicaClient struct {
	Address    string
	SyncAgent  string
	Host       string
	httpClient *http.Client
}

// ControllerClient describes jiva controller
type ControllerClient struct {
	Address    string
	Host       string
	httpClient *http.Client
}

// RevertInput defines snapshot name to be reverted
type RevertInput struct {
	Resource
	Name string `json:"name"`
}

// SnapshotInput is Input struct to create
// snapshot
type SnapshotInput struct {
	Resource
	Name        string            `json:"name"`
	UserCreated bool              `json:"usercreated"`
	Created     string            `json:"created"`
	Labels      map[string]string `json:"labels"`
}

// SnapshotOutput defines resource of volume
type SnapshotOutput struct {
	Resource
}

// Resource defines disk info
type Resource struct {
	Id      string            `json:"id,omitempty"`
	Type    string            `json:"type,omitempty"`
	Links   map[string]string `json:"links"`
	Actions map[string]string `json:"actions"`
}

// Collection defines attributes for VolumeCollection
type Collection struct {
	Type         string                 `json:"type,omitempty"`
	ResourceType string                 `json:"resourceType,omitempty"`
	Links        map[string]string      `json:"links,omitempty"`
	CreateTypes  map[string]string      `json:"createTypes,omitempty"`
	Actions      map[string]string      `json:"actions,omitempty"`
	SortLinks    map[string]string      `json:"sortLinks,omitempty"`
	Pagination   *Pagination            `json:"pagination,omitempty"`
	Sort         *Sort                  `json:"sort,omitempty"`
	Filters      map[string][]Condition `json:"filters,omitempty"`
}

// Sort contains sort data
type Sort struct {
	Name    string `json:"name,omitempty"`
	Order   string `json:"order,omitempty"`
	Reverse string `json:"reverse,omitempty"`
}

// Condition defines set of cond. with modifier and value
type Condition struct {
	Modifier string      `json:"modifier,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

// Pagination defines attributes for pagination
type Pagination struct {
	Marker   string `json:"marker,omitempty"`
	First    string `json:"first,omitempty"`
	Previous string `json:"previous,omitempty"`
	Next     string `json:"next,omitempty"`
	Limit    *int64 `json:"limit,omitempty"`
	Total    *int64 `json:"total,omitempty"`
	Partial  bool   `json:"partial,omitempty"`
}

// Filter returns filtered results
func Filter(list []string, check func(string) bool) []string {
	result := make([]string, 0, len(list))
	for _, i := range list {
		if check(i) {
			result = append(result, i)
		}
	}
	return result
}

// Contains checks if string presents in string array
func Contains(arr []string, val string) bool {
	for _, a := range arr {
		if a == val {
			return true
		}
	}
	return false
}

// IsHeadDisk checks if Disk is a Head
func IsHeadDisk(diskName string) bool {
	if strings.HasPrefix(diskName, headPrefix) && strings.HasSuffix(diskName, headSuffix) {
		return true
	}
	return false
}
