package command

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"sort"

	"strings"
	"syscall"
	"time"

	"github.com/openebs/maya/cmd/maya-apiserver/app/config"
	"github.com/openebs/maya/cmd/maya-apiserver/app/server"

	"github.com/hashicorp/go-syslog"
	"github.com/hashicorp/logutils"
	"github.com/mitchellh/cli"
	"github.com/openebs/maya/kit/flaghelper"
	"github.com/openebs/maya/kit/loghelper"
)

// gracefulTimeout controls how long we wait before forcefully terminating
const gracefulTimeout = 5 * time.Second

// UpCommand is a cli implementation that runs a Maya server.
// The command will not end unless a shutdown message is sent on the
// ShutdownCh. If two messages are sent on the ShutdownCh it will forcibly
// exit.
type UpCommand struct {
	Revision          string
	Version           string
	VersionPrerelease string
	Ui                cli.Ui
	ShutdownCh        <-chan struct{}

	args []string

	// TODO
	// Check if both maya & httpServer instances are required ?
	// Can httpServer or maya embed one of the other ?
	// Need to take care of shuting down & graceful exit scenarios !!
	maya       *server.MayaApiServer
	httpServer *server.HTTPServer
	logFilter  *logutils.LevelFilter
	logOutput  io.Writer
}

func (c *UpCommand) readMayaConfig() *config.MayaConfig {
	var configPath []string

	// Make a new, empty config.
	cmdConfig := &config.MayaConfig{
		Ports: &config.Ports{},
	}

	flags := flag.NewFlagSet("up", flag.ContinueOnError)
	flags.Usage = func() { c.Ui.Error(c.Help()) }

	// options
	flags.Var((*flaghelper.StringFlag)(&configPath), "config", "config")
	flags.StringVar(&cmdConfig.BindAddr, "bind", "", "")
	flags.StringVar(&cmdConfig.DataDir, "data-dir", "", "")
	flags.StringVar(&cmdConfig.LogLevel, "log-level", "", "")

	if err := flags.Parse(c.args); err != nil {
		return nil
	}

	// Load the configuration
	mconfig := config.DefaultMayaConfig()

	for _, path := range configPath {
		current, err := config.LoadMayaConfig(path)
		if err != nil {
			c.Ui.Error(fmt.Sprintf(
				"Error loading configuration from %s: %s", path, err))
			return nil
		}

		// The user asked us to load some config here but we didn't find any,
		// so we'll complain but continue.
		if current == nil || reflect.DeepEqual(current, &config.MayaConfig{}) {
			c.Ui.Warn(fmt.Sprintf("No configuration loaded from %s", path))
		}

		if mconfig == nil {
			mconfig = current
		} else {
			mconfig = mconfig.Merge(current)
		}
	}

	// Merge any CLI options over config file options
	mconfig = mconfig.Merge(cmdConfig)

	// Set the version info
	mconfig.Revision = c.Revision
	mconfig.Version = c.Version
	mconfig.VersionPrerelease = c.VersionPrerelease

	// Normalize binds, ports, addresses, and advertise
	if err := mconfig.NormalizeAddrs(); err != nil {
		c.Ui.Error(err.Error())
		return nil
	}

	// Verify the paths are absolute.
	dirs := map[string]string{
		"data-dir": mconfig.DataDir,
	}
	for k, dir := range dirs {
		if dir == "" {
			continue
		}

		if !filepath.IsAbs(dir) {
			c.Ui.Error(fmt.Sprintf("%s must be given as an absolute path: got %v", k, dir))
			return nil
		}
	}

	return mconfig
}

