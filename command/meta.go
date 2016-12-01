package command

import (
	"bufio"
	"flag"
	"io"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/mitchellh/colorstring"
)

const (

	// Constants for CLI identifier length
	shortId = 8
	fullId  = 36
)

// FlagSetFlags is an enum to define what flags are present in the
// default FlagSet returned by Meta.FlagSet.
type FlagSetFlags uint

const (
	FlagSetNone    FlagSetFlags = 0
	FlagSetClient  FlagSetFlags = 1 << iota
	FlagSetDefault              = FlagSetClient
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
// Nomad command implements. The exact behavior of FlagSet can be configured
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
  -address=<addr>
    The address of the server.
  -region=<region>
    The region of the servers to forward commands to.    
  -no-color
    Disables colored command output.
  -ca-cert=<path>           
    Path to a PEM encoded CA cert file to use to verify the 
    server SSL certificate.
  -ca-path=<path>           
    Path to a directory of PEM encoded CA cert files to verify 
    the server SSL certificate. If both -ca-cert and 
    -ca-path are specified, -ca-cert is used.
  -client-cert=<path>       
    Path to a PEM encoded client certificate for TLS authentication 
    to the server. Must also specify -client-key.
  -client-key=<path>        
    Path to an unencrypted PEM encoded private key matching the 
    client certificate from -client-cert.
  -tls-skip-verify        
    Do not verify TLS certificate. This is highly not recommended.
`
	return strings.TrimSpace(helpText)
}
