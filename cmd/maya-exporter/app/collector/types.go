// Copyright © 2017-2019 The OpenEBS Authors
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

package collector

import (
	v1 "github.com/openebs/maya/pkg/stats/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

type volumeStatus int

const (
	_ volumeStatus = iota
	// Offline is the status of volume when no io's have been served
	// or volume may be in RO state (only for jiva)
	Offline
	// Degraded is the status of volume when volume is
	// performing in degraded mode but all features may available
	Degraded
	// Healthy is the status of volume when volume is serving io's
	// and all features are available or volume may be in RW state
	// (for jiva)
	Healthy
	// Unknown is the status of volume when no info is available
	Unknown
)

const (
	errored         = v1.ReplicaMode("ERR")
	writeOnly       = v1.ReplicaMode("WO")
	readOnly        = v1.ReplicaMode("RO")
	degraded        = v1.ReplicaMode("Degraded")
	readWrite       = v1.ReplicaMode("RW")
	healthy         = v1.ReplicaMode("Healthy")
	targetOffline   = v1.TargetMode("Offline")
	targetDegraded  = v1.TargetMode("Degraded")
	targetHealthy   = v1.TargetMode("Healthy")
	targetReadOnly  = v1.TargetMode("RO")
	targetReadWrite = v1.TargetMode("RW")
	host            = "127.0.0.1"
	port            = ":9500"
	endpoint        = "/v1/stats"
	jivaIQN         = "iqn.2016-09.com.openebs.jiva:"
	protocol        = "http://"
)

// Volume interface defines the interfaces that has methods to be
// implemented by various storage engines e.g. cstor, jiva etc.
type Volume interface {
	Getter
	Parser
}

// Parser interface defines the method that to be implemented by the
// Cstor and Jiva. parse() is used to parse
// the response into the Metrics struct.
type Parser interface {
	parse(stats v1.VolumeStats, metrics *metrics) stats
}

// Getter interface defines the method that to be implemented by
// the cstor and jiva. getter() is used
// to collect the stats from the Jiva and Cstor.
type Getter interface {
	get() (v1.VolumeStats, error)
}

// Collector is an interface that is used for initializing
// various collectors for collecting volume and pool metrics
type Collector interface {
	Describe(chan<- *prometheus.Desc)
	Collect(chan<- prometheus.Metric)
}