// setupLoggers is used to setup the logGate, logWriter, and our logOutput
func (c *UpCommand) setupLoggers(mconfig *config.MayaConfig) (*loghelper.Writer, *loghelper.LogRegistrar, io.Writer) {
	// Setup logging. First create the gated log writer, which will
	// store logs until we're ready to show them. Then create the level
	// filter, filtering logs of the specified level.
	logGate := &loghelper.Writer{
		Writer: &cli.UiWriter{Ui: c.Ui},
	}

	c.logFilter = loghelper.LevelFilter()
	c.logFilter.MinLevel = logutils.LogLevel(strings.ToUpper(mconfig.LogLevel))
	c.logFilter.Writer = logGate
	if !loghelper.ValidateLevelFilter(c.logFilter.MinLevel, c.logFilter) {
		c.Ui.Error(fmt.Sprintf(
			"Invalid log level: %s. Valid log levels are: %v",
			c.logFilter.MinLevel, c.logFilter.Levels))
		return nil, nil, nil
	}

	// Check if syslog is enabled
	var syslog io.Writer
	if mconfig.EnableSyslog {
		l, err := gsyslog.NewLogger(gsyslog.LOG_NOTICE, mconfig.SyslogFacility, "mapiserver")
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Syslog setup failed: %v", err))
			return nil, nil, nil
		}
		syslog = &loghelper.SyslogWriter{l, c.logFilter}
	}

	// Create a log writer, and wrap a logOutput around it
	logWriter := loghelper.NewLogRegistrar(512)
	var logOutput io.Writer
	if syslog != nil {
		logOutput = io.MultiWriter(c.logFilter, logWriter, syslog)
	} else {
		logOutput = io.MultiWriter(c.logFilter, logWriter)
	}
	c.logOutput = logOutput
	log.SetOutput(logOutput)
	return logGate, logWriter, logOutput
}

// setupMayaServer is used to start Maya server
func (c *UpCommand) setupMayaServer(mconfig *config.MayaConfig, logOutput io.Writer) error {
	c.Ui.Output("Starting maya api server ...")

	// Setup maya service i.e. maya api server
	maya, err := server.NewMayaApiServer(mconfig, logOutput)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error starting maya api server: %s", err))
		return err
	}

	c.maya = maya

	// Setup the HTTP server
	http, err := server.NewHTTPServer(maya, mconfig, logOutput)
	if err != nil {
		maya.Shutdown()
		c.Ui.Error(fmt.Sprintf("Error starting http server: %s", err))
		return err
	}

	c.httpServer = http

	return nil
}

// Run does tasks related to mayaserver.
func (c *UpCommand) Run(args []string) int {
	c.Ui = &cli.PrefixedUi{
		OutputPrefix: "==> ",
		InfoPrefix:   "    ",
		ErrorPrefix:  "==> ",
		Ui:           c.Ui,
	}

	// Parse our configs
	c.args = args
	mconfig := c.readMayaConfig()
	if mconfig == nil {
		return 1
	}

	// Setup the log outputs
	logGate, _, logOutput := c.setupLoggers(mconfig)
	if logGate == nil {
		return 1
	}

	// Log config files
	if len(mconfig.Files) > 0 {
		c.Ui.Info(fmt.Sprintf("Loaded configuration from %s", strings.Join(mconfig.Files, ", ")))
	} else {
		c.Ui.Info("No configuration files loaded")
	}

	// Setup Maya server
	if err := c.setupMayaServer(mconfig, logOutput); err != nil {
		return 1
	}
	defer c.maya.Shutdown()

	// Check and shut down at the end
	defer func() {
		if c.httpServer != nil {
			c.httpServer.Shutdown()
		}
	}()

	// Compile Maya server information for output later
	info := make(map[string]string)
	info["version"] = fmt.Sprintf("%s%s", mconfig.Version, mconfig.VersionPrerelease)
	info["log level"] = mconfig.LogLevel
	info["region"] = fmt.Sprintf("%s (DC: %s)", mconfig.Region, mconfig.Datacenter)

	// Sort the keys for output
	infoKeys := make([]string, 0, len(info))
	for key := range info {
		infoKeys = append(infoKeys, key)
	}
	sort.Strings(infoKeys)

	// Maya server configuration output
	padding := 18
	c.Ui.Output("Maya api server configuration:\n")
	for _, k := range infoKeys {
		c.Ui.Info(fmt.Sprintf(
			"%s%s: %s",
			strings.Repeat(" ", padding-len(k)),
			strings.Title(k),
			info[k]))
	}
	c.Ui.Output("")

	// Output the header that the server has started
	c.Ui.Output("Maya api server started! Log data will stream in below:\n")

	// Enable log streaming
	logGate.Flush()

	// Wait for exit
	return c.handleSignals(mconfig)
}

