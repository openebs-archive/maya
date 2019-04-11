package driver

// Config struct fills the parameters of request or user input
type Config struct {
	DriverName string
	PluginType string
	Version    string
	Endpoint   string
	NodeID     string
	Token      string
	RestURL    string
}

//NewConfig returns config struct to initialize new driver
func NewConfig() *Config {
	return &Config{}
}
