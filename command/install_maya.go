package command

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// InstallMayaCommand is a command implementation struct to setup Master Node
type InstallMayaCommand struct {
	// To control this CLI's display
	M Meta

	// OS command to execute
	Cmd *exec.Cmd

	// all servers excluding self
	memberIps string

	// all servers including self
	serverCount int

	// all servers ipv4, in a comma separated format
	allServersIpv4 string

	// self ip address
	selfIP string

	// self hostname
	selfHostname string

	//Contains the version info
	nomad  string
	consul string

	//flag variable for config
	conf   string
	config []Config
}

// Help shows helpText for a particular CLI command
func (c *InstallMayaCommand) Help() string {
	helpText := `
	Usage: maya setup-omm <config>

	Configure this machine as OpenEBS Maya Master (omm) 
	OMM is a clustered management server node that can either be
	run in VMs or Physical Hosts and is responsible for managing 
	and scheduling OpenEBS hosts and VSMs. 

	OMM also comes with an clustered configuration store. 

	OMM can be clustered with other local or remote OMMs.

	General Options:

	` + generalOptionsUsage() + `

	OpenEBS Maya Master (omm) setup Options:

	-config=<config yaml file>
	Congifuration file in Yaml format,example config file is available
	at example/maya_config.yaml dir.

	-omm-ips=<IP Address(es) of peer OMMs>
	Comma separated list of IP addresses of all management nodes
	participating in the cluster.

	NOTE: Do not include the IP address of this local machine i.e.
	the machine where this command is being run.

	If not provided, this machine will be added as the first node
	in the cluster. 

	-self-ip=<IP Address>
	The IP Address of this local machine i.e. the machine where
	this command is being run. This is required when the machine
	has many private IPs and you want to use a specific IP.
	`
	return strings.TrimSpace(helpText)
}

// Synopsis shows short information related to CLI command
func (c *InstallMayaCommand) Synopsis() string {
	return "Configure OpenEBS Maya Master on this machine."
}

// Run holds the flag values for CLI subcommands
func (c *InstallMayaCommand) Run(args []string) int {
	var runop int

	flags := c.M.FlagSet("setup-omm", FlagSetClient)
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }

	flags.StringVar(&c.memberIps, "omm-ips", "", "")
	flags.StringVar(&c.selfIP, "self-ip", "", "")
	flags.StringVar(&c.conf, "config", "", "Config file for setup omm")

	if err := flags.Parse(args); err != nil {
		return 1
	}
	// There are no extra arguments
	oargs := flags.Args()
	if len(oargs) != 0 && len(strings.TrimSpace(c.conf)) == 0 {
		c.M.Ui.Error(c.Help())
		return 1
	}

	if c.conf != "" {
		config := getConfig(c.conf)

		ok, errs := config.validate()
		if !ok {
			PrintValidationErrors(errs)
			fmt.Printf("file validation error prevents installation from proceeding:\n")
			return 1
		}

		c.selfIP = config.Args[0].Addr
		c.nomad = config.Spec.Bin[0].Version
		c.consul = config.Spec.Bin[1].Version
	}

	if c.Cmd != nil {
		// execute the provided command
		return execute(c.Cmd, c.M.Ui)
	}

	//Check if scripts were already downloaded
	if runop = c.verifyBootstrap(); runop != 0 {
		if runop = c.bootTheInstall(); runop != 0 {
			return runop
		}
	}

	if runop = c.init(); runop != 0 {
		return runop
	}

	if runop = c.installConsul(); runop != 0 {
		return runop
	}

	if runop = c.setConsulAsServer(); runop != 0 {
		return runop
	}

	if runop = c.startConsul(); runop != 0 {
		return runop
	}

	if runop = c.installNomad(); runop != 0 {
		return runop
	}

	if runop = c.setNomadAsServer(); runop != 0 {
		return runop
	}

	if runop = c.startNomad(); runop != 0 {
		return runop
	}
	if runop = c.installMayaserver(); runop != 0 {
		return runop
	}
	if runop = c.startMayaserver(); runop != 0 {
		return runop
	}

	return runop
}

func (c *InstallMayaCommand) installConsul() int {
	var runop int
	c.Cmd = exec.Command("sh", InstallConsulScript, c.consul)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error installing consul")
	}

	return runop
}

