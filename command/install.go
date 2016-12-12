package command

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/mitchellh/cli"
)

type Installer interface {
	Install() int
}

type InstallCommand struct {
	// To control this CLI's display
	Ui cli.Ui

	// OS command to execute
	Cmd *exec.Cmd
}

type MayaAsNomadInstaller struct {
	InstallCommand

	// if the installation is in server mode or client mode
	is_master bool

	// all maya master ips, in a comma separated format
	master_ips string

	// self ip address
	self_ip string

	// the provided master ips in a format understood by Nomad & Consul
	fmt_master_ips string

	// the provided master ips with rpc ports in a format understood by Nomad
	fmt_master_ipnports string
}

// The public command
func (c *MayaAsNomadInstaller) Install() int {

	var runop int = 0

	if runop = c.bootstrap(); runop != 0 {
		return runop
	}

	if runop = c.verifyBootstrap(); runop != 0 {
		return runop
	}

	if runop = c.initAsClient(); runop != 0 {
		return runop
	}

	if runop = c.installConsul(); runop != 0 {
		return runop
	}

	if runop = c.setConsulAsClient(); runop != 0 {
		return runop
	}

	if runop = c.startConsulAsClient(); runop != 0 {
		return runop
	}

	if runop = c.installNomad(); runop != 0 {
		return runop
	}

	if runop = c.setNomadAsClient(); runop != 0 {
		return runop
	}

	if runop = c.startNomadAsClient(); runop != 0 {
		return runop
	}

	return runop
}

func (c *MayaAsNomadInstaller) bootstrap() int {

	var runop int = 0

	c.Cmd = exec.Command("curl", "-sSL", BootstrapScriptPath, "-o", BootstrapScript)

	if runop = execute(c.Cmd, c.Ui); runop != 0 {
		c.Ui.Error(fmt.Sprintf("Install failed: Bootstrap failed: Could not fetch file: %s", BootstrapScriptPath))

		c.Cmd = exec.Command("rm", "-rf", BootstrapScript)
		execute(c.Cmd, c.Ui)

		return runop
	}

	c.Cmd = exec.Command("sh", "./"+BootstrapScript)
	runop = execute(c.Cmd, c.Ui)

	c.Cmd = exec.Command("rm", "-rf", BootstrapScript)
	execute(c.Cmd, c.Ui)

	if runop != 0 {
		c.Ui.Error("Install failed: Error while bootstraping")
	}

	return runop
}

func (c *MayaAsNomadInstaller) verifyBootstrap() int {

	var runop int = 0

	c.Cmd = exec.Command("ls", MayaScriptsPath)

	if runop = execute(c.Cmd, c.Ui); runop != 0 {
		c.Ui.Error(fmt.Sprintf("Install failed: Bootstrap failed: Missing path: %s", MayaScriptsPath))
	}

	return runop
}

// Set the instance variables i.e. properties of
// MayaAsNomadInstaller
func (c *MayaAsNomadInstaller) initAsClient() int {

	var runop int = 0
	var master_iparr []string

	if len(strings.TrimSpace(c.self_ip)) == 0 {
		c.Cmd = exec.Command("sh", GetPrivateIPScript)

		if runop = execute(c.Cmd, c.Ui, &c.self_ip); runop != 0 {
			c.Ui.Error("Install failed: Error fetching local IP address")
			return runop
		}
	}

	if len(strings.TrimSpace(c.self_ip)) == 0 {
		c.Ui.Error("Install failed: IP address could not be determined")
		return 1
	}

	if len(strings.TrimSpace(c.master_ips)) > 0 {
		master_iparr = strings.Split(strings.TrimSpace(c.master_ips), ",")
	}

	for _, master_ip := range master_iparr {
		if len(c.fmt_master_ips) > 0 {
			c.fmt_master_ips = c.fmt_master_ips + ","
			c.fmt_master_ipnports = c.fmt_master_ipnports + ","
		}
		c.fmt_master_ips = c.fmt_master_ips + `"` + master_ip + `"`
		c.fmt_master_ipnports = c.fmt_master_ipnports + `"` + master_ip + `:4647"`
	}

	return runop
}

func (c *MayaAsNomadInstaller) installConsul() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", InstallConsulScript)

	if runop = execute(c.Cmd, c.Ui); runop != 0 {
		c.Ui.Error("Install failed: Error installing consul")
	}

	return runop
}

func (c *MayaAsNomadInstaller) installNomad() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", InstallNomadScript)

	if runop = execute(c.Cmd, c.Ui); runop != 0 {
		c.Ui.Error("Install failed: Error installing nomad")
	}

	return runop
}

func (c *MayaAsNomadInstaller) startConsulAsClient() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", StartConsulClientScript)

	if runop := execute(c.Cmd, c.Ui); runop != 0 {
		c.Ui.Error("Install failed: Systemd failed: Error starting consul in client mode")
	}

	return runop
}

func (c *MayaAsNomadInstaller) startNomadAsClient() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", StartNomadClientScript)

	if runop := execute(c.Cmd, c.Ui); runop != 0 {
		c.Ui.Error("Install failed: Systemd failed: Error starting nomad in client mode")
	}

	return runop
}

func (c *MayaAsNomadInstaller) setConsulAsClient() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", SetConsulAsClientScript, c.self_ip, c.fmt_master_ips)

	if runop = execute(c.Cmd, c.Ui); runop != 0 {
		c.Ui.Error("Install failed: Error setting consul as client")
	}

	return runop
}

func (c *MayaAsNomadInstaller) setNomadAsClient() int {

	var runop int = 0

	c.Cmd = exec.Command("sh", SetNomadAsClientScript, c.self_ip, c.fmt_master_ipnports)

	if runop = execute(c.Cmd, c.Ui); runop != 0 {
		c.Ui.Error("Install failed: Error setting nomad as client")
	}

	return runop
}
