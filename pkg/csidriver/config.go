package driver

// Config struct fills the parameters of request or user input
type Config struct {
	// DriverName to be registered at CSI
	DriverName string
	// PluginType helps in specifying whether it is a node plugin or controller
	// Identity has to be run with both controller and node plugin
	// Same binry contains all the three plugins
	PluginType string
	// Version specifies the version of the CSI controller/node driver
	Version string
	// Endpoint on which requests are made by kubelet or external provisioner
	// Controller/node plugin will listen on this
	// This will be a unix based socket
	Endpoint string
	// NodeID helps in differentiating the nodes on which node deivers are
	// running. This is useful in case of topologies and publishing /
	// unpublishing volumes on nodes
	NodeID string
	// A REST Server is exposed on this URL for internal operations and Day2-ops
	RestURL string
}

//NewConfig returns config struct to initialize new driver
func NewConfig() *Config {
	return &Config{}
}
