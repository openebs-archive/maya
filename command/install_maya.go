package command

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type InstallMayaCommand struct {
	// To control this CLI's display
	M Meta

	// OS command to execute; <optional>
	Cmd *exec.Cmd

	// all servers excluding self
	member_ips string

	// all servers including self
	server_count int

	// all servers ipv4, in a comma separated format
	all_servers_ipv4 string

	// self ip address
	self_ip string

	// self hostname
	self_hostname string
}

func (c *InstallMayaCommand) Help() string {
	helpText := `
Usage: maya install-maya

  Installs maya server on this machine. In other words, the machine
  where this command is run will become a maya server.

General Options:

  ` + generalOptionsUsage() + `

Install Maya Options:

  -member-ips=<IP Address(es) of all server members>
    Comma separated list of IP addresses of all server members
    participating in the cluster.
    
    NOTE: Do not include the IP address of this local machine i.e.
    the machine where this command is being run.

  -self-ip=<IP Address>
    The IP Address of this local machine i.e. the machine where
    this command is being run. This is required when the machine
    has many private IPs and you want to use a specific IP.
`
	return strings.TrimSpace(helpText)
}

func (c *InstallMayaCommand) Synopsis() string {
	return "Installs maya server on this machine."
}

func (c *InstallMayaCommand) Run(args []string) int {
	var runop int

	flags := c.M.FlagSet("install-maya", FlagSetClient)
	flags.Usage = func() { c.M.Ui.Output(c.Help()) }

	flags.StringVar(&c.member_ips, "member-ips", "", "")
	flags.StringVar(&c.self_ip, "self-ip", "", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	// There are no extra arguments
	oargs := flags.Args()
	if len(oargs) != 0 {
		c.M.Ui.Error(c.Help())
		return 1
	}

	if c.Cmd != nil {
		// execute the provided command
		return execute(c.Cmd, c.M.Ui)
	}

	if runop = c.bootTheInstall(); runop != 0 {
		return runop
	}

	if runop = c.verifyBootstrap(); runop != 0 {
		return runop
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

	return runop
}

func (c *InstallMayaCommand) installConsul() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", InstallConsulScript)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error installing consul")
	}

	return runop
}

func (c *InstallMayaCommand) installNomad() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", InstallNomadScript)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error installing nomad")
	}

	return runop
}

func (c *InstallMayaCommand) verifyBootstrap() int {

	var runop int = 0

	c.Cmd = exec.Command("ls", MayaScriptsPath)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error(fmt.Sprintf("Install failed: Bootstrap failed: Missing path: %s", MayaScriptsPath))
	}

	return runop
}

func (c *InstallMayaCommand) bootTheInstall() int {

	var runop int = 0

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

	var runop int = 0
	var server_members []string

	c.Cmd = exec.Command("hostname")

	if runop = execute(c.Cmd, c.M.Ui, &c.self_hostname); runop != 0 {
		c.M.Ui.Error("Install failed: hostname could not be determined")
		return runop
	}

	if len(strings.TrimSpace(c.self_ip)) == 0 {
		c.Cmd = exec.Command("sh", GetPrivateIPScript)

		if runop = execute(c.Cmd, c.M.Ui, &c.self_ip); runop != 0 {
			c.M.Ui.Error("Install failed: Error fetching local IP address")
			return runop
		}
	}

	if len(strings.TrimSpace(c.self_ip)) == 0 {
		c.M.Ui.Error("Install failed: IP address could not be determined")
		return 1
	}

	// server count will be count(members) + self
	c.server_count = 1
	if len(strings.TrimSpace(c.member_ips)) > 0 {
		server_members = strings.Split(strings.TrimSpace(c.member_ips), ",")
		c.server_count = len(server_members) + 1
	}

	c.all_servers_ipv4 = `"` + c.self_ip + `"`

	for _, server_ip := range server_members {
		c.all_servers_ipv4 = c.all_servers_ipv4 + `,"` + server_ip + `"`
	}

	return runop
}

func (c *InstallMayaCommand) setConsulAsServer() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", SetConsulAsServerScript, c.self_ip, c.self_hostname, c.all_servers_ipv4, strconv.Itoa(c.server_count))

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error setting consul as server")
	}

	return runop
}

func (c *InstallMayaCommand) setNomadAsServer() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", SetNomadAsServerScript, c.self_ip, c.self_hostname, c.all_servers_ipv4, strconv.Itoa(c.server_count))

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error setting nomad as server")
	}

	return runop
}

func (c *InstallMayaCommand) startConsul() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", StartConsulServerScript)

	if runop := execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Systemd failed: Error starting consul")
	}

	return runop
}
func (c *InstallMayaCommand) startNomad() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", StartNomadServerScript)

	if runop := execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Systemd failed: Error starting nomad")
	}

	return runop
}
