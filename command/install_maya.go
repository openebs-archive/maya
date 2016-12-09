package command

import (
	"fmt"
	"os/exec"
	"strings"
)

type InstallMayaCommand struct {
	// To control this CLI's display
	M Meta

	// OS command to execute; <optional>
	Cmd *exec.Cmd

	member_ips string

	self_ip string
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
    Command separated list of IP addresses of all server members
    participating in the cluster.
    
    NOTE: Do not include the IP address of this local machine i.e.
    the machine where this command is being run.

  -self-ip=<IP Address>
    The IP Address of this local machine i.e. the machine where
    this command is being run.
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

	// install related steps
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
		c.M.Ui.Error("Failed to bootstrap the install")
		return runop
	}

	c.Cmd = exec.Command("ls", MayaScriptsPath)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error(fmt.Sprintf("Install failed: Missing path: %s", MayaScriptsPath))
		return runop
	}

	c.Cmd = exec.Command("sh", InstallConsulScript)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error installing consul")
		return runop
	}

	c.setServerCount()

	if runop = c.setIP(); runop != 0 {
		return runop
	}

	if runop = c.setConsulAsServer(); runop != 0 {
		return runop
	}

	if runop = c.startConsul(); runop != 0 {
		return runop
	}

	return runop
}

func (c *InstallMayaCommand) setIP() int {

	var runop int = 0

	if c.self_ip == "" {
		// Derive the self ip
		c.Cmd = exec.Command("sh", GetPrivateIPScript)

		if runop = execute(c.Cmd, c.M.Ui, c.self_ip); runop != 0 {
			c.M.Ui.Error("Install failed: Error fetching local IP address")
		}
	}

	c.M.Ui.Output(fmt.Sprintf("Self IP: %s", c.self_ip))
	return runop
}

func (c *InstallMayaCommand) setServerCount() {

	var server_count int
	var server_members []string

	if c.member_ips == "" {
		// This will be the only server as there are no members
		server_count = 1
		c.M.Ui.Output(fmt.Sprintf("tot servers: %d", server_count))
	} else {
		// server count will be count(members) + self
		server_members = strings.Split(strings.TrimSpace(c.member_ips), ",")
		server_count = len(server_members) + 1
		c.M.Ui.Output(fmt.Sprintf("tot servers: %d", server_count))
	}

}

func (c *InstallMayaCommand) setConsulAsServer() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", SetConsulAsServerScript)

	if runop = execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Error setting consul as server")
	}

	return runop
}

func (c *InstallMayaCommand) startConsul() int {

	var runop int = 0

	c.Cmd = exec.Command("start", "consul")

	if runop := execute(c.Cmd, c.M.Ui); runop != 0 {
		c.M.Ui.Error("Install failed: Upstart failed: Error starting consul")
	}

	return runop
}