func (c *InstallMayaCommand) installNomad() int {

	var runop int

	c.Cmd = exec.Command("sh", InstallNomadScript, c.nomad)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error installing nomad")
	}

	return runop
}

func (c *InstallMayaCommand) installMayaserver() int {

	var runop int

	c.Cmd = exec.Command("bash", InstallMayaserverScript)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error installing mayaserver")
	}

	return runop
}

func (c *InstallMayaCommand) verifyBootstrap() int {
	//TODO: Enhance this logic to verify if there are updated scripts
	var runop int

	c.Cmd = exec.Command("ls", MayaScriptsPath)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error(fmt.Sprintf("Install failed: Bootstrap failed: Missing path: %s", MayaScriptsPath))
	}

	return runop
}

func (c *InstallMayaCommand) bootTheInstall() int {

	var runop int

	c.Cmd = exec.Command("curl", "-sSL", BootstrapScriptPath, "-o", BootstrapScript)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error(fmt.Sprintf("Failed to fetch file: %s", BootstrapScriptPath))

		c.Cmd = exec.Command("rm", "-rf", BootstrapScript)
		execute(c.Cmd, c.M.Ui)

		return runop
	}

	c.Cmd = exec.Command("sh", "./"+BootstrapScript)
	runop = execute(c.Cmd, c.M.Ui)

	c.Cmd = exec.Command("rm", "-rf", BootstrapScript)
	execute(c.Cmd, c.M.Ui)

	if runop != 0 {
		c.M.Ui.Error("Install failed: Error while bootstraping")
	}

	return runop
}

func (c *InstallMayaCommand) init() int {

	var runop int
	var serverMembers []string

	c.Cmd = exec.Command("hostname")

	if runop = execute(c.Cmd, c.M.Ui, &c.selfHostname); runop != 0 {
		c.M.Ui.Error("Install failed: hostname could not be determined")
		return runop
	}

	if len(strings.TrimSpace(c.selfIP)) == 0 {
		c.Cmd = exec.Command("sh", GetPrivateIPScript)

		if runop = execute(c.Cmd, c.M.Ui, &c.selfIP); runop != 0 {
			c.M.Ui.Error("Install failed: Error fetching local IP address")
			return runop
		}
	}

	if len(strings.TrimSpace(c.selfIP)) == 0 {
		c.M.Ui.Error("Install failed: IP address could not be determined")
		return 1
	}

	// server count will be count(members) + self
	c.serverCount = 1
	if len(strings.TrimSpace(c.memberIps)) > 0 {
		serverMembers = strings.Split(strings.TrimSpace(c.memberIps), ",")
		c.serverCount = len(serverMembers) + 1
	}

	c.allServersIpv4 = `"` + c.selfIP + `"`

	for _, serverIP := range serverMembers {
		c.allServersIpv4 = c.allServersIpv4 + `,"` + serverIP + `"`
	}

	return runop
}

func (c *InstallMayaCommand) setConsulAsServer() int {

	var runop int

	c.Cmd = exec.Command("sh", SetConsulAsServerScript, c.selfIP, c.selfHostname, c.allServersIpv4, strconv.Itoa(c.serverCount))

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error setting consul as server")
	}

	return runop
}

func (c *InstallMayaCommand) setNomadAsServer() int {

	var runop int

	c.Cmd = exec.Command("bash", SetNomadAsServerScript, c.selfIP, c.selfHostname, c.allServersIpv4, strconv.Itoa(c.serverCount))

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error setting nomad as server")
	}

	return runop
}

func (c *InstallMayaCommand) startConsul() int {

	var runop int

	c.Cmd = exec.Command("sh", StartConsulServerScript)

	if runop := execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Systemd failed: Error starting consul")
	}

	return runop
}
func (c *InstallMayaCommand) startNomad() int {

	var runop int

	c.Cmd = exec.Command("sh", StartNomadServerScript)

	if runop := execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Systemd failed: Error starting nomad")
	}

	return runop
}

func (c *InstallMayaCommand) startMayaserver() int {

	var runop int

	c.Cmd = exec.Command("sh", StartMayaServerScript, c.selfIP)

	if runop := execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Systemd failed: Error starting mayaserver")
	}

	return runop
}
