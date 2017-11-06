package command

import (
	"bufio"
	"flag"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/nomad/api"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/colorstring"
)

// Names of environment variables used to supply various
// config options to the Nomad CLI.
const (
	// EnvNomadAddress supplies Nomad Address config option
	EnvNomadAddress = "NOMAD_ADDR"
	// EnvNomadRegion supplies Nomad Region config option
	EnvNomadRegion = "NOMAD_REGION"

	// Constants for CLI identifier length
	shortID = 8
	fullID  = 36
)

// FlagSetFlags is an enum to define what flags are present in the
// default FlagSet returned by Meta.FlagSet.
type FlagSetFlags uint

const (
	// FlagSetNone is a constant of type FlagSetFlags
	FlagSetNone FlagSetFlags = 0
	// FlagSetClient is a constant of type FlagSetFlags
	FlagSetClient FlagSetFlags = 1 << iota
	// FlagSetDefault is a constant of type FlagSetFlags
	FlagSetDefault = FlagSetClient
)

// Meta contains the meta-options and functionality that nearly every
// Nomad command inherits.
type Meta struct {
	Ui cli.Ui

	// These are set by the command line flags.
	flagAddress string

	// Whether to not-colorize output
	noColor bool

	// The region to send API requests
	region string

	caCert     string
	caPath     string
	clientCert string
	clientKey  string
	insecure   bool
}

// FlagSet returns a FlagSet with the common flags that every
// command implements. The exact behavior of FlagSet can be configured
// using the flags as the second parameter, for example to disable
// server settings on the commands that don't talk to a server.
func (m *Meta) FlagSet(n string, fs FlagSetFlags) *flag.FlagSet {
	f := flag.NewFlagSet(n, flag.ContinueOnError)

	// FlagSetClient is used to enable the settings for specifying
	// client connectivity options.
	if fs&FlagSetClient != 0 {
		f.StringVar(&m.flagAddress, "address", "", "")
		f.StringVar(&m.region, "region", "", "")
		f.BoolVar(&m.noColor, "no-color", false, "")
		f.StringVar(&m.caCert, "ca-cert", "", "")
		f.StringVar(&m.caPath, "ca-path", "", "")
		f.StringVar(&m.clientCert, "client-cert", "", "")
		f.StringVar(&m.clientKey, "client-key", "", "")
		f.BoolVar(&m.insecure, "insecure", false, "")
		f.BoolVar(&m.insecure, "tls-skip-verify", false, "")

	}

	// Create an io.Writer that writes to our UI properly for errors.
	// This is kind of a hack, but it does the job. Basically: create
	// a pipe, use a scanner to break it into lines, and output each line
	// to the UI. Do this forever.
	errR, errW := io.Pipe()
	errScanner := bufio.NewScanner(errR)
	go func() {
		for errScanner.Scan() {
			m.Ui.Error(errScanner.Text())
		}
	}()
	f.SetOutput(errW)

	return f
}

// Client is used to initialize and return a new API client using
// the default command line arguments and env vars.
func (m *Meta) Client() (*api.Client, error) {
	config := api.DefaultConfig()
	if v := os.Getenv(EnvNomadAddress); v != "" {
		config.Address = v
	}
	if m.flagAddress != "" {
		config.Address = m.flagAddress
	}
	if v := os.Getenv(EnvNomadRegion); v != "" {
		config.Region = v
	}
	if m.region != "" {
		config.Region = m.region
	}
	// If we need custom TLS configuration, then set it
	if m.caCert != "" || m.caPath != "" || m.clientCert != "" || m.clientKey != "" || m.insecure {
		t := &api.TLSConfig{
			CACert:     m.caCert,
			CAPath:     m.caPath,
			ClientCert: m.clientCert,
			ClientKey:  m.clientKey,
			Insecure:   m.insecure,
		}
		config.TLSConfig = t
	}

	return api.NewClient(config)
}

// Colorize colorizes output
func (m *Meta) Colorize() *colorstring.Colorize {
	return &colorstring.Colorize{
		Colors:  colorstring.DefaultColors,
		Disable: m.noColor,
		Reset:   true,
	}
}

// generalOptionsUsage returns the help string for the global options.
func generalOptionsUsage() string {
	helpText := `
	`
	return strings.TrimSpace(helpText)
}
