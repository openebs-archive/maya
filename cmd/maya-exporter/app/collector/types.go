package collector

import (
	v1 "github.com/openebs/maya/pkg/apis/openebs.io/stats"
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

// Volume interface defines the interfaces that has methods to be
// implemented by the cstor and jiva.
type Volume interface {
	Getter
	Parser
}

// Parser interface defines the method that to be implemented by the
// Cstor and Jiva. parse() is used to parse
// the response into the Metrics struct.
type Parser interface {
	parse(stats v1.VolumeStats) stats
}

// Getter interface defines the method that to be implemented by
// the cstor and jiva. getter() is used
// to collect the stats from the Jiva and Cstor.
type Getter interface {
	get() (v1.VolumeStats, error)
}
