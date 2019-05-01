// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/openebs/maya/pkg/util"
)

// ZpoolStatus is pool's status
type ZpoolStatus string

const (
	Binary                          = "zpool" // Binary represents zpool binary
	Offline             ZpoolStatus = "OFFLINE"
	Online              ZpoolStatus = "ONLINE"
	Degraded            ZpoolStatus = "DEGRADED"
	Faulted             ZpoolStatus = "FAULTED"
	Removed             ZpoolStatus = "REMOVED"
	Unavail             ZpoolStatus = "UNAVAIL"
	NoPoolAvailable     ZpoolStatus = "no pools available"
	InCompleteStdoutErr             = "Couldn't receive complete output"
)

var (
	// Status is map of zpool status with values
	Status = map[ZpoolStatus]float64{
		Offline:         0,
		Online:          1,
		Degraded:        2,
		Faulted:         3,
		Removed:         4,
		Unavail:         5,
		NoPoolAvailable: 6,
	}
)

// Stats is used to store the values of parsed stats
// of zpool list -Hp command
type Stats struct {
	Status              ZpoolStatus // Status represents status of a Pool
	Name                string
	Used                string // Used size of Pools
	Free                string // Free size of Pools
	Size                string // Size of pool
	UsedCapacityPercent string // Used size of pools in precent
}

// Run is wrapper over RunCommandWithTimeoutContext for running zpool commands
func Run(timeout time.Duration, runner util.Runner, args ...string) ([]byte, error) {
	status, err := runner.RunCommandWithTimeoutContext(timeout, Binary, args...)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// IsNotAvailable checks whether any pool is availble or not.
func IsNotAvailable(str string) bool {
	return strings.Contains(str, string(NoPoolAvailable))
}

func isValid(stats string) ([]string, bool) {
	statsList := strings.Fields(stats)
	if len(statsList) < 9 {
		return nil, false
	}
	return statsList, true
}

// ListParser parses output of zpool list -Hp
// Command: zpool list -Hp
// Output:
// cstor-5ce4639a-2dc1-11e9-bbe3-42010a80017a	10670309376	716288	10669593088	-	0	0	1.00 ONLINE	-
func ListParser(output []byte) (Stats, error) {
	str := string(output)
	if IsNotAvailable(str) {
		return Stats{}, errors.New(string(NoPoolAvailable))
	}

	stats, ok := isValid(str)
	if !ok {
		return Stats{}, errors.New(InCompleteStdoutErr)
	}

	return Stats{
		Name:                stats[0],
		Size:                stats[1],
		Used:                stats[2],
		Free:                stats[3],
		UsedCapacityPercent: stats[6],
		Status:              ZpoolStatus(stats[8]),
	}, nil
}