// handleSignals blocks until we get an exit-causing signal
func (c *UpCommand) handleSignals(mconfig *config.MayaConfig) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGPIPE)

	// Wait for a signal
WAIT:
	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case <-c.ShutdownCh:
		sig = os.Interrupt
	}
	c.Ui.Output(fmt.Sprintf("Caught signal: %v", sig))

	// Skip any SIGPIPE signal (See issue #1798)
	if sig == syscall.SIGPIPE {
		goto WAIT
	}

	// Check if this is a SIGHUP
	if sig == syscall.SIGHUP {
		if conf := c.handleReload(mconfig); conf != nil {
			// Update the value only, not address
			*mconfig = *conf
		}
		goto WAIT
	}

	// Check if we should do a graceful leave
	graceful := false
	if sig == os.Interrupt && mconfig.LeaveOnInt {
		graceful = true
	} else if sig == syscall.SIGTERM && mconfig.LeaveOnTerm {
		graceful = true
	}

	// Bail fast if not doing a graceful leave
	if !graceful {
		return 1
	}

	// Attempt a graceful leave
	gracefulCh := make(chan struct{})
	c.Ui.Output("Gracefully shutting maya api server...")
	go func() {
		if err := c.maya.Leave(); err != nil {
			c.Ui.Error(fmt.Sprintf("Error: %s", err))
			return
		}
		close(gracefulCh)
	}()

	// Wait for leave or another signal
	select {
	case <-signalCh:
		return 1
	case <-time.After(gracefulTimeout):
		return 1
	case <-gracefulCh:
		return 0
	}
}

// handleReload is invoked when we should reload our configs, e.g. SIGHUP
// TODO
// The current reload code is very basic.
// Add ways to reload the orchestrator & plugins without shuting down the
// process
func (c *UpCommand) handleReload(mconfig *config.MayaConfig) *config.MayaConfig {

	c.Ui.Output("Reloading maya api server configuration...")

	newConf := c.readMayaConfig()
	if newConf == nil {
		c.Ui.Error(fmt.Sprintf("Failed to reload config"))
		return mconfig
	}

	// Change the log level
	minLevel := logutils.LogLevel(strings.ToUpper(newConf.LogLevel))
	if loghelper.ValidateLevelFilter(minLevel, c.logFilter) {
		c.logFilter.SetMinLevel(minLevel)
	} else {
		c.Ui.Error(fmt.Sprintf(
			"Invalid log level: %s. Valid log levels are: %v",
			minLevel, c.logFilter.Levels))

		// Keep the current log level
		newConf.LogLevel = mconfig.LogLevel
	}

	return newConf
}

// Synopsis returns that maya api server started
func (c *UpCommand) Synopsis() string {
	return "Starts maya api server"
}

// Help returns the various help tags and other options.
func (c *UpCommand) Help() string {
	helpText := `
Usage: m-apiserver up [options]

  Starts maya api server and runs until an interrupt is received.

  The maya api server's configuration primarily comes from the config
  files used, but a subset of the options may also be passed directly
  as CLI arguments, listed below.

General Options :

  -bind=<addr>
    The address the agent will bind to for all of its various network
    services. The individual services that run bind to individual
    ports on this address. Defaults to the loopback 127.0.0.1.

  -config=<path>
    The path to either a single config file or a directory of config
    files to use for configuring maya api server. This option may be
    specified multiple times. If multiple config files are used, the
    values from each will be merged together. During merging, values
    from files found later in the list are merged over values from
    previously parsed files.

  -data-dir=<path>
    The data directory used to store state and other persistent data.
    On client machines this is used to house allocation data such as
    downloaded artifacts used by drivers. On server nodes, the data
    dir is also used to store the replicated log.

  -log-level=<level>
    Specify the verbosity level of maya api server's logs. Valid values include
    DEBUG, INFO, and WARN, in decreasing order of verbosity. The
    default is INFO.

  -node=<name>
    The name of the local agent. This name is used to identify the node
    in the cluster. The name must be unique per region. The default is
    the current hostname of the machine.
 `
	return strings.TrimSpace(helpText)
}
