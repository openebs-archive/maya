package zpool

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/openebs/maya/pkg/util"
)

const (
	NoPoolsAvailable = "no pools available"
	Zpool            = "./zpool"
	Offline          = "OFFLINE"
	Online           = "ONLINE"
	Degraded         = "DEGRADED"
	Faulted          = "FAULTED"
	Removed          = "REMOVED"
	Unavail          = "UNAVAIL"
)

var (
	Status map[string]float64 = map[string]float64{Offline: 0, Online: 1, Degraded: 2, Faulted: 3, Removed: 4, Unavail: 5, NoPoolsAvailable: 6}
)

type Stats struct {
	Status              string
	Used                string
	Free                string
	Size                string
	UsedCapacityPercent string
}

func Run(runner util.Runner, args ...string) ([]byte, error) {
	status, err := runner.RunCombinedOutput(Zpool, "list", "-Hp")
	if err != nil {
		return nil, err
	}
	return status, nil
}

func isNoPoolAvailable(str string) bool {
	return strings.Contains(str, NoPoolsAvailable)
}

func ListParser(output []byte) (Stats, error) {
	str := string(output)
	if isNoPoolAvailable(str) {
		return Stats{}, errors.New(NoPoolsAvailable)
	}
	stats := strings.Fields(string(output))
	return Stats{
		Size:                stats[1],
		Used:                stats[2],
		Free:                stats[3],
		Status:              stats[8],
		UsedCapacityPercent: stats[6],
	}, nil
}
