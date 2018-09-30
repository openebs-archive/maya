package v1

const (
	// BytesToGB used to convert bytes to GB
	BytesToGB = 1073741824
	// BytesToMB used to convert bytes to MB
	BytesToMB = 1048567
	// BytesToKB used to convert bytes to KB
	BytesToKB = 1024
	// MicSec used to convert to microsec to second
	MicSec = 1000000
	// MinWidth used in tabwriter
	MinWidth = 0
	// MaxWidth used in tabwriter
	MaxWidth = 0
	// Padding used in tabwriter
	Padding = 4

	// ControllerPort : Jiva volume controller listens on this for various api
	// requests.
	ControllerPort string = ":9501"
	// InfoAPI is the api for getting the volume access modes.
	InfoAPI string = "/replicas"
	// ReplicaPort : Jiva volume replica listens on this for various api
	// requests
	ReplicaPort string = ":9502"
	// StatsAPI is api to query about the volume stats from both replica and
	// controller.
	StatsAPI string = "/stats"
)
