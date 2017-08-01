package command

import "github.com/rancher/go-rancher/client"

type Volumes struct {
	client.Resource
	Name         string `json:"name"`
	ReplicaCount int    `json:"replicaCount"`
	Endpoint     string `json:"endpoint"`
}

type VolumeCollection struct {
	client.Collection
	Data []Volumes `json:"data"`
}

type Replica struct {
	client.Resource
	Address string `json:"address"`
	Mode    string `json:"mode"`
}

type InfoReplica struct {
	client.Resource
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

type DiskInfo struct {
	Name        string   `json:"name"`
	Parent      string   `json:"parent"`
	Children    []string `json:"children"`
	Removed     bool     `json:"removed"`
	UserCreated bool     `json:"usercreated"`
	Created     string   `json:"created"`
	Size        string   `json:"size"`
}

type ReplicaCollection struct {
	client.Collection
	Data []Replica `json:"data"`
}

func Filter(list []string, check func(string) bool) []string {
	result := make([]string, 0, len(list))
	for _, i := range list {
		if check(i) {
			result = append(result, i)
		}
	}
	return result
}

func Contains(arr []string, val string) bool {
	for _, a := range arr {
		if a == val {
			return true
		}
	}
	return false
}
