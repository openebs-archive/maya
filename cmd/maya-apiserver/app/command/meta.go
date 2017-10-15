package command

// This is an adaptation of Hashicorp's Nomad library.
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
	// FlagSetNone sets FlagSetFlags to 0
	FlagSetNone FlagSetFlags = 0
	// FlagSetClient sets FlagSetFlags to 1 and varies
	FlagSetClient FlagSetFlags = 1 << iota
	// FlagSetDefault sets default value
	FlagSetDefault = FlagSetClient
)

// Meta contains the meta-options and functionality that nearly every
// Maya server command inherits.
type Meta struct {
	Ui cli.Ui

	// Whether to not-colorize output
	noColor bool
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
		f.BoolVar(&m.noColor, "no-color", false, "")
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

// Colorize returns all the including fields.
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
  -no-color
    Disables colored command output.
`
	return strings.TrimSpace(helpText)
}
